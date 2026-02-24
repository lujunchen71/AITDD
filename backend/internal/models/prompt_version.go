package models

import (
	"time"

	"gorm.io/gorm"
)

// PromptVersion 提示词版本
type PromptVersion struct {
	ID            string `json:"id" gorm:"primaryKey;type:text"`
	EntityType    string `json:"entityType" gorm:"not null;type:text;check:entity_type IN ('module', 'task')"`
	EntityID      string `json:"entityId" gorm:"not null;type:text;index"`
	Version       int    `json:"version" gorm:"not null"`
	Prompt        string `json:"prompt" gorm:"not null;type:text"`
	ChangeSummary string `json:"changeSummary" gorm:"type:text"`
	CreatedBy     string `json:"createdBy" gorm:"type:text"`
	CreatedAt     int64  `json:"createdAt" gorm:"not null"`
}

// BeforeCreate 创建前钩子
func (p *PromptVersion) BeforeCreate(_ *gorm.DB) error {
	if p.CreatedAt == 0 {
		p.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (PromptVersion) TableName() string {
	return "prompt_versions"
}
