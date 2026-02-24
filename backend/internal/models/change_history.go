package models

import (
	"time"

	"gorm.io/gorm"
)

// ChangeHistory 变更历史
type ChangeHistory struct {
	ID         string `json:"id" gorm:"primaryKey;type:text"`
	EntityType string `json:"entityType" gorm:"not null;type:text;index"` // project, module, task, dependency
	EntityID   string `json:"entityId" gorm:"not null;type:text;index"`
	Action     string `json:"action" gorm:"not null;type:text"` // create, update, delete
	Changes    string `json:"changes" gorm:"type:text"` // JSON object
	ChangedBy  string `json:"changedBy" gorm:"type:text"`
	CreatedAt  int64  `json:"createdAt" gorm:"not null"`
}

// EntityType 实体类型枚举
const (
	EntityTypeProject    = "project"
	EntityTypeModule     = "module"
	EntityTypeTask       = "task"
	EntityTypeDependency = "dependency"
)

// ActionType 操作类型枚举
const (
	ActionTypeCreate = "create"
	ActionTypeUpdate = "update"
	ActionTypeDelete = "delete"
)

// BeforeCreate 创建前钩子
func (c *ChangeHistory) BeforeCreate(_ *gorm.DB) error {
	if c.CreatedAt == 0 {
		c.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (ChangeHistory) TableName() string {
	return "change_history"
}
