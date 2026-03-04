package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// AnalysisGenerator 分析生成器
// 用于生成AI分析请求和解析AI响应
type AnalysisGenerator struct {
	qaConfig *QAConfig
	logger   *MCPLogger
}

// AnalysisRequest 分析请求
// 包含发送给AI的完整分析请求信息
type AnalysisRequest struct {
	AnalysisType string                 `json:"analysis_type"` // task, module, project
	Data         interface{}            `json:"data"`
	Questions    []Question             `json:"questions"`
	SystemPrompt string                 `json:"system_prompt"`
	UserPrompt   string                 `json:"user_prompt"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisResult 分析结果
// 包含AI分析的完整结果
type AnalysisResult struct {
	AnalysisID    string    `json:"analysis_id"`
	AnalysisType  string    `json:"analysis_type"`
	TargetID      uint      `json:"target_id"`
	TargetName    string    `json:"target_name"`
	Issues        []Issue   `json:"issues"`
	Suggestions   []string  `json:"suggestions"`
	Score         int       `json:"score"` // 0-100
	Confidence    float64   `json:"confidence"` // 0.0-1.0
	AnalyzedAt    time.Time `json:"analyzed_at"`
	RawResponse   string    `json:"raw_response"`
	Summary       string    `json:"summary"`
}

// Issue 发现的问题
type Issue struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Severity    string `json:"severity"` // critical, warning, info
	Description string `json:"description"`
	Location    string `json:"location"`
	Suggestion  string `json:"suggestion"`
	QuestionID  string `json:"question_id,omitempty"`
}

// AnalysisOptions 分析选项
type AnalysisOptions struct {
	IncludeRawResponse bool                   `json:"include_raw_response"`
	CustomQuestions    []Question             `json:"custom_questions,omitempty"`
	AdditionalContext  map[string]interface{} `json:"additional_context,omitempty"`
	MaxIssues          int                    `json:"max_issues"`
	MinSeverity        string                 `json:"min_severity"` // 只返回此严重级别及以上的问题
}

// DefaultAnalysisOptions 默认分析选项
func DefaultAnalysisOptions() *AnalysisOptions {
	return &AnalysisOptions{
		IncludeRawResponse: true,
		MaxIssues:          50,
		MinSeverity:        SeverityInfo,
	}
}

// NewAnalysisGenerator 创建新的分析生成器
func NewAnalysisGenerator(qaConfig *QAConfig) *AnalysisGenerator {
	if qaConfig == nil {
		qaConfig = GetDefaultQAConfig()
	}
	return &AnalysisGenerator{
		qaConfig: qaConfig,
		logger:   GetMCPLogger(),
	}
}

// NewAnalysisGeneratorWithLogger 创建带自定义日志的分析生成器
func NewAnalysisGeneratorWithLogger(qaConfig *QAConfig, logger *MCPLogger) *AnalysisGenerator {
	if qaConfig == nil {
		qaConfig = GetDefaultQAConfig()
	}
	return &AnalysisGenerator{
		qaConfig: qaConfig,
		logger:   logger,
	}
}

// GenerateTaskAnalysis 生成任务分析请求
// 根据任务数据生成完整的AI分析请求
func (g *AnalysisGenerator) GenerateTaskAnalysis(ctx context.Context, data *TaskAnalysisData, opts *AnalysisOptions) (*AnalysisRequest, error) {
	if data == nil {
		return nil, NewParamError("任务数据不能为空")
	}

	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	g.logger.Debug("生成任务分析请求", map[string]interface{}{
		"task_id":   data.TaskID,
		"task_name": data.TaskName,
	})

	// 获取任务相关问题
	questions := g.qaConfig.GetTaskQuestions()
	if len(opts.CustomQuestions) > 0 {
		questions = append(questions, opts.CustomQuestions...)
	}

	// 构建分析请求
	request := &AnalysisRequest{
		AnalysisType: QuestionTypeTask,
		Data:         data,
		Questions:    questions,
		SystemPrompt: g.BuildSystemPrompt(QuestionTypeTask),
		UserPrompt:   g.buildTaskUserPrompt(data, questions, opts),
		Metadata: map[string]interface{}{
			"task_id":   data.TaskID,
			"task_name": data.TaskName,
			"timestamp": time.Now().UnixMilli(),
		},
	}

	return request, nil
}

// GenerateModuleAnalysis 生成模块分析请求
// 根据模块数据生成完整的AI分析请求
func (g *AnalysisGenerator) GenerateModuleAnalysis(ctx context.Context, data *ModuleAnalysisData, opts *AnalysisOptions) (*AnalysisRequest, error) {
	if data == nil {
		return nil, NewParamError("模块数据不能为空")
	}

	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	g.logger.Debug("生成模块分析请求", map[string]interface{}{
		"module_id":   data.ModuleID,
		"module_name": data.ModuleName,
	})

	// 获取模块相关问题
	questions := g.qaConfig.GetModuleQuestions()
	if len(opts.CustomQuestions) > 0 {
		questions = append(questions, opts.CustomQuestions...)
	}

	// 构建分析请求
	request := &AnalysisRequest{
		AnalysisType: QuestionTypeModule,
		Data:         data,
		Questions:    questions,
		SystemPrompt: g.BuildSystemPrompt(QuestionTypeModule),
		UserPrompt:   g.buildModuleUserPrompt(data, questions, opts),
		Metadata: map[string]interface{}{
			"module_id":   data.ModuleID,
			"module_name": data.ModuleName,
			"timestamp":   time.Now().UnixMilli(),
		},
	}

	return request, nil
}

// GenerateProjectAnalysis 生成项目分析请求
// 根据项目数据生成完整的AI分析请求
func (g *AnalysisGenerator) GenerateProjectAnalysis(ctx context.Context, data *ProjectAnalysisData, opts *AnalysisOptions) (*AnalysisRequest, error) {
	if data == nil {
		return nil, NewParamError("项目数据不能为空")
	}

	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	g.logger.Debug("生成项目分析请求", map[string]interface{}{
		"project_id":   data.ProjectID,
		"project_name": data.ProjectName,
	})

	// 获取项目相关问题
	questions := g.qaConfig.GetProjectQuestions()
	if len(opts.CustomQuestions) > 0 {
		questions = append(questions, opts.CustomQuestions...)
	}

	// 构建分析请求
	request := &AnalysisRequest{
		AnalysisType: QuestionTypeProject,
		Data:         data,
		Questions:    questions,
		SystemPrompt: g.BuildSystemPrompt(QuestionTypeProject),
		UserPrompt:   g.buildProjectUserPrompt(data, questions, opts),
		Metadata: map[string]interface{}{
			"project_id":   data.ProjectID,
			"project_name": data.ProjectName,
			"timestamp":    time.Now().UnixMilli(),
		},
	}

	return request, nil
}

// BuildSystemPrompt 构建系统提示词
// 根据分析类型生成相应的系统提示词
func (g *AnalysisGenerator) BuildSystemPrompt(analysisType string) string {
	var sb strings.Builder

	sb.WriteString("你是一个专业的软件工程分析助手，负责分析")
	switch analysisType {
	case QuestionTypeTask:
		sb.WriteString("任务（Task）")
	case QuestionTypeModule:
		sb.WriteString("模块（Module）")
	case QuestionTypeProject:
		sb.WriteString("项目（Project）")
	default:
		sb.WriteString("软件工程")
	}
	sb.WriteString("的质量和完整性。\n\n")

	sb.WriteString("## 你的职责\n")
	sb.WriteString("1. 仔细审查提供的数据和信息\n")
	sb.WriteString("2. 根据给定的问题清单进行系统分析\n")
	sb.WriteString("3. 识别潜在的问题和风险\n")
	sb.WriteString("4. 提供具体、可操作的改进建议\n")
	sb.WriteString("5. 给出整体质量评分（0-100分）\n\n")

	sb.WriteString("## 输出格式\n")
	sb.WriteString("请严格按照以下JSON格式输出分析结果：\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"分析摘要（一句话概述主要发现）\",\n")
	sb.WriteString("  \"score\": 85,\n")
	sb.WriteString("  \"confidence\": 0.9,\n")
	sb.WriteString("  \"issues\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"id\": \"issue_001\",\n")
	sb.WriteString("      \"category\": \"completeness\",\n")
	sb.WriteString("      \"severity\": \"critical|warning|info\",\n")
	sb.WriteString("      \"description\": \"问题描述\",\n")
	sb.WriteString("      \"location\": \"问题位置\",\n")
	sb.WriteString("      \"suggestion\": \"改进建议\",\n")
	sb.WriteString("      \"question_id\": \"对应的问题ID\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"suggestions\": [\n")
	sb.WriteString("    \"整体改进建议1\",\n")
	sb.WriteString("    \"整体改进建议2\"\n")
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## 严重级别说明\n")
	sb.WriteString("- **critical**: 严重问题，必须立即解决\n")
	sb.WriteString("- **warning**: 警告问题，建议尽快解决\n")
	sb.WriteString("- **info**: 信息提示，可以考虑改进\n\n")

	sb.WriteString("## 注意事项\n")
	sb.WriteString("1. 保持客观、专业的分析态度\n")
	sb.WriteString("2. 问题要有具体依据，不要臆测\n")
	sb.WriteString("3. 建议要具体可行，避免空泛\n")
	sb.WriteString("4. 评分要综合考虑所有因素\n")
	sb.WriteString("5. 只输出JSON，不要包含其他文字\n")

	return sb.String()
}

// buildTaskUserPrompt 构建任务分析的用户提示词
func (g *AnalysisGenerator) buildTaskUserPrompt(data *TaskAnalysisData, questions []Question, opts *AnalysisOptions) string {
	var sb strings.Builder

	sb.WriteString("## 任务信息\n\n")
	sb.WriteString(fmt.Sprintf("**任务ID**: %d\n", data.TaskID))
	sb.WriteString(fmt.Sprintf("**任务名称**: %s\n", data.TaskName))
	sb.WriteString(fmt.Sprintf("**状态**: %s\n", data.Status))

	if data.ModuleName != "" {
		sb.WriteString(fmt.Sprintf("**所属模块**: %s\n", data.ModuleName))
	}
	if data.ProjectName != "" {
		sb.WriteString(fmt.Sprintf("**所属项目**: %s\n", data.ProjectName))
	}

	sb.WriteString("\n### 任务描述\n")
	if data.TaskDescription != "" {
		sb.WriteString(data.TaskDescription)
	} else {
		sb.WriteString("*无描述*\n")
	}

	sb.WriteString("\n### 验收标准\n")
	if data.AcceptanceCriteria != "" {
		sb.WriteString(data.AcceptanceCriteria)
	} else {
		sb.WriteString("*未定义*\n")
	}

	sb.WriteString("\n### 测试要求\n")
	if data.TestRequirements != "" {
		sb.WriteString(data.TestRequirements)
	} else {
		sb.WriteString("*未定义*\n")
	}

	if len(data.CodePaths) > 0 {
		sb.WriteString("\n### 代码路径\n")
		for _, path := range data.CodePaths {
			sb.WriteString(fmt.Sprintf("- %s\n", path))
		}
	}

	if len(data.Dependencies) > 0 {
		sb.WriteString("\n### 依赖关系\n")
		for _, dep := range data.Dependencies {
			if dep.UpstreamTaskName != "" {
				sb.WriteString(fmt.Sprintf("- 上游: %s\n", dep.UpstreamTaskName))
			}
			if dep.DownstreamTaskName != "" {
				sb.WriteString(fmt.Sprintf("- 下游: %s\n", dep.DownstreamTaskName))
			}
		}
	}

	sb.WriteString("\n## 分析问题\n\n")
	sb.WriteString("请针对以下问题进行分析：\n\n")
	for i, q := range questions {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, q.Severity, q.Question))
	}

	if opts.AdditionalContext != nil {
		sb.WriteString("\n## 附加信息\n\n")
		for k, v := range opts.AdditionalContext {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}

	return sb.String()
}

// buildModuleUserPrompt 构建模块分析的用户提示词
func (g *AnalysisGenerator) buildModuleUserPrompt(data *ModuleAnalysisData, questions []Question, opts *AnalysisOptions) string {
	var sb strings.Builder

	sb.WriteString("## 模块信息\n\n")
	sb.WriteString(fmt.Sprintf("**模块ID**: %d\n", data.ModuleID))
	sb.WriteString(fmt.Sprintf("**模块名称**: %s\n", data.ModuleName))
	sb.WriteString(fmt.Sprintf("**状态**: %s\n", data.Status))

	if data.ProjectName != "" {
		sb.WriteString(fmt.Sprintf("**所属项目**: %s\n", data.ProjectName))
	}

	sb.WriteString("\n### 模块描述\n")
	if data.ModuleDescription != "" {
		sb.WriteString(data.ModuleDescription)
	} else {
		sb.WriteString("*无描述*\n")
	}

	if len(data.Tasks) > 0 {
		sb.WriteString("\n### 包含的任务\n")
		for _, task := range data.Tasks {
			sb.WriteString(fmt.Sprintf("- %s [%s]\n", task.TaskName, task.Status))
		}
		sb.WriteString(fmt.Sprintf("\n**任务总数**: %d\n", len(data.Tasks)))
	}

	if len(data.Dependencies) > 0 {
		sb.WriteString("\n### 依赖关系\n")
		for _, dep := range data.Dependencies {
			sb.WriteString(fmt.Sprintf("- 类型: %s\n", dep.Type))
		}
	}

	sb.WriteString("\n## 分析问题\n\n")
	sb.WriteString("请针对以下问题进行分析：\n\n")
	for i, q := range questions {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, q.Severity, q.Question))
	}

	if opts.AdditionalContext != nil {
		sb.WriteString("\n## 附加信息\n\n")
		for k, v := range opts.AdditionalContext {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}

	return sb.String()
}

// buildProjectUserPrompt 构建项目分析的用户提示词
func (g *AnalysisGenerator) buildProjectUserPrompt(data *ProjectAnalysisData, questions []Question, opts *AnalysisOptions) string {
	var sb strings.Builder

	sb.WriteString("## 项目信息\n\n")
	sb.WriteString(fmt.Sprintf("**项目ID**: %d\n", data.ProjectID))
	sb.WriteString(fmt.Sprintf("**项目名称**: %s\n", data.ProjectName))

	sb.WriteString("\n### 项目描述/宪法\n")
	if data.Constitution != "" {
		sb.WriteString(data.Constitution)
	} else {
		sb.WriteString("*无描述*\n")
	}

	sb.WriteString(fmt.Sprintf("\n**模块总数**: %d\n", data.TotalModules))
	sb.WriteString(fmt.Sprintf("**任务总数**: %d\n", data.TotalTasks))

	if len(data.Modules) > 0 {
		sb.WriteString("\n### 模块概览\n")
		for _, module := range data.Modules {
			sb.WriteString(fmt.Sprintf("- %s [%s] - %d 个任务\n", module.ModuleName, module.Status, module.TaskCount))
		}
	}

	if len(data.Tasks) > 0 {
		sb.WriteString("\n### 任务状态分布\n")
		statusCount := make(map[string]int)
		for _, task := range data.Tasks {
			statusCount[task.Status]++
		}
		for status, count := range statusCount {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", status, count))
		}
	}

	sb.WriteString("\n## 分析问题\n\n")
	sb.WriteString("请针对以下问题进行分析：\n\n")
	for i, q := range questions {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, q.Severity, q.Question))
	}

	if opts.AdditionalContext != nil {
		sb.WriteString("\n## 附加信息\n\n")
		for k, v := range opts.AdditionalContext {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}

	return sb.String()
}

// ParseAIResponse 解析AI响应为结构化结果
// 将AI返回的JSON字符串解析为AnalysisResult结构
func (g *AnalysisGenerator) ParseAIResponse(response string, analysisType string, targetID uint, targetName string) (*AnalysisResult, error) {
	g.logger.Debug("解析AI响应", map[string]interface{}{
		"analysis_type": analysisType,
		"target_id":     targetID,
		"response_len":  len(response),
	})

	// 提取JSON部分
	jsonStr := g.extractJSON(response)
	if jsonStr == "" {
		return nil, NewBusinessError("无法从AI响应中提取JSON")
	}

	// 解析JSON
	var parsed struct {
		Summary     string   `json:"summary"`
		Score       int      `json:"score"`
		Confidence  float64  `json:"confidence"`
		Issues      []Issue  `json:"issues"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		g.logger.Error("解析AI响应JSON失败", err, map[string]interface{}{
			"json_str": jsonStr,
		})
		return nil, NewInternalError("解析AI响应失败", err)
	}

	// 验证和修正分数
	if parsed.Score < 0 {
		parsed.Score = 0
	} else if parsed.Score > 100 {
		parsed.Score = 100
	}

	// 验证和修正置信度
	if parsed.Confidence < 0 {
		parsed.Confidence = 0
	} else if parsed.Confidence > 1 {
		parsed.Confidence = 1
	}
	if parsed.Confidence == 0 {
		parsed.Confidence = 0.8 // 默认置信度
	}

	// 生成分析ID
	analysisID := fmt.Sprintf("%s_%d_%d", analysisType, targetID, time.Now().UnixMilli())

	// 构建结果
	result := &AnalysisResult{
		AnalysisID:   analysisID,
		AnalysisType: analysisType,
		TargetID:     targetID,
		TargetName:   targetName,
		Summary:      parsed.Summary,
		Score:        parsed.Score,
		Confidence:   parsed.Confidence,
		Issues:       g.validateIssues(parsed.Issues),
		Suggestions:  parsed.Suggestions,
		AnalyzedAt:   time.Now(),
		RawResponse:  response,
	}

	g.logger.Info("AI响应解析完成", map[string]interface{}{
		"analysis_id":    result.AnalysisID,
		"score":          result.Score,
		"issues_count":   len(result.Issues),
		"suggestions":    len(result.Suggestions),
	})

	return result, nil
}

// extractJSON 从响应中提取JSON部分
func (g *AnalysisGenerator) extractJSON(response string) string {
	// 尝试提取 ```json ... ``` 块
	jsonBlockRegex := regexp.MustCompile("(?s)```json\\s*({.*?})\\s*```")
	if matches := jsonBlockRegex.FindStringSubmatch(response); len(matches) > 1 {
		return matches[1]
	}

	// 尝试提取 ``` ... ``` 块
	codeBlockRegex := regexp.MustCompile("(?s)```\\s*({.*?})\\s*```")
	if matches := codeBlockRegex.FindStringSubmatch(response); len(matches) > 1 {
		return matches[1]
	}

	// 尝试直接匹配JSON对象
	jsonRegex := regexp.MustCompile("(?s){.*}")
	if matches := jsonRegex.FindString(response); matches != "" {
		return matches
	}

	return ""
}

// validateIssues 验证和修正问题列表
func (g *AnalysisGenerator) validateIssues(issues []Issue) []Issue {
	validIssues := make([]Issue, 0, len(issues))

	for _, issue := range issues {
		// 验证严重级别
		if issue.Severity != SeverityCritical && issue.Severity != SeverityWarning && issue.Severity != SeverityInfo {
			issue.Severity = SeverityInfo // 默认为信息级别
		}

		// 确保有ID
		if issue.ID == "" {
			issue.ID = fmt.Sprintf("issue_%d", time.Now().UnixNano())
		}

		// 确保有描述
		if issue.Description == "" {
			continue // 跳过没有描述的问题
		}

		validIssues = append(validIssues, issue)
	}

	return validIssues
}

// GenerateAnalysisID 生成分析ID
func GenerateAnalysisID(analysisType string, targetID uint) string {
	return fmt.Sprintf("%s_%d_%d", analysisType, targetID, time.Now().UnixMilli())
}

// MergeAnalysisResults 合并多个分析结果
func MergeAnalysisResults(results []*AnalysisResult) *AnalysisResult {
	if len(results) == 0 {
		return nil
	}

	if len(results) == 1 {
		return results[0]
	}

	merged := &AnalysisResult{
		AnalysisID:   fmt.Sprintf("merged_%d", time.Now().UnixMilli()),
		AnalysisType: "merged",
		Issues:       make([]Issue, 0),
		Suggestions:  make([]string, 0),
		AnalyzedAt:   time.Now(),
	}

	totalScore := 0
	totalConfidence := 0.0
	issueMap := make(map[string]bool)

	for _, r := range results {
		totalScore += r.Score
		totalConfidence += r.Confidence

		// 合并问题（去重）
		for _, issue := range r.Issues {
			key := issue.Category + "_" + issue.Description
			if !issueMap[key] {
				issueMap[key] = true
				merged.Issues = append(merged.Issues, issue)
			}
		}

		// 合并建议
		merged.Suggestions = append(merged.Suggestions, r.Suggestions...)
	}

	// 计算平均分数和置信度
	merged.Score = totalScore / len(results)
	merged.Confidence = totalConfidence / float64(len(results))

	// 生成摘要
	criticalCount := 0
	warningCount := 0
	for _, issue := range merged.Issues {
		switch issue.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityWarning:
			warningCount++
		}
	}
	merged.Summary = fmt.Sprintf("发现 %d 个严重问题，%d 个警告，整体评分 %d 分", criticalCount, warningCount, merged.Score)

	return merged
}

// FilterIssuesBySeverity 按严重级别过滤问题
func FilterIssuesBySeverity(issues []Issue, minSeverity string) []Issue {
	severityOrder := map[string]int{
		SeverityCritical: 3,
		SeverityWarning:  2,
		SeverityInfo:     1,
	}

	minLevel := severityOrder[minSeverity]
	filtered := make([]Issue, 0)

	for _, issue := range issues {
		if severityOrder[issue.Severity] >= minLevel {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

// GetAnalysisSummary 获取分析结果摘要
func GetAnalysisSummary(result *AnalysisResult) string {
	if result == nil {
		return "无分析结果"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 分析结果摘要\n\n"))
	sb.WriteString(fmt.Sprintf("**分析对象**: %s (%s)\n", result.TargetName, result.AnalysisType))
	sb.WriteString(fmt.Sprintf("**评分**: %d/100\n", result.Score))
	sb.WriteString(fmt.Sprintf("**置信度**: %.0f%%\n", result.Confidence*100))
	sb.WriteString(fmt.Sprintf("**分析时间**: %s\n\n", result.AnalyzedAt.Format("2006-01-02 15:04:05")))

	// 统计问题
	criticalCount := 0
	warningCount := 0
	infoCount := 0
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityWarning:
			warningCount++
		case SeverityInfo:
			infoCount++
		}
	}

	sb.WriteString("### 问题统计\n")
	sb.WriteString(fmt.Sprintf("- 严重: %d\n", criticalCount))
	sb.WriteString(fmt.Sprintf("- 警告: %d\n", warningCount))
	sb.WriteString(fmt.Sprintf("- 信息: %d\n\n", infoCount))

	if result.Summary != "" {
		sb.WriteString("### 摘要\n")
		sb.WriteString(result.Summary)
		sb.WriteString("\n\n")
	}

	if len(result.Suggestions) > 0 {
		sb.WriteString("### 改进建议\n")
		for i, s := range result.Suggestions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s))
		}
	}

	return sb.String()
}

// ToJSON 将分析结果转换为JSON字符串
func (r *AnalysisResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML 将分析结果转换为YAML格式字符串
func (r *AnalysisResult) ToYAML() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("analysis_id: %s\n", r.AnalysisID))
	sb.WriteString(fmt.Sprintf("analysis_type: %s\n", r.AnalysisType))
	sb.WriteString(fmt.Sprintf("target_id: %d\n", r.TargetID))
	sb.WriteString(fmt.Sprintf("target_name: %s\n", r.TargetName))
	sb.WriteString(fmt.Sprintf("score: %d\n", r.Score))
	sb.WriteString(fmt.Sprintf("confidence: %.2f\n", r.Confidence))
	sb.WriteString(fmt.Sprintf("analyzed_at: %s\n", r.AnalyzedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("summary: %s\n", r.Summary))

	if len(r.Issues) > 0 {
		sb.WriteString("issues:\n")
		for _, issue := range r.Issues {
			sb.WriteString(fmt.Sprintf("  - id: %s\n", issue.ID))
			sb.WriteString(fmt.Sprintf("    category: %s\n", issue.Category))
			sb.WriteString(fmt.Sprintf("    severity: %s\n", issue.Severity))
			sb.WriteString(fmt.Sprintf("    description: %s\n", issue.Description))
			if issue.Location != "" {
				sb.WriteString(fmt.Sprintf("    location: %s\n", issue.Location))
			}
			if issue.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("    suggestion: %s\n", issue.Suggestion))
			}
		}
	}

	if len(r.Suggestions) > 0 {
		sb.WriteString("suggestions:\n")
		for _, s := range r.Suggestions {
			sb.WriteString(fmt.Sprintf("  - %s\n", s))
		}
	}

	return sb.String()
}

// BuildTaskUserPromptFromMap 直接从任务 map 数据构建任务分析用户提示词
// 不依赖数字 ID，直接使用任务的 map 数据，适用于 compile_dynamic 的直接遍历场景
func (g *AnalysisGenerator) BuildTaskUserPromptFromMap(taskData map[string]interface{}, questions []Question) string {
	var sb strings.Builder

	// 提取任务字段
	taskName, _ := taskData["name"].(string)
	taskPathName, _ := taskData["pathName"].(string)
	taskDescription, _ := taskData["description"].(string)
	taskStatus, _ := taskData["status"].(string)
	taskPrompt, _ := taskData["prompt"].(string)
	taskTests, _ := taskData["tests"].(string)
	codePaths, _ := taskData["codePaths"].(string)

	sb.WriteString("## 当前任务信息\n\n")
	if taskPathName != "" {
		sb.WriteString(fmt.Sprintf("**任务路径**: %s\n", taskPathName))
	}
	if taskName != "" {
		sb.WriteString(fmt.Sprintf("**任务名称**: %s\n", taskName))
	}
	if taskStatus != "" {
		sb.WriteString(fmt.Sprintf("**状态**: %s\n", taskStatus))
	}

	if taskDescription != "" {
		sb.WriteString("\n### 任务描述\n")
		sb.WriteString(taskDescription)
		sb.WriteString("\n")
	}

	if taskPrompt != "" {
		sb.WriteString("\n### 实现提示词（Prompt）\n")
		sb.WriteString(taskPrompt)
		sb.WriteString("\n")
	}

	if taskTests != "" {
		sb.WriteString("\n### 测试要求\n")
		sb.WriteString(taskTests)
		sb.WriteString("\n")
	}

	if codePaths != "" {
		sb.WriteString("\n### 代码路径\n")
		sb.WriteString(codePaths)
		sb.WriteString("\n")
	}

	// 上游契约
	if upstreamContract, ok := taskData["upstreamContractDetail"]; ok && upstreamContract != nil {
		sb.WriteString("\n### 上游接口契约\n")
		sb.WriteString(fmt.Sprintf("%v\n", upstreamContract))
	}

	// 下游契约
	if downstreamContract, ok := taskData["downstreamContractDetail"]; ok && downstreamContract != nil {
		sb.WriteString("\n### 下游接口契约\n")
		sb.WriteString(fmt.Sprintf("%v\n", downstreamContract))
	}

	sb.WriteString("\n## 分析问题\n\n")
	sb.WriteString("请针对以下问题进行分析：\n\n")
	for i, q := range questions {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, q.Question))
	}

	return sb.String()
}

// ParseAIResponseByPath 解析AI响应为结构化结果（使用 pathName 标识，不使用数字 ID）
// 适用于不通过数字 ID 进行分析的场景
func (g *AnalysisGenerator) ParseAIResponseByPath(response string, analysisType string, targetPathName string, targetName string) (*AnalysisResult, error) {
	g.logger.Debug("解析AI响应（by path）", map[string]interface{}{
		"analysis_type":    analysisType,
		"target_path_name": targetPathName,
		"response_len":     len(response),
	})

	// 提取JSON部分
	jsonStr := g.extractJSON(response)
	if jsonStr == "" {
		return nil, NewBusinessError("无法从AI响应中提取JSON")
	}

	// 解析JSON
	var parsed struct {
		Summary     string   `json:"summary"`
		Score       int      `json:"score"`
		Confidence  float64  `json:"confidence"`
		Issues      []Issue  `json:"issues"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		g.logger.Error("解析AI响应JSON失败（by path）", err, map[string]interface{}{
			"json_str": jsonStr,
		})
		return nil, NewInternalError("解析AI响应失败", err)
	}

	// 验证和修正分数
	if parsed.Score < 0 {
		parsed.Score = 0
	} else if parsed.Score > 100 {
		parsed.Score = 100
	}

	// 验证和修正置信度
	if parsed.Confidence < 0 {
		parsed.Confidence = 0
	} else if parsed.Confidence > 1 {
		parsed.Confidence = 1
	}
	if parsed.Confidence == 0 {
		parsed.Confidence = 0.8 // 默认置信度
	}

	// 生成分析ID（使用 pathName 代替数字 ID）
	analysisID := fmt.Sprintf("%s_%s_%d", analysisType, strings.ReplaceAll(targetPathName, "/", "_"), time.Now().UnixMilli())

	// 构建结果（TargetID 设为 0，使用 TargetName 和 AnalysisID 标识）
	result := &AnalysisResult{
		AnalysisID:   analysisID,
		AnalysisType: analysisType,
		TargetID:     0,
		TargetName:   targetName,
		Summary:      parsed.Summary,
		Score:        parsed.Score,
		Confidence:   parsed.Confidence,
		Issues:       g.validateIssues(parsed.Issues),
		Suggestions:  parsed.Suggestions,
		AnalyzedAt:   time.Now(),
		RawResponse:  response,
	}

	g.logger.Info("AI响应解析完成（by path）", map[string]interface{}{
		"analysis_id":  result.AnalysisID,
		"target_path":  targetPathName,
		"score":        result.Score,
		"issues_count": len(result.Issues),
	})

	return result, nil
}
