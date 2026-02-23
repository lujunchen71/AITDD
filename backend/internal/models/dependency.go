package models

import "time"

// Dependency 任务依赖关系
type Dependency struct {
	ID             string `gorm:"primaryKey;type:text"`
	UpstreamTaskID string `gorm:"not null;type:text;index"`
	DownstreamTaskID string `gorm:"not null;type:text;index"`
	ContractSummary string `gorm:"type:text"`
	CreatedAt      int64  `gorm:"not null"`
	Version        int    `gorm:"not null;default:1"`
	SyncStatus     string `gorm:"not null;default:'SYNCED';type:text"`
}

// BeforeCreate 创建前钩子
func (d *Dependency) BeforeCreate() error {
	if d.CreatedAt == 0 {
		d.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (Dependency) TableName() string {
	return "dependencies"
}
