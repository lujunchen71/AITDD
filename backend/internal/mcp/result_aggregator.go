package mcp

import (
	"fmt"
	"sync"
	"time"
)

// ResultAggregator 结果聚合器
// 用于将多个分析结果合并为单一的综合结果
type ResultAggregator struct {
	logger *MCPLogger
	mu     sync.RWMutex
}

// AggregatedResult 聚合分析结果
// 包含多个分析结果的综合信息
type AggregatedResult struct {
	Summary          string             `json:"summary"`
	TotalResults     int                `json:"total_results"`
	SuccessCount     int                `json:"success_count"`
	FailedCount      int                `json:"failed_count"`
	Issues           []Issue            `json:"issues"`
	Suggestions      []string           `json:"suggestions"`
	TotalScore       int                `json:"total_score"`
	AvgScore         float64            `json:"avg_score"`
	HighScore        int                `json:"high_score"`
	LowScore         int                `json:"low_score"`
	IssuesBySeverity map[string]int     `json:"issues_by_severity"`
	IssuesByCategory map[string]int     `json:"issues_by_category"`
	AnalyzedAt       time.Time          `json:"analyzed_at"`
	TotalDuration    time.Duration      `json:"total_duration"`
	AvgDuration      time.Duration      `json:"avg_duration"`
	RawResponses     []string           `json:"raw_responses,omitempty"`
	ByType           map[string]int     `json:"by_type"`
}

// NewResultAggregator 创建新的结果聚合器
func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{
		logger: GetMCPLogger(),
	}
}

// NewResultAggregatorWithLogger 创建带自定义日志的结果聚合器
func NewResultAggregatorWithLogger(logger *MCPLogger) *ResultAggregator {
	return &ResultAggregator{
		logger: logger,
	}
}

// Aggregate 聚合多个分析结果
// results: 多个分析结果
// 返回聚合后的综合结果
func (a *ResultAggregator) Aggregate(results []*AnalysisResult) *AggregatedResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(results) == 0 {
		return &AggregatedResult{
			Summary:          "无分析结果",
			TotalResults:     0,
			Issues:           []Issue{},
			Suggestions:      []string{},
			IssuesBySeverity: map[string]int{},
			IssuesByCategory: map[string]int{},
			ByType:           map[string]int{},
			AnalyzedAt:       time.Now(),
		}
	}

	aggregated := &AggregatedResult{
		TotalResults:     len(results),
		Issues:           []Issue{},
		Suggestions:      []string{},
		IssuesBySeverity: map[string]int{},
		IssuesByCategory: map[string]int{},
		ByType:           map[string]int{},
		AnalyzedAt:       results[0].AnalyzedAt,
		HighScore:        0,
		LowScore:         100,
	}

	totalScore := 0
	totalDuration := time.Duration(0)
	seenSuggestions := make(map[string]bool)
	rawResponses := []string{}

	for _, result := range results {
		if result == nil {
			aggregated.FailedCount++
			continue
		}

		aggregated.SuccessCount++

		// 统计分数
		totalScore += result.Score
		if result.Score > aggregated.HighScore {
			aggregated.HighScore = result.Score
		}
		if result.Score < aggregated.LowScore {
			aggregated.LowScore = result.Score
		}

		// 收集问题
		for _, issue := range result.Issues {
			aggregated.Issues = append(aggregated.Issues, issue)
			aggregated.IssuesBySeverity[issue.Severity]++
			aggregated.IssuesByCategory[issue.Category]++
		}

		// 收集建议（去重）
		for _, suggestion := range result.Suggestions {
			if !seenSuggestions[suggestion] {
				seenSuggestions[suggestion] = true
				aggregated.Suggestions = append(aggregated.Suggestions, suggestion)
			}
		}

		// 统计分析类型
		aggregated.ByType[result.AnalysisType]++

		// 收集原始响应
		if result.RawResponse != "" {
			rawResponses = append(rawResponses, result.RawResponse)
		}
	}

	// 计算平均分
	if aggregated.SuccessCount > 0 {
		aggregated.TotalScore = totalScore
		aggregated.AvgScore = float64(totalScore) / float64(aggregated.SuccessCount)
		aggregated.TotalDuration = totalDuration
		aggregated.AvgDuration = totalDuration / time.Duration(aggregated.SuccessCount)
	}

	// 如果没有成功结果，重置高低分
	if aggregated.SuccessCount == 0 {
		aggregated.HighScore = 0
		aggregated.LowScore = 0
	}

	aggregated.RawResponses = rawResponses

	// 生成汇总文本
	aggregated.Summary = a.generateSummary(aggregated)

	a.logger.Info("结果聚合完成", map[string]interface{}{
		"total_results":  aggregated.TotalResults,
		"success_count":  aggregated.SuccessCount,
		"failed_count":   aggregated.FailedCount,
		"total_issues":   len(aggregated.Issues),
		"avg_score":      aggregated.AvgScore,
	})

	return aggregated
}

// AggregateFromExecutorResults 从执行器结果聚合
// executorResults: 并发执行器的结果
// analysisResults: 解析后的分析结果（与executorResults对应）
// 返回聚合后的综合结果
func (a *ResultAggregator) AggregateFromExecutorResults(executorResults []*ExecutorResult, analysisResults []*AnalysisResult) *AggregatedResult {
	// 将分析结果聚合
	aggregated := a.Aggregate(analysisResults)

	// 补充执行器统计信息
	totalDuration := time.Duration(0)
	for _, er := range executorResults {
		if er != nil {
			totalDuration += er.Duration
		}
	}

	if len(executorResults) > 0 {
		aggregated.TotalDuration = totalDuration
		if aggregated.SuccessCount > 0 {
			aggregated.AvgDuration = totalDuration / time.Duration(aggregated.SuccessCount)
		}
	}

	return aggregated
}

// generateSummary 生成汇总文本
func (a *ResultAggregator) generateSummary(aggregated *AggregatedResult) string {
	var summary string

	summary += fmt.Sprintf("## 分析汇总\n\n")
	summary += fmt.Sprintf("- 总分析数量: %d\n", aggregated.TotalResults)
	summary += fmt.Sprintf("- 成功: %d, 失败: %d\n", aggregated.SuccessCount, aggregated.FailedCount)

	if aggregated.SuccessCount > 0 {
		summary += fmt.Sprintf("- 平均分: %.1f (最高: %d, 最低: %d)\n",
			aggregated.AvgScore, aggregated.HighScore, aggregated.LowScore)
	}

	summary += fmt.Sprintf("- 发现问题总数: %d\n", len(aggregated.Issues))
	summary += fmt.Sprintf("- 建议数量: %d\n", len(aggregated.Suggestions))

	if len(aggregated.IssuesBySeverity) > 0 {
		summary += "\n### 问题按严重级别分布\n"
		if count, ok := aggregated.IssuesBySeverity[SeverityCritical]; ok {
			summary += fmt.Sprintf("- 严重(critical): %d\n", count)
		}
		if count, ok := aggregated.IssuesBySeverity[SeverityWarning]; ok {
			summary += fmt.Sprintf("- 警告(warning): %d\n", count)
		}
		if count, ok := aggregated.IssuesBySeverity[SeverityInfo]; ok {
			summary += fmt.Sprintf("- 信息(info): %d\n", count)
		}
	}

	if len(aggregated.ByType) > 0 {
		summary += "\n### 按分析类型分布\n"
		for analysisType, count := range aggregated.ByType {
			summary += fmt.Sprintf("- %s: %d\n", analysisType, count)
		}
	}

	return summary
}

// FilterBySeverity 按严重级别过滤问题
// aggregated: 聚合结果
// minSeverity: 最低严重级别（critical > warning > info）
// 返回过滤后的问题列表
func (a *ResultAggregator) FilterBySeverity(aggregated *AggregatedResult, minSeverity string) []Issue {
	a.mu.RLock()
	defer a.mu.RUnlock()

	severityLevel := map[string]int{
		SeverityCritical: 3,
		SeverityWarning:  2,
		SeverityInfo:     1,
	}

	minLevel := severityLevel[minSeverity]
	if minLevel == 0 {
		minLevel = 1
	}

	filtered := []Issue{}
	for _, issue := range aggregated.Issues {
		level := severityLevel[issue.Severity]
		if level >= minLevel {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

// GetTopIssues 获取最重要的问题
// aggregated: 聚合结果
// maxCount: 最多返回的问题数量
// 返回按严重级别排序的问题列表
func (a *ResultAggregator) GetTopIssues(aggregated *AggregatedResult, maxCount int) []Issue {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 按严重级别排序：critical > warning > info
	severityOrder := map[string]int{
		SeverityCritical: 3,
		SeverityWarning:  2,
		SeverityInfo:     1,
	}

	// 收集各级别的问题
	critical := []Issue{}
	warning := []Issue{}
	info := []Issue{}

	for _, issue := range aggregated.Issues {
		switch severityOrder[issue.Severity] {
		case 3:
			critical = append(critical, issue)
		case 2:
			warning = append(warning, issue)
		default:
			info = append(info, issue)
		}
	}

	// 按优先级排序组合
	sorted := append(critical, warning...)
	sorted = append(sorted, info...)

	if maxCount > 0 && len(sorted) > maxCount {
		return sorted[:maxCount]
	}

	return sorted
}
