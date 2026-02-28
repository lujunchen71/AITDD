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

// APIClient API 客户端
type APIClient struct {
	baseURL    string
	httpClient *http.Client
	cache      *MCPCache
	logger     *MCPLogger
	timeout    time.Duration
}

// DefaultAPITimeout 默认 API 超时时间
const DefaultAPITimeout = 30 * time.Second

// NewAPIClient 创建新的 API 客户端
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultAPITimeout,
		},
		cache:   GetMCPCache(),
		logger:  GetMCPLogger(),
		timeout: DefaultAPITimeout,
	}
}

// SetTimeout 设置超时时间
func (c *APIClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// ==================== 通用 HTTP 方法 ====================

// Get 发送 GET 请求
func (c *APIClient) Get(ctx context.Context, path string) (map[string]interface{}, error) {
	return c.doRequest(ctx, "GET", path, nil)
}

// Post 发送 POST 请求
func (c *APIClient) Post(ctx context.Context, path string, body interface{}) (map[string]interface{}, error) {
	return c.doRequest(ctx, "POST", path, body)
}

// Put 发送 PUT 请求
func (c *APIClient) Put(ctx context.Context, path string, body interface{}) (map[string]interface{}, error) {
	return c.doRequest(ctx, "PUT", path, body)
}

// Delete 发送 DELETE 请求
func (c *APIClient) Delete(ctx context.Context, path string) (map[string]interface{}, error) {
	return c.doRequest(ctx, "DELETE", path, nil)
}

// doRequest 执行 HTTP 请求
func (c *APIClient) doRequest(ctx context.Context, method, path string, body interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, NewParamError("failed to marshal request body", map[string]interface{}{
				"error": err.Error(),
			})
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, NewNetworkError("failed to create request", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error(fmt.Sprintf("API request failed: %s %s", method, path), err)
		return nil, NewNetworkErrorf(err, "API request failed: %s %s", method, path)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime).Milliseconds()
	LogAPICall(method, path, resp.StatusCode, duration)

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError("failed to read response body", err)
	}

	// 检查状态码
	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return nil, NewBusinessError(msg, errResp)
			}
		}
		return nil, NewBusinessError(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, NewInternalError("failed to parse response JSON", err)
	}

	return result, nil
}

// ==================== 缓存失效 ====================

// InvalidateProject 使项目缓存失效
func (c *APIClient) InvalidateProject(pathName string) {
	c.cache.DeleteProject(pathName)
	c.logger.Debug(fmt.Sprintf("Cache invalidated for project: %s", pathName))
}

// InvalidateModule 使模块缓存失效
func (c *APIClient) InvalidateModule(pathName string) {
	c.cache.DeleteModule(pathName)
	c.logger.Debug(fmt.Sprintf("Cache invalidated for module: %s", pathName))
}

// InvalidateTask 使任务缓存失效
func (c *APIClient) InvalidateTask(pathName string) {
	c.cache.DeleteTask(pathName)
	c.logger.Debug(fmt.Sprintf("Cache invalidated for task: %s", pathName))
}

// InvalidateByPath 使路径前缀相关的所有缓存失效
func (c *APIClient) InvalidateByPath(pathPrefix string) {
	c.cache.InvalidateByPath(pathPrefix)
	c.logger.Debug(fmt.Sprintf("Cache invalidated for path prefix: %s", pathPrefix))
}
