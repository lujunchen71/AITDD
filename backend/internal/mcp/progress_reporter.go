package mcp

import (
	"encoding/json"
	"sync"
	"time"
)

// ProgressReporter 进度报告器
// 用于实时报告分析进度，支持线程安全的状态管理
type ProgressReporter struct {
	analysisID   string
	total        int
	completed    int
	status       string              // pending, running, completed, failed
	message      string
	results      []*PartialResult
	startTime    time.Time
	logger       *MCPLogger
	mu           sync.RWMutex
}

// ProgressStatus 进度状态
// 包含完整的进度信息，用于JSON序列化
type ProgressStatus struct {
	AnalysisID   string           `json:"analysis_id"`
	Status       string           `json:"status"`
	Total        int              `json:"total"`
	Completed    int              `json:"completed"`
	Percentage   int              `json:"percentage"`
	Message      string           `json:"message"`
	ElapsedTime  time.Duration      `json:"elapsed_time"`
	Results      []*PartialResult `json:"results,omitempty"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// PartialResult 部分结果
// 用于记录单个任务的分析结果
type PartialResult struct {
	TaskID    string `json:"task_id"`
	Type      string `json:"type"`
	TargetID  uint   `json:"target_id"`
	Status    string `json:"status"` // success, failed, skipped
	Message   string `json:"message"`
	Duration  time.Duration `json:"duration,omitempty"`
	Score     int       `json:"score,omitempty"`
}

// 进度状态常量
const (
	ProgressStatusPending   = "pending"
	ProgressStatusRunning    = "running"
	ProgressStatusCompleted = "completed"
	ProgressStatusFailed    = "failed"
)

// 部分结果状态常量
const (
	PartialResultSuccess  = "success"
	PartialResultFailed   = "failed"
	PartialResultSkipped   = "skipped"
)

// NewProgressReporter 创建新的进度报告器
// analysisID: 分析ID，// total: 总任务数
func NewProgressReporter(analysisID string, total int) *ProgressReporter {
	return &ProgressReporter{
		analysisID: analysisID,
		total:      total,
		completed: 0,
		status:      ProgressStatusPending,
		message:    "等待开始",
		results:     make([]*PartialResult, 0),
		startTime:  time.Now(),
		logger:     GetMCPLogger(),
	}
}

// NewProgressReporterWithLogger 创建带自定义日志的进度报告器
func NewProgressReporterWithLogger(analysisID string, total int, logger *MCPLogger) *ProgressReporter {
	return &ProgressReporter{
		analysisID: analysisID,
		total:      total,
		completed: 0,
		status:      ProgressStatusPending,
		message:    "等待开始",
		results:     make([]*PartialResult, 0),
		startTime:  time.Now(),
		logger:     logger,
	}
}

// Start 开始分析
// 将状态设置为running
func (r *ProgressReporter) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = ProgressStatusRunning
	r.message = "分析进行中"
	r.startTime = time.Now()

	r.logger.Info("分析开始", map[string]interface{}{
		"analysis_id": r.analysisID,
		"total_tasks": r.total,
	})
}

// Complete 完成分析
// 将状态设置为completed
func (r *ProgressReporter) Complete() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = ProgressStatusCompleted
	r.message = "分析完成"
	r.completed = r.total

	r.logger.Info("分析完成", map[string]interface{}{
		"analysis_id":   r.analysisID,
		"total_tasks":    r.total,
		"completed":    r.completed,
		"elapsed_time":  time.Since(r.startTime).String(),
	})
}

// Fail 分析失败
// 将状态设置为failed并记录失败消息
func (r *ProgressReporter) Fail(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = ProgressStatusFailed
	r.message = message

	r.logger.Error("分析失败", nil, map[string]interface{}{
		"analysis_id": r.analysisID,
		"message":      message,
		"elapsed_time": time.Since(r.startTime).String(),
	})
}

// Increment 增加完成计数
// 记录一个任务的分析结果并增加完成计数
func (r *ProgressReporter) Increment(taskID string, result *PartialResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.completed++
	if r.completed > r.total {
		r.completed = r.total
	}

	if result != nil {
		r.results = append(r.results, result)
	}

	percentage := r.calculatePercentage()

	r.logger.Debug("任务完成", map[string]interface{}{
		"analysis_id": r.analysisID,
		"task_id":     taskID,
		"completed":  r.completed,
		"total":      r.total,
		"percentage": percentage,
	})
}

// GetStatus 获取当前状态
// 返回进度的快照
func (r *ProgressReporter) GetStatus() *ProgressStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	elapsed := time.Since(r.startTime)
	percentage := r.calculatePercentage()

	return &ProgressStatus{
		AnalysisID:  r.analysisID,
		Status:      r.status,
		Total:       r.total,
		Completed:   r.completed,
		Percentage:  percentage,
		Message:     r.message,
		ElapsedTime: elapsed,
		Results:     r.results,
		UpdatedAt:   time.Now(),
	}
}

// SetMessage 设置状态消息
// 用于更新当前进度消息
func (r *ProgressReporter) SetMessage(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.message = message
	r.logger.Debug("进度消息更新", map[string]interface{}{
		"analysis_id": r.analysisID,
		"message":      message,
	})
}

// IsRunning 检查是否正在运行
func (r *ProgressReporter) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.status == ProgressStatusRunning
}

// IsCompleted 检查是否已完成
func (r *ProgressReporter) IsCompleted() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.status == ProgressStatusCompleted
}

// IsFailed 检查是否失败
func (r *ProgressReporter) IsFailed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.status == ProgressStatusFailed
}

// GetCompletedCount 获取已完成数量
func (r *ProgressReporter) GetCompletedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.completed
}

// GetRemainingCount 获取剩余数量
func (r *ProgressReporter) GetRemainingCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.total - r.completed
}

// GetElapsedTime 获取已用时间
func (r *ProgressReporter) GetElapsedTime() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return time.Since(r.startTime)
}

// GetEstimatedTimeRemaining 估算剩余时间
// 基于已完成任务的平均时间估算剩余时间
func (r *ProgressReporter) GetEstimatedTimeRemaining() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.completed == 0 {
		return 0 // 还没有完成的任务，无法估算
	}

	elapsed := time.Since(r.startTime)
	avgTimePerTask := elapsed / time.Duration(r.completed)
	remaining := avgTimePerTask * time.Duration(r.total-r.completed)

	return remaining
}

// calculatePercentage 计算完成百分比
func (r *ProgressReporter) calculatePercentage() int {
	if r.total == 0 {
		return 0
	}
	percentage := (r.completed * 100) / r.total
	if percentage > 100 {
		percentage = 100
	}
	return percentage
}

// ToJSON 将进度状态转换为JSON字符串
func (s *ProgressStatus) ToJSON() (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreatePartialResult 创建部分结果的辅助函数
func CreatePartialResult(taskID string, resultType string, targetID uint) *PartialResult {
	return &PartialResult{
		TaskID:   taskID,
		Type:     resultType,
		TargetID: targetID,
		Status:  ProgressStatusPending,
	}
}

// MarkSuccess 标记部分结果为成功
func (r *PartialResult) MarkSuccess(message string, score int) {
	r.Status = PartialResultSuccess
	r.Message = message
	r.Score = score
}

// MarkFailed 标记部分结果为失败
func (r *PartialResult) MarkFailed(message string) {
	r.Status = PartialResultFailed
	r.Message = message
}

// MarkSkipped 标记部分结果为跳过
func (r *PartialResult) MarkSkipped(message string) {
	r.Status = PartialResultSkipped
	r.Message = message
}

// ProgressReporterRegistry 进度报告器注册表
// 用于管理多个进度报告器实例
type ProgressReporterRegistry struct {
	reporters map[string]*ProgressReporter
	mu        sync.RWMutex
}

// 全局进度报告器注册表
var globalProgressRegistry = &ProgressReporterRegistry{
	reporters: make(map[string]*ProgressReporter),
}
var progressRegistryOnce sync.Once

// GetProgressReporterRegistry 获取全局进度报告器注册表
func GetProgressReporterRegistry() *ProgressReporterRegistry {
	progressRegistryOnce.Do(func() {
		globalProgressRegistry = &ProgressReporterRegistry{
			reporters: make(map[string]*ProgressReporter),
		}
	})
	return globalProgressRegistry
}

// Register 注册进度报告器
func (pr *ProgressReporterRegistry) Register(reporter *ProgressReporter) error {
	if reporter == nil {
		return NewParamError("报告器不能为空")
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()

	if _, exists := pr.reporters[reporter.analysisID]; exists {
		return NewBusinessErrorf("分析ID %s 已存在", reporter.analysisID)
	}

	pr.reporters[reporter.analysisID] = reporter
	return nil
}

// Unregister 注销进度报告器
func (pr *ProgressReporterRegistry) Unregister(analysisID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	delete(pr.reporters, analysisID)
}

// Get 获取进度报告器
func (pr *ProgressReporterRegistry) Get(analysisID string) (*ProgressReporter, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	reporter, exists := pr.reporters[analysisID]
	return reporter, exists
}

// GetAll 获取所有进度报告器
func (pr *ProgressReporterRegistry) GetAll() map[string]*ProgressReporter {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	// 返回副本
	result := make(map[string]*ProgressReporter)
	for k, v := range pr.reporters {
		result[k] = v
	}
	return result
}

// Clear 清除所有进度报告器
func (pr *ProgressReporterRegistry) Clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.reporters = make(map[string]*ProgressReporter)
}

// Count 获取进度报告器数量
func (pr *ProgressReporterRegistry) Count() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.reporters)
}

// CreateReporter 创建并注册进度报告器的便捷方法
func CreateReporter(analysisID string, total int) (*ProgressReporter, error) {
	registry := GetProgressReporterRegistry()
	reporter := NewProgressReporter(analysisID, total)
	if err := registry.Register(reporter); err != nil {
		return nil, err
	}
	return reporter, nil
}

// GetReporter 获取进度报告器的便捷方法
func GetReporter(analysisID string) (*ProgressReporter, bool) {
	registry := GetProgressReporterRegistry()
	return registry.Get(analysisID)
}

// RemoveReporter 移除进度报告器的便捷方法
func RemoveReporter(analysisID string) {
	registry := GetProgressReporterRegistry()
	registry.Unregister(analysisID)
}
