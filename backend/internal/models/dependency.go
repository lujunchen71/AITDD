package models

import (
	"time"

	"gorm.io/gorm"
)

// Dependency 任务依赖关系
type Dependency struct {
	ID             string `json:"id" gorm:"primaryKey;type:text"`
	UpstreamTaskID string `json:"upstreamTaskId" gorm:"not null;type:text;index"`
	DownstreamTaskID string `json:"downstreamTaskId" gorm:"not null;type:text;index"`
	ContractSummary string `json:"contractSummary" gorm:"type:text"`
	CreatedAt      int64  `json:"createdAt" gorm:"not null"`
	Version        int    `json:"version" gorm:"not null;default:1"`
	SyncStatus     string `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// BeforeCreate 创建前钩子
func (d *Dependency) BeforeCreate(_ *gorm.DB) error {
	if d.CreatedAt == 0 {
		d.CreatedAt = time.Now().UnixMilli()
	}
	return nil
}

// TableName 指定表名
func (Dependency) TableName() string {
	return "dependencies"
}
