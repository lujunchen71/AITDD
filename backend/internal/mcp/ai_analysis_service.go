package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AIAnalysisService AI分析服务
// 整合所有组件，提供任务、模块和项目的AI分析功能
type AIAnalysisService struct {
	configManager   *SubAgentConfigManager
	qaConfigManager *QAConfigManager
	aiClient        *AIClient
	dataCollector   *DataCollector
	generator       *AnalysisGenerator
	executor        *ConcurrentExecutor
	aggregator      *ResultAggregator
	logger          *MCPLogger

	// 进度报告器存储，用于跟踪正在进行的分析
	reporters map[string]*ProgressReporter
	reportMu  sync.RWMutex

	// 取消函数存储，用于取消正在进行的分析
	cancels  map[string]context.CancelFunc
	cancelMu sync.RWMutex

	// 分析结果存储
	analysisResults map[string][]*AnalysisResult
	resultMu        sync.RWMutex
}

// ServiceAnalysisOptions AI分析服务分析选项
type ServiceAnalysisOptions struct {
	AnalysisType    string     // task, module, project
	TargetIDs       []uint     // 要分析的任务/模块/项目ID
	CustomQuestions []Question // 自定义问题
	Concurrent      bool       // 是否并发执行
	MaxIssues       int        // 最多返回的问题数量
	MinSeverity     string     // 最低严重级别
}

// NewAIAnalysisService 创建AI分析服务
// apiBaseURL: REST API的基础URL
func NewAIAnalysisService(apiBaseURL string) *AIAnalysisService {
	configManager := GetSubAgentConfigManager()
	qaConfigManager := GetQAConfigManager()
	config := configManager.GetConfig()
	logger := GetMCPLogger()

	aiClient := NewAIClientWithLogger(config, logger)
	dataCollector := NewDataCollectorWithLogger(apiBaseURL, logger)
	qaConfig := qaConfigManager.GetConfig()
	generator := NewAnalysisGeneratorWithLogger(qaConfig, logger)
	executor := NewConcurrentExecutorWithLogger(config, aiClient, logger)
	aggregator := NewResultAggregatorWithLogger(logger)

	return &AIAnalysisService{
		configManager:   configManager,
		qaConfigManager: qaConfigManager,
		aiClient:        aiClient,
		dataCollector:   dataCollector,
		generator:       generator,
		executor:        executor,
		aggregator:      aggregator,
		logger:          logger,
		reporters:       make(map[string]*ProgressReporter),
		cancels:         make(map[string]context.CancelFunc),
		analysisResults: make(map[string][]*AnalysisResult),
	}
}

// NewAIAnalysisServiceWithComponents 使用指定组件创建AI分析服务（用于测试）
func NewAIAnalysisServiceWithComponents(
	configManager *SubAgentConfigManager,
	qaConfigManager *QAConfigManager,
	aiClient *AIClient,
	dataCollector *DataCollector,
	generator *AnalysisGenerator,
	executor *ConcurrentExecutor,
	aggregator *ResultAggregator,
	logger *MCPLogger,
) *AIAnalysisService {
	return &AIAnalysisService{
		configManager:   configManager,
		qaConfigManager: qaConfigManager,
		aiClient:        aiClient,
		dataCollector:   dataCollector,
		generator:       generator,
		executor:        executor,
		aggregator:      aggregator,
		logger:          logger,
		reporters:       make(map[string]*ProgressReporter),
		cancels:         make(map[string]context.CancelFunc),
		analysisResults: make(map[string][]*AnalysisResult),
	}
}

// AnalyzeTask 分析单个任务
// ctx: 上下文
// taskID: 任务ID
// opts: 分析选项（可为nil，使用默认选项）
// 返回分析结果
func (s *AIAnalysisService) AnalyzeTask(ctx context.Context, taskID uint, opts *AnalysisOptions) (*AnalysisResult, error) {
	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	s.logger.Info("开始分析任务", map[string]interface{}{
		"task_id": taskID,
	})

	// 收集任务数据
	taskData, err := s.dataCollector.CollectTaskData(ctx, taskID)
	if err != nil {
		return nil, NewNetworkErrorf(err, "收集任务 %d 数据失败", taskID)
	}

	// 生成分析请求
	analysisReq, err := s.generator.GenerateTaskAnalysis(ctx, taskData, opts)
	if err != nil {
		return nil, err
	}

	// 转换为AI请求
	aiReq := &AIRequest{
		Messages: []Message{
			{Role: "system", Content: analysisReq.SystemPrompt},
			{Role: "user", Content: analysisReq.UserPrompt},
		},
	}

	// 调用AI
	aiResp, err := s.aiClient.Call(ctx, aiReq)
	if err != nil {
		return nil, err
	}

	// 解析结果
	result, err := s.generator.ParseAIResponse(aiResp.Content, QuestionTypeTask, taskID, taskData.TaskName)
	if err != nil {
		return nil, err
	}

	s.logger.Info("任务分析完成", map[string]interface{}{
		"task_id": taskID,
		"score":   result.Score,
		"issues":  len(result.Issues),
	})

	return result, nil
}

// AnalyzeModule 分析单个模块
// ctx: 上下文
// moduleID: 模块ID
// opts: 分析选项（可为nil，使用默认选项）
// 返回分析结果
func (s *AIAnalysisService) AnalyzeModule(ctx context.Context, moduleID uint, opts *AnalysisOptions) (*AnalysisResult, error) {
	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	s.logger.Info("开始分析模块", map[string]interface{}{
		"module_id": moduleID,
	})

	// 收集模块数据
	moduleData, err := s.dataCollector.CollectModuleData(ctx, moduleID)
	if err != nil {
		return nil, NewNetworkErrorf(err, "收集模块 %d 数据失败", moduleID)
	}

	// 生成分析请求
	analysisReq, err := s.generator.GenerateModuleAnalysis(ctx, moduleData, opts)
	if err != nil {
		return nil, err
	}

	// 转换为AI请求
	aiReq := &AIRequest{
		Messages: []Message{
			{Role: "system", Content: analysisReq.SystemPrompt},
			{Role: "user", Content: analysisReq.UserPrompt},
		},
	}

	// 调用AI
	aiResp, err := s.aiClient.Call(ctx, aiReq)
	if err != nil {
		return nil, err
	}

	// 解析结果
	result, err := s.generator.ParseAIResponse(aiResp.Content, QuestionTypeModule, moduleID, moduleData.ModuleName)
	if err != nil {
		return nil, err
	}

	s.logger.Info("模块分析完成", map[string]interface{}{
		"module_id": moduleID,
		"score":     result.Score,
		"issues":    len(result.Issues),
	})

	return result, nil
}

// AnalyzeProject 分析单个项目
// ctx: 上下文
// projectID: 项目ID
// opts: 分析选项（可为nil，使用默认选项）
// 返回分析结果
func (s *AIAnalysisService) AnalyzeProject(ctx context.Context, projectID uint, opts *AnalysisOptions) (*AnalysisResult, error) {
	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	s.logger.Info("开始分析项目", map[string]interface{}{
		"project_id": projectID,
	})

	// 收集项目数据
	projectData, err := s.dataCollector.CollectProjectData(ctx, projectID)
	if err != nil {
		return nil, NewNetworkErrorf(err, "收集项目 %d 数据失败", projectID)
	}

	// 生成分析请求
	analysisReq, err := s.generator.GenerateProjectAnalysis(ctx, projectData, opts)
	if err != nil {
		return nil, err
	}

	// 转换为AI请求
	aiReq := &AIRequest{
		Messages: []Message{
			{Role: "system", Content: analysisReq.SystemPrompt},
			{Role: "user", Content: analysisReq.UserPrompt},
		},
	}

	// 调用AI
	aiResp, err := s.aiClient.Call(ctx, aiReq)
	if err != nil {
		return nil, err
	}

	// 解析结果
	result, err := s.generator.ParseAIResponse(aiResp.Content, QuestionTypeProject, projectID, projectData.ProjectName)
	if err != nil {
		return nil, err
	}

	s.logger.Info("项目分析完成", map[string]interface{}{
		"project_id": projectID,
		"score":      result.Score,
		"issues":     len(result.Issues),
	})

	return result, nil
}

// AnalyzeBatch 批量分析
// ctx: 上下文
// opts: 服务分析选项，包含分析类型和目标ID列表
// 返回分析ID，可通过GetProgress获取进度
func (s *AIAnalysisService) AnalyzeBatch(ctx context.Context, opts *ServiceAnalysisOptions) (string, error) {
	if opts == nil {
		return "", NewParamError("分析选项不能为空")
	}

	if len(opts.TargetIDs) == 0 {
		return "", NewParamError("目标ID列表不能为空")
	}

	if opts.AnalysisType == "" {
		opts.AnalysisType = QuestionTypeTask
	}

	// 生成分析ID（使用时间戳+随机数）
	analysisID := fmt.Sprintf("analysis_%d", time.Now().UnixNano())

	s.logger.Info("开始批量分析", map[string]interface{}{
		"analysis_id":   analysisID,
		"analysis_type": opts.AnalysisType,
		"target_count":  len(opts.TargetIDs),
		"concurrent":    opts.Concurrent,
	})

	// 创建进度报告器
	reporter := NewProgressReporter(analysisID, len(opts.TargetIDs))
	s.reportMu.Lock()
	s.reporters[analysisID] = reporter
	s.reportMu.Unlock()

	// 创建可取消的上下文
	cancelCtx, cancel := context.WithCancel(ctx)
	s.cancelMu.Lock()
	s.cancels[analysisID] = cancel
	s.cancelMu.Unlock()

	// 转换分析选项
	analysisOpts := &AnalysisOptions{
		IncludeRawResponse: true,
		CustomQuestions:    opts.CustomQuestions,
		MaxIssues:          opts.MaxIssues,
		MinSeverity:        opts.MinSeverity,
	}

	if analysisOpts.MaxIssues == 0 {
		analysisOpts.MaxIssues = 50
	}

	if analysisOpts.MinSeverity == "" {
		analysisOpts.MinSeverity = SeverityInfo
	}

	// 在后台执行分析
	go func() {
		defer func() {
			// 清理取消函数
			s.cancelMu.Lock()
			delete(s.cancels, analysisID)
			s.cancelMu.Unlock()
		}()

		reporter.Start()
		results := make([]*AnalysisResult, 0, len(opts.TargetIDs))
		hasAnySuccess := false

		for _, targetID := range opts.TargetIDs {
			// 检查是否已取消
			select {
			case <-cancelCtx.Done():
				reporter.Fail("分析被取消")
				return
			default:
			}

			var result *AnalysisResult
			var err error

			switch opts.AnalysisType {
			case QuestionTypeTask:
				result, err = s.AnalyzeTask(cancelCtx, targetID, analysisOpts)
			case QuestionTypeModule:
				result, err = s.AnalyzeModule(cancelCtx, targetID, analysisOpts)
			case QuestionTypeProject:
				result, err = s.AnalyzeProject(cancelCtx, targetID, analysisOpts)
			default:
				err = NewParamErrorf("不支持的分析类型: %s", opts.AnalysisType)
			}

			partialResult := &PartialResult{
				TaskID:   fmt.Sprintf("%s_%d", opts.AnalysisType, targetID),
				Type:     opts.AnalysisType,
				TargetID: targetID,
			}

			if err != nil {
				partialResult.Status = PartialResultFailed
				partialResult.Message = err.Error()
				s.logger.Error("分析目标失败", err, map[string]interface{}{
					"analysis_id": analysisID,
					"target_id":   targetID,
					"target_type": opts.AnalysisType,
				})
			} else {
				hasAnySuccess = true
				partialResult.Status = PartialResultSuccess
				partialResult.Message = fmt.Sprintf("分析完成，得分: %d", result.Score)
				partialResult.Score = result.Score
				results = append(results, result)
			}

			reporter.Increment(partialResult.TaskID, partialResult)
		}

		// 存储结果
		s.resultMu.Lock()
		s.analysisResults[analysisID] = results
		s.resultMu.Unlock()

		if hasAnySuccess || len(results) > 0 {
			reporter.Complete()
		} else {
			reporter.Fail("所有分析任务都失败了")
		}

		// 聚合结果
		if len(results) > 0 {
			aggregated := s.aggregator.Aggregate(results)
			s.logger.Info("批量分析聚合完成", map[string]interface{}{
				"analysis_id":   analysisID,
				"total_results": aggregated.TotalResults,
				"avg_score":     aggregated.AvgScore,
			})
		}
	}()

	return analysisID, nil
}

// GetProgress 获取分析进度
// analysisID: 分析ID
// 返回进度状态，如果分析ID不存在则返回nil
func (s *AIAnalysisService) GetProgress(analysisID string) *ProgressStatus {
	s.reportMu.RLock()
	reporter, exists := s.reporters[analysisID]
	s.reportMu.RUnlock()

	if !exists {
		return nil
	}

	return reporter.GetStatus()
}

// GetAnalysisResults 获取分析结果
// analysisID: 分析ID
// 返回分析结果列表，如果分析ID不存在则返回nil
func (s *AIAnalysisService) GetAnalysisResults(analysisID string) []*AnalysisResult {
	s.resultMu.RLock()
	defer s.resultMu.RUnlock()

	return s.analysisResults[analysisID]
}

// GetAggregatedResults 获取聚合分析结果
// analysisID: 分析ID
// 返回聚合结果，如果分析ID不存在则返回nil
func (s *AIAnalysisService) GetAggregatedResults(analysisID string) *AggregatedResult {
	s.resultMu.RLock()
	results, exists := s.analysisResults[analysisID]
	s.resultMu.RUnlock()

	if !exists {
		return nil
	}

	return s.aggregator.Aggregate(results)
}

// CancelAnalysis 取消分析
// analysisID: 分析ID
// 返回是否成功取消
func (s *AIAnalysisService) CancelAnalysis(analysisID string) bool {
	s.cancelMu.RLock()
	cancel, exists := s.cancels[analysisID]
	s.cancelMu.RUnlock()

	if !exists {
		return false
	}

	cancel()
	s.logger.Info("分析已取消", map[string]interface{}{
		"analysis_id": analysisID,
	})

	return true
}

// AnalyzeByID 根据ID和类型分析目标（compile_dynamic集成使用）
// ctx: 上下文
// targetID: 目标ID（任务/模块/项目ID）
// analysisType: 分析类型 (task/module/project)
// opts: 分析选项（可为nil）
// 返回分析结果和聚合结果
func (s *AIAnalysisService) AnalyzeByID(ctx context.Context, targetID uint, analysisType string, opts *AnalysisOptions) (*AnalysisResult, *AggregatedResult, error) {
	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	s.logger.Info("根据ID分析", map[string]interface{}{
		"target_id":     targetID,
		"analysis_type": analysisType,
	})

	var result *AnalysisResult
	var err error

	switch analysisType {
	case QuestionTypeTask:
		result, err = s.AnalyzeTask(ctx, targetID, opts)
	case QuestionTypeModule:
		result, err = s.AnalyzeModule(ctx, targetID, opts)
	case QuestionTypeProject:
		result, err = s.AnalyzeProject(ctx, targetID, opts)
	default:
		return nil, nil, NewParamErrorf("不支持的分析类型: %s", analysisType)
	}

	if err != nil {
		return nil, nil, err
	}

	// 聚合单个结果
	aggregated := s.aggregator.Aggregate([]*AnalysisResult{result})

	return result, aggregated, nil
}

// AnalyzeMultiple 分析多个目标（compile_dynamic集成使用）
// ctx: 上下文
// targetIDs: 目标ID列表
// analysisType: 分析类型 (task/module/project)
// opts: 分析选项（可为nil）
// 返回分析结果列表和聚合结果
func (s *AIAnalysisService) AnalyzeMultiple(ctx context.Context, targetIDs []uint, analysisType string, opts *AnalysisOptions) ([]*AnalysisResult, *AggregatedResult, error) {
	if len(targetIDs) == 0 {
		return nil, nil, NewParamError("目标ID列表不能为空")
	}

	if opts == nil {
		opts = DefaultAnalysisOptions()
	}

	s.logger.Info("分析多个目标", map[string]interface{}{
		"target_count":  len(targetIDs),
		"analysis_type": analysisType,
	})

	results := make([]*AnalysisResult, 0, len(targetIDs))
	var lastErr error

	for _, targetID := range targetIDs {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return results, s.aggregator.Aggregate(results), NewTimeoutError("批量分析", int64(0))
		default:
		}

		var result *AnalysisResult
		var err error

		switch analysisType {
		case QuestionTypeTask:
			result, err = s.AnalyzeTask(ctx, targetID, opts)
		case QuestionTypeModule:
			result, err = s.AnalyzeModule(ctx, targetID, opts)
		case QuestionTypeProject:
			result, err = s.AnalyzeProject(ctx, targetID, opts)
		default:
			return nil, nil, NewParamErrorf("不支持的分析类型: %s", analysisType)
		}

		if err != nil {
			lastErr = err
			s.logger.Error("分析目标失败", err, map[string]interface{}{
				"target_id":   targetID,
				"target_type": analysisType,
			})
		} else {
			results = append(results, result)
		}
	}

	// 聚合结果
	aggregated := s.aggregator.Aggregate(results)

	// 如果有任何结果，返回成功；全部失败则返回最后的错误
	if len(results) == 0 && lastErr != nil {
		return nil, nil, lastErr
	}

	return results, aggregated, nil
}
