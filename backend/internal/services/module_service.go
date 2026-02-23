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
