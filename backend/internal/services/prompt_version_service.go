package services

import (
	"errors"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PromptVersionService 提示词版本服务
type PromptVersionService struct{}

// NewPromptVersionService 创建提示词版本服务
func NewPromptVersionService() *PromptVersionService {
	return &PromptVersionService{}
}

// CreatePromptVersion 创建提示词版本
func (s *PromptVersionService) CreatePromptVersion(entityType, entityID, prompt, changeSummary, createdBy string) (*models.PromptVersion, error) {
	// 获取当前最新版本
	var lastVersion models.PromptVersion
	result := database.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("version DESC").First(&lastVersion)

	newVersion := 1
	if result.Error == nil {
		newVersion = lastVersion.Version + 1
	}

	version := models.PromptVersion{
		ID:            uuid.New().String(),
		EntityType:    entityType,
		EntityID:      entityID,
		Version:       newVersion,
		Prompt:        prompt,
		ChangeSummary: changeSummary,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now().UnixMilli(),
	}

	if err := database.DB.Create(&version).Error; err != nil {
		return nil, err
	}

	return &version, nil
}

// GetPromptVersions 获取实体的所有提示词版本
func (s *PromptVersionService) GetPromptVersions(entityType, entityID string) ([]models.PromptVersion, error) {
	var versions []models.PromptVersion

	if err := database.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("version ASC").Find(&versions).Error; err != nil {
		return nil, err
	}

	return versions, nil
}

// GetPromptVersion 获取指定版本
func (s *PromptVersionService) GetPromptVersion(entityType, entityID string, version int) (*models.PromptVersion, error) {
	var v models.PromptVersion

	if err := database.DB.Where("entity_type = ? AND entity_id = ? AND version = ?", entityType, entityID, version).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &v, nil
}

// GetLatestPromptVersion 获取最新版本
func (s *PromptVersionService) GetLatestPromptVersion(entityType, entityID string) (*models.PromptVersion, error) {
	var v models.PromptVersion

	if err := database.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("version DESC").First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &v, nil
}

// CompareVersions 比较两个版本的差异
func (s *PromptVersionService) CompareVersions(entityType, entityID string, version1, version2 int) (map[string]interface{}, error) {
	v1, err := s.GetPromptVersion(entityType, entityID, version1)
	if err != nil {
		return nil, err
	}
	if v1 == nil {
		return nil, errors.New("版本 1 不存在")
	}

	v2, err := s.GetPromptVersion(entityType, entityID, version2)
	if err != nil {
		return nil, err
	}
	if v2 == nil {
		return nil, errors.New("版本 2 不存在")
	}

	return map[string]interface{}{
		"version1": map[string]interface{}{
			"version":       v1.Version,
			"prompt":        v1.Prompt,
			"changeSummary": v1.ChangeSummary,
			"createdAt":     v1.CreatedAt,
			"createdBy":     v1.CreatedBy,
		},
		"version2": map[string]interface{}{
			"version":       v2.Version,
			"prompt":        v2.Prompt,
			"changeSummary": v2.ChangeSummary,
			"createdAt":     v2.CreatedAt,
			"createdBy":     v2.CreatedBy,
		},
	}, nil
}
