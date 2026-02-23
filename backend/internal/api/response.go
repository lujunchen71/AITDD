package api

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// 错误码常量
const (
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeVersionConflict     = "VERSION_CONFLICT"
	ErrCodeLocked              = "LOCKED"
	ErrCodeValidationError     = "VALIDATION_ERROR"
	ErrCodeCircularDependency  = "CIRCULAR_DEPENDENCY"
	ErrCodeInternalError       = "INTERNAL_ERROR"
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code string, message string, details ...map[string]interface{}) {
	var detailMap map[string]interface{}
	if len(details) > 0 {
		detailMap = details[0]
	}

	c.JSON(httpStatus, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: detailMap,
		},
		Timestamp: time.Now().UnixMilli(),
	})
}

// NotFound 404响应
func NotFound(c *gin.Context, message string) {
	Error(c, 404, ErrCodeNotFound, message)
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, message string, details map[string]interface{}) {
	Error(c, 400, ErrCodeValidationError, message, details)
}

// VersionConflict 版本冲突响应
func VersionConflict(c *gin.Context, message string) {
	Error(c, 409, ErrCodeVersionConflict, message)
}

// Locked 锁定响应
func Locked(c *gin.Context, message string, details map[string]interface{}) {
	Error(c, 423, ErrCodeLocked, message, details)
}

// InternalError 内部错误响应
func InternalError(c *gin.Context, message string) {
	Error(c, 500, ErrCodeInternalError, message)
}
