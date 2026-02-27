package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigService 配置服务
type ConfigService struct {
	configPath string
}

// AppConfig 应用配置
type AppConfig struct {
	ProjectName string     `json:"projectName"`
	PathName    string     `json:"pathName"`
	PluginType  string     `json:"pluginType"`
	ServerPort  int        `json:"serverPort"`
	Database    DBConfig   `json:"database"`
	Sync        SyncConfig `json:"sync"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	Enabled   bool   `json:"enabled"`
	RemoteURL string `json:"remoteUrl"`
}

// NewConfigService 创建配置服务
func NewConfigService(configPath string) *ConfigService {
	return &ConfigService{
		configPath: configPath,
	}
}

// Load 加载配置
func (s *ConfigService) Load() (*AppConfig, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// Save 保存配置
func (s *ConfigService) Save(config *AppConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// CreateDefault 创建默认配置
func (s *ConfigService) CreateDefault(projectName, pluginType string) error {
	config := &AppConfig{
		ProjectName: projectName,
		PathName:    projectName,
		PluginType:  pluginType,
		ServerPort:  34567,
		Database: DBConfig{
			Type: "sqlite",
			Path: ".aitdd/data/aitdd.db",
		},
		Sync: SyncConfig{
			Enabled:   false,
			RemoteURL: "",
		},
	}

	// 确保目录存在
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	return s.Save(config)
}

// Update 更新配置
func (s *ConfigService) Update(updates func(*AppConfig)) error {
	config, err := s.Load()
	if err != nil {
		return err
	}

	updates(config)

	return s.Save(config)
}
