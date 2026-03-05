package models

import (
	"time"

	"gorm.io/gorm"
)

// Module 模块
type Module struct {
	ID                       string  `json:"id" gorm:"primaryKey;type:text"`
	ParentID                 *string `json:"parentId" gorm:"type:text;index"`
	ProjectID                string  `json:"projectId" gorm:"not null;type:text;index"`
	Name                     string  `json:"name" gorm:"not null;type:text"`
	PathName                 string  `json:"pathName" gorm:"unique;type:text"`
	Description              string  `json:"description" gorm:"type:text"`
	Prompt                   string  `json:"prompt" gorm:"type:text"`
	Status                   string  `json:"status" gorm:"not null;default:'designing';type:text;index"`
	TestCoverage             float64 `json:"testCoverage" gorm:"default:0"`
	BugLog                   string  `json:"bugLog" gorm:"column:bug_log;type:text"` // JSON: BugLog 结构
	UpstreamContractSummary  string  `json:"upstreamContractSummary" gorm:"type:text"`
	DownstreamContractSummary string  `json:"downstreamContractSummary" gorm:"type:text"`
	Locked                   bool    `json:"locked" gorm:"not null;default:false"`
	LockedBy                 *string `json:"lockedBy" gorm:"type:text"`
	LockedAt                 *int64  `json:"lockedAt" gorm:"type:integer"`
	LockExpiresAt            *int64  `json:"lockExpiresAt" gorm:"type:integer"`
	// 位置字段
	PositionX                *float64 `json:"positionX" gorm:"type:real"`
	PositionY                *float64 `json:"positionY" gorm:"type:real"`
	PositionUpdatedAt        *int64   `json:"positionUpdatedAt" gorm:"type:integer"`
	CreatedAt                int64   `json:"createdAt" gorm:"not null"`
	UpdatedAt                int64   `json:"updatedAt" gorm:"not null"`
	Version                  int     `json:"version" gorm:"not null;default:1"`
	SyncStatus               string  `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// ModuleStatus 模块状态枚举
const (
	ModuleStatusDesigning   = "designing"
	ModuleStatusDeveloping  = "developing"
	ModuleStatusCompleted   = "completed"
	ModuleStatusDeprecated  = "deprecated"
)

// BeforeCreate 创建前钩子
func (m *Module) BeforeCreate(_ *gorm.DB) error {
	if m.CreatedAt == 0 {
		m.CreatedAt = time.Now().UnixMilli()
	}
	if m.UpdatedAt == 0 {
		m.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Module) BeforeUpdate(_ *gorm.DB) error {
	m.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Module) TableName() string {
	return "modules"
}
