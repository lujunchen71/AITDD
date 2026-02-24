package services

import (
	"errors"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ModuleService 模块服务
type ModuleService struct{}

// NewModuleService 创建模块服务
func NewModuleService() *ModuleService {
	return &ModuleService{}
}

// GetModules 获取模块列表
func (s *ModuleService) GetModules(projectID string, parentID *string, status string) ([]models.Module, error) {
	var modules []models.Module

	query := database.DB.Model(&models.Module{}).Where("project_id = ?", projectID)

	if parentID != nil {
		if *parentID == "" {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&modules).Error; err != nil {
		return nil, err
	}

	return modules, nil
}

// GetModule 获取单个模块
func (s *ModuleService) GetModule(id string) (*models.Module, error) {
	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &module, nil
}

// CreateModule 创建模块
func (s *ModuleService) CreateModule(projectID string, parentID *string, name, description, prompt string) (*models.Module, error) {
	module := models.Module{
		ID:          uuid.New().String(),
		ParentID:    parentID,
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		Prompt:      prompt,
		Status:      models.ModuleStatusDesigning,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		Version:     1,
		SyncStatus:  models.SyncStatusSynced,
	}

	if err := database.DB.Create(&module).Error; err != nil {
		return nil, err
	}

	return &module, nil
}

// UpdateModule 更新模块
func (s *ModuleService) UpdateModule(id string, updates map[string]interface{}, version int) (*models.Module, error) {
	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模块不存在")
		}
		return nil, err
	}

	// 版本检查
	if module.Version != version {
		return nil, errors.New("版本冲突")
	}

	// 检查锁定
	if module.Locked {
		return nil, errors.New("模块已被锁定")
	}

	// 更新字段
	updates["version"] = module.Version + 1
	updates["updated_at"] = time.Now().UnixMilli()

	if err := database.DB.Model(&module).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取
	return s.GetModule(id)
}

// DeleteModule 删除模块
func (s *ModuleService) DeleteModule(id string) error {
	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("模块不存在")
		}
		return err
	}

	// 检查锁定
	if module.Locked {
		return errors.New("模块已被锁定")
	}

	// 检查子模块
	var childCount int64
	database.DB.Model(&models.Module{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		return errors.New("存在子模块，无法删除")
	}

	// 检查任务
	var taskCount int64
	database.DB.Model(&models.Task{}).Where("module_id = ?", id).Count(&taskCount)
	if taskCount > 0 {
		return errors.New("存在关联任务，无法删除")
	}

	return database.DB.Delete(&module).Error
}

// GetModuleTree 获取模块树
func (s *ModuleService) GetModuleTree(projectID string) ([]map[string]interface{}, error) {
	modules, err := s.GetModules(projectID, nil, "")
	if err != nil {
		return nil, err
	}

	// 构建树结构
	return s.buildTree(modules, nil), nil
}

// buildTree 递归构建树
func (s *ModuleService) buildTree(modules []models.Module, parentID *string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, module := range modules {
		if (parentID == nil && module.ParentID == nil) || (parentID != nil && module.ParentID != nil && *module.ParentID == *parentID) {
			node := map[string]interface{}{
				"id":          module.ID,
				"name":        module.Name,
				"description": module.Description,
				"status":      module.Status,
				"version":     module.Version,
				"createdAt":   module.CreatedAt,
				"updatedAt":   module.UpdatedAt,
			}

			// 递归获取子模块
			pid := module.ID
			children := s.buildTree(modules, &pid)
			if len(children) > 0 {
				node["children"] = children
			}

			result = append(result, node)
		}
	}

	return result
}

// GetModuleDependencies 获取模块的依赖列表
func (s *ModuleService) GetModuleDependencies(moduleID string) ([]models.ModuleDependencyDetail, error) {
	var dependencies []models.ModuleDependency

	if err := database.DB.Where("module_id = ?", moduleID).Find(&dependencies).Error; err != nil {
		return nil, err
	}

	var result []models.ModuleDependencyDetail
	for _, dep := range dependencies {
		detail := models.ModuleDependencyDetail{
			ModuleDependency: dep,
		}

		// 获取被依赖模块的信息
		var module models.Module
		if err := database.DB.First(&module, "id = ?", dep.DependsOnModuleID).Error; err == nil {
			detail.DependsOnModule = &module
		}

		result = append(result, detail)
	}

	return result, nil
}

// GetModuleDependents 获取依赖此模块的模块列表
func (s *ModuleService) GetModuleDependents(moduleID string) ([]models.ModuleDependentDetail, error) {
	var dependencies []models.ModuleDependency

	if err := database.DB.Where("depends_on_module_id = ?", moduleID).Find(&dependencies).Error; err != nil {
		return nil, err
	}

	var result []models.ModuleDependentDetail
	for _, dep := range dependencies {
		detail := models.ModuleDependentDetail{
			ModuleDependency: dep,
		}

		// 获取依赖方模块的信息
		var module models.Module
		if err := database.DB.First(&module, "id = ?", dep.ModuleID).Error; err == nil {
			detail.Module = &module
		}

		result = append(result, detail)
	}

	return result, nil
}

// CreateModuleDependency 创建模块依赖
func (s *ModuleService) CreateModuleDependency(moduleID, dependsOnModuleID, dependencyType, contractSummary string) (*models.ModuleDependency, error) {
	// 检查模块是否存在
	var module models.Module
	if err := database.DB.First(&module, "id = ?", moduleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模块不存在")
		}
		return nil, err
	}

	// 检查被依赖模块是否存在
	var dependsOnModule models.Module
	if err := database.DB.First(&dependsOnModule, "id = ?", dependsOnModuleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("被依赖的模块不存在")
		}
		return nil, err
	}

	// 不能依赖自己
	if moduleID == dependsOnModuleID {
		return nil, errors.New("模块不能依赖自己")
	}

	// 检查是否已存在依赖关系
	var existingCount int64
	database.DB.Model(&models.ModuleDependency{}).Where("module_id = ? AND depends_on_module_id = ?", moduleID, dependsOnModuleID).Count(&existingCount)
	if existingCount > 0 {
		return nil, errors.New("依赖关系已存在")
	}

	// 检查是否会形成循环依赖
	if s.wouldCreateCycle(moduleID, dependsOnModuleID) {
		return nil, errors.New("会形成循环依赖")
	}

	// 设置默认依赖类型
	if dependencyType == "" {
		dependencyType = models.ModuleDependencyRequired
	}

	dependency := models.ModuleDependency{
		ID:                uuid.New().String(),
		ModuleID:          moduleID,
		DependsOnModuleID: dependsOnModuleID,
		DependencyType:    dependencyType,
		ContractSummary:   contractSummary,
		CreatedAt:         time.Now().UnixMilli(),
		UpdatedAt:         time.Now().UnixMilli(),
		Version:          1,
		SyncStatus:       models.SyncStatusSynced,
	}

	if err := database.DB.Create(&dependency).Error; err != nil {
		return nil, err
	}

	return &dependency, nil
}

// wouldCreateCycle 检查是否会形成循环依赖
func (s *ModuleService) wouldCreateCycle(moduleID, dependsOnModuleID string) bool {
	// 如果 dependsOnModuleID 已经直接或间接依赖于 moduleID，则会形成循环
	visited := make(map[string]bool)
	return s.hasDependencyPath(dependsOnModuleID, moduleID, visited)
}

// hasDependencyPath 检查是否存在依赖路径
func (s *ModuleService) hasDependencyPath(fromModuleID, toModuleID string, visited map[string]bool) bool {
	if fromModuleID == toModuleID {
		return true
	}

	if visited[fromModuleID] {
		return false
	}
	visited[fromModuleID] = true

	var dependencies []models.ModuleDependency
	database.DB.Where("module_id = ?", fromModuleID).Find(&dependencies)

	for _, dep := range dependencies {
		if s.hasDependencyPath(dep.DependsOnModuleID, toModuleID, visited) {
			return true
		}
	}

	return false
}

// DeleteModuleDependency 删除模块依赖
func (s *ModuleService) DeleteModuleDependency(moduleID, dependencyID string) error {
	var dependency models.ModuleDependency
	if err := database.DB.First(&dependency, "id = ? AND module_id = ?", dependencyID, moduleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("依赖关系不存在")
		}
		return err
	}

	return database.DB.Delete(&dependency).Error
}

// UpdateModuleDependency 更新模块依赖
func (s *ModuleService) UpdateModuleDependency(moduleID, dependencyID, dependencyType, contractSummary string) (*models.ModuleDependency, error) {
	var dependency models.ModuleDependency
	if err := database.DB.First(&dependency, "id = ? AND module_id = ?", dependencyID, moduleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("依赖关系不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
	}

	if dependencyType != "" {
		updates["dependency_type"] = dependencyType
	}
	if contractSummary != "" {
		updates["contract_summary"] = contractSummary
	}

	if err := database.DB.Model(&dependency).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取
	if err := database.DB.First(&dependency, "id = ?", dependencyID).Error; err != nil {
		return nil, err
	}

	return &dependency, nil
}
