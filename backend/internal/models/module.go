package models

import "time"

// Module 模块
type Module struct {
	ID                       string  `gorm:"primaryKey;type:text"`
	ParentID                 *string `gorm:"type:text;index"`
	ProjectID                string  `gorm:"not null;type:text;index"`
	Name                     string  `gorm:"not null;type:text"`
	Description              string  `gorm:"type:text"`
	Prompt                   string  `gorm:"type:text"`
	Status                   string  `gorm:"not null;default:'designing';type:text;index"`
	TestCoverage             float64 `gorm:"default:0"`
	UpstreamContractSummary  string  `gorm:"type:text"`
	DownstreamContractSummary string  `gorm:"type:text"`
	Locked                   bool    `gorm:"not null;default:false"`
	LockedBy                 *string `gorm:"type:text"`
	LockedAt                 *int64  `gorm:"type:integer"`
	LockExpiresAt            *int64  `gorm:"type:integer"`
	CreatedAt                int64   `gorm:"not null"`
	UpdatedAt                int64   `gorm:"not null"`
	Version                  int     `gorm:"not null;default:1"`
	SyncStatus               string  `gorm:"not null;default:'SYNCED';type:text"`
}

// ModuleStatus 模块状态枚举
const (
	ModuleStatusDesigning   = "designing"
	ModuleStatusDeveloping  = "developing"
	ModuleStatusCompleted   = "completed"
	ModuleStatusDeprecated  = "deprecated"
)

// BeforeCreate 创建前钩子
func (m *Module) BeforeCreate() error {
	if m.CreatedAt == 0 {
		m.CreatedAt = time.Now().UnixMilli()
	}
	if m.UpdatedAt == 0 {
		m.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Module) BeforeUpdate() error {
	m.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Module) TableName() string {
	return "modules"
}
