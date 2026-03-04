package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIClient AI客户端
type AIClient struct {
	config     *SubAgentConfig
	httpClient *http.Client
	logger     *MCPLogger
}

// AIRequest AI请求结构
type AIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse AI响应结构
type AIResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
	Usage   *Usage `json:"usage,omitempty"`
}

// Usage 令牌使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIResponse OpenAI API响应结构
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// NewAIClient 创建新的AI客户端
func NewAIClient(config *SubAgentConfig) *AIClient {
	if config == nil {
		config = GetDefaultSubAgentConfig()
	}

	// 创建HTTP客户端，配置超时
	httpClient := &http.Client{
		Timeout:   time.Duration(config.Timeout.Request) * time.Second,
		Transport: &http.Transport{},
	}

	return &AIClient{
		config:     config,
		httpClient: httpClient,
		logger:     GetMCPLogger(),
	}
}

// NewAIClientWithLogger 创建带自定义日志的AI客户端
func NewAIClientWithLogger(config *SubAgentConfig, logger *MCPLogger) *AIClient {
	if config == nil {
		config = GetDefaultSubAgentConfig()
	}

	httpClient := &http.Client{
		Timeout:   time.Duration(config.Timeout.Request) * time.Second,
		Transport: &http.Transport{},
	}

	return &AIClient{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Call 调用AI API
func (c *AIClient) Call(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	return c.CallWithModel(ctx, req, "")
}

// CallWithModel 使用指定模型调用AI API
func (c *AIClient) CallWithModel(ctx context.Context, req *AIRequest, modelName string) (*AIResponse, error) {
	if req == nil {
		return nil, NewParamError("请求不能为空")
	}

	// 获取模型配置
	var model *ModelConfig
	var err error

	if modelName != "" {
		model, err = c.getModelByName(modelName)
	} else {
		model, err = GetPrimaryModel(c.config)
	}

	if err != nil {
		return nil, err
	}

	// 执行带重试的调用
	return c.callWithRetry(ctx, req, model)
}

// callWithRetry 带重试逻辑的API调用
func (c *AIClient) callWithRetry(ctx context.Context, req *AIRequest, model *ModelConfig) (*AIResponse, error) {
	var lastErr error
	maxRetries := c.config.Retry.MaxRetries
	backoffMs := c.config.Retry.BackoffMs

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, NewTimeoutError("AI API调用", 1)
		default:
		}

		// 执行调用
		resp, err := c.doCall(ctx, req, model)
		if err == nil {
			return resp, nil
		}

		// 记录错误
		lastErr = err
		c.logger.Warn(fmt.Sprintf("AI API调用失败，尝试 %d/%d", attempt+1, maxRetries+1), map[string]interface{}{
			"error": err.Error(),
			"model": model.Name,
		})

		// 如果是参数错误，不重试
		if IsParamError(err) {
			return nil, err
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxRetries {
			// 指数退避
			waitTime := time.Duration(backoffMs*(1<<attempt)) * time.Millisecond
			c.logger.Debug(fmt.Sprintf("等待 %v 后重试", waitTime))

			select {
			case <-ctx.Done():
				return nil, NewTimeoutError("AI API调用", 1)
			case <-time.After(waitTime):
				continue
			}
		}
	}

	return nil, lastErr
}

// doCall 执行实际的HTTP调用
func (c *AIClient) doCall(ctx context.Context, req *AIRequest, model *ModelConfig) (*AIResponse, error) {
	// 展开API密钥中的环境变量
	apiKey := model.ExpandAPIKey()

	// 构建请求体：优先级：req.Model > model.Model > model.Name
	modelName := model.Name
	if model.Model != "" {
		modelName = model.Model
	}
	if req.Model != "" {
		modelName = req.Model
	}
	requestBody := map[string]interface{}{
		"model":    modelName,
		"messages": req.Messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, NewInternalError("序列化请求失败", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", model.Endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, NewNetworkError("创建请求失败", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 记录请求开始
	timer := c.logger.StartTimer("ai_api_call")
	defer timer.End()

	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, NewNetworkError("发送请求失败", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, NewNetworkError("读取响应失败", err)
	}

	// 检查HTTP状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(httpResp.StatusCode, body)
	}

	// 解析OpenAI响应
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, NewInternalError("解析响应失败", err)
	}

	// 检查API错误
	if openAIResp.Error != nil {
		return nil, NewBusinessErrorf("API错误: %s", openAIResp.Error.Message)
	}

	// 检查是否有选择
	if len(openAIResp.Choices) == 0 {
		return nil, NewBusinessError("API返回空响应")
	}

	// 提取内容
	content := openAIResp.Choices[0].Message.Content

	c.logger.Info("AI API调用成功", map[string]interface{}{
		"model":         model.Name,
		"content_len":   len(content),
		"finish_reason": openAIResp.Choices[0].FinishReason,
	})

	return &AIResponse{
		Content: content,
		Usage:   openAIResp.Usage,
	}, nil
}

// handleHTTPError 处理HTTP错误
func (c *AIClient) handleHTTPError(statusCode int, body []byte) error {
	// 尝试解析错误响应
	var errResp struct {
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	errorMsg := string(body)
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
		errorMsg = errResp.Error.Message
	}

	switch statusCode {
	case 400:
		return NewParamErrorf("请求参数错误: %s", errorMsg)
	case 401:
		return NewAuthError(fmt.Sprintf("认证失败: %s", errorMsg))
	case 403:
		return NewAuthError(fmt.Sprintf("权限不足: %s", errorMsg))
	case 404:
		return NewNotFoundError("API端点", errorMsg)
	case 429:
		return NewBusinessErrorf("请求过于频繁: %s", errorMsg)
	case 500, 502, 503:
		return NewNetworkErrorf(nil, "服务器错误 (%d): %s", statusCode, errorMsg)
	default:
		return NewNetworkErrorf(nil, "HTTP错误 (%d): %s", statusCode, errorMsg)
	}
}

// getModelByName 根据名称获取模型配置
func (c *AIClient) getModelByName(name string) (*ModelConfig, error) {
	for i := range c.config.Models {
		if c.config.Models[i].Name == name {
			return &c.config.Models[i], nil
		}
	}
	return nil, NewNotFoundError("模型", name)
}

// CallPrimary 调用主模型
func (c *AIClient) CallPrimary(ctx context.Context, messages []Message) (*AIResponse, error) {
	req := &AIRequest{
		Messages: messages,
	}
	return c.Call(ctx, req)
}

// CallFallback 调用备用模型
func (c *AIClient) CallFallback(ctx context.Context, messages []Message) (*AIResponse, error) {
	fallbackModel, err := GetFallbackModel(c.config)
	if err != nil {
		return nil, err
	}

	req := &AIRequest{
		Messages: messages,
	}
	return c.callWithRetry(ctx, req, fallbackModel)
}

// SimpleCall 简单调用（使用系统提示和用户消息）
func (c *AIClient) SimpleCall(ctx context.Context, systemPrompt, userMessage string) (*AIResponse, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}
	return c.CallPrimary(ctx, messages)
}

// SimpleCallWithHistory 带历史记录的简单调用
func (c *AIClient) SimpleCallWithHistory(ctx context.Context, systemPrompt string, history []Message, userMessage string) (*AIResponse, error) {
	messages := make([]Message, 0, len(history)+2)
	messages = append(messages, Message{Role: "system", Content: systemPrompt})
	messages = append(messages, history...)
	messages = append(messages, Message{Role: "user", Content: userMessage})
	return c.CallPrimary(ctx, messages)
}

// GetContent 仅获取响应内容
func (c *AIClient) GetContent(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	resp, err := c.SimpleCall(ctx, systemPrompt, userMessage)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// StreamCall 流式调用（返回通道）
func (c *AIClient) StreamCall(ctx context.Context, req *AIRequest) (<-chan *AIResponse, error) {
	// 注意：这是一个简化的实现，实际流式调用需要使用 SSE 或 WebSocket
	// 这里仅作为接口预留
	resultCh := make(chan *AIResponse, 1)

	go func() {
		defer close(resultCh)
		resp, err := c.Call(ctx, req)
		if err != nil {
			resultCh <- &AIResponse{Error: err.Error()}
			return
		}
		resultCh <- resp
	}()

	return resultCh, nil
}

// QuickCall 快速调用AI（使用全局客户端）
func QuickCall(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	config, err := LoadSubAgentConfig()
	if err != nil {
		return "", err
	}
	client := NewAIClient(config)
	return client.GetContent(ctx, systemPrompt, userMessage)
}
