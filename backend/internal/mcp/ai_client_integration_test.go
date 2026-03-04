//go:build integration
// +build integration

package mcp

import (
	"context"
	"testing"
	"time"
)

// TestDeepSeekAPIIntegration 测试DeepSeek API连接
// 运行方式: cd backend && go test -v -tags integration ./internal/mcp/ -run TestDeepSeekAPIIntegration
func TestDeepSeekAPIIntegration(t *testing.T) {
	// 使用硬编码的测试配置
	config := &SubAgentConfig{
		Models: []ModelConfig{
			{
				Name:      "primary",
				APIKey:    "sk-239103d32e3e40c8a2cc1b954269a838",
				Endpoint:  "https://api.deepseek.com/v1/chat/completions",
				Model:     "deepseek-chat",
				IsPrimary: true,
			},
		},
		RateLimit: RateLimitConfig{
			MaxConcurrent:  3,
			RequestsPerMin: 20,
		},
		Timeout: TimeoutConfig{
			Connect: 10,
			Request: 60,
		},
		Retry: RetryConfig{
			MaxRetries: 3,
			BackoffMs:  1000,
		},
	}

	client := NewAIClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("基础连接测试", func(t *testing.T) {
		req := &AIRequest{
			Model: "deepseek-chat",
			Messages: []Message{
				{Role: "system", Content: "You are a helpful assistant. Reply in Chinese."},
				{Role: "user", Content: "请用一句话介绍你自己。"},
			},
		}

		resp, err := client.Call(ctx, req)
		if err != nil {
			t.Fatalf("API调用失败: %v", err)
		}

		if resp.Content == "" {
			t.Fatal("响应内容为空")
		}

		t.Logf("✅ DeepSeek API响应: %s", resp.Content)
	})

	t.Run("代码分析能力测试", func(t *testing.T) {
		resp, err := client.GetContent(ctx,
			"你是一个代码审查专家，请用JSON格式回答。",
			`请分析以下任务描述是否合理，返回JSON格式：{"score": 1-100, "issues": [], "suggestions": []}
任务描述：实现用户登录功能，包括用户名密码验证。`)

		if err != nil {
			t.Fatalf("API调用失败: %v", err)
		}

		t.Logf("✅ 分析结果: %s", resp)
	})
}
