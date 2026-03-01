package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// BatchMockAPIHandler 批量操作模拟 API 处理器
type BatchMockAPIHandler struct {
	modules map[string]map[string]interface{}
	tasks   map[string]map[string]interface{}
}

// NewBatchMockAPIHandler 创建批量操作模拟 API 处理器
func NewBatchMockAPIHandler() *BatchMockAPIHandler {
	return &BatchMockAPIHandler{
		modules: map[string]map[string]interface{}{
			"test-project/auth": {
				"id":          "module-1",
				"name":        "Auth Module",
				"pathName":    "test-project/auth",
				"description": "Authentication module",
				"prompt":      "Implement authentication",
				"status":      "developing",
				"version":     1,
			},
			"test-project/api": {
				"id":          "module-2",
				"name":        "API Module",
				"pathName":    "test-project/api",
				"description": "API module",
				"prompt":      "Implement API",
				"status":      "pending",
				"version":     2,
			},
		},
		tasks: map[string]map[string]interface{}{
			"test-project/auth/login": {
				"id":          "task-1",
				"name":        "Login Task",
				"pathName":    "test-project/auth/login",
				"description": "Login functionality",
				"status":      "in_progress",
				"tests":       `[{"target": "验证登录", "api": "test_login()"}]`,
				"version":     1,
			},
			"test-project/auth/logout": {
				"id":          "task-2",
				"name":        "Logout Task",
				"pathName":    "test-project/auth/logout",
				"description": "Logout functionality",
				"status":      "pending",
				"version":     1,
			},
		},
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *BatchMockAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := r.URL.Path

	// 处理模块查询
	if strings.HasPrefix(path, "/api/v1/modules/by-path/") {
		pathName := strings.TrimPrefix(path, "/api/v1/modules/by-path/")
		if r.Method == "GET" {
			if module, ok := h.modules[pathName]; ok {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"module": module,
					},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "module not found"})
			return
		} else if r.Method == "PUT" {
			if module, ok := h.modules[pathName]; ok {
				var updateData map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				// 检查版本
				providedVersion := int(updateData["version"].(float64))
				currentVersion := module["version"].(int)
				if providedVersion != currentVersion {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]string{
						"error":   "version conflict",
						"message": "版本冲突，请获取新版本后重试",
					})
					return
				}
				// 更新模块
				for k, v := range updateData {
					module[k] = v
				}
				module["version"] = currentVersion + 1
				h.modules[pathName] = module
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"module": module,
					},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// 处理任务查询
	if strings.HasPrefix(path, "/api/v1/tasks/by-path/") {
		pathName := strings.TrimPrefix(path, "/api/v1/tasks/by-path/")
		if r.Method == "GET" {
			if task, ok := h.tasks[pathName]; ok {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"task": task,
					},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
			return
		} else if r.Method == "PUT" {
			if task, ok := h.tasks[pathName]; ok {
				var updateData map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				// 检查版本
				providedVersion := int(updateData["version"].(float64))
				currentVersion := task["version"].(int)
				if providedVersion != currentVersion {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]string{
						"error":   "version conflict",
						"message": "版本冲突，请获取新版本后重试",
					})
					return
				}
				// 更新任务
				for k, v := range updateData {
					task[k] = v
				}
				task["version"] = currentVersion + 1
				h.tasks[pathName] = task
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"task": task,
					},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
}

// ==================== query_module 批量查询测试 ====================

func TestHandleQueryModuleImpl_BatchQuery(t *testing.T) {
	mockHandler := NewBatchMockAPIHandler()
	testServer := httptest.NewServer(mockHandler)
	defer testServer.Close()

	mcpServer := &MCPServer{}

	t.Run("批量查询多个模块", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"queries": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth",
					"fields":   []interface{}{"name", "status", "version"},
				},
				map[string]interface{}{
					"pathName": "test-project/api",
					"fields":   []interface{}{"description", "prompt"},
				},
			},
		})

		// 设置 API URL
		mcpServer.configManager = NewTestConfigManager(testServer.URL + "/api/v1")

		result, err := mcpServer.handleQueryModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		// 检查输出内容
		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证树形结构
		if !strings.Contains(content, "test-project:") {
			t.Error("输出应包含项目名称")
		}
		if !strings.Contains(content, "auth:") {
			t.Error("输出应包含 auth 模块")
		}
		if !strings.Contains(content, "api:") {
			t.Error("输出应包含 api 模块")
		}
	})

	t.Run("空queries数组应返回错误", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"queries": []interface{}{},
		})

		result, err := mcpServer.handleQueryModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		content := getResultText(result)
		if !strings.Contains(content, "不能为空") {
			t.Errorf("空数组应返回错误信息，实际: %s", content)
		}
	})

	t.Run("缺少queries参数应返回错误", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleQueryModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		content := getResultText(result)
		if !strings.Contains(content, "缺少") {
			t.Errorf("缺少参数应返回错误信息，实际: %s", content)
		}
	})

	t.Run("部分路径不存在", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"queries": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth",
					"fields":   []interface{}{"name"},
				},
				map[string]interface{}{
					"pathName": "test-project/not-exist",
					"fields":   []interface{}{"name"},
				},
			},
		})

		result, err := mcpServer.handleQueryModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 应该包含存在的模块和错误信息
		if !strings.Contains(content, "auth:") {
			t.Error("输出应包含存在的 auth 模块")
		}
	})
}

// ==================== query_task 批量查询测试 ====================

func TestHandleQueryTaskImpl_BatchQuery(t *testing.T) {
	mockHandler := NewBatchMockAPIHandler()
	testServer := httptest.NewServer(mockHandler)
	defer testServer.Close()

	mcpServer := &MCPServer{}
	mcpServer.configManager = NewTestConfigManager(testServer.URL + "/api/v1")

	t.Run("批量查询多个任务", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"queries": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth/login",
					"fields":   []interface{}{"name", "status", "tests"},
				},
				map[string]interface{}{
					"pathName": "test-project/auth/logout",
					"fields":   []interface{}{"name", "description"},
				},
			},
		})

		result, err := mcpServer.handleQueryTaskImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if result == nil {
			t.Fatal("结果不应为 nil")
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证树形结构
		if !strings.Contains(content, "test-project:") {
			t.Error("输出应包含项目名称")
		}
		if !strings.Contains(content, "auth:") {
			t.Error("输出应包含 auth 模块")
		}
		if !strings.Contains(content, "login:") {
			t.Error("输出应包含 login 任务")
		}
		if !strings.Contains(content, "logout:") {
			t.Error("输出应包含 logout 任务")
		}
	})

	t.Run("拒绝模块路径（斜杠数量不足）", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"queries": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth", // 只有一个斜杠，不是任务路径
					"fields":   []interface{}{"name"},
				},
			},
		})

		result, err := mcpServer.handleQueryTaskImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		content := getResultText(result)
		if !strings.Contains(content, "必须包含两个斜杠") {
			t.Errorf("应拒绝模块路径，实际: %s", content)
		}
	})

	t.Run("空queries数组应返回错误", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"queries": []interface{}{},
		})

		result, err := mcpServer.handleQueryTaskImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		content := getResultText(result)
		if !strings.Contains(content, "不能为空") {
			t.Errorf("空数组应返回错误信息，实际: %s", content)
		}
	})
}

// ==================== modify_module 批量修改测试 ====================

func TestHandleModifyModuleImpl_BatchModify(t *testing.T) {
	mockHandler := NewBatchMockAPIHandler()
	testServer := httptest.NewServer(mockHandler)
	defer testServer.Close()

	mcpServer := &MCPServer{}
	mcpServer.configManager = NewTestConfigManager(testServer.URL + "/api/v1")

	t.Run("批量修改多个模块成功", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"operations": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth",
					"version":  1,
					"data": map[string]interface{}{
						"status": "completed",
					},
				},
				map[string]interface{}{
					"pathName": "test-project/api",
					"version":  2,
					"data": map[string]interface{}{
						"status": "developing",
					},
				},
			},
		})

		result, err := mcpServer.handleModifyModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证成功
		if !strings.Contains(content, "success: true") {
			t.Errorf("批量修改应成功，实际: %s", content)
		}
		if !strings.Contains(content, "newVersion: 2") {
			t.Error("应返回新版本号")
		}
	})

	t.Run("版本冲突应拒绝整个批量操作", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"operations": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth",
					"version":  1,
					"data": map[string]interface{}{
						"status": "completed",
					},
				},
				map[string]interface{}{
					"pathName": "test-project/api",
					"version":  1, // 错误版本，当前是 2
					"data": map[string]interface{}{
						"status": "developing",
					},
				},
			},
		})

		result, err := mcpServer.handleModifyModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证冲突信息
		if !strings.Contains(content, "success: false") {
			t.Errorf("版本冲突应导致失败，实际: %s", content)
		}
		if !strings.Contains(content, "VERSION_CONFLICT") {
			t.Error("应返回版本冲突错误")
		}
		if !strings.Contains(content, "conflicts:") {
			t.Error("应包含冲突详情")
		}
		if !strings.Contains(content, "checks:") {
			t.Error("应包含检查结果")
		}
	})

	t.Run("空operations数组应返回错误", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"operations": []interface{}{},
		})

		result, err := mcpServer.handleModifyModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		if !strings.Contains(content, "不能为空") {
			t.Errorf("空数组应返回错误信息，实际: %s", content)
		}
	})

	t.Run("缺少operations参数应返回错误", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{})

		result, err := mcpServer.handleModifyModuleImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		if !strings.Contains(content, "缺少") {
			t.Errorf("缺少参数应返回错误信息，实际: %s", content)
		}
	})
}

// ==================== modify_task 批量修改测试 ====================

func TestHandleModifyTaskImpl_BatchModify(t *testing.T) {
	mockHandler := NewBatchMockAPIHandler()
	testServer := httptest.NewServer(mockHandler)
	defer testServer.Close()

	mcpServer := &MCPServer{}
	mcpServer.configManager = NewTestConfigManager(testServer.URL + "/api/v1")

	t.Run("批量修改多个任务成功", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"operations": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth/login",
					"version":  1,
					"data": map[string]interface{}{
						"status": "completed",
					},
				},
				map[string]interface{}{
					"pathName": "test-project/auth/logout",
					"version":  1,
					"data": map[string]interface{}{
						"status": "in_progress",
					},
				},
			},
		})

		result, err := mcpServer.handleModifyTaskImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证成功
		if !strings.Contains(content, "success: true") {
			t.Errorf("批量修改应成功，实际: %s", content)
		}
	})

	t.Run("版本冲突应拒绝整个批量操作", func(t *testing.T) {
		req := CreateTestRequest(map[string]interface{}{
			"operations": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth/login",
					"version":  1,
					"data": map[string]interface{}{
						"status": "completed",
					},
				},
				map[string]interface{}{
					"pathName": "test-project/auth/logout",
					"version":  99, // 错误版本
					"data": map[string]interface{}{
						"status": "in_progress",
					},
				},
			},
		})

		result, err := mcpServer.handleModifyTaskImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证冲突信息
		if !strings.Contains(content, "success: false") {
			t.Errorf("版本冲突应导致失败，实际: %s", content)
		}
		if !strings.Contains(content, "VERSION_CONFLICT") {
			t.Error("应返回版本冲突错误")
		}
	})

	t.Run("JSON字段处理", func(t *testing.T) {
		// 重新创建 mock handler 以确保版本从1开始
		mockHandler := NewBatchMockAPIHandler()
		localServer := httptest.NewServer(mockHandler)
		defer localServer.Close()

		localMcpServer := &MCPServer{}
		localMcpServer.configManager = NewTestConfigManager(localServer.URL + "/api/v1")

		req := CreateTestRequest(map[string]interface{}{
			"operations": []interface{}{
				map[string]interface{}{
					"pathName": "test-project/auth/login",
					"version":  1,
					"data": map[string]interface{}{
						"tests": []interface{}{
							map[string]interface{}{
								"target": "验证登录",
								"api":    "test_login()",
							},
						},
						"codePaths": []interface{}{"src/auth/login.ts"},
					},
				},
			},
		})

		result, err := localMcpServer.handleModifyTaskImpl(context.Background(), req)
		if err != nil {
			t.Fatalf("修改失败: %v", err)
		}

		content := getResultText(result)
		t.Logf("输出:\n%s", content)

		// 验证成功
		if !strings.Contains(content, "success: true") {
			t.Errorf("批量修改应成功，实际: %s", content)
		}
	})
}

// ==================== 辅助函数 ====================

// NewTestConfigManager 创建测试配置管理器
func NewTestConfigManager(apiURL string) *ConfigManager {
	return &ConfigManager{
		config: &Config{
			ApiBaseUrl: apiURL,
		},
	}
}

// getResultText 获取结果文本
func getResultText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return ""
}
