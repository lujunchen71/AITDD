package services

import (
	"errors"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DependencyService 依赖服务
type DependencyService struct{}

// NewDependencyService 创建依赖服务
func NewDependencyService() *DependencyService {
	return &DependencyService{}
}

// GetDependencies 获取依赖列表
func (s *DependencyService) GetDependencies(moduleID string, dependsOn string) ([]models.Dependency, error) {
	var dependencies []models.Dependency

	query := database.DB.Model(&models.Dependency{})

	if moduleID != "" {
		query = query.Where("module_id = ?", moduleID)
	}
	if dependsOn != "" {
		query = query.Where("depends_on = ?", dependsOn)
	}

	if err := query.Find(&dependencies).Error; err != nil {
		return nil, err
	}

	return dependencies, nil
}

// CreateDependency 创建依赖
func (s *DependencyService) CreateDependency(moduleID, dependsOn, dependency string) (*models.Dependency, error) {
	// 检查是否已存在
	var existing models.Dependency
	if err := database.DB.Where("module_id = ? AND depends_on = ?", moduleID, dependsOn).First(&existing).Error; err == nil {
		return nil, errors.New("依赖关系已存在")
	}

	// 检查循环依赖
	if s.wouldCreateCycle(moduleID, dependsOn) {
		return nil, errors.New("创建此依赖将导致循环依赖")
	}

	dep := models.Dependency{
		ID:         uuid.New().String(),
		ModuleID:   moduleID,
		DependsOn:  dependsOn,
		Dependency: dependency,
		CreatedAt:  time.Now().UnixMilli(),
	}

	if err := database.DB.Create(&dep).Error; err != nil {
		return nil, err
	}

	return &dep, nil
}

// DeleteDependency 删除依赖
func (s *DependencyService) DeleteDependency(id string) error {
	var dep models.Dependency
	if err := database.DB.First(&dep, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("依赖不存在")
		}
		return err
	}

	return database.DB.Delete(&dep).Error
}

// wouldCreateCycle 检查是否会创建循环依赖
func (s *DependencyService) wouldCreateCycle(moduleID, dependsOn string) bool {
	// 如果 dependsOn 依赖于 moduleID（直接或间接），则创建循环
	visited := make(map[string]bool)
	return s.hasPath(dependsOn, moduleID, visited)
}

// hasPath 检查从 start 到 target 是否存在路径
func (s *DependencyService) hasPath(start, target string, visited map[string]bool) bool {
	if start == target {
		return true
	}

	if visited[start] {
		return false
	}
	visited[start] = true

	var dependencies []models.Dependency
	database.DB.Where("module_id = ?", start).Find(&dependencies)

	for _, dep := range dependencies {
		if s.hasPath(dep.DependsOn, target, visited) {
			return true
		}
	}

	return false
}

// GetModuleDependencies 获取模块的所有依赖（直接和间接）
func (s *DependencyService) GetModuleDependencies(moduleID string) ([]models.Dependency, error) {
	var dependencies []models.Dependency

	// 递归获取所有依赖
	s.collectDependencies(moduleID, &dependencies, make(map[string]bool))

	return dependencies, nil
}

// collectDependencies 递归收集依赖
func (s *DependencyService) collectDependencies(moduleID string, deps *[]models.Dependency, visited map[string]bool) {
	if visited[moduleID] {
		return
	}
	visited[moduleID] = true

	var directDeps []models.Dependency
	database.DB.Where("module_id = ?", moduleID).Find(&directDeps)

	for _, dep := range directDeps {
		*deps = append(*deps, dep)
		s.collectDependencies(dep.DependsOn, deps, visited)
	}
}
