package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// CompileDynamicAIParams compile_dynamic工具的AI分析参数
// 这些配置内容由AI助手从项目根目录的 .aitdd/ 文件夹中读取后传入
type CompileDynamicAIParams struct {
	// ProjectName 项目名称
	ProjectName string `json:"project_name"`
	// SubAgentConfig AI子代理配置内容（从 .aitdd/sub_agent.json 读取）
	SubAgentConfig interface{} `json:"sub_agent_config"`
	// QAConfig 问题配置内容（从 .aitdd/qa.json 读取）
	QAConfig interface{} `json:"qa_config"`
}

// ParseAIParams 从compile_dynamic工具参数中解析AI分析参数
func ParseAIParams(args map[string]interface{}) (*CompileDynamicAIParams, error) {
	params := &CompileDynamicAIParams{}

	if v, ok := args["project_name"]; ok {
		if s, ok := v.(string); ok {
			params.ProjectName = s
		}
	}

	if v, ok := args["sub_agent_config"]; ok {
		params.SubAgentConfig = v
	}

	if v, ok := args["qa_config"]; ok {
		params.QAConfig = v
	}

	return params, nil
}

// RunAIAnalysisForCompile 在compile_dynamic工具中执行AI分析
// 直接遍历所有节点（任务、模块）进行分析，不依赖数字 ID
// projectName: 项目名称（pathName格式）
// subAgentConfigRaw: 来自参数的sub_agent配置（JSON字符串或map）
// qaConfigRaw: 来自参数的qa配置（JSON字符串或map）
// tasks: 已收集的任务列表（map[string]interface{}）
func RunAIAnalysisForCompile(ctx context.Context, projectName string, subAgentConfigRaw interface{}, qaConfigRaw interface{}, tasks []map[string]interface{}) (string, error) {
	// 解析 SubAgentConfig
	var subAgentConfig *SubAgentConfig
	if subAgentConfigRaw != nil {
		var err error
		subAgentConfig, err = parseSubAgentConfig(subAgentConfigRaw)
		if err != nil {
			return "", fmt.Errorf("解析SubAgentConfig失败: %w", err)
		}
	}

	// 解析 QAConfig
	var qaConfig *QAConfig
	if qaConfigRaw != nil {
		var err error
		qaConfig, err = parseQAConfig(qaConfigRaw)
		if err != nil {
			return "", fmt.Errorf("解析QAConfig失败: %w", err)
		}
	}
	if qaConfig == nil {
		qaConfig = GetDefaultQAConfig()
	}

	// 如果没有任务，返回提示
	if len(tasks) == 0 {
		return fmt.Sprintf("项目 %s 没有找到任务节点，跳过AI分析", projectName), nil
	}

	// 使用 subAgentConfig 创建 AI 客户端（如果配置了的话）
	var aiClient *AIClient
	if subAgentConfig != nil && len(subAgentConfig.Models) > 0 {
		aiClient = NewAIClient(subAgentConfig)
	} else {
		aiClient = NewAIClientWithLogger(GetSubAgentConfigManager().GetConfig(), GetMCPLogger())
	}

	logger := GetMCPLogger()
	generator := NewAnalysisGeneratorWithLogger(qaConfig, logger)
	taskQuestions := qaConfig.GetTaskQuestions()

	// 确定并发数（从 sub_agent_config.rate_limit.max_concurrent 读取）
	maxConcurrent := 3 // 默认并发数
	if subAgentConfig != nil && subAgentConfig.RateLimit.MaxConcurrent > 0 {
		maxConcurrent = subAgentConfig.RateLimit.MaxConcurrent
	}

	logger.Info("开始并发 AI 分析", map[string]interface{}{
		"task_count":     len(tasks),
		"max_concurrent": maxConcurrent,
	})

	// 并发执行 AI 分析
	type taskResult struct {
		index  int
		result *AnalysisResult
		err    string
	}

	resultCh := make(chan taskResult, len(tasks))
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, taskData := range tasks {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		wg.Add(1)
		go func(idx int, td map[string]interface{}) {
			defer wg.Done()

			// 获取并发令牌
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				resultCh <- taskResult{index: idx, err: "上下文已取消"}
				return
			default:
			}

			taskName, _ := td["name"].(string)
			taskPathName, _ := td["pathName"].(string)
			if taskPathName == "" {
				taskPathName = taskName
			}

			// 构建提示词
			userPrompt := generator.BuildTaskUserPromptFromMap(td, taskQuestions)
			systemPrompt := generator.BuildSystemPrompt(QuestionTypeTask)

			aiReq := &AIRequest{
				Messages: []Message{
					{Role: "system", Content: systemPrompt},
					{Role: "user", Content: userPrompt},
				},
			}

			aiResp, err := aiClient.Call(ctx, aiReq)
			if err != nil {
				resultCh <- taskResult{index: idx, err: fmt.Sprintf("[%s] AI调用失败: %v", taskPathName, err)}
				return
			}

			targetLabel := taskPathName
			if targetLabel == "" {
				targetLabel = taskName
			}
			result, err := generator.ParseAIResponseByPath(aiResp.Content, QuestionTypeTask, taskPathName, targetLabel)
			if err != nil {
				resultCh <- taskResult{index: idx, err: fmt.Sprintf("[%s] 解析AI响应失败: %v", taskPathName, err)}
				return
			}

			resultCh <- taskResult{index: idx, result: result}
		}(i, taskData)
	}

done:
	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 按原始顺序收集结果
	resultMap := make(map[int]*AnalysisResult)
	var analyzeErrors []string

	for tr := range resultCh {
		if tr.err != "" {
			analyzeErrors = append(analyzeErrors, tr.err)
		} else if tr.result != nil {
			resultMap[tr.index] = tr.result
		}
	}

	// 按顺序构建结果列表
	results := make([]*AnalysisResult, 0, len(resultMap))
	for i := 0; i < len(tasks); i++ {
		if r, ok := resultMap[i]; ok {
			results = append(results, r)
		}
	}

	logger.Info("并发 AI 分析完成", map[string]interface{}{
		"success_count": len(results),
		"error_count":   len(analyzeErrors),
	})

	aggregator := NewResultAggregatorWithLogger(logger)
	aggregated := aggregator.Aggregate(results)

	output := formatCompileAIReport(results, aggregated, projectName, maxConcurrent, len(tasks))
	if len(analyzeErrors) > 0 {
		output += "\n### 分析错误\n"
		for _, errMsg := range analyzeErrors {
			output += fmt.Sprintf("- %s\n", errMsg)
		}
	}
	return output, nil
}

// parseSubAgentConfig 从原始输入解析SubAgentConfig
func parseSubAgentConfig(raw interface{}) (*SubAgentConfig, error) {
	var jsonBytes []byte
	var err error

	switch v := raw.(type) {
	case string:
		jsonBytes = []byte(v)
	case map[string]interface{}:
		jsonBytes, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的sub_agent_config格式: %T", raw)
	}

	config := &SubAgentConfig{}
	if err := json.Unmarshal(jsonBytes, config); err != nil {
		return nil, fmt.Errorf("解析sub_agent_config JSON失败: %w", err)
	}

	return config, nil
}

// parseQAConfig 从原始输入解析QAConfig
func parseQAConfig(raw interface{}) (*QAConfig, error) {
	var jsonBytes []byte
	var err error

	switch v := raw.(type) {
	case string:
		jsonBytes = []byte(v)
	case map[string]interface{}:
		jsonBytes, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的qa_config格式: %T", raw)
	}

	config := &QAConfig{}
	if err := json.Unmarshal(jsonBytes, config); err != nil {
		return nil, fmt.Errorf("解析qa_config JSON失败: %w", err)
	}

	return config, nil
}

// analysisTypeLabel 获取分析类型的中文标签
func analysisTypeLabel(analysisType string) string {
	switch analysisType {
	case QuestionTypeTask:
		return "任务"
	case QuestionTypeModule:
		return "模块"
	case QuestionTypeProject:
		return "项目"
	default:
		return analysisType
	}
}

// formatCompileAIReport 格式化AI分析报告为字符串
// 每个分析结果单独展示，清晰标注类型（任务/模块/项目）和路径名
func formatCompileAIReport(results []*AnalysisResult, aggregated *AggregatedResult, projectName string, maxConcurrent int, totalTasks int) string {
	output := fmt.Sprintf("## AI设计合理性分析报告 - %s\n\n", projectName)

	// 显示并发配置信息
	output += fmt.Sprintf("**并发配置**: 最大并发数 = %d  |  任务总数 = %d\n\n", maxConcurrent, totalTasks)

	if len(results) == 0 {
		output += "AI分析完成，但未生成报告\n"
		return output
	}

	// 汇总统计
	if aggregated != nil {
		output += fmt.Sprintf("**分析对象总数**: %d  |  **平均评分**: %.1f/100  |  严重: %d  |  警告: %d\n\n",
			aggregated.TotalResults,
			aggregated.AvgScore,
			aggregated.IssuesBySeverity[SeverityCritical],
			aggregated.IssuesBySeverity[SeverityWarning],
		)
	}

	output += "---\n\n"

	// 逐个展示分析结果
	for _, r := range results {
		typeLabel := analysisTypeLabel(r.AnalysisType)

		// 从 AnalysisID 中提取 pathName（格式: "task_QuickSearch_核心模块_类型定义_timestamp"）
		// TargetName 只是名称，我们用它显示
		output += fmt.Sprintf("### 【%s】%s  (评分: %d/100)\n\n", typeLabel, r.TargetName, r.Score)

		if r.Summary != "" {
			output += fmt.Sprintf("**摘要**: %s\n\n", r.Summary)
		}

		// 显示严重问题
		criticalIssues := filterIssuesBySeverity(r.Issues, SeverityCritical)
		if len(criticalIssues) > 0 {
			output += "**严重问题**:\n"
			for _, issue := range criticalIssues {
				output += fmt.Sprintf("- [%s] %s\n", issue.Category, issue.Description)
				if issue.Suggestion != "" {
					output += fmt.Sprintf("  → 建议: %s\n", issue.Suggestion)
				}
			}
			output += "\n"
		}

		// 显示警告问题
		warningIssues := filterIssuesBySeverity(r.Issues, SeverityWarning)
		if len(warningIssues) > 0 {
			output += "**警告**:\n"
			for _, issue := range warningIssues {
				output += fmt.Sprintf("- [%s] %s\n", issue.Category, issue.Description)
			}
			output += "\n"
		}

		// 显示建议
		if len(r.Suggestions) > 0 {
			output += "**优化建议**:\n"
			maxSugg := 3
			if len(r.Suggestions) < maxSugg {
				maxSugg = len(r.Suggestions)
			}
			for _, s := range r.Suggestions[:maxSugg] {
				output += fmt.Sprintf("- %s\n", s)
			}
			output += "\n"
		}

		output += "---\n\n"
	}

	return output
}

// filterIssuesBySeverity 按严重程度过滤问题
func filterIssuesBySeverity(issues []Issue, severity string) []Issue {
	var filtered []Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// ==================== 辅助函数（保留供其他文件使用）====================

// parseIDList 解析ID列表参数
// 支持 []interface{}、[]float64、[]int 等格式
func parseIDList(raw interface{}) ([]uint, error) {
	if raw == nil {
		return nil, fmt.Errorf("ID列表为nil")
	}

	var ids []uint

	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			switch id := item.(type) {
			case float64:
				ids = append(ids, uint(id))
			case int:
				ids = append(ids, uint(id))
			case int64:
				ids = append(ids, uint(id))
			default:
				return nil, fmt.Errorf("无效的ID类型: %T", item)
			}
		}
	case []float64:
		for _, id := range v {
			ids = append(ids, uint(id))
		}
	case []int:
		for _, id := range v {
			ids = append(ids, uint(id))
		}
	default:
		return nil, fmt.Errorf("无效的ID列表类型: %T", raw)
	}

	return ids, nil
}
