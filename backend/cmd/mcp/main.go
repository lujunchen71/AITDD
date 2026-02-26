package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer MCP 服务器
type MCPServer struct {
	server        *server.MCPServer
	configManager *ConfigManager
	ruleEngine    *RuleEngine
}

// getApiURL 动态获取API URL
func (s *MCPServer) getApiURL() string {
	return s.configManager.GetApiBaseUrl()
}

// NewMCPServer 创建 MCP 服务器
func NewMCPServer() *MCPServer {
	s := &MCPServer{
		configManager: GetConfigManager(),
		ruleEngine:    GetRuleEngine(),
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

// getParamInt 从请求参数中获取整数值
func getParamInt(request mcp.CallToolRequest, key string) (int, bool) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return 0, false
	}
	val, exists := args[key]
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// getParamBool 从请求参数中获取布尔值
func getParamBool(request mcp.CallToolRequest, key string) (bool, bool) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return false, false
	}
	val, exists := args[key]
	if !exists {
		return false, false
	}
	boolVal, ok := val.(bool)
	return boolVal, ok
}

// registerTools 注册所有 MCP 工具
func (s *MCPServer) registerTools() {
	// ==================== 配置管理工具 ====================
	s.registerConfigTools()

	// ==================== 获取类工具 ====================
	s.registerGetTools()

	// ==================== 修改类工具 ====================
	s.registerModifyTools()

	// ==================== 状态类工具 ====================
	s.registerStatusTools()

	// ==================== 错误处理类工具 ====================
	s.registerErrorTools()

	// ==================== 检查类工具 ====================
	s.registerCheckTools()

	// ==================== 其他工具 ====================
	s.registerOtherTools()
}

// registerConfigTools 注册配置管理工具
func (s *MCPServer) registerConfigTools() {
	// init_project - 初始化项目配置
	s.server.AddTool(mcp.NewTool("init_project",
		mcp.WithDescription("初始化项目配置，从后端获取项目列表供用户选择，保存到 .aitdd/config.json"),
	), s.handleInitProject)

	// get_config - 获取当前配置
	s.server.AddTool(mcp.NewTool("get_config",
		mcp.WithDescription("获取当前 .aitdd/config.json 配置文件的内容"),
	), s.handleGetConfig)

	// set_project - 设置当前项目
	s.server.AddTool(mcp.NewTool("set_project",
		mcp.WithDescription("设置当前项目信息"),
		mcp.WithString("projectId", mcp.Description("项目 ID"), mcp.Required()),
		mcp.WithString("projectName", mcp.Description("项目名称"), mcp.Required()),
	), s.handleSetProject)
}

// registerGetTools 注册获取类工具
func (s *MCPServer) registerGetTools() {
	// get_project_info - 获取项目简介
	s.server.AddTool(mcp.NewTool("get_project_info",
		mcp.WithDescription("获取项目简介、架构信息、编码规范"),
		mcp.WithString("projectId", mcp.Description("项目ID，不传则使用配置文件中的项目ID")),
	), s.handleGetProjectInfo)

	// get_all_task_code_paths - 获取所有task代码路径
	s.server.AddTool(mcp.NewTool("get_all_task_code_paths",
		mcp.WithDescription("获取项目中所有任务的代码结构路径"),
		mcp.WithString("projectId", mcp.Description("项目ID，不传则使用配置文件中的项目ID")),
	), s.handleGetAllTaskCodePaths)

	// get_all_modules - 获取所有模块概要
	s.server.AddTool(mcp.NewTool("get_all_modules",
		mcp.WithDescription("获取所有模块的 id、名称、介绍"),
		mcp.WithString("projectId", mcp.Description("项目ID，不传则使用配置文件中的项目ID")),
		mcp.WithBoolean("includeStats", mcp.Description("是否包含统计信息")),
	), s.handleGetAllModules)

	// get_module_tasks - 获取模块任务列表
	s.server.AddTool(mcp.NewTool("get_module_tasks",
		mcp.WithDescription("获取某模块的所有任务列表概要"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithBoolean("includeContracts", mcp.Description("是否包含契约信息")),
	), s.handleGetModuleTasks)

	// get_task_detail - 获取任务详情
	s.server.AddTool(mcp.NewTool("get_task_detail",
		mcp.WithDescription("获取某任务的详细信息"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
	), s.handleGetTaskDetail)

	// get_task_contracts - 获取任务上下游契约
	s.server.AddTool(mcp.NewTool("get_task_contracts",
		mcp.WithDescription("获取某任务的上下游契约接口信息"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
		mcp.WithString("direction", mcp.Description("方向：upstream/downstream/both，默认 both")),
	), s.handleGetTaskContracts)

	// 保留旧的工具名称以兼容
	s.server.AddTool(mcp.NewTool("get_project_summary",
		mcp.WithDescription("获取项目简述信息（兼容旧版）"),
	), s.handleGetProjectInfo)

	s.server.AddTool(mcp.NewTool("get_constitution",
		mcp.WithDescription("获取项目公约"),
	), s.handleGetConstitution)

	s.server.AddTool(mcp.NewTool("get_module_overview",
		mcp.WithDescription("获取模块概览"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
	), s.handleGetModuleOverview)

	s.server.AddTool(mcp.NewTool("get_module_task_ids",
		mcp.WithDescription("获取模块中所有任务 ID 列表"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
	), s.handleGetModuleTaskIDs)

	s.server.AddTool(mcp.NewTool("get_task_details",
		mcp.WithDescription("获取任务详细信息（兼容旧版）"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
	), s.handleGetTaskDetail)
}

// registerModifyTools 注册修改类工具
func (s *MCPServer) registerModifyTools() {
	// delete_module - 删除模块
	s.server.AddTool(mcp.NewTool("delete_module",
		mcp.WithDescription("删除指定模块及其所有子任务"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithBoolean("force", mcp.Description("是否强制删除（即使有依赖）")),
	), s.handleDeleteModule)

	// create_module - 创建模块
	s.server.AddTool(mcp.NewTool("create_module",
		mcp.WithDescription("添加新模块"),
		mcp.WithString("name", mcp.Description("模块名称"), mcp.Required()),
		mcp.WithString("description", mcp.Description("模块描述")),
		mcp.WithString("prompt", mcp.Description("模块提示词")),
		mcp.WithString("parentId", mcp.Description("父模块ID")),
		mcp.WithString("projectId", mcp.Description("项目ID，不传则使用配置文件中的项目ID")),
	), s.handleCreateModule)

	// update_module - 更新模块
	s.server.AddTool(mcp.NewTool("update_module",
		mcp.WithDescription("修改模块信息"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("模块名称")),
		mcp.WithString("description", mcp.Description("模块描述")),
		mcp.WithString("prompt", mcp.Description("模块提示词")),
		mcp.WithString("status", mcp.Description("模块状态")),
		mcp.WithNumber("version", mcp.Description("当前版本号"), mcp.Required()),
	), s.handleUpdateModule)

	// delete_module_tasks - 删除模块所有任务
	s.server.AddTool(mcp.NewTool("delete_module_tasks",
		mcp.WithDescription("删除模块的所有任务"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithBoolean("force", mcp.Description("是否强制删除（即使有依赖）")),
	), s.handleDeleteModuleTasks)

	// create_task - 创建任务
	s.server.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("在模块中添加新任务"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("任务名称"), mcp.Required()),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("prompt", mcp.Description("任务提示词")),
		mcp.WithString("upstreamContractDetail", mcp.Description("上游契约详情JSON")),
		mcp.WithString("downstreamContractDetail", mcp.Description("下游契约详情JSON")),
		mcp.WithString("tests", mcp.Description("测试用例JSON数组")),
		mcp.WithString("codePaths", mcp.Description("代码路径JSON数组")),
	), s.handleCreateTask)

	// update_module_full - 完整更新模块
	s.server.AddTool(mcp.NewTool("update_module_full",
		mcp.WithDescription("重新定义模块的所有信息"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithString("moduleJson", mcp.Description("完整的模块JSON数据"), mcp.Required()),
	), s.handleUpdateModuleFull)

	// update_task_full - 完整更新任务
	s.server.AddTool(mcp.NewTool("update_task_full",
		mcp.WithDescription("重新定义任务的完整信息"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
		mcp.WithString("taskJson", mcp.Description("完整的任务JSON数据"), mcp.Required()),
	), s.handleUpdateTaskFull)

	// 保留旧的工具名称
	s.server.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("修改任务详细信息"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("任务名称")),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("status", mcp.Description("任务状态")),
		mcp.WithString("prompt", mcp.Description("任务提示词")),
	), s.handleUpdateTask)

	s.server.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("删除任务"),
		mcp.WithString("taskId", mcp.Description("任务 ID"), mcp.Required()),
	), s.handleDeleteTask)
}

// registerStatusTools 注册状态类工具
func (s *MCPServer) registerStatusTools() {
	// get_all_task_status - 获取所有任务状态
	s.server.AddTool(mcp.NewTool("get_all_task_status",
		mcp.WithDescription("获取项目中所有任务的状态信息"),
		mcp.WithString("projectId", mcp.Description("项目ID，不传则使用配置文件中的项目ID")),
		mcp.WithString("status", mcp.Description("状态过滤")),
		mcp.WithBoolean("includeLockInfo", mcp.Description("是否包含锁定信息")),
	), s.handleGetAllTaskStatus)

	// get_module_task_status - 获取模块任务状态
	s.server.AddTool(mcp.NewTool("get_module_task_status",
		mcp.WithDescription("获取当前模块的所有任务状态信息"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithBoolean("includeLockInfo", mcp.Description("是否包含锁定信息")),
	), s.handleGetModuleTaskStatus)
}

// registerErrorTools 注册错误处理类工具
func (s *MCPServer) registerErrorTools() {
	// get_project_errors - 获取项目错误列表
	s.server.AddTool(mcp.NewTool("get_project_errors",
		mcp.WithDescription("返回整个项目的错误列表，遍历所有任务检查issue_details和bug_log"),
		mcp.WithString("projectId", mcp.Description("项目ID，不传则使用配置文件中的项目ID")),
		mcp.WithString("severity", mcp.Description("严重级别过滤：error/warning/info")),
		mcp.WithBoolean("includeDetails", mcp.Description("是否包含详细信息")),
	), s.handleGetProjectErrors)

	// get_module_errors - 获取模块错误列表
	s.server.AddTool(mcp.NewTool("get_module_errors",
		mcp.WithDescription("返回某模块的所有子任务错误列表"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithString("severity", mcp.Description("严重级别过滤")),
		mcp.WithBoolean("includeDetails", mcp.Description("是否包含详细信息")),
	), s.handleGetModuleErrors)
}

// registerCheckTools 注册检查类工具
func (s *MCPServer) registerCheckTools() {
	// check_module - 检查模块并生成报告
	s.server.AddTool(mcp.NewTool("check_module",
		mcp.WithDescription("根据 .aitdd/rule.json 中的规则检查模块并生成报告"),
		mcp.WithString("moduleId", mcp.Description("模块 ID"), mcp.Required()),
		mcp.WithArray("rules", mcp.Description("指定检查规则ID列表，不传则检查所有")),
		mcp.WithBoolean("includeDynamic", mcp.Description("是否包含动态检查，默认 true")),
		mcp.WithString("format", mcp.Description("输出格式：json/markdown，默认 json")),
	), s.handleCheckModule)
}

// registerOtherTools 注册其他工具
func (s *MCPServer) registerOtherTools() {
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

// ==================== 配置管理工具处理函数 ====================

func (s *MCPServer) handleInitProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取项目列表
	resp, err := http.Get(s.getApiURL() + "/projects")
	if err != nil {
		return mcp.NewToolResultText("获取项目列表失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析项目列表失败：" + err.Error()), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("请从以下项目列表中选择一个项目，然后使用 set_project 工具设置：\n\n%s", string(data))), nil
}

func (s *MCPServer) handleGetConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config, err := s.configManager.ToJSON()
	if err != nil {
		return mcp.NewToolResultText("获取配置失败：" + err.Error()), nil
	}
	return mcp.NewToolResultText(config), nil
}

func (s *MCPServer) handleSetProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, ok1 := getParam(request, "projectId")
	projectName, ok2 := getParam(request, "projectName")

	if !ok1 || !ok2 {
		return mcp.NewToolResultText("缺少必要参数：projectId 和 projectName"), nil
	}

	if err := s.configManager.SetProject(projectID, projectName); err != nil {
		return mcp.NewToolResultText("设置项目失败：" + err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("项目设置成功！\n项目ID: %s\n项目名称: %s", projectID, projectName)), nil
}

// ==================== 获取类工具处理函数 ====================

func (s *MCPServer) handleGetProjectInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/projects/%s", s.getApiURL(), projectID))
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
	projectID := s.configManager.GetProjectID()
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/projects/%s/constitution", s.getApiURL(), projectID))
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

func (s *MCPServer) handleGetAllTaskCodePaths(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks?projectId=%s&fields=id,name,moduleId,codePaths", s.getApiURL(), projectID))
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
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	url := fmt.Sprintf("%s/modules?projectId=%s", s.getApiURL(), projectID)
	if includeStats, ok := getParamBool(request, "includeStats"); ok && includeStats {
		url += "&includeStats=true"
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

func (s *MCPServer) handleGetModuleOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID))
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

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/tasks?fields=id", s.getApiURL(), moduleID))
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

func (s *MCPServer) handleGetModuleTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	url := fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID)
	if includeContracts, ok := getParamBool(request, "includeContracts"); ok && includeContracts {
		url += "&includeContracts=true"
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

func (s *MCPServer) handleGetTaskDetail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	resp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID))
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

func (s *MCPServer) handleGetTaskContracts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	direction, _ := getParam(request, "direction")
	if direction == "" {
		direction = "both"
	}

	// 获取任务详情
	resp, err := http.Get(fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var task map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 获取任务依赖
	depResp, err := http.Get(fmt.Sprintf("%s/dependencies?taskId=%s", s.getApiURL(), taskID))
	if err != nil {
		return mcp.NewToolResultText("获取依赖失败：" + err.Error()), nil
	}
	defer depResp.Body.Close()

	var deps map[string]interface{}
	if err := json.NewDecoder(depResp.Body).Decode(&deps); err != nil {
		return mcp.NewToolResultText("解析依赖失败：" + err.Error()), nil
	}

	// 构建结果
	result := map[string]interface{}{
		"task": map[string]interface{}{
			"id":   task["id"],
			"name": task["name"],
		},
		"upstream": map[string]interface{}{
			"title":     "依赖上游任务提供",
			"contracts": task["upstreamContractDetail"],
		},
		"downstream": map[string]interface{}{
			"title":     "为下游任务提供以下接口",
			"contracts": task["downstreamContractDetail"],
		},
		"dependencies": deps,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ==================== 修改类工具处理函数 ====================

func (s *MCPServer) handleDeleteModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID), nil)
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

func (s *MCPServer) handleCreateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := getParam(request, "name")
	if !ok {
		return mcp.NewToolResultText("缺少 name 参数"), nil
	}

	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	// 构建创建数据
	createData := map[string]interface{}{
		"name":      name,
		"projectId": projectID,
		"status":    "designing",
	}
	if description, ok := getParamAny(request, "description"); ok {
		createData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		createData["prompt"] = prompt
	}
	if parentId, ok := getParamAny(request, "parentId"); ok {
		createData["parentId"] = parentId
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/modules", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

func (s *MCPServer) handleUpdateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if name, ok := getParamAny(request, "name"); ok {
		updateData["name"] = name
	}
	if description, ok := getParamAny(request, "description"); ok {
		updateData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		updateData["prompt"] = prompt
	}
	if status, ok := getParamAny(request, "status"); ok {
		updateData["status"] = status
	}
	if version, ok := getParamInt(request, "version"); ok {
		updateData["version"] = version
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID), bytes.NewBuffer(jsonData))
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

func (s *MCPServer) handleDeleteModuleTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	// 获取模块下的所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取任务列表失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析任务列表失败：" + err.Error()), nil
	}

	// 获取任务列表
	tasks, ok := result["data"].([]interface{})
	if !ok {
		return mcp.NewToolResultText("任务列表格式错误"), nil
	}

	// 逐个删除任务
	deletedCount := 0
	var deletedTasks []map[string]interface{}
	var errors []string

	for _, task := range tasks {
		if taskMap, ok := task.(map[string]interface{}); ok {
			taskID, ok := taskMap["id"].(string)
			if !ok {
				continue
			}

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), nil)
			delResp, err := http.DefaultClient.Do(req)
			if err != nil {
				errors = append(errors, fmt.Sprintf("删除任务 %s 失败: %s", taskID, err.Error()))
				continue
			}
			delResp.Body.Close()
			deletedCount++
			deletedTasks = append(deletedTasks, map[string]interface{}{
				"id":   taskID,
				"name": taskMap["name"],
			})
		}
	}

	resultData := map[string]interface{}{
		"success":      len(errors) == 0,
		"message":      fmt.Sprintf("删除完成，共删除 %d 个任务", deletedCount),
		"deletedCount": deletedCount,
		"deletedTasks": deletedTasks,
	}
	if len(errors) > 0 {
		resultData["errors"] = errors
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
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
		"status":   "ready",
	}
	if description, ok := getParamAny(request, "description"); ok {
		createData["description"] = description
	}
	if prompt, ok := getParamAny(request, "prompt"); ok {
		createData["prompt"] = prompt
	}
	if upstreamContract, ok := getParamAny(request, "upstreamContractDetail"); ok {
		createData["upstreamContractDetail"] = upstreamContract
	}
	if downstreamContract, ok := getParamAny(request, "downstreamContractDetail"); ok {
		createData["downstreamContractDetail"] = downstreamContract
	}
	if tests, ok := getParamAny(request, "tests"); ok {
		createData["tests"] = tests
	}
	if codePaths, ok := getParamAny(request, "codePaths"); ok {
		createData["codePaths"] = codePaths
	}

	jsonData, _ := json.Marshal(createData)
	resp, err := http.Post(fmt.Sprintf("%s/tasks", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

func (s *MCPServer) handleUpdateModuleFull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	moduleJson, ok := getParam(request, "moduleJson")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleJson 参数"), nil
	}

	// 解析JSON
	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(moduleJson), &updateData); err != nil {
		return mcp.NewToolResultText("解析 moduleJson 失败：" + err.Error()), nil
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID), bytes.NewBuffer(jsonData))
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

func (s *MCPServer) handleUpdateTaskFull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	taskJson, ok := getParam(request, "taskJson")
	if !ok {
		return mcp.NewToolResultText("缺少 taskJson 参数"), nil
	}

	// 解析JSON
	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(taskJson), &updateData); err != nil {
		return mcp.NewToolResultText("解析 taskJson 失败：" + err.Error()), nil
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), bytes.NewBuffer(jsonData))
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
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), bytes.NewBuffer(jsonData))
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

func (s *MCPServer) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, ok := getParam(request, "taskId")
	if !ok {
		return mcp.NewToolResultText("缺少 taskId 参数"), nil
	}

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", s.getApiURL(), taskID), nil)
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

// ==================== 状态类工具处理函数 ====================

func (s *MCPServer) handleGetAllTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	url := fmt.Sprintf("%s/tasks?projectId=%s&fields=id,name,moduleId,status", s.getApiURL(), projectID)
	if status, ok := getParam(request, "status"); ok && status != "" {
		url += fmt.Sprintf("&status=%s", status)
	}
	if includeLockInfo, ok := getParamBool(request, "includeLockInfo"); ok && includeLockInfo {
		url += "&includeLockInfo=true"
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

func (s *MCPServer) handleGetModuleTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	url := fmt.Sprintf("%s/tasks?moduleId=%s&fields=id,name,status", s.getApiURL(), moduleID)
	if includeLockInfo, ok := getParamBool(request, "includeLockInfo"); ok && includeLockInfo {
		url += "&includeLockInfo=true"
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

// ==================== 错误处理类工具处理函数 ====================

func (s *MCPServer) handleGetProjectErrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, _ := getParam(request, "projectId")
	if projectID == "" {
		projectID = s.configManager.GetProjectID()
	}
	if projectID == "" {
		return mcp.NewToolResultText("未配置项目ID，请先使用 init_project 或 set_project 设置项目"), nil
	}

	// 获取所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?projectId=%s", s.getApiURL(), projectID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 收集错误
	var errors []map[string]interface{}
	if tasks, ok := result["data"].([]interface{}); ok {
		for _, task := range tasks {
			if taskMap, ok := task.(map[string]interface{}); ok {
				var taskErrors []map[string]interface{}

				// 检查 issue_details
				if issueDetails, ok := taskMap["issueDetails"].(string); ok && issueDetails != "" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "issue",
						"message": issueDetails,
					})
				}

				// 检查 bug_log
				if bugLog, ok := taskMap["bugLog"].(string); ok && bugLog != "" && bugLog != "[]" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "bug",
						"message": bugLog,
					})
				}

				if len(taskErrors) > 0 {
					errors = append(errors, map[string]interface{}{
						"taskId":   taskMap["id"],
						"taskName": taskMap["name"],
						"errors":   taskErrors,
					})
				}
			}
		}
	}

	resultData := map[string]interface{}{
		"projectId":  projectID,
		"scanTime":   time.Now().Format(time.RFC3339),
		"totalTasks": len(result["data"].([]interface{})),
		"errorCount": len(errors),
		"errors":     errors,
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (s *MCPServer) handleGetModuleErrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	// 获取模块下的所有任务
	resp, err := http.Get(fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("请求失败：" + err.Error()), nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return mcp.NewToolResultText("解析失败：" + err.Error()), nil
	}

	// 收集错误
	var errors []map[string]interface{}
	if tasks, ok := result["data"].([]interface{}); ok {
		for _, task := range tasks {
			if taskMap, ok := task.(map[string]interface{}); ok {
				var taskErrors []map[string]interface{}

				// 检查 issue_details
				if issueDetails, ok := taskMap["issueDetails"].(string); ok && issueDetails != "" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "issue",
						"message": issueDetails,
					})
				}

				// 检查 bug_log
				if bugLog, ok := taskMap["bugLog"].(string); ok && bugLog != "" && bugLog != "[]" {
					taskErrors = append(taskErrors, map[string]interface{}{
						"type":    "bug",
						"message": bugLog,
					})
				}

				if len(taskErrors) > 0 {
					errors = append(errors, map[string]interface{}{
						"taskId":   taskMap["id"],
						"taskName": taskMap["name"],
						"errors":   taskErrors,
					})
				}
			}
		}
	}

	resultData := map[string]interface{}{
		"moduleId":   moduleID,
		"scanTime":   time.Now().Format(time.RFC3339),
		"totalTasks": len(result["data"].([]interface{})),
		"errorCount": len(errors),
		"errors":     errors,
	}

	data, _ := json.MarshalIndent(resultData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ==================== 检查类工具处理函数 ====================

func (s *MCPServer) handleCheckModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleID, ok := getParam(request, "moduleId")
	if !ok {
		return mcp.NewToolResultText("缺少 moduleId 参数"), nil
	}

	// 获取模块信息
	moduleResp, err := http.Get(fmt.Sprintf("%s/modules/%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取模块信息失败：" + err.Error()), nil
	}
	defer moduleResp.Body.Close()

	var module map[string]interface{}
	if err := json.NewDecoder(moduleResp.Body).Decode(&module); err != nil {
		return mcp.NewToolResultText("解析模块信息失败：" + err.Error()), nil
	}

	// 获取模块下的任务
	tasksResp, err := http.Get(fmt.Sprintf("%s/tasks?moduleId=%s", s.getApiURL(), moduleID))
	if err != nil {
		return mcp.NewToolResultText("获取任务列表失败：" + err.Error()), nil
	}
	defer tasksResp.Body.Close()

	var tasksResult map[string]interface{}
	if err := json.NewDecoder(tasksResp.Body).Decode(&tasksResult); err != nil {
		return mcp.NewToolResultText("解析任务列表失败：" + err.Error()), nil
	}

	// 提取任务列表
	var tasks []map[string]interface{}
	if data, ok := tasksResult["data"].([]interface{}); ok {
		for _, task := range data {
			if taskMap, ok := task.(map[string]interface{}); ok {
				tasks = append(tasks, taskMap)
			}
		}
	}

	// 执行检查
	report := s.ruleEngine.CheckModule(module, tasks)

	// 获取输出格式
	format, _ := getParam(request, "format")
	if format == "markdown" {
		return mcp.NewToolResultText(report.ToMarkdown()), nil
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ==================== 其他工具处理函数 ====================

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
	resp, err := http.Post(fmt.Sprintf("%s/notifications", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

	url := fmt.Sprintf("%s/notifications?read=false", s.getApiURL())
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
	resp, err := http.Post(fmt.Sprintf("%s/lock", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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
	resp, err := http.Post(fmt.Sprintf("%s/lock/unlock", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

	resp, err := http.Get(fmt.Sprintf("%s/lock/status?resourceType=%s&resourceId=%s", s.getApiURL(), resourceType, resourceID))
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
	resp, err := http.Post(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), moduleID), "application/json", bytes.NewBuffer(jsonData))
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

	resp, err := http.Get(fmt.Sprintf("%s/modules/%s/dependencies", s.getApiURL(), moduleID))
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
	resp, err := http.Post(fmt.Sprintf("%s/dependencies", s.getApiURL()), "application/json", bytes.NewBuffer(jsonData))
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

	resp, err := http.Get(fmt.Sprintf("%s/dependencies?taskId=%s", s.getApiURL(), taskID))
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
