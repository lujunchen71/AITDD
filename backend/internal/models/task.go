package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务
type Task struct {
	ID                     string  `json:"id" gorm:"primaryKey;type:text"`
	ModuleID               string  `json:"moduleId" gorm:"not null;type:text;index"`
	Name                   string  `json:"name" gorm:"not null;type:text"`
	PathName               string  `json:"pathName" gorm:"unique;type:text"`
	Description            string  `json:"description" gorm:"type:text"`
	Status                 string  `json:"status" gorm:"not null;default:'ready';type:text;index"`
	Assignee               *string `json:"assignee" gorm:"type:text"`
	UpstreamContractDetail string  `json:"upstreamContractDetail" gorm:"type:text"`
	DownstreamContractDetail string  `json:"downstreamContractDetail" gorm:"type:text"`
	Prompt                 string  `json:"prompt" gorm:"type:text"`
	Tests                  string  `json:"tests" gorm:"type:text"` // JSON array of TestItem objects
	TestResult             string  `json:"testResult" gorm:"type:text"` // JSON array of test evidence strings
	BugLog                 string  `json:"bugLog" gorm:"column:bug_log;type:text"` // JSON array
	CodePaths              string  `json:"codePaths" gorm:"type:text"` // JSON array
	HumanAssistance        string  `json:"humanAssistance" gorm:"type:text"` // JSON object
	IssueDetails           string  `json:"issueDetails" gorm:"column:issue_details;type:text;default:''"` // 新增字段
	Locked                 bool    `json:"locked" gorm:"not null;default:false"`
	LockedBy               *string `json:"lockedBy" gorm:"type:text"`
	LockedAt               *int64  `json:"lockedAt" gorm:"type:integer"`
	LockExpiresAt          *int64  `json:"lockExpiresAt" gorm:"type:integer"`
	CreatedAt              int64   `json:"createdAt" gorm:"not null"`
	UpdatedAt              int64   `json:"updatedAt" gorm:"not null"`
	Version                int     `json:"version" gorm:"not null;default:1"`
	SyncStatus             string  `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
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
func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().UnixMilli()
	}
	if t.UpdatedAt == 0 {
		t.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (t *Task) BeforeUpdate(_ *gorm.DB) error {
	t.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}
