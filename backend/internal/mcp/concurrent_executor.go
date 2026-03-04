package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ConcurrentExecutor 并发执行器
// 用于控制AI请求的并发执行，实现信号量控制和速率限制
type ConcurrentExecutor struct {
	config      *SubAgentConfig
	aiClient    *AIClient
	semaphore   chan struct{}     // 信号量控制并发数
	rateLimiter  *rate.Limiter     // 速率限制器
	logger       *MCPLogger
	mu           sync.RWMutex
}

// ExecutorTask 执行任务
// 包含任务的基本信息和AI请求
type ExecutorTask struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`        // task, module, project
	TargetID    uint        `json:"target_id"`
	TargetName  string      `json:"target_name"`
	Request     *AIRequest  `json:"request"`
}

// ExecutorResult 执行结果
// 包含任务执行的结果信息
type ExecutorResult struct {
	TaskID      string        `json:"task_id"`
	Success     bool          `json:"success"`
	Response    *AIResponse  `json:"response,omitempty"`
	Result       *AnalysisResult `json:"result,omitempty"`
	Error       error         `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	RetryCount  int           `json:"retry_count"`
}

// ExecutorStats 执行器统计信息
type ExecutorStats struct {
	TotalTasks     int           `json:"total_tasks"`
	CompletedTasks int           `json:"completed_tasks"`
	FailedTasks    int           `json:"failed_tasks"`
	TotalDuration  time.Duration `json:"total_duration"`
	AvgDuration    time.Duration `json:"avg_duration"`
}

// NewConcurrentExecutor 创建新的并发执行器
// config: 子代理配置，包含并发数和速率限制配置
// aiClient: AI客户端，用于调用AI API
func NewConcurrentExecutor(config *SubAgentConfig, aiClient *AIClient) *ConcurrentExecutor {
	if config == nil {
		config = GetDefaultSubAgentConfig()
	}

	if aiClient == nil {
		aiClient = NewAIClient(config)
	}

	// 创建信号量，控制最大并发数
	maxConcurrent := config.RateLimit.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 3 // 默认并发数
	}

	// 创建速率限制器
	requestsPerMin := config.RateLimit.RequestsPerMin
	if requestsPerMin <= 0 {
		requestsPerMin = 20 // 默认每分钟请求数
	}
	// 将每分钟请求数转换为每秒速率
	ratePerSecond := float64(requestsPerMin) / 60.0

	return &ConcurrentExecutor{
		config:     config,
		aiClient:   aiClient,
		semaphore:  make(chan struct{}, maxConcurrent),
		rateLimiter: rate.NewLimiter(rate.Limit(ratePerSecond), maxConcurrent),
		logger:     GetMCPLogger(),
	}
}

// NewConcurrentExecutorWithLogger 创建带自定义日志的并发执行器
func NewConcurrentExecutorWithLogger(config *SubAgentConfig, aiClient *AIClient, logger *MCPLogger) *ConcurrentExecutor {
	executor := NewConcurrentExecutor(config, aiClient)
	executor.logger = logger
	return executor
}

// Execute 执行多个任务
// 使用并发控制执行多个AI请求任务
// 返回所有任务的执行结果
func (e *ConcurrentExecutor) Execute(ctx context.Context, tasks []*ExecutorTask) []*ExecutorResult {
	if len(tasks) == 0 {
		return []*ExecutorResult{}
	}

	e.logger.Info("开始并发执行任务", map[string]interface{}{
		"task_count":     len(tasks),
		"max_concurrent": e.config.RateLimit.MaxConcurrent,
		"rate_limit":     e.config.RateLimit.RequestsPerMin,
	})

	startTime := time.Now()
	results := make([]*ExecutorResult, len(tasks))
	var wg sync.WaitGroup

	// 为每个任务启动goroutine
	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t *ExecutorTask) {
			defer wg.Done()
			results[idx] = e.ExecuteSingle(ctx, t)
		}(i, task)
	}

	// 等待所有任务完成
	wg.Wait()

	// 计算统计信息
	totalDuration := time.Since(startTime)
	stats := e.calculateStats(results, totalDuration)

	e.logger.Info("并发执行完成", map[string]interface{}{
		"total_duration":   totalDuration.String(),
		"completed_tasks":  stats.CompletedTasks,
		"failed_tasks":     stats.FailedTasks,
		"avg_duration":     stats.AvgDuration.String(),
	})

	return results
}

// ExecuteSingle 执行单个任务
// 包含信号量获取、速率限制等待和AI调用
func (e *ConcurrentExecutor) ExecuteSingle(ctx context.Context, task *ExecutorTask) *ExecutorResult {
	if task == nil {
		return &ExecutorResult{
			Success: false,
			Error:   NewParamError("任务不能为空"),
		}
	}

	startTime := time.Now()
	result := &ExecutorResult{
		TaskID: task.ID,
	}

	e.logger.Debug("开始执行任务", map[string]interface{}{
		"task_id":   task.ID,
		"task_type": task.Type,
		"target_id": task.TargetID,
	})

	// 获取信号量
	if err := e.acquireSemaphore(ctx); err != nil {
		result.Success = false
		result.Error = err
		result.Duration = time.Since(startTime)
		return result
	}
	defer e.releaseSemaphore()

	// 等待速率限制
	if err := e.waitForRateLimit(ctx); err != nil {
		result.Success = false
		result.Error = err
		result.Duration = time.Since(startTime)
		return result
	}

	// 执行AI调用
	response, err := e.executeWithRetry(ctx, task)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = err
		e.logger.Error("任务执行失败", err, map[string]interface{}{
			"task_id":   task.ID,
			"duration":  result.Duration.String(),
		})
	} else {
		result.Success = true
		result.Response = response
		e.logger.Debug("任务执行成功", map[string]interface{}{
			"task_id":   task.ID,
			"duration":  result.Duration.String(),
		})
	}

	return result
}

// ExecuteWithContext 执行任务并支持结果回调
// 使用回调函数处理每个完成的任务结果
func (e *ConcurrentExecutor) ExecuteWithContext(ctx context.Context, tasks []*ExecutorTask, callback func(*ExecutorResult)) {
	if len(tasks) == 0 {
		return
	}

	e.logger.Info("开始并发执行任务（带回调）", map[string]interface{}{
		"task_count": len(tasks),
	})

	var wg sync.WaitGroup
	resultChan := make(chan *ExecutorResult, len(tasks))

	// 启动任务
	for _, task := range tasks {
		wg.Add(1)
		go func(t *ExecutorTask) {
			defer wg.Done()
			result := e.ExecuteSingle(ctx, t)
			resultChan <- result
		}(task)
	}

	// 启动结果处理goroutine
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 处理结果
	for result := range resultChan {
		if callback != nil {
			callback(result)
		}
	}
}

// executeWithRetry 带重试的执行
func (e *ConcurrentExecutor) executeWithRetry(ctx context.Context, task *ExecutorTask) (*AIResponse, error) {
	if task.Request == nil {
		return nil, NewParamError("任务请求不能为空")
	}

	maxRetries := e.config.Retry.MaxRetries
	backoffMs := e.config.Retry.BackoffMs

	var lastErr error
	var retryCount int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, NewTimeoutError("任务执行", 0)
		default:
		}

		// 执行AI调用
		response, err := e.aiClient.Call(ctx, task.Request)
		if err == nil {
			return response, nil
		}

		lastErr = err
		retryCount = attempt

		// 如果是参数错误，不重试
		if IsParamError(err) {
			return nil, err
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxRetries {
			waitTime := time.Duration(backoffMs*(1<<attempt)) * time.Millisecond
			e.logger.Debug("任务执行失败，准备重试", map[string]interface{}{
				"task_id":     task.ID,
				"attempt":     attempt + 1,
				"max_retries": maxRetries,
				"wait_time":   waitTime.String(),
				"error":       err.Error(),
			})

			select {
			case <-ctx.Done():
				return nil, NewTimeoutError("任务执行", 0)
			case <-time.After(waitTime):
				continue
			}
		}
	}

	return nil, NewBusinessErrorf("任务执行失败，重试 %d 次后仍失败: %v", retryCount, lastErr)
}

// acquireSemaphore 获取信号量
// 如果超过最大并发数，将阻塞等待
func (e *ConcurrentExecutor) acquireSemaphore(ctx context.Context) error {
	select {
	case e.semaphore <- struct{}{}:
		e.logger.Debug("获取信号量成功")
		return nil
	case <-ctx.Done():
		return NewTimeoutError("等待信号量", 0)
	}
}

// releaseSemaphore 释放信号量
func (e *ConcurrentExecutor) releaseSemaphore() {
	<-e.semaphore
	e.logger.Debug("释放信号量")
}

// waitForRateLimit 等待速率限制
// 使用令牌桶算法控制请求速率
func (e *ConcurrentExecutor) waitForRateLimit(ctx context.Context) error {
	err := e.rateLimiter.Wait(ctx)
	if err != nil {
		return NewTimeoutError("等待速率限制", 0)
	}
	e.logger.Debug("速率限制检查通过")
	return nil
}

// calculateStats 计算执行统计信息
func (e *ConcurrentExecutor) calculateStats(results []*ExecutorResult, totalDuration time.Duration) *ExecutorStats {
	stats := &ExecutorStats{
		TotalTasks:    len(results),
		TotalDuration: totalDuration,
	}

	var totalTaskDuration time.Duration

	for _, result := range results {
		if result.Success {
			stats.CompletedTasks++
		} else {
			stats.FailedTasks++
		}
		totalTaskDuration += result.Duration
	}

	if len(results) > 0 {
		stats.AvgDuration = totalTaskDuration / time.Duration(len(results))
	}

	return stats
}

// GetStats 获取执行器当前统计信息
func (e *ConcurrentExecutor) GetStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"max_concurrent":   e.config.RateLimit.MaxConcurrent,
		"requests_per_min": e.config.RateLimit.RequestsPerMin,
		"max_retries":      e.config.Retry.MaxRetries,
		"current_running":  len(e.semaphore),
	}
}

// SetRateLimit 动态设置速率限制
func (e *ConcurrentExecutor) SetRateLimit(requestsPerMin int) error {
	if requestsPerMin <= 0 {
		return NewParamError("每分钟请求数必须大于0")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	ratePerSecond := float64(requestsPerMin) / 60.0
	e.rateLimiter.SetLimit(rate.Limit(ratePerSecond))

	e.logger.Info("更新速率限制", map[string]interface{}{
		"requests_per_min": requestsPerMin,
	})

	return nil
}

// SetMaxConcurrent 动态设置最大并发数
// 注意：此操作会重新创建信号量通道
func (e *ConcurrentExecutor) SetMaxConcurrent(maxConcurrent int) error {
	if maxConcurrent <= 0 {
		return NewParamError("最大并发数必须大于0")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 创建新的信号量通道
	e.semaphore = make(chan struct{}, maxConcurrent)
	e.rateLimiter.SetBurst(maxConcurrent)

	e.logger.Info("更新最大并发数", map[string]interface{}{
		"max_concurrent": maxConcurrent,
	})

	return nil
}

// Shutdown 关闭执行器
// 等待所有正在执行的任务完成
func (e *ConcurrentExecutor) Shutdown(ctx context.Context) error {
	e.logger.Info("关闭并发执行器")

	// 等待所有任务完成（信号量通道为空）
	for {
		select {
		case <-ctx.Done():
			return NewTimeoutError("关闭执行器", 0)
		default:
			if len(e.semaphore) == 0 {
				e.logger.Info("并发执行器已关闭")
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// CreateTask 创建执行任务的辅助函数
func CreateTask(taskType string, targetID uint, targetName string, request *AIRequest) *ExecutorTask {
	return &ExecutorTask{
		ID:         fmt.Sprintf("%s_%d_%d", taskType, targetID, time.Now().UnixNano()),
		Type:       taskType,
		TargetID:   targetID,
		TargetName: targetName,
		Request:    request,
	}
}

// CreateTaskFromAnalysis 从分析请求创建执行任务
func CreateTaskFromAnalysis(analysisReq *AnalysisRequest) *ExecutorTask {
	if analysisReq == nil {
		return nil
	}

	targetID := uint(0)
	targetName := ""
	if analysisReq.Metadata != nil {
		if id, ok := analysisReq.Metadata["task_id"].(float64); ok {
			targetID = uint(id)
		}
		if id, ok := analysisReq.Metadata["module_id"].(float64); ok {
			targetID = uint(id)
		}
		if id, ok := analysisReq.Metadata["project_id"].(float64); ok {
			targetID = uint(id)
		}
		if name, ok := analysisReq.Metadata["task_name"].(string); ok {
			targetName = name
		}
		if name, ok := analysisReq.Metadata["module_name"].(string); ok {
			targetName = name
		}
		if name, ok := analysisReq.Metadata["project_name"].(string); ok {
			targetName = name
		}
	}

	return &ExecutorTask{
		ID:         GenerateAnalysisID(analysisReq.AnalysisType, targetID),
		Type:       analysisReq.AnalysisType,
		TargetID:   targetID,
		TargetName: targetName,
		Request: &AIRequest{
			Model: analysisReq.Metadata["model"].(string),
			Messages: []Message{
				{Role: "system", Content: analysisReq.SystemPrompt},
				{Role: "user", Content: analysisReq.UserPrompt},
			},
		},
	}
}

// BatchExecuteResults 批量执行结果
type BatchExecuteResults struct {
	Results     []*ExecutorResult `json:"results"`
	Stats       *ExecutorStats    `json:"stats"`
	Duration    time.Duration     `json:"duration"`
	SuccessRate float64           `json:"success_rate"`
}

// ToJSON 转换为JSON字符串
func (r *ExecutorResult) ToJSON() (string, error) {
	return toJSON(r)
}

// ToJSON 转换为JSON字符串
func (s *ExecutorStats) ToJSON() (string, error) {
	return toJSON(s)
}

// toJSON 辅助函数，将对象转换为JSON
func toJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
