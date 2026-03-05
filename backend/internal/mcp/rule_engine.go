package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Rule 规则定义
type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // error, warning, info
	Enabled     bool   `json:"enabled"`
}

// RuleCategory 规则分类
type RuleCategory struct {
	Description string `json:"description"`
	Rules       []Rule `json:"rules"`
}

// RuleConfig 规则配置
type RuleConfig struct {
	Version     string                   `json:"version"`
	Description string                   `json:"description"`
	Rules       map[string]RuleCategory `json:"rules"`
}

// CheckResult 检查结果
type CheckResult struct {
	RuleID       string                 `json:"ruleId"`
	RuleName     string                 `json:"ruleName"`
	Severity     string                 `json:"severity"`
	ResourceType string                 `json:"resourceType"` // module, task, dependency
	ResourceID   string                 `json:"resourceId"`
	ResourceName string                 `json:"resourceName"`
	ModuleID     string                 `json:"moduleId,omitempty"`
	ModuleName   string                 `json:"moduleName,omitempty"`
	Message      string                 `json:"message"`
	Suggestion   string                 `json:"suggestion"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// CheckReport 检查报告
type CheckReport struct {
	ModuleID     string        `json:"moduleId"`
	ModuleName   string        `json:"moduleName"`
	ModuleStatus string        `json:"moduleStatus"`
	CheckTime    string        `json:"checkTime"`
	Summary      CheckSummary  `json:"summary"`
	Errors       []CheckResult `json:"errors"`
	Warnings     []CheckResult `json:"warnings"`
	Suggestions  []CheckResult `json:"suggestions"`
}

// CheckSummary 检查摘要
type CheckSummary struct {
	Total       int `json:"total"`
	Errors      int `json:"errorCount"`
	Warnings    int `json:"warningCount"`
	Suggestions int `json:"suggestionCount"`
}

// RuleEngine 规则引擎
type RuleEngine struct {
	config     *RuleConfig
	configPath string
	mu         sync.RWMutex
}

// 全局规则引擎实例
var globalRuleEngine *RuleEngine
var ruleOnce sync.Once

// GetRuleEngine 获取规则引擎单例
func GetRuleEngine() *RuleEngine {
	ruleOnce.Do(func() {
		configPath := ".aitdd/rule.json"
		globalRuleEngine = &RuleEngine{
			configPath: configPath,
		}
		_ = globalRuleEngine.Load()
	})
	return globalRuleEngine
}

// NewRuleEngine 创建新的规则引擎（用于测试或自定义路径）
func NewRuleEngine(configPath string) *RuleEngine {
	e := &RuleEngine{
		configPath: configPath,
	}
	_ = e.Load()
	return e
}

// Load 加载规则配置
func (e *RuleEngine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(e.configPath); os.IsNotExist(err) {
		// 使用默认规则
		e.config = e.getDefaultConfig()
		return nil
	}

	// 读取文件
	data, err := os.ReadFile(e.configPath)
	if err != nil {
		e.config = e.getDefaultConfig()
		return fmt.Errorf("读取规则文件失败: %w", err)
	}

	// 解析JSON
	var config RuleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		e.config = e.getDefaultConfig()
		return fmt.Errorf("解析规则文件失败: %w", err)
	}

	e.config = &config
	return nil
}

// getDefaultConfig 获取默认规则配置
func (e *RuleEngine) getDefaultConfig() *RuleConfig {
	return &RuleConfig{
		Version:     "1.0.0",
		Description: "默认检查规则",
		Rules: map[string]RuleCategory{
			"static": {
				Description: "静态检查规则",
				Rules: []Rule{
					{ID: "S-01", Name: "提示词完整性检查", Description: "检查任务提示词是否包含必要的描述", Severity: "error", Enabled: true},
					{ID: "S-02", Name: "上游契约完整性", Description: "检查上游契约是否正确声明", Severity: "warning", Enabled: true},
					{ID: "S-03", Name: "下游契约完整性", Description: "检查下游契约是否正确声明", Severity: "warning", Enabled: true},
					{ID: "S-04", Name: "测试用例存在性", Description: "检查任务是否定义了测试用例", Severity: "warning", Enabled: true},
					{ID: "S-05", Name: "代码路径声明", Description: "检查任务是否声明了代码路径", Severity: "info", Enabled: true},
					{ID: "S-06", Name: "模块描述完整性", Description: "检查模块是否有描述信息", Severity: "warning", Enabled: true},
					{ID: "S-07", Name: "模块提示词完整性", Description: "检查模块是否有提示词", Severity: "info", Enabled: true},
					{ID: "E-S-09", Name: "模块循环引用错误", Description: "检测跨模块任务引用形成的循环依赖", Severity: "error", Enabled: true},
										{ID: "E-S-10", Name: "孤立任务错误", Description: "检测没有任何上下游依赖关系的孤立任务", Severity: "error", Enabled: true},
										{ID: "E-S-11", Name: "任务循环引用错误", Description: "检测任务间存在的循环引用", Severity: "error", Enabled: true},
										{ID: "E-S-12", Name: "孤立模块错误", Description: "检测没有任何跨模块依赖的孤立模块", Severity: "error", Enabled: true},
										{ID: "E-S-13", Name: "测试用例为空", Description: "检测任务未定义测试用例", Severity: "error", Enabled: true},
				},
			},
			"dynamic": {
				Description: "动态检查规则",
				Rules: []Rule{
					{ID: "D-01", Name: "循环依赖检测", Description: "检测任务间是否存在循环依赖", Severity: "error", Enabled: true},
					{ID: "D-02", Name: "孤立任务检测", Description: "检测没有依赖关系的孤立任务", Severity: "warning", Enabled: true},
					{ID: "D-03", Name: "契约一致性检查", Description: "检查上下游契约是否匹配", Severity: "error", Enabled: true},
					{ID: "D-04", Name: "任务状态一致性", Description: "检查任务状态与依赖关系是否一致", Severity: "warning", Enabled: true},
					{ID: "D-07", Name: "测试失败检测", Description: "检测测试是否失败", Severity: "error", Enabled: true},
					{ID: "D-08", Name: "阻塞任务检测", Description: "检测被阻塞的任务", Severity: "warning", Enabled: true},
				},
			},
		},
	}
}

// GetRules 获取所有规则
func (e *RuleEngine) GetRules() map[string]RuleCategory {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.Rules
}

// GetEnabledRules 获取启用的规则
func (e *RuleEngine) GetEnabledRules(category string) []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var rules []Rule
	if cat, ok := e.config.Rules[category]; ok {
		for _, rule := range cat.Rules {
			if rule.Enabled {
				rules = append(rules, rule)
			}
		}
	}
	return rules
}

// CheckModule 检查模块
func (e *RuleEngine) CheckModule(module map[string]interface{}, tasks []map[string]interface{}) *CheckReport {
	report := &CheckReport{
		ModuleID:     getString(module, "id"),
		ModuleName:   getString(module, "name"),
		ModuleStatus: getString(module, "status"),
		CheckTime:    time.Now().Format(time.RFC3339),
		Summary:      CheckSummary{},
		Errors:       []CheckResult{},
		Warnings:     []CheckResult{},
		Suggestions:  []CheckResult{},
	}

	// 执行静态检查
	e.executeStaticChecks(module, tasks, report)

	// 执行动态检查
	e.executeDynamicChecks(module, tasks, report)

	// 计算摘要
	report.Summary.Errors = len(report.Errors)
	report.Summary.Warnings = len(report.Warnings)
	report.Summary.Suggestions = len(report.Suggestions)
	report.Summary.Total = report.Summary.Errors + report.Summary.Warnings + report.Summary.Suggestions

	return report
}

// executeStaticChecks 执行静态检查
func (e *RuleEngine) executeStaticChecks(module map[string]interface{}, tasks []map[string]interface{}, report *CheckReport) {
	staticRules := e.GetEnabledRules("static")

	for _, rule := range staticRules {
		switch rule.ID {
		case "S-06":
			e.checkModuleDescription(module, rule, report)
		case "S-07":
			e.checkModulePrompt(module, rule, report)
		}
	}

	// 检查任务
	for _, task := range tasks {
		for _, rule := range staticRules {
			switch rule.ID {
			case "S-01":
				e.checkTaskPrompt(task, module, rule, report)
			case "S-02":
				e.checkUpstreamContract(task, module, rule, report)
			case "S-03":
				e.checkDownstreamContract(task, module, rule, report)
			case "S-04":
				e.checkTestCases(task, module, rule, report)
			case "S-05":
				e.checkCodePaths(task, module, rule, report)
			}
		}
	}
}

// executeDynamicChecks 执行动态检查
func (e *RuleEngine) executeDynamicChecks(module map[string]interface{}, tasks []map[string]interface{}, report *CheckReport) {
	dynamicRules := e.GetEnabledRules("dynamic")

	for _, rule := range dynamicRules {
		switch rule.ID {
		case "D-02":
			e.checkOrphanTasks(tasks, module, rule, report)
		case "D-07":
			e.checkTestResult(tasks, module, rule, report)
		case "D-08":
			e.checkBlockedTasks(tasks, module, rule, report)
		}
	}
}

// 静态检查实现

func (e *RuleEngine) checkModuleDescription(module map[string]interface{}, rule Rule, report *CheckReport) {
	desc := getString(module, "description")
	if len(desc) < 20 {
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "module",
			ResourceID:   getString(module, "id"),
			ResourceName: getString(module, "name"),
			Message:      fmt.Sprintf("模块描述长度不足20字符，当前%d字符", len(desc)),
			Suggestion:   "请补充模块描述，至少20字符",
		}
		e.addResult(report, result)
	}
}

func (e *RuleEngine) checkModulePrompt(module map[string]interface{}, rule Rule, report *CheckReport) {
	prompt := getString(module, "prompt")
	if len(prompt) < 50 {
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "module",
			ResourceID:   getString(module, "id"),
			ResourceName: getString(module, "name"),
			Message:      fmt.Sprintf("模块提示词长度不足50字符，当前%d字符", len(prompt)),
			Suggestion:   "请补充模块提示词，至少50字符",
		}
		e.addResult(report, result)
	}
}

func (e *RuleEngine) checkTaskPrompt(task map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	prompt := getString(task, "prompt")
	if len(prompt) < 50 {
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "task",
			ResourceID:   getString(task, "id"),
			ResourceName: getString(task, "name"),
			ModuleID:     getString(module, "id"),
			ModuleName:   getString(module, "name"),
			Message:      fmt.Sprintf("任务提示词长度不足50字符，当前%d字符", len(prompt)),
			Suggestion:   "请补充任务提示词，至少50字符",
		}
		e.addResult(report, result)
	}
}

func (e *RuleEngine) checkUpstreamContract(task map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	contract := getString(task, "upstreamContractDetail")
	if contract == "" || contract == "{}" || contract == `{"title":"","list":[]}` {
		// 只有当任务有上游依赖时才检查
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "task",
			ResourceID:   getString(task, "id"),
			ResourceName: getString(task, "name"),
			ModuleID:     getString(module, "id"),
			ModuleName:   getString(module, "name"),
			Message:      "任务未定义上游契约详情",
			Suggestion:   "如果任务有上游依赖，请定义上游契约详情",
		}
		e.addResult(report, result)
	}
}

func (e *RuleEngine) checkDownstreamContract(task map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	contract := getString(task, "downstreamContractDetail")
	if contract == "" || contract == "{}" || contract == `{"title":"","list":[]}` {
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "task",
			ResourceID:   getString(task, "id"),
			ResourceName: getString(task, "name"),
			ModuleID:     getString(module, "id"),
			ModuleName:   getString(module, "name"),
			Message:      "任务未定义下游契约详情",
			Suggestion:   "请定义下游契约详情，说明为下游任务提供的接口",
		}
		e.addResult(report, result)
	}
}

func (e *RuleEngine) checkTestCases(task map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	tests := getArray(task, "tests")
	if len(tests) == 0 {
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "task",
			ResourceID:   getString(task, "id"),
			ResourceName: getString(task, "name"),
			ModuleID:     getString(module, "id"),
			ModuleName:   getString(module, "name"),
			Message:      "任务未定义测试用例",
			Suggestion:   "为任务添加测试用例",
		}
		e.addResult(report, result)
	}
}

func (e *RuleEngine) checkCodePaths(task map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	codePaths := getString(task, "codePaths")
	if codePaths == "" || codePaths == "[]" {
		result := CheckResult{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			Severity:     rule.Severity,
			ResourceType: "task",
			ResourceID:   getString(task, "id"),
			ResourceName: getString(task, "name"),
			ModuleID:     getString(module, "id"),
			ModuleName:   getString(module, "name"),
			Message:      "任务未声明代码路径",
			Suggestion:   "为任务声明相关的代码文件路径",
		}
		e.addResult(report, result)
	}
}

// 动态检查实现

func (e *RuleEngine) checkOrphanTasks(tasks []map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	// 简化实现：检查没有依赖关系的任务
	for _, task := range tasks {
		// 这里可以添加更复杂的依赖检查逻辑
		_ = task // 占位
	}
}

func (e *RuleEngine) checkTestResult(tasks []map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	for _, task := range tasks {
		testResult := getString(task, "testResult")
		if testResult == "failed" {
			result := CheckResult{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				Severity:     rule.Severity,
				ResourceType: "task",
				ResourceID:   getString(task, "id"),
				ResourceName: getString(task, "name"),
				ModuleID:     getString(module, "id"),
				ModuleName:   getString(module, "name"),
				Message:      "任务测试失败",
				Suggestion:   "检查并修复测试失败的问题",
			}
			e.addResult(report, result)
		}
	}
}

func (e *RuleEngine) checkBlockedTasks(tasks []map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	for _, task := range tasks {
		status := getString(task, "status")
		if status == "blocked" {
			result := CheckResult{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				Severity:     rule.Severity,
				ResourceType: "task",
				ResourceID:   getString(task, "id"),
				ResourceName: getString(task, "name"),
				ModuleID:     getString(module, "id"),
				ModuleName:   getString(module, "name"),
				Message:      "任务处于阻塞状态",
				Suggestion:   "检查阻塞原因并尝试解决",
			}
			e.addResult(report, result)
		}
	}
}

// addResult 添加检查结果到报告
func (e *RuleEngine) addResult(report *CheckReport, result CheckResult) {
	switch result.Severity {
	case "error":
		report.Errors = append(report.Errors, result)
	case "warning":
		report.Warnings = append(report.Warnings, result)
	case "info":
		report.Suggestions = append(report.Suggestions, result)
	}
}

// 辅助函数

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getArray(m map[string]interface{}, key string) []interface{} {
	if val, ok := m[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			return arr
		}
		// 尝试解析JSON字符串
		if str, ok := val.(string); ok && str != "" {
			var arr []interface{}
			if err := json.Unmarshal([]byte(str), &arr); err == nil {
				return arr
			}
		}
	}
	return nil
}

// ToMarkdown 将检查报告转换为Markdown格式
func (report *CheckReport) ToMarkdown() string {
	var md string

	md += fmt.Sprintf("# 模块检查报告\n\n")
	md += fmt.Sprintf("## 基本信息\n")
	md += fmt.Sprintf("- **模块名称**: %s\n", report.ModuleName)
	md += fmt.Sprintf("- **模块ID**: %s\n", report.ModuleID)
	md += fmt.Sprintf("- **状态**: %s\n", report.ModuleStatus)
	md += fmt.Sprintf("- **检查时间**: %s\n\n", report.CheckTime)

	md += fmt.Sprintf("## 检查摘要\n\n")
	md += fmt.Sprintf("| 严重级别 | 数量 |\n")
	md += fmt.Sprintf("|---------|------|\n")
	md += fmt.Sprintf("| 错误 | %d |\n", report.Summary.Errors)
	md += fmt.Sprintf("| 警告 | %d |\n", report.Summary.Warnings)
	md += fmt.Sprintf("| 信息 | %d |\n\n", report.Summary.Suggestions)

	if len(report.Errors) > 0 {
		md += fmt.Sprintf("## 详细结果\n\n")
		md += fmt.Sprintf("### 错误 (%d)\n\n", len(report.Errors))
		for _, err := range report.Errors {
			md += fmt.Sprintf("#### [%s] %s\n", err.RuleID, err.RuleName)
			md += fmt.Sprintf("- **资源**: 任务 - %s\n", err.ResourceName)
			md += fmt.Sprintf("- **消息**: %s\n", err.Message)
			md += fmt.Sprintf("- **建议**: %s\n\n", err.Suggestion)
		}
	}

	if len(report.Warnings) > 0 {
		md += fmt.Sprintf("### 警告 (%d)\n\n", len(report.Warnings))
		for _, warn := range report.Warnings {
			md += fmt.Sprintf("#### [%s] %s\n", warn.RuleID, warn.RuleName)
			md += fmt.Sprintf("- **资源**: 任务 - %s\n", warn.ResourceName)
			md += fmt.Sprintf("- **消息**: %s\n", warn.Message)
			md += fmt.Sprintf("- **建议**: %s\n\n", warn.Suggestion)
		}
	}

	if len(report.Suggestions) > 0 {
		md += fmt.Sprintf("### 信息 (%d)\n\n", len(report.Suggestions))
		for _, info := range report.Suggestions {
			md += fmt.Sprintf("#### [%s] %s\n", info.RuleID, info.RuleName)
			md += fmt.Sprintf("- **资源**: 任务 - %s\n", info.ResourceName)
			md += fmt.Sprintf("- **消息**: %s\n", info.Message)
			md += fmt.Sprintf("- **建议**: %s\n\n", info.Suggestion)
		}
	}

	md += fmt.Sprintf("## 改进建议\n\n")
	if report.Summary.Errors > 0 {
		md += fmt.Sprintf("1. 优先修复 %d 个错误级别的问题\n", report.Summary.Errors)
	}
	if report.Summary.Warnings > 0 {
		md += fmt.Sprintf("2. 处理 %d 个警告级别的问题\n", report.Summary.Warnings)
	}
	if report.Summary.Suggestions > 0 {
		md += fmt.Sprintf("3. 考虑 %d 个信息级别的建议\n", report.Summary.Suggestions)
	}

	return md
}

// GetRule 获取规则配置（返回 map 格式）
func (e *RuleEngine) GetRule() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.config == nil {
		return nil
	}

	// 转换为 map
	data, _ := json.Marshal(e.config)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// SaveRule 保存规则配置（接受 map 参数）
func (e *RuleEngine) SaveRule(rule map[string]interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 转换为 RuleConfig
	data, _ := json.Marshal(rule)
	var config RuleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析规则失败: %w", err)
	}

	e.config = &config

	// 保存到文件
	if e.configPath != "" {
		data, err := json.MarshalIndent(e.config, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化规则失败: %w", err)
		}
		if err := os.WriteFile(e.configPath, data, 0644); err != nil {
			return fmt.Errorf("写入规则文件失败: %w", err)
		}
	}

	return nil
}

// =============================================
// E-S-09: 模块循环引用错误检测
// =============================================

// 规则常量
const (
	RuleModuleCircularReference = "E-S-09" // 模块循环引用错误
	RuleOrphanTask              = "E-S-10" // 孤立任务错误
	RuleTaskCircularReference   = "E-S-11" // 任务循环引用错误
	RuleModuleOrphan            = "E-S-12" // 孤立模块错误
)

// CrossModuleReference 跨模块引用详情
type CrossModuleReference struct {
	SourceTask string `json:"sourceTask"` // 源任务名称或路径
	TargetTask string `json:"targetTask"` // 目标任务名称或路径
}

// CircularReferenceDetail 循环引用详情
type CircularReferenceDetail struct {
	ModuleX        string                 `json:"moduleX"`        // 模块X名称
	ModuleY        string                 `json:"moduleY"`        // 模块Y名称
	XToYReferences []CrossModuleReference `json:"xToYReferences"` // X→Y的引用列表
	YToXReferences []CrossModuleReference `json:"yToXReferences"` // Y→X的引用列表
	Suggestion     string                 `json:"suggestion"`     // 解决建议
}

// ModuleReferenceInfo 模块引用信息（用于内部计算）
type ModuleReferenceInfo struct {
	ModulePathName string                 // 模块路径名
	ModuleName     string                 // 模块名称
	Tasks          []TaskReferenceInfo    // 任务列表
}

// TaskReferenceInfo 任务引用信息
type TaskReferenceInfo struct {
	TaskPathName string   // 任务路径名
	TaskName     string   // 任务名称
	CodePaths    []string // 代码路径列表
}

// CheckModuleCircularReferenceResult 模块循环引用检查结果
type CheckModuleCircularReferenceResult struct {
	HasCircularReference bool                       `json:"hasCircularReference"` // 是否存在循环引用
	CircularPairs        []CircularReferenceDetail `json:"circularPairs"`        // 循环引用对列表
}

// CheckModuleCircularReference 检查模块间的循环引用
// 参数：modulesData - 模块数据列表，每个模块需包含 pathName, name, tasks 字段
// 返回：检测到的循环引用错误列表
func (e *RuleEngine) CheckModuleCircularReference(modulesData []map[string]interface{}) *CheckModuleCircularReferenceResult {
	result := &CheckModuleCircularReferenceResult{
		HasCircularReference: false,
		CircularPairs:        []CircularReferenceDetail{},
	}

	if len(modulesData) < 2 {
		// 少于两个模块不可能存在循环引用
		return result
	}

	// 1. 构建模块信息列表
	moduleInfos := e.buildModuleReferenceInfos(modulesData)

	// 2. 构建模块间引用关系图
	// key: "模块X路径/模块Y路径", value: 引用列表
	referenceGraph := e.buildModuleReferenceGraph(moduleInfos)

	// 3. 检测循环引用
	e.detectCircularReferences(moduleInfos, referenceGraph, result)

	return result
}

// buildModuleReferenceInfos 构建模块引用信息列表
func (e *RuleEngine) buildModuleReferenceInfos(modulesData []map[string]interface{}) []ModuleReferenceInfo {
	moduleInfos := make([]ModuleReferenceInfo, 0, len(modulesData))

	for _, moduleData := range modulesData {
		modulePathName := getString(moduleData, "pathName")
		moduleName := getString(moduleData, "name")

		moduleInfo := ModuleReferenceInfo{
			ModulePathName: modulePathName,
			ModuleName:     moduleName,
			Tasks:          []TaskReferenceInfo{},
		}

		// 提取任务信息
		tasks := getArray(moduleData, "tasks")
		for _, taskInterface := range tasks {
			if task, ok := taskInterface.(map[string]interface{}); ok {
				taskInfo := TaskReferenceInfo{
					TaskPathName: getString(task, "pathName"),
					TaskName:     getString(task, "name"),
					CodePaths:    parseCodePaths(task),
				}
				moduleInfo.Tasks = append(moduleInfo.Tasks, taskInfo)
			}
		}

		moduleInfos = append(moduleInfos, moduleInfo)
	}

	return moduleInfos
}

// buildModuleReferenceGraph 构建模块间引用关系图
// 返回: map[key]value 其中 key 为 "源模块路径->目标模块路径"，value 为引用详情列表
func (e *RuleEngine) buildModuleReferenceGraph(moduleInfos []ModuleReferenceInfo) map[string][]CrossModuleReference {
	referenceGraph := make(map[string][]CrossModuleReference)

	// 构建模块路径到模块信息的映射
	modulePathMap := make(map[string]ModuleReferenceInfo)
	for _, moduleInfo := range moduleInfos {
		modulePathMap[moduleInfo.ModulePathName] = moduleInfo
	}

	// 遍历所有模块的所有任务，分析跨模块引用
	for _, sourceModule := range moduleInfos {
		for _, task := range sourceModule.Tasks {
			for _, codePath := range task.CodePaths {
				// 检查 codePath 是否指向其他模块的任务
				targetModulePath := e.findTargetModulePath(codePath, sourceModule.ModulePathName, modulePathMap)
				if targetModulePath != "" && targetModulePath != sourceModule.ModulePathName {
					// 发现跨模块引用
					key := sourceModule.ModulePathName + "->" + targetModulePath
					referenceGraph[key] = append(referenceGraph[key], CrossModuleReference{
						SourceTask: task.TaskPathName,
						TargetTask: codePath,
					})
				}
			}
		}
	}

	return referenceGraph
}

// findTargetModulePath 查找 codePath 指向的目标模块路径
func (e *RuleEngine) findTargetModulePath(codePath string, sourceModulePath string, modulePathMap map[string]ModuleReferenceInfo) string {
	// codePath 可能是任务路径，格式为 "项目/模块/任务"
	// 需要提取其中的模块路径

	// 遍历所有模块，检查 codePath 是否以某个模块路径开头
	for targetModulePath := range modulePathMap {
		if targetModulePath != sourceModulePath {
			// 检查 codePath 是否属于目标模块
			// 可能的情况：
			// 1. codePath 是完整任务路径，包含模块路径
			// 2. codePath 是相对路径或任务名称
			if e.isCodePathBelongsToModule(codePath, targetModulePath, modulePathMap[targetModulePath]) {
				return targetModulePath
			}
		}
	}

	return ""
}

// isCodePathBelongsToModule 检查 codePath 是否属于指定模块
func (e *RuleEngine) isCodePathBelongsToModule(codePath string, modulePath string, moduleInfo ModuleReferenceInfo) bool {
	// 1. 检查 codePath 是否以模块路径开头（完整路径）
	if len(codePath) > len(modulePath) && codePath[:len(modulePath)] == modulePath {
		return true
	}

	// 2. 检查 codePath 是否匹配模块中的某个任务名称
	for _, task := range moduleInfo.Tasks {
		if task.TaskName == codePath || task.TaskPathName == codePath {
			return true
		}
	}

	return false
}

// detectCircularReferences 检测循环引用
func (e *RuleEngine) detectCircularReferences(moduleInfos []ModuleReferenceInfo, referenceGraph map[string][]CrossModuleReference, result *CheckModuleCircularReferenceResult) {
	// 检查所有模块对之间是否存在双向引用
	for i := 0; i < len(moduleInfos); i++ {
		for j := i + 1; j < len(moduleInfos); j++ {
			moduleX := moduleInfos[i]
			moduleY := moduleInfos[j]

			keyXToY := moduleX.ModulePathName + "->" + moduleY.ModulePathName
			keyYToX := moduleY.ModulePathName + "->" + moduleX.ModulePathName

			xToYRefs, hasXToY := referenceGraph[keyXToY]
			yToXRefs, hasYToX := referenceGraph[keyYToX]

			// 如果同时存在 X→Y 和 Y→X 的引用，则存在循环引用
			if hasXToY && hasYToX {
				result.HasCircularReference = true
				result.CircularPairs = append(result.CircularPairs, CircularReferenceDetail{
					ModuleX:        moduleX.ModuleName,
					ModuleY:        moduleY.ModuleName,
					XToYReferences: xToYRefs,
					YToXReferences: yToXRefs,
					Suggestion:     "建议引入第三方共享模块来打破循环依赖",
				})
			}
		}
	}
}

// parseCodePaths 解析任务的 codePaths 字段
func parseCodePaths(task map[string]interface{}) []string {
	codePaths := getArray(task, "codePaths")
	if len(codePaths) == 0 {
		// 尝试从字符串解析
		codePathsStr := getString(task, "codePaths")
		if codePathsStr != "" && codePathsStr != "[]" && codePathsStr != "null" {
			var paths []string
			if err := json.Unmarshal([]byte(codePathsStr), &paths); err == nil {
				return paths
			}
		}
	}

	// 转换为字符串数组
	result := make([]string, 0, len(codePaths))
	for _, path := range codePaths {
		if str, ok := path.(string); ok && str != "" {
			result = append(result, str)
		}
	}
	return result
}

// GenerateCircularReferenceReport 生成循环引用检查报告（Markdown格式）
func (result *CheckModuleCircularReferenceResult) GenerateCircularReferenceReport() string {
	if !result.HasCircularReference {
		return "## 模块循环引用检查\n\n✅ 未检测到模块间的循环引用\n"
	}

	var report string
	report += "## 模块循环引用检查\n\n"
	report += fmt.Sprintf("❌ 检测到 %d 对模块存在循环引用\n\n", len(result.CircularPairs))

	for i, pair := range result.CircularPairs {
		report += fmt.Sprintf("### 循环引用 #%d\n\n", i+1)
		report += fmt.Sprintf("**模块**: \"%s\" ↔ \"%s\"\n\n", pair.ModuleX, pair.ModuleY)

		if len(pair.XToYReferences) > 0 {
			report += fmt.Sprintf("**%s → %s 引用**:\n", pair.ModuleX, pair.ModuleY)
			for _, ref := range pair.XToYReferences {
				report += fmt.Sprintf("- 任务 \"%s\" → 任务 \"%s\"\n", ref.SourceTask, ref.TargetTask)
			}
			report += "\n"
		}

		if len(pair.YToXReferences) > 0 {
			report += fmt.Sprintf("**%s → %s 引用**:\n", pair.ModuleY, pair.ModuleX)
			for _, ref := range pair.YToXReferences {
				report += fmt.Sprintf("- 任务 \"%s\" → 任务 \"%s\"\n", ref.SourceTask, ref.TargetTask)
			}
			report += "\n"
		}

		report += fmt.Sprintf("**建议**: %s\n\n", pair.Suggestion)
	}

	return report
}
