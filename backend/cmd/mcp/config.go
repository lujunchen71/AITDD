package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config 配置文件结构
type Config struct {
	ProjectID   string    `json:"projectId"`
	ProjectName string    `json:"projectName"`
	ApiBaseUrl  string    `json:"apiBaseUrl"`
	MCP         MCPConfig `json:"mcp"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

// MCPConfig MCP配置
type MCPConfig struct {
	ServerName  string `json:"serverName"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

// 全局配置管理器实例
var globalConfigManager *ConfigManager
var configOnce sync.Once

// GetConfigManager 获取配置管理器单例
func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		configPath := ".aitdd/config.json"
		globalConfigManager = &ConfigManager{
			configPath: configPath,
		}
		// 尝试加载配置
		_ = globalConfigManager.Load()
	})
	return globalConfigManager
}

// Load 加载配置文件
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// 创建默认配置
		cm.config = &Config{
			ApiBaseUrl: "http://localhost:34567/api/v1",
			MCP: MCPConfig{
				ServerName:  "aitdd-mcp",
				Version:     "1.0.0",
				Description: "AITDD数据库增删改查MCP服务",
			},
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		return cm.saveWithoutLock()
	}

	// 读取文件
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	cm.config = &config
	return nil
}

// Save 保存配置文件
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.saveWithoutLock()
}

// saveWithoutLock 不加锁的保存方法
func (cm *ConfigManager) saveWithoutLock() error {
	// 确保目录存在
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 更新时间
	cm.config.UpdatedAt = time.Now().Format(time.RFC3339)

	// 序列化JSON
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetConfig 获取当前配置
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// SetProject 设置项目信息
func (cm *ConfigManager) SetProject(projectID, projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config.ProjectID = projectID
	cm.config.ProjectName = projectName

	return cm.saveWithoutLock()
}

// SetApiBaseUrl 设置API基础URL
func (cm *ConfigManager) SetApiBaseUrl(url string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config.ApiBaseUrl = url

	return cm.saveWithoutLock()
}

// GetProjectID 获取项目ID
func (cm *ConfigManager) GetProjectID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.ProjectID
}

// GetApiBaseUrl 获取API基础URL
func (cm *ConfigManager) GetApiBaseUrl() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.config.ApiBaseUrl == "" {
		return "http://localhost:34567/api/v1"
	}
	return cm.config.ApiBaseUrl
}

// IsConfigured 检查是否已配置项目
func (cm *ConfigManager) IsConfigured() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config != nil && cm.config.ProjectID != ""
}

// ToJSON 将配置转换为JSON字符串
func (cm *ConfigManager) ToJSON() (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProjectInfo 项目信息（用于项目列表）
type ProjectInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	Projects []ProjectInfo `json:"projects"`
	Total    int           `json:"total"`
}
