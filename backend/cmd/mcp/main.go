package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	APIBaseURL = "http://localhost:34567/api/v1"
)

// MCPServer MCP 服务器
type MCPServer struct {
	server *server.MCPServer
	apiURL string
}

// NewMCPServer 创建 MCP 服务器
func NewMCPServer() *MCPServer {
	s := &MCPServer{
		apiURL: APIBaseURL,
	}

	// 创建 MCP 服务器
	s.server = server.NewMCPServer(
		"aitdd",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
	)

	// 注册工具
	s.registerTools()

	return s
}

// registerTools 注册所有 MCP 工具
func (s *MCPServer) registerTools() {
	// 项目相关工具
	s.server.AddTool(mcp.NewTool("get_project_summary",
		mcp.WithDescription("获取项目简述信息"),
	), s.handleGetProjectSummary)

	s.server.AddTool(mcp.NewTool("get_constitution",
		mcp.WithDescription("获取项目公约"),
	), s.handleGetConstitution)

	// 模块相关工具
	s.server.AddTool(mcp.NewTool("get_all_modules",
		mcp.WithDescription("获取所有模块和子模块"),
		mcp.WithString("projectId", mcp.Description("项目 ID"), mcp.Required()),
	), s.handleGetAllModules)

	s.server.AddTool(mcp.NewTool("get_module_overview",
		mcp.WithDescription("获取模块概览"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
	), s.handleGetModuleOverview)

	s.server.AddTool(mcp.NewTool("get_module_task_ids",
		mcp.WithDescription("获取模块中所有任务 ID 列表"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
	), s.handleGetModuleTaskIDs)

	// 任务相关工具
	s.server.AddTool(mcp.NewTool("get_task_details",
		mcp.WithDescription("获取任务详细信息"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
	), s.handleGetTaskDetails)

	s.server.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("修改任务详细信息"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("任务名称")),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("status", mcp.Description("任务状态")),
		mcp.WithString("prompt", mcp.Description("任务提示词")),
	), s.handleUpdateTask)

	s.server.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("在模块中创建任务"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("任务名称"), mcp.Required()),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("prompt", mcp.Description("任务提示词")),
	), s.handleCreateTask)

	s.server.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("删除任务"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
	), s.handleDeleteTask)

	// 模块相关工具
	s.server.AddTool(mcp.NewTool("delete_module",
		mcp.WithDescription("删除模块"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
	), s.handleDeleteModule)

	// 工具相关
	s.server.AddTool(mcp.NewTool("open_frontend",
		mcp.WithDescription("打开前端网页"),
	), s.handleOpenFrontend)

	// 通知相关
	s.server.AddTool(mcp.NewTool("send_notification",
		mcp.WithDescription("发送通知"),
		mcp.WithString("toTaskId", mcp.Description("接收方任务 ID"), mcp.Required()),
		mcp.WithString("type", mcp.Description("通知类型"), mcp.Required()),
		mcp.WithString("title", mcp.Description("通知标题"), mcp.Required()),
		mcp.WithString("message", mcp.Description("通知内容"), mcp.Required()),
	), s.handleSendNotification)

	s.server.AddTool(mcp.NewTool("read_notifications",
		mcp.WithDescription("读取未读通知"),
		mcp.WithString("toTaskId", mcp.Description("任务 ID")),
	), s.handleReadNotifications)

	// 锁相关
	s.server.AddTool(mcp.NewTool("lock_resource",
		mcp.WithDescription("锁定资源"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("resourceId", mcp.Description("资源 ID"), mcp.Required()),
	), s.handleLockResource)

	s.server.AddTool(mcp.NewTool("unlock_resource",
		mcp.WithDescription("解锁资源"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("resourceId", mcp.Description("资源 ID"), mcp.Required()),
	), s.handleUnlockResource)

	s.server.AddTool(mcp.NewTool("get_lock_status",
		mcp.WithDescription("查询锁定状态"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("resourceId", mcp.Description("资源 ID"), mcp.Required()),
	), s.handleGetLockStatus)

	// 模块依赖相关
	s.server.AddTool(mcp.NewTool("create_module_dependency",
		mcp.WithDescription("创建模块依赖"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithString("dependsOnModuleId", mcp.Description("被依赖的模块 ID"), mcp.Required()),
		mcp.WithString("dependencyType", mcp.Description("依赖类型")),
		mcp.WithString("contractSummary", mcp.Description("契约摘要")),
	), s.handleCreateModuleDependency)

	s.server.AddTool(mcp.NewTool("get_module_dependencies",
		mcp.WithDescription("获取模块依赖"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
	), s.handleGetModuleDependencies)

	// 任务依赖相关
	s.server.AddTool(mcp.NewTool("create_task_dependency",
		mcp.WithDescription("创建任务依赖"),
		mcp.WithString("upstreamTaskId", mcp.Description("上游任务 ID"), mcp.Required()),
		mcp.WithString("downstreamTaskId", mcp.Description("下游任务 ID"), mcp.Required()),
		mcp.WithString("contractSummary", mcp.Description("契约摘要")),
	), s.handleCreateTaskDependency)

	s.server.AddTool(mcp.NewTool("get_task_dependencies",
		mcp.WithDescription("获取任务依赖"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
	), s.handleGetTaskDependencies)
}

// 工具处理函数实现

func (s *MCPServer) handleGetProjectSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	resp, err := http.Get(s.apiURL + "/project")
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetConstitution(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	resp, err := http.Get(s.apiURL + "/project/constitution")
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetAllModules(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	projectID := request.Params.Arguments["projectId"]
	if projectID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 projectId 参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules?projectId=%s", s.apiURL, projectID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetModuleOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	moduleID := request.Params.Arguments["moduleId"]
	if moduleID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 moduleId 参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.apiURL, moduleID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetModuleTaskIDs(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	moduleID := request.Params.Arguments["moduleId"]
	if moduleID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 moduleId 参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks?fields=id", s.apiURL, moduleID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetTaskDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	taskID := request.Params.Arguments["taskId"]
	if taskID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 taskId 参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.apiURL, taskID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	taskID := request.Params.Arguments["taskId"]
	if taskID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 taskId 参数"}}}, nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if name := request.Params.Arguments["name"]; name != nil {
		updateData["name"] = name
	}
	if description := request.Params.Arguments["description"]; description != nil {
		updateData["description"] = description
	}
	if status := request.Params.Arguments["status"]; status != nil {
		updateData["status"] = status
	}
	if prompt := request.Params.Arguments["prompt"]; prompt != nil {
		updateData["prompt"] = prompt
	}

	jsonData, _ := json.Marshal(updateData)
	resp, err := http.Put(fmt.Sprintf("%s/tasks/%s", s.apiURL, taskID), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleCreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	moduleID := request.Params.Arguments["moduleId"]
	if moduleID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 moduleId 参数"}}}, nil
	}

	name := request.Params.Arguments["name"]
	if name == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 name 参数"}}}, nil
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"moduleId": moduleID,
		"name":     name,
	}
	if description := request.Params.Arguments["description"]; description != nil {
		createData["description"] = description
	}
	if prompt := request.Params.Arguments["prompt"]; prompt != nil {
		createData["prompt"] = prompt
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/tasks", s.apiURL), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	taskID := request.Params.Arguments["taskId"]
	if taskID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 taskId 参数"}}}, nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", s.apiURL, taskID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleDeleteModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	moduleID := request.Params.Arguments["moduleId"]
	if moduleID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 moduleId 参数"}}}, nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/modules/%s", s.apiURL, moduleID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleOpenFrontend(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	// 尝试打开浏览器
	url := "http://localhost:5173"
	
	// 不同平台的打开命令
	var cmd string
	switch {
	case isWindows():
		cmd = fmt.Sprintf("start %s", url)
	case isMacOS():
		cmd = fmt.Sprintf("open %s", url)
	default:
		cmd = fmt.Sprintf("xdg-open %s", url)
	}

	// 执行命令 (这里简化处理，实际需要使用 os/exec)
	_ = cmd

	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "已尝试打开前端页面：" + url}}}, nil
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

func isMacOS() bool {
	// 简化的判断
	return false
}

func (s *MCPServer) handleSendNotification(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	toTaskID := request.Params.Arguments["toTaskId"]
	if toTaskID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 toTaskId 参数"}}}, nil
	}

	// 构建通知数据
	notificationData := map[string]interface{}{
		"toTaskId": toTaskID,
		"type":     request.Params.Arguments["type"],
		"title":    request.Params.Arguments["title"],
		"message":  request.Params.Arguments["message"],
	}

	jsonData, _ := json.Marshal(notificationData)
	resp, err := http.Post(fmt.Sprintf("%s/notifications", s.apiURL), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleReadNotifications(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	toTaskID := request.Params.Arguments["toTaskId"]
	
	url := fmt.Sprintf("%s/notifications?read=false", s.apiURL)
	if toTaskID != nil {
		url += fmt.Sprintf("&toTaskId=%s", toTaskID)
	}

	resp, err := http.Get(url)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleLockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	resourceType := request.Params.Arguments["resourceType"]
	resourceID := request.Params.Arguments["resourceId"]

	if resourceType == nil || resourceID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少必要参数"}}}, nil
	}

	lockData := map[string]interface{}{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	jsonData, _ := json.Marshal(lockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock", s.apiURL), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleUnlockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	resourceType := request.Params.Arguments["resourceType"]
	resourceID := request.Params.Arguments["resourceId"]

	if resourceType == nil || resourceID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少必要参数"}}}, nil
	}

	unlockData := map[string]interface{}{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	jsonData, _ := json.Marshal(unlockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock/unlock", s.apiURL), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetLockStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	resourceType := request.Params.Arguments["resourceType"]
	resourceID := request.Params.Arguments["resourceId"]

	if resourceType == nil || resourceID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少必要参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/lock/status?resourceType=%s&resourceId=%s", s.apiURL, resourceType, resourceID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleCreateModuleDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	moduleID := request.Params.Arguments["moduleId"]
	dependsOnModuleID := request.Params.Arguments["dependsOnModuleId"]

	if moduleID == nil || dependsOnModuleID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少必要参数"}}}, nil
	}

	depData := map[string]interface{}{
		"dependsOnModuleId": dependsOnModuleID,
	}
	if depType := request.Params.Arguments["dependencyType"]; depType != nil {
		depData["dependencyType"] = depType
	}
	if contract := request.Params.Arguments["contractSummary"]; contract != nil {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/modules/%s/dependencies", s.apiURL, moduleID), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetModuleDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	moduleID := request.Params.Arguments["moduleId"]
	if moduleID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 moduleId 参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependencies", s.apiURL, moduleID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleCreateTaskDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	upstreamTaskID := request.Params.Arguments["upstreamTaskId"]
	downstreamTaskID := request.Params.Arguments["downstreamTaskId"]

	if upstreamTaskID == nil || downstreamTaskID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少必要参数"}}}, nil
	}

	depData := map[string]interface{}{
		"upstreamTaskId":   upstreamTaskID,
		"downstreamTaskId": downstreamTaskID,
	}
	if contract := request.Params.Arguments["contractSummary"]; contract != nil {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/dependencies", s.apiURL), "application/json", jsonData)
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

func (s *MCPServer) handleGetTaskDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.ToolResult, error) {
	taskID := request.Params.Arguments["taskId"]
	if taskID == nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "缺少 taskId 参数"}}}, nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/dependencies?taskId=%s", s.apiURL, taskID))
	if err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "请求失败：" + err.Error()}}}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: "解析失败：" + err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.ToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(data)}}}, nil
}

// Run 运行 MCP 服务器
func (s *MCPServer) Run() error {
	log.Println("Starting AITDD MCP Server...")
	return s.server.Start(context.Background())
}

func main() {
	server := NewMCPServer()
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
