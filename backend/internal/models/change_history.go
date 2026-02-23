package models

import "time"

// ChangeHistory 变更历史
type ChangeHistory struct {
	ID         string `gorm:"primaryKey;type:text"`
	EntityType string `gorm:"not null;type:text;index"` // project, module, task, dependency
	EntityID   string `gorm:"not null;type:text;index"`
	Action     string `gorm:"not null;type:text"` // create, update, delete
	Changes    string `gorm:"type:text"` // JSON object
	ChangedBy  string `gorm:"type:text"`
	CreatedAt  int64  `gorm:"not null"`
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
func (c *ChangeHistory) BeforeCreate() error {
	if c.CreatedAt == 0 {
		c.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (ChangeHistory) TableName() string {
	return "change_history"
}
