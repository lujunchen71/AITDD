package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// 问题严重级别常量
const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
	SeverityError    = "error" // 兼容新格式中的 "error" 级别（等同于 critical）
)

// 问题类型常量
const (
	QuestionTypeTask    = "task"
	QuestionTypeModule  = "module"
	QuestionTypeProject = "project"
)

// QAConfig 问题配置
// 支持两种格式：
//  1. 新格式（推荐）：{ "questions": { "task": [...], "module": [...], "project": [...] } }
//  2. 旧格式（兼容）：{ "task_questions": [...], "module_questions": [...], "project_questions": [...] }
type QAConfig struct {
	// 新格式：嵌套 questions 字段
	Questions *QAQuestions `json:"questions,omitempty"`
	// 旧格式（兼容）：扁平字段
	TaskQuestions    []Question `json:"task_questions,omitempty"`
	ModuleQuestions  []Question `json:"module_questions,omitempty"`
	ProjectQuestions []Question `json:"project_questions,omitempty"`
}

// QAQuestions 新格式的问题列表容器
type QAQuestions struct {
	Task    []Question `json:"task"`
	Module  []Question `json:"module"`
	Project []Question `json:"project"`
}

// Question 问题定义
// ID 字段为可选，如未提供则不使用 ID 进行标识
type Question struct {
	ID       string `json:"id,omitempty"`
	Category string `json:"category"`
	Question string `json:"question"`
	Severity string `json:"severity"` // critical/error, warning, info
}

// GetTaskQuestions 获取任务问题列表（自动适配新旧格式）
func (c *QAConfig) GetTaskQuestions() []Question {
	if c == nil {
		return nil
	}
	if c.Questions != nil {
		return c.Questions.Task
	}
	return c.TaskQuestions
}

// GetModuleQuestions 获取模块问题列表（自动适配新旧格式）
func (c *QAConfig) GetModuleQuestions() []Question {
	if c == nil {
		return nil
	}
	if c.Questions != nil {
		return c.Questions.Module
	}
	return c.ModuleQuestions
}

// GetProjectQuestions 获取项目问题列表（自动适配新旧格式）
func (c *QAConfig) GetProjectQuestions() []Question {
	if c == nil {
		return nil
	}
	if c.Questions != nil {
		return c.Questions.Project
	}
	return c.ProjectQuestions
}

// IsEffectiveSeverity 判断 severity 是否为有效的严重级别（兼容 error/critical）
func IsEffectiveSeverity(severity string) bool {
	return severity == SeverityCritical ||
		severity == SeverityError ||
		severity == SeverityWarning ||
		severity == SeverityInfo
}

// NormalizeSeverity 将 "error" 统一映射为 "critical"
func NormalizeSeverity(severity string) string {
	if severity == SeverityError {
		return SeverityCritical
	}
	return severity
}

// QAConfigManager 问题配置管理器
type QAConfigManager struct {
	configPath string
	config     *QAConfig
	mu         sync.RWMutex
}

// 全局问题配置管理器实例
var globalQAConfigManager *QAConfigManager
var qaConfigOnce sync.Once

// getQAConfigPath 获取问题配置文件路径
func getQAConfigPath() string {
	// 首先检查环境变量
	if envPath := os.Getenv("AITDD_QA_CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "backend/config/qa_config.json"
	}

	// 从当前目录向上查找 backend/config 目录
	dir := cwd
	for {
		configPath := filepath.Join(dir, "backend", "config", "qa_config.json")
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
			return filepath.Join(cwd, "backend", "config", "qa_config.json")
		}
		dir = parent
	}
}

// GetQAConfigManager 获取问题配置管理器单例
func GetQAConfigManager() *QAConfigManager {
	qaConfigOnce.Do(func() {
		configPath := getQAConfigPath()
		globalQAConfigManager = &QAConfigManager{
			configPath: configPath,
		}
		// 尝试加载配置
		_ = globalQAConfigManager.Load()
	})
	return globalQAConfigManager
}

// NewQAConfigManager 创建新的问题配置管理器（用于测试或自定义路径）
func NewQAConfigManager(configPath string) *QAConfigManager {
	cm := &QAConfigManager{
		configPath: configPath,
	}
	_ = cm.Load()
	return cm
}

// GetDefaultQAConfig 获取默认问题配置（使用新格式，无 ID 字段）
func GetDefaultQAConfig() *QAConfig {
	return &QAConfig{
		Questions: &QAQuestions{
			Task: []Question{
				{Category: "prompt", Question: "请分析当前任务的提示词是否清晰、完整，能否指导开发人员正确实现功能？", Severity: SeverityError},
				{Category: "test", Question: "请分析当前任务的测试用例是否合理，能否覆盖主要功能场景？", Severity: SeverityWarning},
				{Category: "contract", Question: "请分析当前任务的上游接口定义是否完整，是否包含所有必要的输入信息？", Severity: SeverityError},
				{Category: "contract", Question: "请分析当前任务的接口描述是否清晰，参数和返回值是否定义明确？", Severity: SeverityWarning},
			},
			Module: []Question{
				{Category: "dependency", Question: "请分析当前模块的依赖关系是否合理，是否存在不必要的依赖或缺失的依赖？", Severity: SeverityError},
				{Category: "description", Question: "请分析当前模块的功能描述是否合理，是否清晰定义了模块的职责边界？", Severity: SeverityWarning},
				{Category: "task", Question: "请分析当前模块的任务粒度是否合理，是否存在过大或过小的任务划分？", Severity: SeverityWarning},
				{Category: "architecture", Question: "请分析当前模块的职责范围是否正确，是否与其他模块存在职责重叠？", Severity: SeverityError},
				{Category: "description", Question: "请分析当前模块的职责描述是否清晰，能否让开发人员理解模块的核心功能？", Severity: SeverityWarning},
			},
			Project: []Question{
				{Category: "requirement", Question: "请分析当前项目是否有需要澄清的地方，需求是否明确？", Severity: SeverityInfo},
				{Category: "architecture", Question: "请分析当前项目的架构选型是否合理，技术栈选择是否适合项目需求？", Severity: SeverityError},
				{Category: "architecture", Question: "请分析当前项目的模块耦合性设计是否合理，模块间依赖是否清晰？", Severity: SeverityWarning},
				{Category: "consistency", Question: "请分析当前项目的接口设计风格是否一致，是否遵循统一的设计规范？", Severity: SeverityWarning},
				{Category: "maintainability", Question: "请分析当前项目的整体设计是否易于维护和扩展？", Severity: SeverityInfo},
			},
		},
	}
}

// LoadQAConfig 从配置文件加载问题配置（全局函数）
func LoadQAConfig() (*QAConfig, error) {
	manager := GetQAConfigManager()
	return manager.GetConfig(), nil
}

// Load 加载配置文件
func (m *QAConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// 创建默认配置
		m.config = GetDefaultQAConfig()
		Debug("问题配置文件不存在，使用默认配置", map[string]interface{}{"path": m.configPath})
		return m.saveWithoutLock()
	}

	// 读取文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		Error("读取问题配置文件失败", err, map[string]interface{}{"path": m.configPath})
		return fmt.Errorf("读取问题配置文件失败: %w", err)
	}

	// 解析JSON
	var config QAConfig
	if err := json.Unmarshal(data, &config); err != nil {
		Error("解析问题配置文件失败", err, map[string]interface{}{"path": m.configPath})
		return fmt.Errorf("解析问题配置文件失败: %w", err)
	}

	m.config = &config
	Info("问题配置加载成功", map[string]interface{}{
		"path":             m.configPath,
		"task_questions":   len(config.GetTaskQuestions()),
		"module_questions": len(config.GetModuleQuestions()),
		"project_questions": len(config.GetProjectQuestions()),
	})
	return nil
}

// SaveQAConfig 保存问题配置到文件（全局函数）
func SaveQAConfig(config *QAConfig) error {
	manager := GetQAConfigManager()
	return manager.SaveConfig(config)
}

// Save 保存当前配置
func (m *QAConfigManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveWithoutLock()
}

// SaveConfig 保存指定配置
func (m *QAConfigManager) SaveConfig(config *QAConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	return m.saveWithoutLock()
}

// saveWithoutLock 不加锁的保存方法
func (m *QAConfigManager) saveWithoutLock() error {
	// 确保目录存在
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		Error("创建问题配置目录失败", err, map[string]interface{}{"dir": dir})
		return fmt.Errorf("创建问题配置目录失败: %w", err)
	}

	// 序列化JSON
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		Error("序列化问题配置失败", err)
		return fmt.Errorf("序列化问题配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		Error("写入问题配置文件失败", err, map[string]interface{}{"path": m.configPath})
		return fmt.Errorf("写入问题配置文件失败: %w", err)
	}

	Info("问题配置保存成功", map[string]interface{}{"path": m.configPath})
	return nil
}

// GetConfig 获取当前配置
func (m *QAConfigManager) GetConfig() *QAConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return GetDefaultQAConfig()
	}
	return m.config
}

// GetQuestionsByType 按类型获取问题
func GetQuestionsByType(config *QAConfig, questionType string) ([]Question, error) {
	if config == nil {
		return nil, NewParamError("配置不能为空")
	}

	switch questionType {
	case QuestionTypeTask:
		return config.GetTaskQuestions(), nil
	case QuestionTypeModule:
		return config.GetModuleQuestions(), nil
	case QuestionTypeProject:
		return config.GetProjectQuestions(), nil
	default:
		return nil, NewParamErrorf("未知的问题类型: %s", questionType)
	}
}

// GetAllQuestions 获取所有问题
func GetAllQuestions(config *QAConfig) []Question {
	if config == nil {
		return nil
	}

	taskQ := config.GetTaskQuestions()
	moduleQ := config.GetModuleQuestions()
	projectQ := config.GetProjectQuestions()

	all := make([]Question, 0, len(taskQ)+len(moduleQ)+len(projectQ))
	all = append(all, taskQ...)
	all = append(all, moduleQ...)
	all = append(all, projectQ...)
	return all
}

// GetQuestionsByCategory 按类别获取问题
func GetQuestionsByCategory(config *QAConfig, category string) []Question {
	if config == nil {
		return nil
	}

	var result []Question
	for _, q := range GetAllQuestions(config) {
		if q.Category == category {
			result = append(result, q)
		}
	}
	return result
}

// GetQuestionsBySeverity 按严重级别获取问题（兼容 error/critical）
func GetQuestionsBySeverity(config *QAConfig, severity string) []Question {
	if config == nil {
		return nil
	}

	var result []Question
	for _, q := range GetAllQuestions(config) {
		normalizedQ := NormalizeSeverity(q.Severity)
		normalizedTarget := NormalizeSeverity(severity)
		if normalizedQ == normalizedTarget {
			result = append(result, q)
		}
	}
	return result
}

// GetCriticalQuestions 获取所有关键问题（包含 error 级别）
func GetCriticalQuestions(config *QAConfig) []Question {
	if config == nil {
		return nil
	}
	var result []Question
	for _, q := range GetAllQuestions(config) {
		if q.Severity == SeverityCritical || q.Severity == SeverityError {
			result = append(result, q)
		}
	}
	return result
}

// GetWarningQuestions 获取所有警告问题
func GetWarningQuestions(config *QAConfig) []Question {
	return GetQuestionsBySeverity(config, SeverityWarning)
}

// GetInfoQuestions 获取所有信息级别问题
func GetInfoQuestions(config *QAConfig) []Question {
	return GetQuestionsBySeverity(config, SeverityInfo)
}

// ValidateQAConfig 验证问题配置有效性
func ValidateQAConfig(config *QAConfig) error {
	if config == nil {
		return NewParamError("配置不能为空")
	}

	// 验证任务问题
	if err := validateQuestions(config.GetTaskQuestions(), "task"); err != nil {
		return err
	}

	// 验证模块问题
	if err := validateQuestions(config.GetModuleQuestions(), "module"); err != nil {
		return err
	}

	// 验证项目问题
	if err := validateQuestions(config.GetProjectQuestions(), "project"); err != nil {
		return err
	}

	Debug("问题配置验证通过")
	return nil
}

// validateQuestions 验证问题列表（ID 为可选字段）
func validateQuestions(questions []Question, questionType string) error {
	seenIDs := make(map[string]bool)

	for i, q := range questions {
		// ID 为可选字段，如果提供则检查重复
		if q.ID != "" {
			if seenIDs[q.ID] {
				return NewParamErrorf("%s 问题 ID 重复: %s", questionType, q.ID)
			}
			seenIDs[q.ID] = true
		}

		// 检查问题内容
		if q.Question == "" {
			label := q.ID
			if label == "" {
				label = fmt.Sprintf("index_%d", i)
			}
			return NewParamErrorf("%s 问题 %s 内容不能为空", questionType, label)
		}

		// 检查严重级别
		if !IsEffectiveSeverity(q.Severity) {
			label := q.ID
			if label == "" {
				label = fmt.Sprintf("index_%d", i)
			}
			return NewParamErrorf("%s 问题 %s 严重级别无效: %s", questionType, label, q.Severity)
		}
	}

	return nil
}

// AddQuestion 添加问题
func (m *QAConfigManager) AddQuestion(questionType string, question Question) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 只有在提供了ID的情况下才检查重复
	if question.ID != "" && m.isIDExists(question.ID) {
		return NewBusinessErrorf("问题 ID 已存在: %s", question.ID)
	}

	// 确保使用新格式
	if m.config.Questions == nil {
		m.config.Questions = &QAQuestions{
			Task:    m.config.TaskQuestions,
			Module:  m.config.ModuleQuestions,
			Project: m.config.ProjectQuestions,
		}
		m.config.TaskQuestions = nil
		m.config.ModuleQuestions = nil
		m.config.ProjectQuestions = nil
	}

	switch questionType {
	case QuestionTypeTask:
		m.config.Questions.Task = append(m.config.Questions.Task, question)
	case QuestionTypeModule:
		m.config.Questions.Module = append(m.config.Questions.Module, question)
	case QuestionTypeProject:
		m.config.Questions.Project = append(m.config.Questions.Project, question)
	default:
		return NewParamErrorf("未知的问题类型: %s", questionType)
	}

	return m.saveWithoutLock()
}

// RemoveQuestion 移除问题（按 ID 移除，仅适用于有 ID 的问题）
func (m *QAConfigManager) RemoveQuestion(questionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	found := false

	if m.config.Questions != nil {
		m.config.Questions.Task, found = removeQuestionFromList(m.config.Questions.Task, questionID, found)
		m.config.Questions.Module, found = removeQuestionFromList(m.config.Questions.Module, questionID, found)
		m.config.Questions.Project, found = removeQuestionFromList(m.config.Questions.Project, questionID, found)
	} else {
		// 旧格式
		m.config.TaskQuestions, found = removeQuestionFromList(m.config.TaskQuestions, questionID, found)
		m.config.ModuleQuestions, found = removeQuestionFromList(m.config.ModuleQuestions, questionID, found)
		m.config.ProjectQuestions, found = removeQuestionFromList(m.config.ProjectQuestions, questionID, found)
	}

	if !found {
		return NewNotFoundError("问题", questionID)
	}

	return m.saveWithoutLock()
}

// removeQuestionFromList 从列表中移除问题（按 ID）
func removeQuestionFromList(questions []Question, id string, found bool) ([]Question, bool) {
	newList := make([]Question, 0, len(questions))
	for _, q := range questions {
		if q.ID == id {
			found = true
			continue
		}
		newList = append(newList, q)
	}
	return newList, found
}

// UpdateQuestion 更新问题（按 ID 更新，仅适用于有 ID 的问题）
func (m *QAConfigManager) UpdateQuestion(question Question) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if question.ID == "" {
		return NewParamError("更新问题需要提供 ID")
	}

	found := false

	if m.config.Questions != nil {
		// 新格式
		for i := range m.config.Questions.Task {
			if m.config.Questions.Task[i].ID == question.ID {
				m.config.Questions.Task[i] = question
				found = true
				break
			}
		}
		if !found {
			for i := range m.config.Questions.Module {
				if m.config.Questions.Module[i].ID == question.ID {
					m.config.Questions.Module[i] = question
					found = true
					break
				}
			}
		}
		if !found {
			for i := range m.config.Questions.Project {
				if m.config.Questions.Project[i].ID == question.ID {
					m.config.Questions.Project[i] = question
					found = true
					break
				}
			}
		}
	} else {
		// 旧格式
		for i := range m.config.TaskQuestions {
			if m.config.TaskQuestions[i].ID == question.ID {
				m.config.TaskQuestions[i] = question
				found = true
				break
			}
		}
		if !found {
			for i := range m.config.ModuleQuestions {
				if m.config.ModuleQuestions[i].ID == question.ID {
					m.config.ModuleQuestions[i] = question
					found = true
					break
				}
			}
		}
		if !found {
			for i := range m.config.ProjectQuestions {
				if m.config.ProjectQuestions[i].ID == question.ID {
					m.config.ProjectQuestions[i] = question
					found = true
					break
				}
			}
		}
	}

	if !found {
		return NewNotFoundError("问题", question.ID)
	}

	return m.saveWithoutLock()
}

// isIDExists 检查ID是否已存在
func (m *QAConfigManager) isIDExists(id string) bool {
	for _, q := range m.config.GetTaskQuestions() {
		if q.ID == id {
			return true
		}
	}
	for _, q := range m.config.GetModuleQuestions() {
		if q.ID == id {
			return true
		}
	}
	for _, q := range m.config.GetProjectQuestions() {
		if q.ID == id {
			return true
		}
	}
	return false
}

// ToJSON 将配置转换为JSON字符串
func (m *QAConfigManager) ToJSON() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetQuestionByID 根据ID获取问题
func GetQuestionByID(config *QAConfig, id string) (*Question, error) {
	if config == nil {
		return nil, NewParamError("配置不能为空")
	}

	for _, q := range GetAllQuestions(config) {
		if q.ID == id {
			return &q, nil
		}
	}

	return nil, NewNotFoundError("问题", id)
}

// QuestionCount 问题统计
type QuestionCount struct {
	Total   int `json:"total"`
	Task    int `json:"task"`
	Module  int `json:"module"`
	Project int `json:"project"`
	Critical int `json:"critical"`
	Warning int `json:"warning"`
	Info    int `json:"info"`
}

// GetQuestionCount 获取问题统计
func GetQuestionCount(config *QAConfig) *QuestionCount {
	if config == nil {
		return &QuestionCount{}
	}

	return &QuestionCount{
		Total:    len(GetAllQuestions(config)),
		Task:     len(config.GetTaskQuestions()),
		Module:   len(config.GetModuleQuestions()),
		Project:  len(config.GetProjectQuestions()),
		Critical: len(GetCriticalQuestions(config)),
		Warning:  len(GetWarningQuestions(config)),
		Info:     len(GetInfoQuestions(config)),
	}
}
