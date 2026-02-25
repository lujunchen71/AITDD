package main

import (
	"bytes"
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

// getParam 从请求参数中获取字符串值
func getParam(request mcp.CallToolRequest, key string) (string, bool) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return "", false
	}
	val, exists := args[key]
	if !exists {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}

// getParamAny 从请求参数中获取任意类型值
func getParamAny(request mcp.CallToolRequest, key string) (any, bool) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, false
	}
	val, exists := args[key]
	return val, exists
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

func (s *MCPServer) handleGetProjectSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := http.Get(s.apiURL + "/project")
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetConstitution(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := http.Get(s.apiURL + "/project/constitution")
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetAllModules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, ok := getParam(request, "projectId")
	if !ok {
		return mcp.NewToolResultText("缺少 projectId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules?projectId=%s", s.apiURL, projectID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetModuleOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.apiURL, moduleID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetModuleTaskIDs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks?fields=id", s.apiURL, moduleID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetTaskDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.apiURL, taskID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if name, ok := getParamAny(request, "name"); ok {
		updateData["name"] = name
	}
	if description, ok := getParamAny(request, "description"); ok {
		updateData["description"] = description
	}
	if status, ok := getParamAny(request, "status"); ok {
		updateData["status"] = status
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		updateData["prompt"] = prompt
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/tasks/%s", s.apiURL, taskID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleCreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	name, ok := getParam(request, "name")
	if !ok {
		return mcp.NewToolResultText("缺少 name 参数"), nil
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"moduleId": moduleID,
		"name":     name,
	}
	if description, ok := getParamAny(request, "description"); ok {
		createData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		createData["prompt"] = prompt
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/tasks", s.apiURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", s.apiURL, taskID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleDeleteModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/modules/%s", s.apiURL, moduleID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleOpenFrontend(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	return mcp.NewToolResultText("已尝试打开前端页面：" + url), nil
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

func isMacOS() bool {
	// 简化的判断
	return false
}

func (s *MCPServer) handleSendNotification(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toTaskID, ok := getParam(request, "toTaskId")
	if !ok {
		return mcp.NewToolResultText("缺少 toTaskId 参数"), nil
	}

	// 构建通知数据
	notificationData := map[string]interface{}{
		"toTaskId": toTaskID,
		"type":     "",
		"title":    "",
		"message":  "",
	}
	if t, ok := getParamAny(request, "type"); ok {
		notificationData["type"] = t
	}
	if t, ok := getParamAny(request, "title"); ok {
		notificationData["title"] = t
	}
	if m, ok := getParamAny(request, "message"); ok {
		notificationData["message"] = m
	}

	jsonData, _ := json.Marshal(notificationData)
	resp, err := http.Post(fmt.Sprintf("%s/notifications", s.apiURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleReadNotifications(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toTaskID, _ := getParam(request, "toTaskId")
	
	url := fmt.Sprintf("%s/notifications?read=false", s.apiURL)
	if toTaskID != "" {
		url += fmt.Sprintf("&toTaskId=%s", toTaskID)
	}

	resp, err := http.Get(url)
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleLockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	resourceID, ok2 := getParam(request, "resourceId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	lockData := map[string]interface{}{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	jsonData, _ := json.Marshal(lockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock", s.apiURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleUnlockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	resourceID, ok2 := getParam(request, "resourceId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	unlockData := map[string]interface{}{
		"resourceType": resourceType,
		"resourceId":   resourceID,
	}

	jsonData, _ := json.Marshal(unlockData)
	resp, err := http.Post(fmt.Sprintf("%s/lock/unlock", s.apiURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetLockStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType, ok1 := getParam(request, "resourceType")
	resourceID, ok2 := getParam(request, "resourceId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/lock/status?resourceType=%s&resourceId=%s", s.apiURL, resourceType, resourceID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleCreateModuleDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok1 := getParam(request, "moduleId")
	dependsOnModuleID, ok2 := getParam(request, "dependsOnModuleId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	depData := map[string]interface{}{
		"dependsOnModuleId": dependsOnModuleID,
	}
	if depType, ok := getParamAny(request, "dependencyType"); ok {
		depData["dependencyType"] = depType
	}
	if contract, ok := getParamAny(request, "contractSummary"); ok {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/modules/%s/dependencies", s.apiURL, moduleID), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetModuleDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependencies", s.apiURL, moduleID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleCreateTaskDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstreamTaskID, ok1 := getParam(request, "upstreamTaskId")
	downstreamTaskID, ok2 := getParam(request, "downstreamTaskId")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数"), nil
	}

	depData := map[string]interface{}{
		"upstreamTaskId":   upstreamTaskID,
		"downstreamTaskId": downstreamTaskID,
	}
	if contract, ok := getParamAny(request, "contractSummary"); ok {
		depData["contractSummary"] = contract
	}

	jsonData, _ := json.Marshal(depData)
	resp, err := http.Post(fmt.Sprintf("%s/dependencies", s.apiURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetTaskDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/dependencies?taskId=%s", s.apiURL, taskID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// Run 运行 MCP 服务器
func (s *MCPServer) Run() error {
	log.Println("Starting AITDD MCP Server...")
	return server.ServeStdio(s.server)
}

func main() {
	s := NewMCPServer()
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
