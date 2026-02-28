package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 错误码常量
const (
	ErrCodeParamError    = "PARAM_ERROR"
	ErrCodeNetworkError  = "NETWORK_ERROR"
	ErrCodeBusinessError = "BUSINESS_ERROR"
	ErrCodeNotFoundError = "NOT_FOUND_ERROR"
	ErrCodeAuthError     = "AUTH_ERROR"
	ErrCodeTimeoutError  = "TIMEOUT_ERROR"
	ErrCodeInternalError = "INTERNAL_ERROR"
)

// MCPError 统一的 MCP 错误结构
type MCPError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Context    map[string]interface{} `json:"context,omitempty"`
	StatusCode int                    `json:"statusCode"`
	Cause      error                  `json:"-"`
}

// Error 实现 error 接口
func (e *MCPError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s", e.Code, e.Message))
	if len(e.Context) > 0 {
		sb.WriteString(fmt.Sprintf(" | context: %v", e.Context))
	}
	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" | cause: %v", e.Cause))
	}
	return sb.String()
}

// Unwrap 支持错误解包
func (e *MCPError) Unwrap() error {
	return e.Cause
}

// ToYAML 将错误转换为 YAML 格式字符串（用于 MCP 响应）
func (e *MCPError) ToYAML() string {
	var sb strings.Builder
	sb.WriteString("error:\n")
	sb.WriteString(fmt.Sprintf("  code: %s\n", e.Code))
	sb.WriteString(fmt.Sprintf("  message: %s\n", e.Message))
	if len(e.Context) > 0 {
		sb.WriteString("  context:\n")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
		}
	}
	return sb.String()
}

// ToJSON 将错误转换为 JSON 格式字符串
func (e *MCPError) ToJSON() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf(`{"code":"%s","message":"%s"}`, e.Code, e.Message)
	}
	return string(data)
}

// ==================== 错误构造函数 ====================

// NewParamError 创建参数错误
func NewParamError(msg string, ctx ...map[string]interface{}) *MCPError {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	return &MCPError{
		Code:       ErrCodeParamError,
		Message:    msg,
		Context:    context,
		StatusCode: 400,
	}
}

// NewParamErrorf 创建带格式的参数错误
func NewParamErrorf(format string, args ...interface{}) *MCPError {
	return &MCPError{
		Code:       ErrCodeParamError,
		Message:    fmt.Sprintf(format, args...),
		StatusCode: 400,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(msg string, err error) *MCPError {
	return &MCPError{
		Code:       ErrCodeNetworkError,
		Message:    msg,
		Cause:      err,
		StatusCode: 503,
		Context: map[string]interface{}{
			"originalError": err.Error(),
		},
	}
}

// NewNetworkErrorf 创建带格式的网络错误
func NewNetworkErrorf(err error, format string, args ...interface{}) *MCPError {
	return &MCPError{
		Code:       ErrCodeNetworkError,
		Message:    fmt.Sprintf(format, args...),
		Cause:      err,
		StatusCode: 503,
	}
}

// NewBusinessError 创建业务错误
func NewBusinessError(msg string, ctx ...map[string]interface{}) *MCPError {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	return &MCPError{
		Code:       ErrCodeBusinessError,
		Message:    msg,
		Context:    context,
		StatusCode: 422,
	}
}

// NewBusinessErrorf 创建带格式的业务错误
func NewBusinessErrorf(format string, args ...interface{}) *MCPError {
	return &MCPError{
		Code:       ErrCodeBusinessError,
		Message:    fmt.Sprintf(format, args...),
		StatusCode: 422,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(resourceType string, identifier string) *MCPError {
	return &MCPError{
		Code:       ErrCodeNotFoundError,
		Message:    fmt.Sprintf("%s not found", resourceType),
		StatusCode: 404,
		Context: map[string]interface{}{
			"resourceType": resourceType,
			"identifier":   identifier,
		},
	}
}

// NewAuthError 创建认证错误
func NewAuthError(msg string) *MCPError {
	return &MCPError{
		Code:       ErrCodeAuthError,
		Message:    msg,
		StatusCode: 401,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(operation string, timeoutMs int64) *MCPError {
	return &MCPError{
		Code:       ErrCodeTimeoutError,
		Message:    fmt.Sprintf("operation '%s' timed out", operation),
		StatusCode: 408,
		Context: map[string]interface{}{
			"operation": operation,
			"timeoutMs": timeoutMs,
		},
	}
}

// NewInternalError 创建内部错误
func NewInternalError(msg string, err error) *MCPError {
	return &MCPError{
		Code:       ErrCodeInternalError,
		Message:    msg,
		Cause:      err,
		StatusCode: 500,
		Context: map[string]interface{}{
			"originalError": err.Error(),
		},
	}
}

// ==================== 错误检查函数 ====================

// IsParamError 检查是否为参数错误
func IsParamError(err error) bool {
	if e, ok := err.(*MCPError); ok {
		return e.Code == ErrCodeParamError
	}
	return false
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	if e, ok := err.(*MCPError); ok {
		return e.Code == ErrCodeNetworkError
	}
	return false
}

// IsBusinessError 检查是否为业务错误
func IsBusinessError(err error) bool {
	if e, ok := err.(*MCPError); ok {
		return e.Code == ErrCodeBusinessError
	}
	return false
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	if e, ok := err.(*MCPError); ok {
		return e.Code == ErrCodeNotFoundError
	}
	return false
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if e, ok := err.(*MCPError); ok {
		return e.Code == ErrCodeTimeoutError
	}
	return false
}

// ==================== 响应辅助函数 ====================

// ErrorResponse 创建错误响应字符串
func ErrorResponse(err error) string {
	if mcpErr, ok := err.(*MCPError); ok {
		return mcpErr.ToYAML()
	}
	// 普通错误转换为内部错误
	return NewInternalError(err.Error(), err).ToYAML()
}

// ErrorResponseWithMsg 创建带自定义消息的错误响应
func ErrorResponseWithMsg(code, msg string, ctx map[string]interface{}) string {
	err := &MCPError{
		Code:       code,
		Message:    msg,
		Context:    ctx,
		StatusCode: 500,
	}
	return err.ToYAML()
}

// SuccessResponse 创建成功响应
func SuccessResponse(data interface{}) string {
	switch v := data.(type) {
	case string:
		return v
	case map[string]interface{}:
		return mapToYAML(v)
	default:
		// 尝试 JSON 序列化
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Sprintf("result: %v", data)
		}
		return fmt.Sprintf("result: %s", string(jsonBytes))
	}
}

// mapToYAML 将 map 转换为简单的 YAML 格式
func mapToYAML(m map[string]interface{}) string {
	var sb strings.Builder
	for k, v := range m {
		switch val := v.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, val))
		case int, int64, float64:
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, val))
		case bool:
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, val))
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("%s:\n", k))
			for sk, sv := range val {
				sb.WriteString(fmt.Sprintf("  %s: %v\n", sk, sv))
			}
		default:
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, val))
		}
	}
	return sb.String()
}

// ==================== 批量操作错误处理 ====================

// BatchResult 批量操作结果
type BatchResult struct {
	Success []string       `json:"success"`
	Failed  []BatchFailure `json:"failed,omitempty"`
	Stats   BatchStats     `json:"stats"`
}

// BatchFailure 批量操作失败项
type BatchFailure struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Code   string `json:"code"`
}

// BatchStats 批量操作统计
type BatchStats struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// NewBatchResult 创建批量操作结果
func NewBatchResult(total int) *BatchResult {
	return &BatchResult{
		Success: make([]string, 0),
		Failed:  make([]BatchFailure, 0),
		Stats: BatchStats{
			Total:   total,
			Success: 0,
			Failed:  0,
		},
	}
}

// AddSuccess 添加成功项
func (r *BatchResult) AddSuccess(id string) {
	r.Success = append(r.Success, id)
	r.Stats.Success++
}

// AddFailure 添加失败项
func (r *BatchResult) AddFailure(id string, err error) {
	failure := BatchFailure{
		ID:     id,
		Reason: err.Error(),
	}
	if mcpErr, ok := err.(*MCPError); ok {
		failure.Code = mcpErr.Code
	} else {
		failure.Code = ErrCodeInternalError
	}
	r.Failed = append(r.Failed, failure)
	r.Stats.Failed++
}

// ToYAML 转换为 YAML 格式
func (r *BatchResult) ToYAML() string {
	var sb strings.Builder
	sb.WriteString("batch_result:\n")
	sb.WriteString(fmt.Sprintf("  total: %d\n", r.Stats.Total))
	sb.WriteString(fmt.Sprintf("  success: %d\n", r.Stats.Success))
	sb.WriteString(fmt.Sprintf("  failed: %d\n", r.Stats.Failed))

	if len(r.Success) > 0 {
		sb.WriteString("  succeeded_items:\n")
		for _, id := range r.Success {
			sb.WriteString(fmt.Sprintf("    - %s\n", id))
		}
	}

	if len(r.Failed) > 0 {
		sb.WriteString("  failed_items:\n")
		for _, f := range r.Failed {
			sb.WriteString(fmt.Sprintf("    - id: %s\n", f.ID))
			sb.WriteString(fmt.Sprintf("      code: %s\n", f.Code))
			sb.WriteString(fmt.Sprintf("      reason: %s\n", f.Reason))
		}
	}

	return sb.String()
}

// IsPartialSuccess 是否部分成功
func (r *BatchResult) IsPartialSuccess() bool {
	return r.Stats.Success > 0 && r.Stats.Failed > 0
}

// IsCompleteSuccess 是否完全成功
func (r *BatchResult) IsCompleteSuccess() bool {
	return r.Stats.Failed == 0
}

// IsCompleteFailure 是否完全失败
func (r *BatchResult) IsCompleteFailure() bool {
	return r.Stats.Success == 0
}
