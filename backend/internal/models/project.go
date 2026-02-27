package models

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目信息
type Project struct {
	ID           string `json:"id" gorm:"primaryKey;type:text"`
	Name         string `json:"name" gorm:"not null;type:text"`
	PathName     string `json:"pathName" gorm:"unique;type:text"`
	Constitution string `json:"constitution" gorm:"type:text"`
	CreatedAt    int64  `json:"createdAt" gorm:"not null"`
	UpdatedAt    int64  `json:"updatedAt" gorm:"not null"`
	Version      int    `json:"version" gorm:"not null;default:1"`
	SyncStatus   string `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// SyncStatus 同步状态枚举
const (
	SyncStatusSynced         = "SYNCED"
	SyncStatusPendingUpload  = "PENDING_UPLOAD"
	SyncStatusPendingDownload = "PENDING_DOWNLOAD"
	SyncStatusConflict       = "CONFLICT"
)

// BeforeCreate 创建前钩子
func (p *Project) BeforeCreate(_ *gorm.DB) error {
	if p.CreatedAt == 0 {
		p.CreatedAt = time.Now().UnixMilli()
	}
	if p.UpdatedAt == 0 {
		p.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (p *Project) BeforeUpdate(_ *gorm.DB) error {
	p.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}
