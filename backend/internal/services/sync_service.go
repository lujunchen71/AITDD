package services

import (
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"gorm.io/gorm"
)

// SyncService 同步服务
type SyncService struct {
	db *gorm.DB
}

// NewSyncService 创建同步服务
func NewSyncService() *SyncService {
	return &SyncService{
		db: database.DB,
	}
}

// SyncStatus 同步状态
type SyncStatus struct {
	LastSyncTime int64  `json:"lastSyncTime"`
	Status       string `json:"status"` // synced, pending, error
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// GetSyncStatus 获取同步状态
func (s *SyncService) GetSyncStatus(projectID string) (*SyncStatus, error) {
	var config models.Config
	err := s.db.Where("project_id = ? AND key = ?", projectID, "sync_status").First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &SyncStatus{Status: "pending"}, nil
		}
		return nil, err
	}

	return &SyncStatus{
		LastSyncTime: config.UpdatedAt,
		Status:       config.Value,
	}, nil
}

// SyncResult 同步结果
type SyncResult struct {
	SyncedItems  int64 `json:"syncedItems"`
	CreatedItems int64 `json:"createdItems"`
	UpdatedItems int64 `json:"updatedItems"`
	DeletedItems int64 `json:"deletedItems"`
}

// SyncProject 同步项目数据
func (s *SyncService) SyncProject(projectID string) (*SyncResult, error) {
	result := &SyncResult{}

	// 获取上次同步时间
	var lastSync int64
	var config models.Config
	err := s.db.Where("project_id = ? AND key = ?", projectID, "last_sync").First(&config).Error
	if err == nil {
		lastSync = config.IntValue
	}

	// 同步模块
	var modules []models.Module
	s.db.Where("project_id = ? AND updated_at > ?", projectID, lastSync).Find(&modules)
	result.UpdatedItems += int64(len(modules))

	// 同步任务
	var tasks []models.Task
	s.db.Joins("JOIN modules ON tasks.module_id = modules.id").
		Where("modules.project_id = ? AND tasks.updated_at > ?", projectID, lastSync).
		Find(&tasks)
	result.UpdatedItems += int64(len(tasks))

	// 更新同步时间
	now := time.Now().UnixMilli()
	if err == gorm.ErrRecordNotFound {
		config = models.Config{
			ProjectID: projectID,
			Key:       "last_sync",
			IntValue:  now,
		}
		s.db.Create(&config)
	} else {
		config.IntValue = now
		s.db.Save(&config)
	}

	// 更新同步状态
	s.db.Where("project_id = ? AND key = ?", projectID, "sync_status").
		Assign(models.Config{
			ProjectID: projectID,
			Key:       "sync_status",
			Value:     "synced",
		}).
		FirstOrCreate(&models.Config{})

	result.SyncedItems = result.CreatedItems + result.UpdatedItems + result.DeletedItems
	return result, nil
}

// ConfigureSync 配置同步
func (s *SyncService) ConfigureSync(projectID, remoteURL, syncInterval string) error {
	configs := []models.Config{
		{ProjectID: projectID, Key: "remote_url", Value: remoteURL},
		{ProjectID: projectID, Key: "sync_interval", Value: syncInterval},
		{ProjectID: projectID, Key: "sync_enabled", Value: "true"},
	}

	for _, cfg := range configs {
		s.db.Where("project_id = ? AND key = ?", projectID, cfg.Key).
			Assign(cfg).
			FirstOrCreate(&models.Config{})
	}

	return nil
}

// GetSyncConfig 获取同步配置
func (s *SyncService) GetSyncConfig(projectID string) (map[string]string, error) {
	var configs []models.Config
	err := s.db.Where("project_id = ?", projectID).
		Where("key IN ?", []string{"remote_url", "sync_interval", "sync_enabled"}).
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.Key] = cfg.Value
	}

	return result, nil
}
