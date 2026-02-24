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
func (s *DependencyService) GetDependencies(upstreamTaskID string, downstreamTaskID string) ([]models.Dependency, error) {
	var dependencies []models.Dependency

	query := database.DB.Model(&models.Dependency{})

	if upstreamTaskID != "" {
		query = query.Where("upstream_task_id = ?", upstreamTaskID)
	}
	if downstreamTaskID != "" {
		query = query.Where("downstream_task_id = ?", downstreamTaskID)
	}

	if err := query.Find(&dependencies).Error; err != nil {
		return nil, err
	}

	return dependencies, nil
}

// CreateDependency 创建依赖
func (s *DependencyService) CreateDependency(upstreamTaskID, downstreamTaskID, contractSummary string) (*models.Dependency, error) {
	// 检查是否已存在
	var existing models.Dependency
	if err := database.DB.Where("upstream_task_id = ? AND downstream_task_id = ?", upstreamTaskID, downstreamTaskID).First(&existing).Error; err == nil {
		return nil, errors.New("依赖关系已存在")
	}

	// 检查循环依赖
	if s.wouldCreateCycle(upstreamTaskID, downstreamTaskID) {
		return nil, errors.New("创建此依赖将导致循环依赖")
	}

	dep := models.Dependency{
		ID:               uuid.New().String(),
		UpstreamTaskID:   upstreamTaskID,
		DownstreamTaskID: downstreamTaskID,
		ContractSummary:  contractSummary,
		CreatedAt:        time.Now().UnixMilli(),
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
func (s *DependencyService) wouldCreateCycle(upstreamTaskID, downstreamTaskID string) bool {
	// 如果 downstreamTaskID 依赖于 upstreamTaskID（直接或间接），则创建循环
	visited := make(map[string]bool)
	return s.hasPath(downstreamTaskID, upstreamTaskID, visited)
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
	database.DB.Where("downstream_task_id = ?", start).Find(&dependencies)

	for _, dep := range dependencies {
		if s.hasPath(dep.UpstreamTaskID, target, visited) {
			return true
		}
	}

	return false
}

// GetTaskDependencies 获取任务的所有依赖（直接和间接）
func (s *DependencyService) GetTaskDependencies(taskID string) ([]models.Dependency, error) {
	var dependencies []models.Dependency

	// 递归获取所有依赖
	s.collectDependencies(taskID, &dependencies, make(map[string]bool))

	return dependencies, nil
}

// collectDependencies 递归收集依赖
func (s *DependencyService) collectDependencies(taskID string, deps *[]models.Dependency, visited map[string]bool) {
	if visited[taskID] {
		return
	}
	visited[taskID] = true

	// 获取当前任务作为下游任务的所有依赖关系
	var directDeps []models.Dependency
	database.DB.Where("downstream_task_id = ?", taskID).Find(&directDeps)

	for _, dep := range directDeps {
		*deps = append(*deps, dep)
		// 递归获取上游任务的依赖
		s.collectDependencies(dep.UpstreamTaskID, deps, visited)
	}
}
