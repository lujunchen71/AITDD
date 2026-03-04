package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SubAgentConfig AI子代理配置
type SubAgentConfig struct {
	Models    []ModelConfig    `json:"models"`
	RateLimit RateLimitConfig  `json:"rate_limit"`
	Timeout   TimeoutConfig    `json:"timeout"`
	Retry     RetryConfig      `json:"retry"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	Endpoint  string `json:"endpoint"`
	Model     string `json:"model"`
	IsPrimary bool   `json:"is_primary"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	MaxConcurrent int `json:"max_concurrent"`
	RequestsPerMin int `json:"requests_per_min"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Connect int `json:"connect"` // 秒
	Request int `json:"request"` // 秒
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int `json:"max_retries"`
	BackoffMs  int `json:"backoff_ms"`
}

// SubAgentConfigManager 子代理配置管理器
type SubAgentConfigManager struct {
	configPath string
	config     *SubAgentConfig
	mu         sync.RWMutex
}

// 全局子代理配置管理器实例
var globalSubAgentConfigManager *SubAgentConfigManager
var subAgentConfigOnce sync.Once

// getDefaultSubAgentConfigPath 获取默认子代理配置文件路径
func getDefaultSubAgentConfigPath() string {
	// 首先检查环境变量
	if envPath := os.Getenv("AITDD_SUB_AGENT_CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "backend/config/sub_agent.json"
	}

	// 从当前目录向上查找 backend/config 目录
	dir := cwd
	for {
		configPath := filepath.Join(dir, "backend", "config", "sub_agent.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		// 检查 backend/config 目录是否存在
		configDir := filepath.Join(dir, "backend", "config")
		if _, err := os.Stat(configDir); err == nil {
			return configPath
		}

		// 向上一级目录
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达根目录，使用当前工作目录
			return filepath.Join(cwd, "backend", "config", "sub_agent.json")
		}
		dir = parent
	}
}

// GetSubAgentConfigManager 获取子代理配置管理器单例
func GetSubAgentConfigManager() *SubAgentConfigManager {
	subAgentConfigOnce.Do(func() {
		configPath := getDefaultSubAgentConfigPath()
		globalSubAgentConfigManager = &SubAgentConfigManager{
			configPath: configPath,
		}
		// 尝试加载配置
		_ = globalSubAgentConfigManager.Load()
	})
	return globalSubAgentConfigManager
}

// NewSubAgentConfigManager 创建新的子代理配置管理器（用于测试或自定义路径）
func NewSubAgentConfigManager(configPath string) *SubAgentConfigManager {
	cm := &SubAgentConfigManager{
		configPath: configPath,
	}
	_ = cm.Load()
	return cm
}

// GetDefaultSubAgentConfig 获取默认子代理配置
func GetDefaultSubAgentConfig() *SubAgentConfig {
	return &SubAgentConfig{
		Models: []ModelConfig{
			{
				Name:      "primary",
				APIKey:    "${AI_API_KEY}",
				Endpoint:  "https://api.openai.com/v1/chat/completions",
				IsPrimary: true,
			},
		},
		RateLimit: RateLimitConfig{
			MaxConcurrent: 3,
			RequestsPerMin: 20,
		},
		Timeout: TimeoutConfig{
			Connect: 10,
			Request: 60,
		},
		Retry: RetryConfig{
			MaxRetries: 3,
			BackoffMs:  1000,
		},
	}
}

// LoadSubAgentConfig 从配置文件加载配置（全局函数）
func LoadSubAgentConfig() (*SubAgentConfig, error) {
	manager := GetSubAgentConfigManager()
	return manager.GetConfig(), nil
}

// Load 加载配置文件
func (m *SubAgentConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// 创建默认配置
		m.config = GetDefaultSubAgentConfig()
		Debug("子代理配置文件不存在，使用默认配置", map[string]interface{}{"path": m.configPath})
		return m.saveWithoutLock()
	}

	// 读取文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		Error("读取子代理配置文件失败", err, map[string]interface{}{"path": m.configPath})
		return fmt.Errorf("读取子代理配置文件失败: %w", err)
	}

	// 解析JSON
	var config SubAgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		Error("解析子代理配置文件失败", err, map[string]interface{}{"path": m.configPath})
		return fmt.Errorf("解析子代理配置文件失败: %w", err)
	}

	m.config = &config
	Info("子代理配置加载成功", map[string]interface{}{"path": m.configPath, "models": len(config.Models)})
	return nil
}

// SaveSubAgentConfig 保存配置到文件（全局函数）
func SaveSubAgentConfig(config *SubAgentConfig) error {
	manager := GetSubAgentConfigManager()
	return manager.SaveConfig(config)
}

// Save 保存当前配置
func (m *SubAgentConfigManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveWithoutLock()
}

// SaveConfig 保存指定配置
func (m *SubAgentConfigManager) SaveConfig(config *SubAgentConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	return m.saveWithoutLock()
}

// saveWithoutLock 不加锁的保存方法
func (m *SubAgentConfigManager) saveWithoutLock() error {
	// 确保目录存在
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		Error("创建子代理配置目录失败", err, map[string]interface{}{"dir": dir})
		return fmt.Errorf("创建子代理配置目录失败: %w", err)
	}

	// 序列化JSON
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		Error("序列化子代理配置失败", err)
		return fmt.Errorf("序列化子代理配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		Error("写入子代理配置文件失败", err, map[string]interface{}{"path": m.configPath})
		return fmt.Errorf("写入子代理配置文件失败: %w", err)
	}

	Info("子代理配置保存成功", map[string]interface{}{"path": m.configPath})
	return nil
}

// ValidateSubAgentConfig 验证配置有效性
func ValidateSubAgentConfig(config *SubAgentConfig) error {
	if config == nil {
		return NewParamError("配置不能为空")
	}

	// 验证模型配置
	if len(config.Models) == 0 {
		return NewParamError("至少需要配置一个模型")
	}

	hasPrimary := false
	for i, model := range config.Models {
		if model.Name == "" {
			return NewParamErrorf("模型 %d 名称不能为空", i)
		}
		if model.Endpoint == "" {
			return NewParamErrorf("模型 %s 端点不能为空", model.Name)
		}
		if model.IsPrimary {
			if hasPrimary {
				return NewParamErrorf("只能有一个主模型，发现多个: %s", model.Name)
			}
			hasPrimary = true
		}
	}

	// 验证速率限制配置
	if config.RateLimit.MaxConcurrent <= 0 {
		return NewParamError("最大并发数必须大于0")
	}
	if config.RateLimit.RequestsPerMin <= 0 {
		return NewParamError("每分钟请求数必须大于0")
	}

	// 验证超时配置
	if config.Timeout.Connect <= 0 {
		return NewParamError("连接超时必须大于0")
	}
	if config.Timeout.Request <= 0 {
		return NewParamError("请求超时必须大于0")
	}

	// 验证重试配置
	if config.Retry.MaxRetries < 0 {
		return NewParamError("最大重试次数不能为负数")
	}
	if config.Retry.BackoffMs <= 0 {
		return NewParamError("退避时间必须大于0")
	}

	Debug("子代理配置验证通过")
	return nil
}

// GetConfig 获取当前配置
func (m *SubAgentConfigManager) GetConfig() *SubAgentConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return GetDefaultSubAgentConfig()
	}
	return m.config
}

// GetPrimaryModel 获取主模型配置
func GetPrimaryModel(config *SubAgentConfig) (*ModelConfig, error) {
	if config == nil {
		return nil, NewParamError("配置不能为空")
	}

	for i := range config.Models {
		if config.Models[i].IsPrimary {
			return &config.Models[i], nil
		}
	}

	// 如果没有标记为主模型，返回第一个模型
	if len(config.Models) > 0 {
		Warn("未找到主模型，使用第一个模型作为主模型")
		return &config.Models[0], nil
	}

	return nil, NewNotFoundError("模型", "primary")
}

// GetFallbackModel 获取备用模型配置
func GetFallbackModel(config *SubAgentConfig) (*ModelConfig, error) {
	if config == nil {
		return nil, NewParamError("配置不能为空")
	}

	for i := range config.Models {
		if !config.Models[i].IsPrimary {
			return &config.Models[i], nil
		}
	}

	return nil, NewNotFoundError("模型", "fallback")
}

// GetAllModels 获取所有模型配置
func GetAllModels(config *SubAgentConfig) ([]ModelConfig, error) {
	if config == nil {
		return nil, NewParamError("配置不能为空")
	}
	return config.Models, nil
}

// ExpandAPIKey 展开API密钥中的环境变量
func (m *ModelConfig) ExpandAPIKey() string {
	apiKey := m.APIKey
	// 如果API密钥是环境变量引用格式 ${VAR_NAME}
	if len(apiKey) > 3 && apiKey[0:2] == "${" && apiKey[len(apiKey)-1] == '}' {
		envVar := apiKey[2 : len(apiKey)-1]
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
		Debug("环境变量未设置，使用原始API密钥", map[string]interface{}{"envVar": envVar})
	}
	return apiKey
}

// ToJSON 将配置转换为JSON字符串
func (m *SubAgentConfigManager) ToJSON() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpdateModel 更新指定模型配置
func (m *SubAgentConfigManager) UpdateModel(name string, updates ModelConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	found := false
	for i := range m.config.Models {
		if m.config.Models[i].Name == name {
			// 更新字段
			if updates.APIKey != "" {
				m.config.Models[i].APIKey = updates.APIKey
			}
			if updates.Endpoint != "" {
				m.config.Models[i].Endpoint = updates.Endpoint
			}
			m.config.Models[i].IsPrimary = updates.IsPrimary
			found = true
			break
		}
	}

	if !found {
		return NewNotFoundError("模型", name)
	}

	return m.saveWithoutLock()
}

// AddModel 添加新模型
func (m *SubAgentConfigManager) AddModel(model ModelConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	for _, existing := range m.config.Models {
		if existing.Name == model.Name {
			return NewBusinessErrorf("模型 %s 已存在", model.Name)
		}
	}

	m.config.Models = append(m.config.Models, model)
	return m.saveWithoutLock()
}

// RemoveModel 移除模型
func (m *SubAgentConfigManager) RemoveModel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	found := false
	newModels := make([]ModelConfig, 0, len(m.config.Models))
	for _, model := range m.config.Models {
		if model.Name == name {
			found = true
			continue
		}
		newModels = append(newModels, model)
	}

	if !found {
		return NewNotFoundError("模型", name)
	}

	m.config.Models = newModels
	return m.saveWithoutLock()
}
