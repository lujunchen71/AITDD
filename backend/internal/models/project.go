package models

import "time"

// Project 项目信息
type Project struct {
	ID           string `gorm:"primaryKey;type:text"`
	Name         string `gorm:"not null;type:text"`
	Constitution string `gorm:"type:text"`
	CreatedAt    int64  `gorm:"not null"`
	UpdatedAt    int64  `gorm:"not null"`
	Version      int    `gorm:"not null;default:1"`
	SyncStatus   string `gorm:"not null;default:'SYNCED';type:text"`
}

// SyncStatus 同步状态枚举
const (
	SyncStatusSynced         = "SYNCED"
	SyncStatusPendingUpload  = "PENDING_UPLOAD"
	SyncStatusPendingDownload = "PENDING_DOWNLOAD"
	SyncStatusConflict       = "CONFLICT"
)

// BeforeCreate 创建前钩子
func (p *Project) BeforeCreate() error {
	if p.CreatedAt == 0 {
		p.CreatedAt = time.Now().UnixMilli()
	}
	if p.UpdatedAt == 0 {
		p.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (p *Project) BeforeUpdate() error {
	p.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}
