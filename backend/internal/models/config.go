package models

import "time"

// Config 配置
type Config struct {
	ID        string `gorm:"primaryKey;type:text"`
	Key       string `gorm:"not null;unique;type:text"`
	Value     string `gorm:"type:text"`
	CreatedAt int64  `gorm:"not null"`
	UpdatedAt int64  `gorm:"not null"`
}

// BeforeCreate 创建前钩子
func (c *Config) BeforeCreate() error {
	if c.CreatedAt == 0 {
		c.CreatedAt = time.Now().UnixMilli()
	}
	if c.UpdatedAt == 0 {
		c.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (c *Config) BeforeUpdate() error {
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
