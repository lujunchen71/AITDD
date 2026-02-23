package models

import "time"

// Task 任务
type Task struct {
	ID                     string  `gorm:"primaryKey;type:text"`
	ModuleID               string  `gorm:"not null;type:text;index"`
	Name                   string  `gorm:"not null;type:text"`
	Description            string  `gorm:"type:text"`
	Status                 string  `gorm:"not null;default:'ready';type:text;index"`
	Assignee               *string `gorm:"type:text"`
	UpstreamContractDetail string  `gorm:"type:text"`
	DownstreamContractDetail string `gorm:"type:text"`
	Prompt                 string  `gorm:"type:text"`
	Tests                  string  `gorm:"type:text"` // JSON array
	Logs                   string  `gorm:"type:text"` // JSON array
	CodePaths              string  `gorm:"type:text"` // JSON array
	HumanAssistance        string  `gorm:"type:text"` // JSON object
	Locked                 bool    `gorm:"not null;default:false"`
	LockedBy               *string `gorm:"type:text"`
	LockedAt               *int64  `gorm:"type:integer"`
	LockExpiresAt          *int64  `gorm:"type:integer"`
	CreatedAt              int64   `gorm:"not null"`
	UpdatedAt              int64   `gorm:"not null"`
	Version                int     `gorm:"not null;default:1"`
	SyncStatus             string  `gorm:"not null;default:'SYNCED';type:text"`
}

// TaskStatus 任务状态枚举
const (
	TaskStatusReady         = "ready"
	TaskStatusClaimed       = "claimed"
	TaskStatusInProgress    = "in_progress"
	TaskStatusPendingReview = "pending_review"
	TaskStatusCompleted     = "completed"
	TaskStatusFailed        = "failed"
	TaskStatusBlocked       = "blocked"
)

// BeforeCreate 创建前钩子
func (t *Task) BeforeCreate() error {
	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().UnixMilli()
	}
	if t.UpdatedAt == 0 {
		t.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (t *Task) BeforeUpdate() error {
	t.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}
