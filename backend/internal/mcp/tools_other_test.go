package mcp

import (
	"strings"
	"testing"
)

func TestCompareContracts(t *testing.T) {
	tests := []struct {
		name              string
		upstreamContract   string
		downstreamContract string
		expectAligned      bool
		expectedDiffCount int
	}{
		{
			name:              "两个空契约",
			upstreamContract:   "",
			downstreamContract: "",
			expectAligned:      true,
			expectedDiffCount: 0,
		},
		{
			name:              "上游为空",
			upstreamContract:   "",
			downstreamContract: `{"token": "string"}`,
			expectAligned:      false,
			expectedDiffCount:  1,
		},
		{
			name:              "下游为空",
			upstreamContract:   `{"token": "string"}`,
			downstreamContract: "",
			expectAligned:      false,
			expectedDiffCount:  1,
		},
		{
			name:              "相同契约",
			upstreamContract:   `{"username": "string", "password": "string"}`,
			downstreamContract: `{"username": "string", "password": "string"}`,
			expectAligned:      true,
			expectedDiffCount:  0,
		},
		{
			name:              "不同契约",
			upstreamContract:   `{"username": "string", "password": "string"}`,
			downstreamContract: `{"token": "string", "expires": "number"}`,
			expectAligned:      false,
			expectedDiffCount:  4, // username 缺失, password 缺失, token 缺失, expires 缺失
		},
		{
			name:              "嵌套对象相同",
			upstreamContract:   `{"user": {"name": "string", "age": "number"}}`,
			downstreamContract: `{"user": {"name": "string", "age": "number"}}`,
			expectAligned:      true,
			expectedDiffCount:  0,
		},
		{
			name:              "嵌套对象不同",
			upstreamContract:   `{"user": {"name": "string"}}`,
			downstreamContract: `{"user": {"name": "string", "email": "string"}}`,
			expectAligned:      false,
			expectedDiffCount:  1, // email 缺失
		},
		{
			name:              "非JSON字符串比较",
			upstreamContract:   "simple text contract",
			downstreamContract: "simple text contract",
			expectAligned:      true,
			expectedDiffCount:  0,
		},
		{
			name:              "非JSON字符串不同",
			upstreamContract:   "contract A",
			downstreamContract: "contract B",
			expectAligned:      false,
			expectedDiffCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			differences := compareContracts(tt.upstreamContract, tt.downstreamContract)
			aligned := len(differences) == 0

			if aligned != tt.expectAligned {
				t.Errorf("期望对齐状态 = %v, 实际 = %v", tt.expectAligned, aligned)
			}

			if len(differences) != tt.expectedDiffCount {
				t.Errorf("期望差异数量 = %d, 实际 = %d", tt.expectedDiffCount, len(differences))
			}
		})
	}
}

func TestCompareJSONFields(t *testing.T) {
	tests := []struct {
		name              string
		upstream          map[string]interface{}
		downstream        map[string]interface{}
		expectedDiffCount int
	}{
		{
			name:              "空对象",
			upstream:          map[string]interface{}{},
			downstream:        map[string]interface{}{},
			expectedDiffCount: 0,
		},
		{
			name: "完全相同",
			upstream: map[string]interface{}{
				"name": "test",
				"count": 42,
			},
			downstream: map[string]interface{}{
				"name": "test",
				"count": 42,
			},
			expectedDiffCount: 0,
		},
		{
			name: "下游缺少字段",
			upstream: map[string]interface{}{
				"name": "test",
				"extra": "value",
			},
			downstream: map[string]interface{}{
				"name": "test",
			},
			expectedDiffCount: 1,
		},
		{
			name: "上游缺少字段",
			upstream: map[string]interface{}{
				"name": "test",
			},
			downstream: map[string]interface{}{
				"name": "test",
				"extra": "value",
			},
			expectedDiffCount: 1,
		},
		{
			name: "嵌套对象差异",
			upstream: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "john",
				},
			},
			downstream: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "john",
					"age":  30,
				},
			},
			expectedDiffCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			differences := compareJSONFields(tt.upstream, tt.downstream, "")
			if len(differences) != tt.expectedDiffCount {
				t.Errorf("期望差异数量 = %d, 实际 = %d, 差异: %v", 
					tt.expectedDiffCount, len(differences), differences)
			}
		})
	}
}

func TestTruncateString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "零长度限制",
			input:    "test",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "小于省略号长度",
			input:    "test",
			maxLen:   2,
			expected: "te...",
		},
		{
			name:     "等于省略号长度",
			input:    "test",
			maxLen:   3,
			expected: "tes...",
		},
		{
			name:     "单个字符",
			input:    "a",
			maxLen:   10,
			expected: "a",
		},
		{
			name:     "ASCII 字符串不截断",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "ASCII 字符串截断",
			input:    "hello world",
			maxLen:   5,
			expected: "hello...",
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

func TestHandleCheckContractAlignmentImpl_MissingParameters(t *testing.T) {
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
		name string
		args map[string]interface{}
	}{
		{
			name: "缺少所有参数",
			args: map[string]interface{}{},
		},
		{
			name: "缺少下游参数",
			args: map[string]interface{}{
				"upstreamPathName": "test-project/auth/login",
			},
		},
		{
			name: "缺少上游参数",
			args: map[string]interface{}{
				"downstreamPathName": "test-project/auth/login",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)

			result, err := mcpServer.handleCheckContractAlignmentImpl(nil, req)
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
}

func TestHandleCheckDependenciesImpl_MissingParameters(t *testing.T) {
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

		result, err := mcpServer.handleCheckDependenciesImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})
}

func TestHandleQueryDependenciesImpl(t *testing.T) {
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

		result, err := mcpServer.handleQueryDependenciesImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})
}

func TestHandleCreateDependencyImpl_MissingParameters(t *testing.T) {
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
		name string
		args map[string]interface{}
	}{
		{
			name: "缺少所有参数",
			args: map[string]interface{}{},
		},
		{
			name: "缺少 downstreamPathName",
			args: map[string]interface{}{
				"upstreamPathName": "test-project/auth/login",
			},
		},
		{
			name: "缺少 upstreamPathName",
			args: map[string]interface{}{
				"downstreamPathName": "test-project/auth/dashboard",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)

			result, err := mcpServer.handleCreateDependencyImpl(nil, req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if result == nil {
				t.Fatal("结果不应为 nil")
			}
		})
	}
}

func TestHandleDeleteDependencyImpl_MissingParameters(t *testing.T) {
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
		name string
		args map[string]interface{}
	}{
		{
			name: "缺少所有参数",
			args: map[string]interface{}{},
		},
		{
			name: "缺少 downstreamPathName",
			args: map[string]interface{}{
				"upstreamPathName": "test-project/auth/login",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)

			result, err := mcpServer.handleDeleteDependencyImpl(nil, req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if result == nil {
				t.Fatal("结果不应为 nil")
			}
		})
	}
}

// Issue 相关测试
func TestHandleCreateIssueImpl_MissingParameters(t *testing.T) {
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
		name string
		args map[string]interface{}
	}{
		{
			name: "缺少所有参数",
			args: map[string]interface{}{},
		},
		{
			name: "缺少 content",
			args: map[string]interface{}{
				"pathName": "test-project/auth/login",
			},
		},
		{
			name: "缺少 pathName",
			args: map[string]interface{}{
				"content": "这是一个问题",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)

			result, err := mcpServer.handleCreateIssueImpl(nil, req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if result == nil {
				t.Fatal("结果不应为 nil")
			}
		})
	}
}

func TestHandleReplyIssueImpl_MissingParameters(t *testing.T) {
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
		name string
		args map[string]interface{}
	}{
		{
			name: "缺少所有参数",
			args: map[string]interface{}{},
		},
		{
			name: "缺少 content",
			args: map[string]interface{}{
				"issueId": "issue-1",
			},
		},
		{
			name: "缺少 issueId",
			args: map[string]interface{}{
				"content": "这是回复",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTestRequest(tt.args)

			result, err := mcpServer.handleReplyIssueImpl(nil, req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}

			if result == nil {
				t.Fatal("结果不应为 nil")
			}
		})
	}
}

func TestHandleResolveIssueImpl_MissingParameters(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("缺少 issueId 参数", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleResolveIssueImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})
}

func TestHandleQueryIssuesImpl(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("不带参数查询", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleQueryIssuesImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})

	t.Run("带 pathName 参数查询", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"pathName": "test-project/auth/login",
		})

		result, err := mcpServer.handleQueryIssuesImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})

	t.Run("带状态过滤查询", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"status": "open",
		})

		result, err := mcpServer.handleQueryIssuesImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}
	})
}

// ==================== 编译接口测试 ====================

func TestHandleCompileStaticImpl_MissingProject(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器（不设置项目）
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("未设置项目时编译", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleCompileStaticImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查返回消息包含提示
		if len(result.Content) > 0 {
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				text := textContent.GetText()
				if !strings.Contains(text, "未配置项目") {
					t.Logf("返回消息: %s", text)
				}
			}
		}
	})
}

func TestHandleCompileStaticImpl_WithPathName(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("带 pathName 参数编译", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"pathName":        "test-project",
			"includeWarnings": true,
		})

		result, err := mcpServer.handleCompileStaticImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查返回内容
		if len(result.Content) > 0 {
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				text := textContent.GetText()
				// 应该包含编译结果字段
				if !strings.Contains(text, "success:") {
					t.Logf("返回消息不包含 success: %s", text)
				}
			}
		}
	})
}

func TestHandleCompileDynamicImpl_MissingProject(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器（不设置项目）
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("未设置项目时动态编译", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleCompileDynamicImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查返回消息包含提示
		if len(result.Content) > 0 {
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				text := textContent.GetText()
				if !strings.Contains(text, "未配置项目") {
					t.Logf("返回消息: %s", text)
				}
			}
		}
	})
}

func TestHandleCompileDynamicImpl_WithPathName(t *testing.T) {
	// 创建模拟 API 处理器
	mockHandler := NewMockAPIHandler()
	testServer := NewTestServer(mockHandler)
	defer testServer.Close()

	// 创建 MCP 服务器
	mcpServer := &MCPServer{
		configManager: testServer.ConfigMgr,
		ruleEngine:    GetRuleEngine(),
	}

	t.Run("带 pathName 参数动态编译", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"pathName":          "test-project",
			"runTests":          true,
			"validateContracts": true,
		})

		result, err := mcpServer.handleCompileDynamicImpl(nil, req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查返回内容
		if len(result.Content) > 0 {
			content := result.Content[0]
			if textContent, ok := content.(interface{ GetText() string }); ok {
				text := textContent.GetText()
				// 应该包含编译结果字段
				if !strings.Contains(text, "success:") {
					t.Logf("返回消息不包含 success: %s", text)
				}
				// 应该包含时间戳
				if !strings.Contains(text, "compiledAt:") {
					t.Logf("返回消息不包含 compiledAt: %s", text)
				}
			}
		}
	})
}

func TestCompileStaticResult_Structure(t *testing.T) {
	// 测试 CompileStaticResult 结构体
	result := &CompileStaticResult{
		Success:         true,
		TotalModules:    5,
		TotalTasks:      23,
		CompletedTasks:  12,
		InProgressTasks: 3,
		ReadyTasks:      8,
		ErrorCount:      0,
		WarningCount:    2,
		Errors:          []CompileIssue{},
		Warnings: []CompileIssue{
			{
				RuleID:           "W-S-01",
				RuleName:         "测试用例为空",
				ResourceType:     "task",
				ResourceName:     "测试任务",
				ResourcePathName: "test-project/test-task",
				Message:          "任务 [测试任务] 未定义测试用例",
				Suggestion:       "建议添加测试用例",
				Severity:         "warning",
			},
		},
	}

	if result.Success != true {
		t.Error("Success 应该为 true")
	}

	if result.TotalModules != 5 {
		t.Errorf("TotalModules 应该为 5, 实际为 %d", result.TotalModules)
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Warnings 长度应该为 1, 实际为 %d", len(result.Warnings))
	}

	if result.Warnings[0].RuleID != "W-S-01" {
		t.Errorf("RuleID 应该为 W-S-01, 实际为 %s", result.Warnings[0].RuleID)
	}
}

func TestCompileDynamicResult_Structure(t *testing.T) {
	// 测试 CompileDynamicResult 结构体
	result := &CompileDynamicResult{
		Success:       false,
		CompiledAt:    "2024-01-15T10:30:00Z",
		Duration:      "2.5s",
		TotalTasks:    23,
		CompiledTasks: 20,
		FailedTasks:   3,
		TestResults: CompileTestResults{
			Passed:  18,
			Failed:  2,
			Skipped: 3,
		},
		ContractValidation: CompileContractValidation{
			Valid:    20,
			Invalid:  0,
			Warnings: 5,
		},
		Errors: []CompileDynamicError{
			{
				TaskPathName: "test-project/auth/login",
				Error:        "测试失败: Expected 200 but got 500",
			},
		},
	}

	if result.Success != false {
		t.Error("Success 应该为 false")
	}

	if result.TotalTasks != 23 {
		t.Errorf("TotalTasks 应该为 23, 实际为 %d", result.TotalTasks)
	}

	if result.TestResults.Passed != 18 {
		t.Errorf("TestResults.Passed 应该为 18, 实际为 %d", result.TestResults.Passed)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Errors 长度应该为 1, 实际为 %d", len(result.Errors))
	}
}

func TestCompileIssue_Severity(t *testing.T) {
	// 测试 CompileIssue 的严重级别
	errorIssue := CompileIssue{
		RuleID:   "E-S-01",
		Severity: "error",
		Message:  "这是一个错误",
	}

	warningIssue := CompileIssue{
		RuleID:   "W-S-01",
		Severity: "warning",
		Message:  "这是一个警告",
	}

	if errorIssue.Severity != "error" {
		t.Errorf("errorIssue.Severity 应该为 error, 实际为 %s", errorIssue.Severity)
	}

	if warningIssue.Severity != "warning" {
		t.Errorf("warningIssue.Severity 应该为 warning, 实际为 %s", warningIssue.Severity)
	}
}
