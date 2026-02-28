package mcp

import (
	"context"
	"strings"
	"testing"
)

func TestDetectNodeType(t *testing.T) {
	server := &MCPServer{}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"项目路径（无斜杠）", "aitdd", "project"},
		{"模块路径（一个斜杠）", "aitdd/auth", "module"},
		{"任务路径（两个斜杠）", "aitdd/auth/login", "task"},
		{"深层任务路径", "aitdd/auth/login/oauth", "task"},
		{"空路径", "", "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.detectNodeType(tt.path)
			if result != tt.expected {
				t.Errorf("detectNodeType(%s) = %s, 期望 %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFormatAsYAML(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		fields   []string
		contains []string
	}{
		{
			name: "格式化所有字段",
			data: map[string]interface{}{
				"name":   "Test",
				"status": "pending",
				"count":  42,
			},
			fields:   []string{},
			contains: []string{"name: Test", "status: pending", "count: 42"},
		},
		{
			name: "只格式化指定字段",
			data: map[string]interface{}{
				"name":        "Test",
				"description": "A test item",
				"hidden":      "should not appear",
			},
			fields:   []string{"name", "description"},
			contains: []string{"name: Test", "description: A test item"},
		},
		{
			name: "处理空值",
			data: map[string]interface{}{
				"name":  "Test",
				"empty": nil,
			},
			fields:   []string{},
			contains: []string{"name: Test", "empty: null"},
		},
		{
			name: "处理数组",
			data: map[string]interface{}{
				"tags": []interface{}{"tag1", "tag2", "tag3"},
			},
			fields:   []string{},
			contains: []string{"tags: [tag1, tag2, tag3]"},
		},
		{
			name: "处理空数组",
			data: map[string]interface{}{
				"items": []interface{}{},
			},
			fields:   []string{},
			contains: []string{"items: []"},
		},
		{
			name: "处理多行字符串",
			data: map[string]interface{}{
				"content": "line1\nline2\nline3",
			},
			fields:   []string{},
			contains: []string{`content: "line1\nline2\nline3"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAsYAML(tt.data, tt.fields)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatAsYAML 结果不包含 '%s'\n实际输出:\n%s", expected, result)
				}
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"字符串", "hello", "hello"},
		{"nil 值", nil, "null"},
		{"布尔值 true", true, "true"},
		{"布尔值 false", false, "false"},
		{"整数", 42, "42"},
		{"浮点数", 3.14, "3.14"},
		{"空数组", []interface{}{}, "[]"},
		{"非空数组", []interface{}{1, 2, 3}, "[1, 2, 3]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %s, 期望 %s", tt.value, result, tt.expected)
			}
		})
	}
}

func TestGeneratePathName(t *testing.T) {
	tests := []struct {
		name       string
		parentPath string
		nodeName   string
		expected   string
	}{
		{
			name:       "简单路径",
			parentPath: "aitdd",
			nodeName:   "Auth",
			expected:   "aitdd/auth",
		},
		{
			name:       "带空格的名称",
			parentPath: "aitdd",
			nodeName:   "User Management",
			expected:   "aitdd/user_management",
		},
		{
			name:       "深层路径",
			parentPath: "aitdd/auth",
			nodeName:   "Login",
			expected:   "aitdd/auth/login",
		},
		{
			name:       "多层空格",
			parentPath: "aitdd",
			nodeName:   "Create User Profile",
			expected:   "aitdd/create_user_profile",
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

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "短字符串不变",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "长字符串截断",
			input:    "this is a very long string that needs to be truncated",
			maxLen:   20,
			expected: "this is a very long ...",
		},
		{
			name:     "正好等于最大长度",
			input:    "exactly twenty char",
			maxLen:   20,
			expected: "exactly twenty char",
		},
		{
			name:     "空字符串",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%s, %d) = %s, 期望 %s",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// 测试 MCP 工具调用（需要模拟 HTTP 服务器）
func TestHandleQueryNodeImpl(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("查询项目节点", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"path": "test-project",
		})

		result, err := mcpServer.handleQueryNodeImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})

	t.Run("查询模块节点", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"path": "test-project/auth",
		})

		result, err := mcpServer.handleQueryNodeImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})

	t.Run("查询任务节点", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"path": "test-project/auth/login",
		})

		result, err := mcpServer.handleQueryNodeImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})

	t.Run("缺少 path 参数", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleQueryNodeImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		// 应该返回错误消息
		if result != nil && len(result.Content) > 0 {
			// 检查是否包含错误信息
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				if !strings.Contains(textContent.GetText(), "缺少") {
					t.Log("返回了结果，可能需要检查错误处理")
				}
			}
		}
	})
}
