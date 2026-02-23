package services

import (
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChangeTracker 变更追踪服务
type ChangeTracker struct {
	db *gorm.DB
}

// NewChangeTracker 创建变更追踪服务
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		db: database.DB,
	}
}

// ChangeRecord 变更记录
type ChangeRecord struct {
	ID         string `json:"id"`
	EntityType string `json:"entityType"` // module, task, dependency
	EntityID   string `json:"entityId"`
	Action     string `json:"action"` // create, update, delete
	Changes    string `json:"changes"`
	ChangedBy  string `json:"changedBy"`
	ChangedAt  int64  `json:"changedAt"`
}

// TrackChange 记录变更
func (ct *ChangeTracker) TrackChange(entityType, entityID, action, changes, changedBy string) error {
	history := models.ChangeHistory{
		ID:         uuid.New().String(),
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Changes:    changes,
		ChangedBy:  changedBy,
		ChangedAt:  time.Now().UnixMilli(),
	}

	return ct.db.Create(&history).Error
}

// GetChanges 获取变更记录
func (ct *ChangeTracker) GetChanges(entityType, entityID string, since int64) ([]ChangeRecord, error) {
	var histories []models.ChangeHistory
	query := ct.db.Model(&models.ChangeHistory{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if entityID != "" {
		query = query.Where("entity_id = ?", entityID)
	}
	if since > 0 {
		query = query.Where("changed_at > ?", since)
	}

	err := query.Order("changed_at DESC").Limit(100).Find(&histories).Error
	if err != nil {
		return nil, err
	}

	records := make([]ChangeRecord, len(histories))
	for i, h := range histories {
		records[i] = ChangeRecord{
			ID:         h.ID,
			EntityType: h.EntityType,
			EntityID:   h.EntityID,
			Action:     h.Action,
			Changes:    h.Changes,
			ChangedBy:  h.ChangedBy,
			ChangedAt:  h.ChangedAt,
		}
	}

	return records, nil
}

// GetPendingChanges 获取待同步的变更
func (ct *ChangeTracker) GetPendingChanges(projectID string, lastSync int64) ([]ChangeRecord, error) {
	var histories []models.ChangeHistory

	// 获取模块变更
	err := ct.db.Raw(`
		SELECT ch.* FROM change_histories ch
		JOIN modules m ON ch.entity_id = m.id
		WHERE m.project_id = ? AND ch.changed_at > ?
		ORDER BY ch.changed_at DESC
	`, projectID, lastSync).Scan(&histories).Error
	if err != nil {
		return nil, err
	}

	records := make([]ChangeRecord, len(histories))
	for i, h := range histories {
		records[i] = ChangeRecord{
			ID:         h.ID,
			EntityType: h.EntityType,
			EntityID:   h.EntityID,
			Action:     h.Action,
			Changes:    h.Changes,
			ChangedBy:  h.ChangedBy,
			ChangedAt:  h.ChangedAt,
		}
	}

	return records, nil
}
