package mcp

import (
	"testing"
)

func TestNewMCPServer(t *testing.T) {
	// 注意：这个测试会创建真实的 MCP 服务器
	// 由于单例模式，可能会影响其他测试
	server := NewMCPServer()

	if server == nil {
		t.Fatal("服务器不应为 nil")
	}

	if server.server == nil {
		t.Error("内部 MCP 服务器不应为 nil")
	}

	if server.configManager == nil {
		t.Error("配置管理器不应为 nil")
	}

	if server.ruleEngine == nil {
		t.Error("规则引擎不应为 nil")
	}
}

func TestMCPServer_GetApiURL(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	apiURL := mcpServer.getApiURL()
	if apiURL == "" {
		t.Error("API URL 不应为空")
	}

	// 验证 API URL 是否与配置管理器返回的一致
	expectedURL := testServer.ConfigMgr.GetApiBaseUrl()
	if apiURL != expectedURL {
		t.Errorf("API URL = %s, 期望 %s", apiURL, expectedURL)
	}
}

func TestGetParam(t *testing.T) {
	tests := []struct {
		name         string
		args         map[string]interface{}
		paramName    string
		expectValue  string
		expectExists bool
	}{
		{
			name: "参数存在",
			args: map[string]interface{}{
				"path": "test-project",
			},
			paramName:    "path",
			expectValue:  "test-project",
			expectExists: true,
		},
		{
			name:         "参数不存在",
			args:         map[string]interface{}{},
			paramName:    "path",
			expectValue:  "",
			expectExists: false,
		},
		{
			name: "参数类型不是字符串",
			args: map[string]interface{}{
				"count": 42,
			},
			paramName:    "count",
			expectValue:  "",
			expectExists: false,
		},
		{
			name: "参数值为 nil",
			args: map[string]interface{}{
				"value": nil,
			},
			paramName:    "value",
			expectValue:  "",
			expectExists: false,
		},
		{
			name: "空字符串参数",
			args: map[string]interface{}{
				"empty": "",
			},
			paramName:    "empty",
			expectValue:  "",
			expectExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)
			value, exists := getParam(req, tt.paramName)

			if value != tt.expectValue {
				t.Errorf("getParam() value = %s, 期望 %s", value, tt.expectValue)
			}

			if exists != tt.expectExists {
				t.Errorf("getParam() exists = %v, 期望 %v", exists, tt.expectExists)
			}
		})
	}
}

func TestMCPServer_DetectNodeType(t *testing.T) {
	server := &MCPServer{}

	// 测试并发安全性
	t.Run("并发调用", func(t *testing.T) {
		done := make(chan bool)
		paths := []string{"project", "project/module", "project/module/task"}

		for i := 0; i < 10; i++ {
			go func() {
				for _, path := range paths {
					_ = server.detectNodeType(path)
				}
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// 测试工具注册
func TestMCPServer_RegisterTools(t *testing.T) {
	// 创建一个 MCP 服务器
	server := NewMCPServer()

	// 验证服务器创建成功
	if server == nil {
		t.Fatal("服务器不应为 nil")
	}

	// 验证内部服务器已初始化
	if server.server == nil {
		t.Error("内部服务器不应为 nil")
	}
}
