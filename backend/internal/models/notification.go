 package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification 通知
type Notification struct {
	ID          string `json:"id" gorm:"primaryKey;type:text"`
	FromTaskID  string `json:"fromTaskId" gorm:"not null;type:text;index"`
	ToTaskID    string `json:"toTaskId" gorm:"not null;type:text;index"`
	Type        string `json:"type" gorm:"not null;type:text"`
	Title       string `json:"title" gorm:"not null;type:text"`
	Content     string `json:"content" gorm:"type:text"`
	Read        bool   `json:"read" gorm:"not null;default:false"`
	CreatedAt   int64  `json:"createdAt" gorm:"not null"`
	SyncStatus  string `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// NotificationType 通知类型枚举
const (
	NotificationTypeTaskCompleted  = "task_completed"
	NotificationTypeTaskFailed     = "task_failed"
	NotificationTypeTaskBlocked    = "task_blocked"
	NotificationTypeReviewRequired = "review_required"
	NotificationTypeContractChange = "contract_change"
)

// BeforeCreate 创建前钩子
func (n *Notification) BeforeCreate(_ *gorm.DB) error {
	if n.CreatedAt == 0 {
		n.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}
