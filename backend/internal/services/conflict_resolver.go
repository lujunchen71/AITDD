package services

import (
	"encoding/json"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"gorm.io/gorm"
)

// ConflictResolver 冲突解决服务
type ConflictResolver struct {
	db *gorm.DB
}

// NewConflictResolver 创建冲突解决服务
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		db: database.DB,
	}
}

// Conflict 冲突记录
type Conflict struct {
	ID           string      `json:"id"`
	EntityType   string      `json:"entityType"`
	EntityID     string      `json:"entityId"`
	LocalData    interface{} `json:"localData"`
	RemoteData   interface{} `json:"remoteData"`
	ConflictType string      `json:"conflictType"` // update_conflict, delete_conflict
	CreatedAt    int64       `json:"createdAt"`
	ResolvedAt   *int64      `json:"resolvedAt,omitempty"`
	Resolution   string      `json:"resolution,omitempty"` // local_wins, remote_wins, merged
}

// DetectConflicts 检测冲突
func (cr *ConflictResolver) DetectConflicts(projectID string, remoteChanges []ChangeRecord) ([]Conflict, error) {
	var conflicts []Conflict

	for _, change := range remoteChanges {
		// 检查本地是否有相同实体的更新
		var localHistory models.ChangeHistory
		err := cr.db.Where("entity_type = ? AND entity_id = ? AND changed_at > ?",
			change.EntityType, change.EntityID, change.ChangedAt).
			First(&localHistory).Error

		if err == nil {
			// 存在本地更新，产生冲突
			conflict := Conflict{
				EntityType:   change.EntityType,
				EntityID:     change.EntityID,
				RemoteData:   change.Changes,
				ConflictType: "update_conflict",
				CreatedAt:    time.Now().UnixMilli(),
			}

			// 获取本地数据
			switch change.EntityType {
			case "module":
				var module models.Module
				if err := cr.db.First(&module, "id = ?", change.EntityID).Error; err == nil {
					conflict.LocalData = module
				}
			case "task":
				var task models.Task
				if err := cr.db.First(&task, "id = ?", change.EntityID).Error; err == nil {
					conflict.LocalData = task
				}
			}

			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

// ResolveConflict 解决冲突
func (cr *ConflictResolver) ResolveConflict(conflictID, resolution string, mergedData interface{}) error {
	// 获取冲突
	var conflict Conflict
	if err := cr.db.Where("id = ?", conflictID).First(&conflict).Error; err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	conflict.ResolvedAt = &now
	conflict.Resolution = resolution

	// 根据解决策略应用更改
	switch resolution {
	case "local_wins":
		// 保持本地数据，不做更改
	case "remote_wins":
		// 应用远程数据
		if err := cr.applyRemoteData(conflict.EntityType, conflict.EntityID, conflict.RemoteData); err != nil {
			return err
		}
	case "merged":
		// 应用合并后的数据
		if err := cr.applyMergedData(conflict.EntityType, conflict.EntityID, mergedData); err != nil {
			return err
		}
	}

	return nil
}

// applyRemoteData 应用远程数据
func (cr *ConflictResolver) applyRemoteData(entityType, entityID string, data interface{}) error {
	switch entityType {
	case "module":
		var module models.Module
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(dataBytes, &module); err != nil {
			return err
		}
		return cr.db.Save(&module).Error
	case "task":
		var task models.Task
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(dataBytes, &task); err != nil {
			return err
		}
		return cr.db.Save(&task).Error
	}
	return nil
}

// applyMergedData 应用合并后的数据
func (cr *ConflictResolver) applyMergedData(entityType, entityID string, data interface{}) error {
	return cr.applyRemoteData(entityType, entityID, data)
}

// GetUnresolvedConflicts 获取未解决的冲突
func (cr *ConflictResolver) GetUnresolvedConflicts(projectID string) ([]Conflict, error) {
	var conflicts []Conflict
	err := cr.db.Where("resolved_at IS NULL").Order("created_at DESC").Find(&conflicts).Error
	return conflicts, err
}
