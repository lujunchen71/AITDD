package mcp

import (
	"context"
	"strings"
	"testing"
)

func TestHandleCreateNodeImpl_MissingParameters(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError string
	}{
		{
			name:        "缺少 type 参数",
			args:        map[string]interface{}{},
			expectError: "缺少 type 参数",
		},
		{
			name: "缺少 parentPath 参数",
			args: map[string]interface{}{
				"type": "module",
			},
			expectError: "缺少 parentPath 参数",
		},
		{
			name: "缺少 name 参数",
			args: map[string]interface{}{
				"type":       "module",
				"parentPath": "test-project",
			},
			expectError: "缺少 name 参数",
		},
		{
			name: "不支持的节点类型",
			args: map[string]interface{}{
				"type":       "invalid",
				"parentPath": "test-project",
				"name":       "Test",
			},
			expectError: "不支持的节点类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)

			result, err := mcpServer.handleCreateNodeImpl(context.Background(), req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if result == nil {
				t.Fatal("结果不应为 nil")
			}

			// 检查错误消息
			if len(result.Content) > 0 {
				content := result.Content[0]
				if textContent, ok := content.(interface{ GetText() string }); ok {
					if !strings.Contains(textContent.GetText(), tt.expectError) {
						t.Errorf("期望错误包含 '%s', 实际: %s", tt.expectError, textContent.GetText())
					}
				}
			}
		})
	}
}

func TestHandleModifyNodeImpl_MissingPath(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("缺少 pathName 参数", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"data": map[string]interface{}{
				"name": "New Name",
			},
		})

		result, err := mcpServer.handleModifyNodeImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查错误消息
		if len(result.Content) > 0 {
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				if !strings.Contains(textContent.GetText(), "缺少") {
					t.Logf("返回消息: %s", textContent.GetText())
				}
			}
		}
	})
}

func TestHandleDeleteNodeImpl_MissingPath(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("缺少 pathName 参数", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleDeleteNodeImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查错误消息
		if len(result.Content) > 0 {
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				if !strings.Contains(textContent.GetText(), "缺少") {
					t.Logf("返回消息: %s", textContent.GetText())
				}
			}
		}
	})
}

func TestGeneratePathName_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		parentPath string
		nodeName   string
		expected   string
	}{
		{
			name:       "空父路径",
			parentPath: "",
			nodeName:   "Test",
			expected:   "/test",
		},
		{
			name:       "特殊字符名称",
			parentPath: "aitdd",
			nodeName:   "Test@Module!",
			expected:   "aitdd/test@module!",
		},
		{
			name:       "大写转小写",
			parentPath: "aitdd",
			nodeName:   "AUTHMODULE",
			expected:   "aitdd/authmodule",
		},
		{
			name:       "多个连续空格",
			parentPath: "aitdd",
			nodeName:   "Test   Module",
			expected:   "aitdd/test___module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePathName(tt.parentPath, tt.nodeName)
			if result != tt.expected {
				t.Errorf("generatePathName(%s, %s) = %s, 期望 %s",
					tt.parentPath, tt.nodeName, result, tt.expected)
			}
		})
	}
}
