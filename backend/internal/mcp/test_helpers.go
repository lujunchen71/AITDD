package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestServer 测试服务器结构
type TestServer struct {
	Server     *httptest.Server
	ConfigPath string
	ConfigMgr  *ConfigManager
	TempDir    string
}

// NewTestServer 创建测试服务器
func NewTestServer(handler http.Handler) *TestServer {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "aitdd-test-*")
	if err != nil {
		panic(err)
	}

	// 创建 .aitdd 目录
	aitddDir := filepath.Join(tempDir, ".aitdd")
	if err := os.MkdirAll(aitddDir, 0755); err != nil {
		panic(err)
	}

	configPath := filepath.Join(aitddDir, "project.json")

	// 创建测试配置
	config := &Config{
		PathName:   "test-project",
		ApiBaseUrl: "http://localhost:34567/api/v1",
	}
	configData, _ := json.Marshal(config)
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		panic(err)
	}

	// 创建配置管理器
	configMgr := NewConfigManager(configPath)

	// 创建测试 HTTP 服务器
	server := httptest.NewServer(handler)

	// 更新配置的 API URL
	configMgr.SetApiBaseUrl(server.URL + "/api/v1")

	return &TestServer{
		Server:     server,
		ConfigPath: configPath,
		ConfigMgr:  configMgr,
		TempDir:    tempDir,
	}
}

// Close 关闭测试服务器
func (ts *TestServer) Close() {
	ts.Server.Close()
	os.RemoveAll(ts.TempDir)
}

// CreateTestRequest 创建测试请求
func CreateTestRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "test_tool",
			Arguments: args,
		},
	}
}

// MockAPIHandler 模拟 API 处理器
type MockAPIHandler struct {
	Projects  []map[string]interface{}
	Modules   []map[string]interface{}
	Tasks     []map[string]interface{}
	Issues    []map[string]interface{}
	DependsOn []map[string]interface{}
}

// NewMockAPIHandler 创建模拟 API 处理器
func NewMockAPIHandler() *MockAPIHandler {
	return &MockAPIHandler{
		Projects: []map[string]interface{}{
			{
				"id":          "project-1",
				"name":        "Test Project",
				"pathName":    "test-project",
				"description": "A test project",
			},
		},
		Modules: []map[string]interface{}{
			{
				"id":          "module-1",
				"name":        "Auth Module",
				"pathName":    "test-project/auth",
				"projectId":   "project-1",
				"description": "Authentication module",
				"prompt":      "Implement authentication",
				"status":      "pending",
			},
		},
		Tasks: []map[string]interface{}{
			{
				"id":                     "task-1",
				"name":                   "Login Task",
				"pathName":               "test-project/auth/login",
				"moduleId":               "module-1",
				"description":            "Login functionality",
				"status":                 "pending",
				"upstreamContractDetail": `{"username": "string", "password": "string"}`,
				"downstreamContractDetail": `{"token": "string", "expires": "number"}`,
			},
		},
		Issues:    []map[string]interface{}{},
		DependsOn: []map[string]interface{}{},
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *MockAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path

	switch {
	case path == "/api/v1/projects":
		h.handleProjects(w, r)
	case path == "/api/v1/projects/project-1" || path == "/api/v1/projects/by-name/Test Project":
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "project-1",
			"name":        "Test Project",
			"pathName":    "test-project",
			"description": "A test project",
		})
	case path == "/api/v1/projects/by-path/test-project":
		json.NewEncoder(w).Encode(map[string]interface{}{
			"project": map[string]interface{}{
				"id":          "project-1",
				"name":        "Test Project",
				"pathName":    "test-project",
				"description": "A test project",
			},
		})
	case path == "/api/v1/modules/module-1" || path == "/api/v1/modules/by-path/test-project/auth":
		for _, m := range h.Modules {
			if m["id"] == "module-1" || m["pathName"] == "test-project/auth" {
				json.NewEncoder(w).Encode(m)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "module not found"})
	case path == "/api/v1/tasks/task-1" || path == "/api/v1/tasks/by-path/test-project/auth/login":
		for _, t := range h.Tasks {
			if t["id"] == "task-1" || t["pathName"] == "test-project/auth/login" {
				json.NewEncoder(w).Encode(t)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
	default:
		// 处理模块列表
		if len(path) > len("/api/v1/projects/") && path[:len("/api/v1/projects/")] == "/api/v1/projects/" {
			h.handleModules(w, r)
			return
		}
		// 处理任务列表
		if len(path) > len("/api/v1/modules/") && path[:len("/api/v1/modules/")] == "/api/v1/modules/" {
			h.handleTasks(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}

func (h *MockAPIHandler) handleProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": h.Projects,
		})
	}
}

func (h *MockAPIHandler) handleModules(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回模块列表
		json.NewEncoder(w).Encode(h.Modules)
	} else if r.Method == "POST" {
		// 创建模块
		var module map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&module); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		module["id"] = "new-module-id"
		h.Modules = append(h.Modules, module)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(module)
	}
}

func (h *MockAPIHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回任务列表
		json.NewEncoder(w).Encode(h.Tasks)
	} else if r.Method == "POST" {
		// 创建任务
		var task map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		task["id"] = "new-task-id"
		h.Tasks = append(h.Tasks, task)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}
