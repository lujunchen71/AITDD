package mcp

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ResultAggregator 结果聚合器
// 用于聚合多个分析结果，生成综合报告
type ResultAggregator struct {
	analysisID    string
	analysisType  string
	results       []*AnalysisResult
	errors        []error
	startTime     time.Time
	logger        *MCPLogger
	mu            sync.RWMutex
}

// AggregatedReport 聚合报告
// 包含所有分析结果的综合信息
type AggregatedReport struct {
	AnalysisID      string     `json:"analysis_id"`
	AnalysisType    string     `json:"analysis_type"`
	TotalAnalyzed   int        `json:"total_analyzed"`
	SuccessCount   int        `json:"success_count"`
	ErrorCount     int        `json:"error_count"`
	AllIssues       []Issue   `json:"all_issues"`
	CriticalIssues []Issue   `json:"critical_issues"`
	WarningIssues  []Issue   `json:"warning_issues"`
	InfoIssues      []Issue   `json:"info_issues"`
	Suggestions     []string   `json:"suggestions"`
	AverageScore    float64    `json:"average_score"`
	Duration        time.Duration `json:"duration"`
	AnalyzedAt      time.Time  `json:"analyzed_at"`
	Summary         string     `json:"summary"`
}

// IssueSeverityStats 问题严重级别统计
type IssueSeverityStats struct {
	CriticalCount int `json:"critical_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

// NewResultAggregator 创建新的结果聚合器
// analysisID: 分析ID
// analysisType: 分析类型 (task, module, project)
func NewResultAggregator(analysisID, analysisType string) *ResultAggregator {
	return &ResultAggregator{
		analysisID:   analysisID,
		analysisType: analysisType,
		results:       make([]*AnalysisResult, 0),
		errors:        make([]error, 0),
		startTime:   time.Now(),
		logger:      GetMCPLogger(),
	}
}

// NewResultAggregatorWithLogger 创建带自定义日志的结果聚合器
func NewResultAggregatorWithLogger(analysisID, analysisType string, logger *MCPLogger) *ResultAggregator {
	return &ResultAggregator{
		analysisID:   analysisID,
		analysisType: analysisType,
		results:       make([]*AnalysisResult, 0),
		errors:        make([]error, 0),
		startTime:   time.Now(),
		logger:      logger,
	}
}

// AddResult 添加分析结果
// 緻加一个成功完成的分析结果到聚合器中func (a *ResultAggregator) AddResult(result *AnalysisResult) {
	if result == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.results = append(a.results, result)

	a.logger.Debug("添加分析结果", map[string]interface{}{
		"analysis_id": result.AnalysisID,
		"target_id":   result.TargetID,
		"target_name": result.TargetName,
		"score":       result.Score,
		"issues_count": len(result.Issues),
	})
}

// AddError 添加错误
// 记录分析过程中发生的错误
func (a *ResultAggregator) AddError(err error) {
	if err == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.errors = append(a.errors, err)

	a.logger.Error("添加分析错误", err, map[string]interface{}{
		"analysis_id": a.analysisID,
		"error":       err.Error(),
	})
}

// Aggregate 生成聚合报告
// 根据所有添加的分析结果生成综合报告
func (a *ResultAggregator) Aggregate() *AggregatedReport {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 计算持续时间
	duration := time.Since(a.startTime)

	// 创建报告
	report := &AggregatedReport{
		AnalysisID:    a.analysisID,
		AnalysisType:  a.analysisType,
		TotalAnalyzed:  len(a.results),
		SuccessCount:  len(a.results),
		ErrorCount:     len(a.errors),
		AllIssues:      make([]Issue, 0),
		CriticalIssues: make([]Issue, 0),
		WarningIssues:  make([]Issue, 0),
		InfoIssues:     make([]Issue, 0),
		Suggestions:    make([]string, 0),
		Duration:       duration,
		AnalyzedAt:     time.Now(),
	}

	// 聚合所有结果
	issueMap := make(map[string]bool) // 用于去重
	for _, result := range a.results {
		// 聚合问题
		for _, issue := range result.Issues {
			key := issue.ID + "_" + issue.Description
			if !issueMap[key] {
				issueMap[key] = true
				report.AllIssues = append(report.AllIssues, issue)
			}
		}

		// 聚合建议
		report.Suggestions = append(report.Suggestions, result.Suggestions...)
	}

	// 分类问题
	report.CriticalIssues, report.WarningIssues, report.InfoIssues = a.classifyIssues(report.AllIssues)

	// 计算平均分数
	report.AverageScore = a.calculateAverageScore()

	// 生成摘要
	report.Summary = a.GetSummary()

	a.logger.Info("生成聚合报告", map[string]interface{}{
		"analysis_id":     a.analysisID,
		"total_analyzed": report.TotalAnalyzed,
		"error_count":    report.ErrorCount,
		"all_issues":      len(report.AllIssues),
		"critical_issues": len(report.CriticalIssues),
		"warning_issues":  len(report.WarningIssues),
		"average_score":   report.AverageScore,
		"duration":        duration.String(),
	})

	return report
}

// classifyIssues 分类问题
// 按严重级别将问题分类
func (a *ResultAggregator) classifyIssues(issues []Issue) (critical, warning, info []Issue) {
	critical = make([]Issue, 0)
	warning = make([]Issue, 0)
	info = make([]Issue, 0)

	for _, issue := range issues {
		switch issue.Severity {
		case SeverityCritical:
			critical = append(critical, issue)
		case SeverityWarning:
			warning = append(warning, issue)
		case SeverityInfo:
			info = append(info, issue)
		default:
			// 未知严重级别，默认为info
			info = append(info, issue)
		}
	}

	return critical, warning, info
}

// calculateAverageScore 计算平均分数
// 计算所有成功分析结果的平均分数
func (a *ResultAggregator) calculateAverageScore() float64 {
	if len(a.results) == 0 {
		return 0
	}

	var totalScore float64
	for _, result := range a.results {
		totalScore += float64(result.Score)
	}

	return totalScore / float64(len(a.results))
}

// GetSummary 生成摘要
// 根据分析结果生成综合摘要
func (a *ResultAggregator) GetSummary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.results) == 0 {
		if len(a.errors) > 0 {
			return fmt.Sprintf("分析失败，共 %d 个错误", len(a.errors))
		}
		return "无分析结果"
	}

	// 统计问题
	stats := a.GetIssueStats()

	// 计算平均分数
	avgScore := a.calculateAverageScore()

	// 构建摘要
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("共分析 %d 个对象， ", len(a.results)))

	if len(a.errors) > 0 {
		sb.WriteString(fmt.Sprintf("其中 %d 个失败, ", len(a.errors)))
	}

	sb.WriteString(fmt.Sprintf("平均评分 %.1f 分. ", avgScore))

	if stats.CriticalCount > 0 {
		sb.WriteString(fmt.Sprintf("发现 %d 个严重问题, ", stats.CriticalCount))
	}
	if stats.WarningCount > 0 {
		sb.WriteString(fmt.Sprintf("%d 个警告. ", stats.WarningCount))
	}
	if stats.InfoCount > 0 {
		sb.WriteString(fmt.Sprintf("%d 个提示信息. ", stats.InfoCount))
	}

	// 添加总体建议
	if stats.CriticalCount > 0 {
		sb.WriteString("建议优先处理严重问题。")
	} else if stats.WarningCount > 0 {
		sb.WriteString("建议尽快处理警告问题.")
	} else {
		sb.WriteString("整体质量良好.")
	}

	return sb.String()
}

// GetIssueStats 获取问题统计
func (a *ResultAggregator) GetIssueStats() *IssueSeverityStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := &IssueSeverityStats{}

	for _, result := range a.results {
		for _, issue := range result.Issues {
			switch issue.Severity {
			case SeverityCritical:
				stats.CriticalCount++
			case SeverityWarning:
				stats.WarningCount++
			case SeverityInfo:
				stats.InfoCount++
			}
		}
	}

	return stats
}

// GetResults 获取所有分析结果
func (a *ResultAggregator) GetResults() []*AnalysisResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 返回副本
	results := make([]*AnalysisResult, len(a.results))
	copy(results, a.results)
	return results
}

// GetErrors 获取所有错误
func (a *ResultAggregator) GetErrors() []error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 返回副本
	errors := make([]error, len(a.errors))
	copy(errors, a.errors)
	return errors
}

// GetDuration 获取分析持续时间
func (a *ResultAggregator) GetDuration() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return time.Since(a.startTime)
}

// GetSuccessRate 获取成功率
func (a *ResultAggregator) GetSuccessRate() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := len(a.results) + len(a.errors)
	if total == 0 {
		return 0
	}

	return float64(len(a.results)) / float64(total) * 100
}

// HasErrors 检查是否有错误
func (a *ResultAggregator) HasErrors() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.errors) > 0
}

// HasCriticalIssues 检查是否有严重问题
func (a *ResultAggregator) HasCriticalIssues() bool {
	stats := a.GetIssueStats()
	return stats.CriticalCount > 0
}

// ToJSON 将聚合报告转换为JSON字符串
func (r *AggregatedReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML 将聚合报告转换为YAML格式字符串
func (r *AggregatedReport) ToYAML() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("analysis_id: %s\n", r.AnalysisID))
	sb.WriteString(fmt.Sprintf("analysis_type: %s\n", r.AnalysisType))
	sb.WriteString(fmt.Sprintf("total_analyzed: %d\n", r.TotalAnalyzed))
	sb.WriteString(fmt.Sprintf("success_count: %d\n", r.SuccessCount))
	sb.WriteString(fmt.Sprintf("error_count: %d\n", r.ErrorCount))
	sb.WriteString(fmt.Sprintf("average_score: %.1f\n", r.AverageScore))
	sb.WriteString(fmt.Sprintf("duration: %s\n", r.Duration.String()))
	sb.WriteString(fmt.Sprintf("analyzed_at: %s\n", r.AnalyzedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("summary: %s\n", r.Summary))

	if len(r.AllIssues) > 0 {
		sb.WriteString("all_issues:\n")
		for _, issue := range r.AllIssues {
			sb.WriteString("  - id: ")
			sb.WriteString(fmt.Sprintf("%s\n", issue.ID))
			sb.WriteString("    severity: ")
			sb.WriteString(fmt.Sprintf("%s\n", issue.Severity))
			sb.WriteString("    category: ")
			sb.WriteString(fmt.Sprintf("%s\n", issue.Category))
			sb.WriteString("    description: ")
			sb.WriteString(fmt.Sprintf("%s\n", issue.Description))
		}
	}

	if len(r.CriticalIssues) > 0 {
		sb.WriteString("critical_issues:\n")
		for _, issue := range r.CriticalIssues {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", issue.ID, issue.Description))
		}
	}

	if len(r.WarningIssues) > 0 {
		sb.WriteString("warning_issues:\n")
		for _, issue := range r.WarningIssues {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", issue.ID, issue.Description))
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

// GetTopIssues 获取最严重的问题
// 返回指定数量的最严重问题
func (r *AggregatedReport) GetTopIssues(count int) []Issue {
	if count <= 0 {
		count = 5
	}

	// 錙所有问题
	allIssues := make([]Issue, len(r.AllIssues))
	copy(allIssues, r.AllIssues)

	// 按严重级别排序
	severityOrder := map[string]int{
		SeverityCritical: 3,
		SeverityWarning:  2,
		SeverityInfo:     1,
	}

	// 排序
	for i := 0; i < len(allIssues)-1; i++ {
		for j := i + 1; j < len(allIssues); j++ {
			if severityOrder[allIssues[i].Severity] < severityOrder[allIssues[j].Severity] {
			 allIssues[i], allIssues[j] = allIssues[j], allIssues[i]
			}
        }
    }

	// 返回前N个
	if len(allIssues) < count {
		count = len(allIssues)
	}
	return allIssues[:count]
}

// MergeReports 合并多个聚合报告
// 将多个聚合报告合并为一个
func MergeReports(reports []*AggregatedReport) *AggregatedReport {
	if len(reports) == 0 {
		return nil
	}

	if len(reports) == 1 {
		return reports[0]
	}

	// 创建新的合并报告
	merged := &AggregatedReport{
		AnalysisID:    fmt.Sprintf("merged_%d", time.Now().UnixMilli()),
		AnalysisType:  "merged",
		AllIssues:      make([]Issue, 0),
		CriticalIssues: make([]Issue, 0),
		WarningIssues:  make([]Issue, 0),
		InfoIssues:     make([]Issue, 0),
		Suggestions:    make([]string, 0),
		AnalyzedAt:     time.Now(),
	}

	// 合并所有报告
	issueMap := make(map[string]bool)
	var totalDuration time.Duration
	var totalScore float64
	var scoreCount int

	for _, report := range reports {
		merged.TotalAnalyzed += report.TotalAnalyzed
		merged.SuccessCount += report.SuccessCount
		merged.ErrorCount += report.ErrorCount

		// 合并问题（去重）
		for _, issue := range report.AllIssues {
			key := issue.ID + "_" + issue.Description
			if !issueMap[key] {
			 issueMap[key] = true
                merged.AllIssues = append(merged.AllIssues, issue)
            }
        }

		// 合并建议
        merged.Suggestions = append(merged.Suggestions, report.Suggestions...)

        // 累加分数和持续时间
        totalScore += report.AverageScore * float64(report.TotalAnalyzed)
        scoreCount += report.TotalAnalyzed
        totalDuration += report.Duration
    }

	// 计算平均分数
	if scoreCount > 0 {
        merged.AverageScore = totalScore / float64(scoreCount)
    }

	// 设置持续时间
    merged.Duration = totalDuration

	// 分类问题
    merged.CriticalIssues, merged.WarningIssues, merged.InfoIssues = classifyIssuesGlobal(merged.AllIssues)

	// 生成摘要
    merged.Summary = generateMergedSummary(merged)

    return merged
}

// classifyIssuesGlobal 全局问题分类函数
func classifyIssuesGlobal(issues []Issue) (critical, warning, info []Issue) {
    critical = make([]Issue, 0)
    warning = make([]Issue, 0)
    info = make([]Issue, 0)

    for _, issue := range issues {
        switch issue.Severity {
        case SeverityCritical:
            critical = append(critical, issue)
        case SeverityWarning:
            warning = append(warning, issue)
        case SeverityInfo:
            info = append(info, issue)
        default:
            info = append(info, issue)
        }
    }

    return critical, warning, info
}

// generateMergedSummary 生成合并摘要
func generateMergedSummary(report *AggregatedReport) string {
    var sb strings.Builder

    sb.WriteString(fmt.Sprintf("合并了 %d 个分析结果, ", report.TotalAnalyzed))

    if report.ErrorCount > 0 {
        sb.WriteString(fmt.Sprintf("其中 %d 个失败. ", report.ErrorCount))
    }

    sb.WriteString(fmt.Sprintf("平均评分 %.1f 分. ", report.AverageScore))

    stats := &IssueSeverityStats{}
    for _, issue := range report.AllIssues {
        switch issue.Severity {
        case SeverityCritical:
            stats.CriticalCount++
        case SeverityWarning:
            stats.WarningCount++
        case SeverityInfo:
            stats.InfoCount++
        }
    }

    if stats.CriticalCount > 0 {
        sb.WriteString(fmt.Sprintf("发现 %d 个严重问题, ", stats.CriticalCount))
    }
    if stats.WarningCount > 0 {
        sb.WriteString(fmt.Sprintf("%d 个警告. ", stats.WarningCount))
    }
    if stats.InfoCount > 0 {
        sb.WriteString(fmt.Sprintf("%d 个提示信息. ", stats.InfoCount))
    }

    if stats.CriticalCount > 0 {
        sb.WriteString("建议优先处理严重问题.")
    } else if stats.WarningCount > 0 {
        sb.WriteString("建议尽快处理警告问题.")
    } else {
        sb.WriteString("整体质量良好.")
    }

    return sb.String()
}

// AggregatorRegistry 聚合器注册表
// 用于管理多个聚合器实例
type AggregatorRegistry struct {
	aggregators map[string]*ResultAggregator
	mu         sync.RWMutex
}

// 全局聚合器注册表
var globalAggregatorRegistry *AggregatorRegistry
var aggregatorRegistryOnce sync.Once

// GetAggregatorRegistry 获取全局聚合器注册表
func GetAggregatorRegistry() *AggregatorRegistry {
	aggregatorRegistryOnce.Do(func() {
		globalAggregatorRegistry = &AggregatorRegistry{
			aggregators: make(map[string]*ResultAggregator),
		}
	})
	return globalAggregatorRegistry
}

// Register 注册聚合器
func (ar *AggregatorRegistry) Register(aggregator *ResultAggregator) error {
	if aggregator == nil {
		return NewParamError("聚合器不能为空")
	}

	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.aggregators[aggregator.analysisID]; exists {
		return NewBusinessErrorf("分析ID %s 已存在", aggregator.analysisID)
	}

	ar.aggregators[aggregator.analysisID] = aggregator
	return nil
}

// Unregister 注销聚合器
func (ar *AggregatorRegistry) Unregister(analysisID string) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	delete(ar.aggregators, analysisID)
}

// Get 获取聚合器
func (ar *AggregatorRegistry) Get(analysisID string) (*ResultAggregator, bool) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	aggregator, exists := ar.aggregators[analysisID]
	return aggregator, exists
}

// GetAll 获取所有聚合器
func (ar *AggregatorRegistry) GetAll() map[string]*ResultAggregator {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	// 返回副本
	result := make(map[string]*ResultAggregator)
	for k, v := range ar.aggregators {
		result[k] = v
	}
	return result
}

// Clear 清除所有聚合器
func (ar *AggregatorRegistry) Clear() {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.aggregators = make(map[string]*ResultAggregator)
}

// Count 获取聚合器数量
func (ar *AggregatorRegistry) Count() int {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	return len(ar.aggregators)
}

// CreateAggregator 创建并注册聚合器的便捷方法
func CreateAggregator(analysisID, analysisType string) (*ResultAggregator, error) {
	registry := GetAggregatorRegistry()
	aggregator := NewResultAggregator(analysisID, analysisType)
	err := registry.Register(aggregator)
	if err != nil {
		return nil, err
	}
	return aggregator, nil
}
