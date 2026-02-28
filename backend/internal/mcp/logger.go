package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var logLevelNames = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
}

// MCPLogger MCP 日志记录器
type MCPLogger struct {
	mu          sync.Mutex
	level       LogLevel
	jsonFormat  bool
	requestID   string
	performance bool
	logger      *log.Logger
}

// 默认日志实例
var defaultLogger *MCPLogger
var loggerOnce sync.Once

// GetMCPLogger 获取默认日志实例
func GetMCPLogger() *MCPLogger {
	loggerOnce.Do(func() {
		defaultLogger = NewMCPLogger(LogLevelInfo, true, true)
	})
	return defaultLogger
}

// NewMCPLogger 创建新的日志实例
func NewMCPLogger(level LogLevel, jsonFormat bool, performance bool) *MCPLogger {
	return &MCPLogger{
		level:       level,
		jsonFormat:  jsonFormat,
		performance: performance,
		logger:      log.New(os.Stdout, "", 0),
	}
}

// SetLevel 设置日志级别
func (l *MCPLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetRequestID 设置请求 ID
func (l *MCPLogger) SetRequestID(requestID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.requestID = requestID
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"requestId,omitempty"`
	Duration  int64                  `json:"durationMs,omitempty"`
	Tool      string                 `json:"tool,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// log 内部日志方法
func (l *MCPLogger) log(level LogLevel, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     logLevelNames[level],
		Message:   message,
		RequestID: l.requestID,
		Fields:    fields,
	}

	if l.jsonFormat {
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			l.logger.Printf("[%s] %s", entry.Level, entry.Message)
			return
		}
		l.logger.Println(string(jsonBytes))
	} else {
		var prefix string
		if l.requestID != "" {
			prefix = fmt.Sprintf("[%s][%s] ", entry.Level, l.requestID)
		} else {
			prefix = fmt.Sprintf("[%s] ", entry.Level)
		}

		var fieldsStr string
		if len(fields) > 0 {
			fieldsStr = fmt.Sprintf(" | %v", fields)
		}

		l.logger.Printf("%s%s%s", prefix, message, fieldsStr)
	}
}

// Debug 调试日志
func (l *MCPLogger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelDebug, message, f)
}

// Info 信息日志
func (l *MCPLogger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelInfo, message, f)
}

// Warn 警告日志
func (l *MCPLogger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelWarn, message, f)
}

// Error 错误日志
func (l *MCPLogger) Error(message string, err error, fields ...map[string]interface{}) {
	f := make(map[string]interface{})
	if len(fields) > 0 {
		f = fields[0]
	}
	if err != nil {
		f["error"] = err.Error()
	}
	l.log(LogLevelError, message, f)
}

// ==================== 性能监控 ====================

// PerformanceTimer 性能计时器
type PerformanceTimer struct {
	logger   *MCPLogger
	tool     string
	start    time.Time
	ended    bool
}

// StartTimer 开始计时
func (l *MCPLogger) StartTimer(tool string) *PerformanceTimer {
	return &PerformanceTimer{
		logger: l,
		tool:   tool,
		start:  time.Now(),
		ended:  false,
	}
}

// End 结束计时并记录
func (t *PerformanceTimer) End() {
	if t.ended {
		return
	}
	t.ended = true

	if !t.logger.performance {
		return
	}

	duration := time.Since(t.start).Milliseconds()

	t.logger.mu.Lock()
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     "PERF",
		Message:   fmt.Sprintf("Tool '%s' completed", t.tool),
		RequestID: t.logger.requestID,
		Duration:  duration,
		Tool:      t.tool,
	}
	t.logger.mu.Unlock()

	if t.logger.jsonFormat {
		jsonBytes, _ := json.Marshal(entry)
		t.logger.logger.Println(string(jsonBytes))
	} else {
		var prefix string
		if entry.RequestID != "" {
			prefix = fmt.Sprintf("[PERF][%s] ", entry.RequestID)
		} else {
			prefix = "[PERF] "
		}
		t.logger.logger.Printf("%s%s took %dms", prefix, t.tool, duration)
	}
}

// EndWithFields 结束计时并记录带额外字段
func (t *PerformanceTimer) EndWithFields(fields map[string]interface{}) {
	if t.ended {
		return
	}
	t.ended = true

	if !t.logger.performance {
		return
	}

	duration := time.Since(t.start).Milliseconds()

	t.logger.mu.Lock()
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     "PERF",
		Message:   fmt.Sprintf("Tool '%s' completed", t.tool),
		RequestID: t.logger.requestID,
		Duration:  duration,
		Tool:      t.tool,
		Fields:    fields,
	}
	t.logger.mu.Unlock()

	if t.logger.jsonFormat {
		jsonBytes, _ := json.Marshal(entry)
		t.logger.logger.Println(string(jsonBytes))
	} else {
		var prefix string
		if entry.RequestID != "" {
			prefix = fmt.Sprintf("[PERF][%s] ", entry.RequestID)
		} else {
			prefix = "[PERF] "
		}
		t.logger.logger.Printf("%s%s took %dms | %v", prefix, t.tool, duration, fields)
	}
}

// ==================== 便捷函数 ====================

// Debug 调试日志（使用默认 logger）
func Debug(message string, fields ...map[string]interface{}) {
	GetMCPLogger().Debug(message, fields...)
}

// Info 信息日志（使用默认 logger）
func Info(message string, fields ...map[string]interface{}) {
	GetMCPLogger().Info(message, fields...)
}

// Warn 警告日志（使用默认 logger）
func Warn(message string, fields ...map[string]interface{}) {
	GetMCPLogger().Warn(message, fields...)
}

// Error 错误日志（使用默认 logger）
func Error(message string, err error, fields ...map[string]interface{}) {
	GetMCPLogger().Error(message, err, fields...)
}

// StartTimer 开始计时（使用默认 logger）
func StartTimer(tool string) *PerformanceTimer {
	return GetMCPLogger().StartTimer(tool)
}

// ==================== 请求追踪 ====================

// WithRequestID 创建带请求 ID 的日志上下文
func WithRequestID(requestID string) *MCPLogger {
	logger := GetMCPLogger()
	logger.SetRequestID(requestID)
	return logger
}

// GenerateRequestID 生成简单的请求 ID
func GenerateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// ==================== 工具调用日志 ====================

// LogToolCall 记录工具调用
func LogToolCall(tool string, args map[string]interface{}) *PerformanceTimer {
	logger := GetMCPLogger()
	logger.Debug(fmt.Sprintf("Tool '%s' called", tool), args)
	return logger.StartTimer(tool)
}

// LogToolResult 记录工具结果
func LogToolResult(tool string, success bool, duration int64, err error) {
	logger := GetMCPLogger()
	fields := map[string]interface{}{
		"success":  success,
		"duration": duration,
	}
	if err != nil {
		fields["error"] = err.Error()
		logger.Error(fmt.Sprintf("Tool '%s' failed", tool), err, fields)
	} else {
		logger.Info(fmt.Sprintf("Tool '%s' succeeded in %dms", tool, duration), fields)
	}
}

// LogCacheHit 记录缓存命中
func LogCacheHit(cacheType, key string) {
	GetMCPLogger().Debug(fmt.Sprintf("Cache hit: %s/%s", cacheType, key))
}

// LogCacheMiss 记录缓存未命中
func LogCacheMiss(cacheType, key string) {
	GetMCPLogger().Debug(fmt.Sprintf("Cache miss: %s/%s", cacheType, key))
}

// LogAPICall 记录 API 调用
func LogAPICall(method, url string, statusCode int, duration int64) {
	GetMCPLogger().Debug(fmt.Sprintf("API %s %s -> %d (%dms)", method, url, statusCode, duration), map[string]interface{}{
		"method":     method,
		"url":        url,
		"statusCode": statusCode,
		"duration":   duration,
	})
}
