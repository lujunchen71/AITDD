package main

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
	Version     string            `json:"version"`
	Description string            `json:"description"`
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
	ModuleID    string        `json:"moduleId"`
	ModuleName  string        `json:"moduleName"`
	ModuleStatus string       `json:"moduleStatus"`
	CheckTime   string        `json:"checkTime"`
	Summary     CheckSummary  `json:"summary"`
	Errors      []CheckResult `json:"errors"`
	Warnings    []CheckResult `json:"warnings"`
	Suggestions []CheckResult `json:"suggestions"`
}

// CheckSummary 检查摘要
type CheckSummary struct {
	Total      int `json:"total"`
	Errors     int `json:"errorCount"`
	Warnings   int `json:"warningCount"`
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
				},
			},
			"dynamic": {
				Description: "动态检查规则",
				Rules: []Rule{
					{ID: "D-01", Name: "循环依赖检测", Description: "检测任务间是否存在循环依赖", Severity: "error", Enabled: true},
					{ID: "D-02", Name: "孤立任务检测", Description: "检测没有依赖关系的孤立任务", Severity: "warning", Enabled: true},
					{ID: "D-03", Name: "契约一致性检查", Description: "检查上下游契约是否匹配", Severity: "error", Enabled: true},
					{ID: "D-04", Name: "任务状态一致性", Description: "检查任务状态与依赖关系是否一致", Severity: "warning", Enabled: true},
					{ID: "D-05", Name: "Bug日志检测", Description: "检测是否有未解决的Bug", Severity: "warning", Enabled: true},
					{ID: "D-06", Name: "Issue详情检测", Description: "检测是否有未处理的Issue", Severity: "warning", Enabled: true},
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
		ModuleID:    getString(module, "id"),
		ModuleName:  getString(module, "name"),
		ModuleStatus: getString(module, "status"),
		CheckTime:   time.Now().Format(time.RFC3339),
		Summary:     CheckSummary{},
		Errors:      []CheckResult{},
		Warnings:    []CheckResult{},
		Suggestions: []CheckResult{},
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
		case "D-05":
			e.checkBugLog(tasks, module, rule, report)
		case "D-06":
			e.checkIssueDetails(tasks, module, rule, report)
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

func (e *RuleEngine) checkBugLog(tasks []map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	for _, task := range tasks {
		bugLog := getString(task, "bugLog")
		if bugLog != "" && bugLog != "[]" {
			result := CheckResult{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				Severity:     rule.Severity,
				ResourceType: "task",
				ResourceID:   getString(task, "id"),
				ResourceName: getString(task, "name"),
				ModuleID:     getString(module, "id"),
				ModuleName:   getString(module, "name"),
				Message:      "任务存在未解决的Bug记录",
				Suggestion:   "检查并解决Bug记录中的问题",
			}
			e.addResult(report, result)
		}
	}
}

func (e *RuleEngine) checkIssueDetails(tasks []map[string]interface{}, module map[string]interface{}, rule Rule, report *CheckReport) {
	for _, task := range tasks {
		issueDetails := getString(task, "issueDetails")
		if issueDetails != "" {
			result := CheckResult{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				Severity:     rule.Severity,
				ResourceType: "task",
				ResourceID:   getString(task, "id"),
				ResourceName: getString(task, "name"),
				ModuleID:     getString(module, "id"),
				ModuleName:   getString(module, "name"),
				Message:      "任务存在未处理的Issue详情",
				Suggestion:   "检查并处理Issue详情中的问题",
			}
			e.addResult(report, result)
		}
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
