package models

import "time"

// Notification 通知
type Notification struct {
	ID          string `gorm:"primaryKey;type:text"`
	FromTaskID  string `gorm:"not null;type:text;index"`
	ToTaskID    string `gorm:"not null;type:text;index"`
	Type        string `gorm:"not null;type:text"`
	Title       string `gorm:"not null;type:text"`
	Content     string `gorm:"type:text"`
	Read        bool   `gorm:"not null;default:false"`
	CreatedAt   int64  `gorm:"not null"`
	SyncStatus  string `gorm:"not null;default:'SYNCED';type:text"`
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
func (n *Notification) BeforeCreate() error {
	if n.CreatedAt == 0 {
		n.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}
