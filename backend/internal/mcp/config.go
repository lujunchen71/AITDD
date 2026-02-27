// Package mcp 提供 MCP (Model Context Protocol) 服务器功能
// 支持通过 SSE 模式集成到 HTTP 服务器中
package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config 配置文件结构（简化版，只保留必要字段）
type Config struct {
	PathName   string `json:"pathName"`
	ApiBaseUrl string `json:"apiBaseUrl"`
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

// findProjectConfigPath 从当前目录向上查找项目配置文件路径
func findProjectConfigPath() string {
	// 首先检查环境变量
	if envPath := os.Getenv("AITDD_CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return ".aitdd/project.json"
	}

	// 从当前目录向上查找.aitdd目录
	dir := cwd
	for {
		configPath := filepath.Join(dir, ".aitdd", "project.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		// 检查.aitdd目录是否存在
		aitddDir := filepath.Join(dir, ".aitdd")
		if _, err := os.Stat(aitddDir); err == nil {
			return configPath
		}

		// 向上一级目录
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达根目录，使用当前工作目录
			return filepath.Join(cwd, ".aitdd", "project.json")
		}
		dir = parent
	}
}

// GetConfigManager 获取配置管理器单例
func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		configPath := findProjectConfigPath()
		globalConfigManager = &ConfigManager{
			configPath: configPath,
		}
		// 尝试加载配置
		_ = globalConfigManager.Load()
	})
	return globalConfigManager
}

// NewConfigManager 创建新的配置管理器（用于测试或自定义路径）
func NewConfigManager(configPath string) *ConfigManager {
	cm := &ConfigManager{
		configPath: configPath,
	}
	_ = cm.Load()
	return cm
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

// SetProject 设置项目信息（只保存 pathName）
func (cm *ConfigManager) SetProject(projectID, projectName, pathName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if pathName != "" {
		cm.config.PathName = pathName
	}

	return cm.saveWithoutLock()
}

// SetApiBaseUrl 设置API基础URL
func (cm *ConfigManager) SetApiBaseUrl(url string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config.ApiBaseUrl = url

	return cm.saveWithoutLock()
}

// GetProjectPathName 获取项目路径名称
func (cm *ConfigManager) GetProjectPathName() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.PathName
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
	return cm.config != nil && cm.config.PathName != ""
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
