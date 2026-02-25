package services

import (
	"strconv"
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

// makeConfigKey 生成配置键
func makeConfigKey(projectID, key string) string {
	return projectID + ":" + key
}

// GetSyncStatus 获取同步状态
func (s *SyncService) GetSyncStatus(projectID string) (*SyncStatus, error) {
	var config models.Config
	err := s.db.Where("key = ?", makeConfigKey(projectID, "sync_status")).First(&config).Error
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
	lastSyncKey := makeConfigKey(projectID, "last_sync")
	err := s.db.Where("key = ?", lastSyncKey).First(&config).Error
	if err == nil {
		lastSync, _ = strconv.ParseInt(config.Value, 10, 64)
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
	nowStr := strconv.FormatInt(now, 10)
	if err == gorm.ErrRecordNotFound {
		config = models.Config{
			Key:   lastSyncKey,
			Value: nowStr,
		}
		s.db.Create(&config)
	} else {
		config.Value = nowStr
		s.db.Save(&config)
	}

	// 更新同步状态
	statusKey := makeConfigKey(projectID, "sync_status")
	s.db.Where("key = ?", statusKey).
		Assign(models.Config{
			Key:   statusKey,
			Value: "synced",
		}).
		FirstOrCreate(&models.Config{})

	result.SyncedItems = result.CreatedItems + result.UpdatedItems + result.DeletedItems
	return result, nil
}

// ConfigureSync 配置同步
func (s *SyncService) ConfigureSync(projectID, remoteURL, syncInterval string) error {
	configs := map[string]string{
		"remote_url":    remoteURL,
		"sync_interval": syncInterval,
		"sync_enabled":  "true",
	}

	for name, value := range configs {
		key := makeConfigKey(projectID, name)
		s.db.Where("key = ?", key).
			Assign(models.Config{
				Key:   key,
				Value: value,
			}).
			FirstOrCreate(&models.Config{})
	}

	return nil
}

// GetSyncConfig 获取同步配置
func (s *SyncService) GetSyncConfig(projectID string) (map[string]string, error) {
	var configs []models.Config
	keys := []string{
		makeConfigKey(projectID, "remote_url"),
		makeConfigKey(projectID, "sync_interval"),
		makeConfigKey(projectID, "sync_enabled"),
	}
	err := s.db.Where("key IN ?", keys).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		// 移除 projectID 前缀
		for _, name := range []string{"remote_url", "sync_interval", "sync_enabled"} {
			if cfg.Key == makeConfigKey(projectID, name) {
				result[name] = cfg.Value
				break
			}
		}
	}

	return result, nil
}
