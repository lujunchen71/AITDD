package models

import (
	"time"

	"gorm.io/gorm"
)

// Config 配置
type Config struct {
	ID        string `json:"id" gorm:"primaryKey;type:text"`
	Key       string `json:"key" gorm:"not null;unique;type:text"`
	Value     string `json:"value" gorm:"type:text"`
	CreatedAt int64  `json:"createdAt" gorm:"not null"`
	UpdatedAt int64  `json:"updatedAt" gorm:"not null"`
}

// BeforeCreate 创建前钩子
func (c *Config) BeforeCreate(_ *gorm.DB) error {
	if c.CreatedAt == 0 {
		c.CreatedAt = time.Now().UnixMilli()
	}
	if c.UpdatedAt == 0 {
		c.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (c *Config) BeforeUpdate(_ *gorm.DB) error {
	c.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (Config) TableName() string {
	return "configs"
}

// 预定义配置键
const (
	ConfigKeyPluginType    = "plugin_type"
	ConfigKeyRemoteSync    = "remote_sync_enabled"
	ConfigKeyRemoteURL     = "remote_url"
	ConfigKeyAuthToken     = "auth_token"
	ConfigKeyServerPort    = "server_port"
)
