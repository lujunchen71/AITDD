package mcp

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer MCP 服务器
type MCPServer struct {
	server        *server.MCPServer
	configManager *ConfigManager
	ruleEngine    *RuleEngine
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

// getApiURL 动态获取API URL
func (s *MCPServer) getApiURL() string {
	return s.configManager.GetApiBaseUrl()
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
		mcp.WithDescription("初始化项目配置。获取数据库中所有项目列表（包含id、名称、简介），返回给AI让用户选择。AI会根据用户提供的项目名称匹配对应的projectId，然后自动调用set_project工具设置当前项目。"),
	), s.handleInitProject)

	// get_config - 获取当前配置
	s.server.AddTool(mcp.NewTool("get_config",
		mcp.WithDescription("获取当前 .aitdd/project.json 配置文件的内容"),
	), s.handleGetConfig)

	// set_project - 设置当前项目
	s.server.AddTool(mcp.NewTool("set_project",
		mcp.WithDescription("设置当前项目。将项目ID、名称和路径名称保存到 .aitdd/project.json 文件中，后续所有MCP工具调用都会使用这个项目ID。通常由AI在init_project后根据用户选择的项目名称自动调用。"),
		mcp.WithString("projectId", mcp.Description("项目 ID"), mcp.Required()),
		mcp.WithString("projectName", mcp.Description("项目名称"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("项目路径名称")),
	), s.handleSetProject)
}

// registerGetTools 注册获取类工具
func (s *MCPServer) registerGetTools() {
	// get_project_info - 获取项目简介
	s.server.AddTool(mcp.NewTool("get_project_info",
		mcp.WithDescription("获取项目简介、架构信息、编码规范"),
		mcp.WithString("pathName", mcp.Description("项目路径名称"), mcp.Required()),
	), s.handleGetProjectInfo)

	// get_all_task_code_paths - 获取所有task代码路径
	s.server.AddTool(mcp.NewTool("get_all_task_code_paths",
		mcp.WithDescription("获取项目中所有任务的代码结构路径"),
		mcp.WithString("pathName", mcp.Description("项目路径名称"), mcp.Required()),
	), s.handleGetAllTaskCodePaths)

	// get_all_modules - 获取所有模块概要
	s.server.AddTool(mcp.NewTool("get_all_modules",
		mcp.WithDescription("获取所有模块的 id、名称、介绍"),
		mcp.WithString("pathName", mcp.Description("项目路径名称"), mcp.Required()),
		mcp.WithBoolean("includeStats", mcp.Description("是否包含统计信息")),
	), s.handleGetAllModules)

	// get_module_tasks - 获取模块任务列表
	s.server.AddTool(mcp.NewTool("get_module_tasks",
		mcp.WithDescription("获取某模块的所有任务列表概要"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithBoolean("includeContracts", mcp.Description("是否包含契约信息")),
	), s.handleGetModuleTasks)

	// get_task_detail - 获取任务详情
	s.server.AddTool(mcp.NewTool("get_task_detail",
		mcp.WithDescription("获取某任务的详细信息"),
		mcp.WithString("pathName", mcp.Description("任务路径名称"), mcp.Required()),
	), s.handleGetTaskDetail)

	// get_task_contracts - 获取任务上下游契约
	s.server.AddTool(mcp.NewTool("get_task_contracts",
		mcp.WithDescription("获取某任务的上下游契约接口信息"),
		mcp.WithString("pathName", mcp.Description("任务路径名称"), mcp.Required()),
		mcp.WithString("direction", mcp.Description("方向：upstream/downstream/both，默认 both")),
	), s.handleGetTaskContracts)

	// 保留旧的工具名称以兼容
	s.server.AddTool(mcp.NewTool("get_project_summary",
		mcp.WithDescription("获取项目简述信息（兼容旧版）"),
		mcp.WithString("pathName", mcp.Description("项目路径名称")),
	), s.handleGetProjectInfo)

	s.server.AddTool(mcp.NewTool("get_constitution",
		mcp.WithDescription("获取项目公约"),
	), s.handleGetConstitution)

	s.server.AddTool(mcp.NewTool("get_module_overview",
		mcp.WithDescription("获取模块概览"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
	), s.handleGetModuleOverview)

	s.server.AddTool(mcp.NewTool("get_module_task_path_names",
		mcp.WithDescription("获取模块中所有任务的 pathName 列表"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
	), s.handleGetModuleTaskPathNames)

	s.server.AddTool(mcp.NewTool("get_task_details",
		mcp.WithDescription("获取任务详细信息（兼容旧版）"),
		mcp.WithString("pathName", mcp.Description("任务路径名称"), mcp.Required()),
	), s.handleGetTaskDetail)
}

// registerModifyTools 注册修改类工具
func (s *MCPServer) registerModifyTools() {
	// delete_module - 删除模块
	s.server.AddTool(mcp.NewTool("delete_module",
		mcp.WithDescription("删除指定模块及其所有子任务"),
		mcp.WithString("pathName", mcp.Description("要删除的模块路径名称"), mcp.Required()),
		mcp.WithBoolean("force", mcp.Description("是否强制删除（即使有依赖）")),
	), s.handleDeleteModule)

	// create_module - 创建模块
	s.server.AddTool(mcp.NewTool("create_module",
		mcp.WithDescription("添加新模块"),
		mcp.WithString("name", mcp.Description("模块名称"), mcp.Required()),
		mcp.WithString("description", mcp.Description("模块描述")),
		mcp.WithString("prompt", mcp.Description("模块提示词")),
		mcp.WithString("pathName", mcp.Description("模块路径名称（可选，不传则使用name自动生成）")),
		mcp.WithString("parentPathName", mcp.Description("父模块路径名称，如果是根模块则不传")),
	), s.handleCreateModule)

	// update_module - 更新模块
	s.server.AddTool(mcp.NewTool("update_module",
		mcp.WithDescription("修改模块信息"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithString("name", mcp.Description("模块名称")),
		mcp.WithString("description", mcp.Description("模块描述")),
		mcp.WithString("prompt", mcp.Description("模块提示词")),
		mcp.WithString("status", mcp.Description("模块状态")),
		mcp.WithNumber("version", mcp.Description("当前版本号"), mcp.Required()),
	), s.handleUpdateModule)

	// delete_module_tasks - 删除模块所有任务
	s.server.AddTool(mcp.NewTool("delete_module_tasks",
		mcp.WithDescription("删除模块的所有任务"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithBoolean("force", mcp.Description("是否强制删除（即使有依赖）")),
	), s.handleDeleteModuleTasks)

	// create_task - 创建任务
	s.server.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("在模块中添加新任务"),
		mcp.WithString("pathName", mcp.Description("所属模块的路径名称"), mcp.Required()),
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
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithString("moduleJson", mcp.Description("完整的模块JSON数据"), mcp.Required()),
	), s.handleUpdateModuleFull)

	// update_task_full - 完整更新任务
	s.server.AddTool(mcp.NewTool("update_task_full",
		mcp.WithDescription("重新定义任务的完整信息"),
		mcp.WithString("pathName", mcp.Description("任务路径名称"), mcp.Required()),
		mcp.WithString("taskJson", mcp.Description("完整的任务JSON数据"), mcp.Required()),
	), s.handleUpdateTaskFull)

	// 保留旧的工具名称
	s.server.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("修改任务详细信息"),
		mcp.WithString("pathName", mcp.Description("任务路径名称"), mcp.Required()),
		mcp.WithString("name", mcp.Description("任务名称")),
		mcp.WithString("description", mcp.Description("任务描述")),
		mcp.WithString("status", mcp.Description("任务状态")),
		mcp.WithString("prompt", mcp.Description("任务提示词")),
	), s.handleUpdateTask)

	s.server.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("删除任务"),
		mcp.WithString("pathName", mcp.Description("要删除的任务路径名称"), mcp.Required()),
	), s.handleDeleteTask)
}

// registerStatusTools 注册状态类工具
func (s *MCPServer) registerStatusTools() {
	// get_all_task_status - 获取所有任务状态
	s.server.AddTool(mcp.NewTool("get_all_task_status",
		mcp.WithDescription("获取项目中所有任务的状态信息"),
		mcp.WithString("pathName", mcp.Description("项目路径名称"), mcp.Required()),
		mcp.WithString("status", mcp.Description("状态过滤")),
		mcp.WithBoolean("includeLockInfo", mcp.Description("是否包含锁定信息")),
	), s.handleGetAllTaskStatus)

	// get_module_task_status - 获取模块任务状态
	s.server.AddTool(mcp.NewTool("get_module_task_status",
		mcp.WithDescription("获取当前模块的所有任务状态信息"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithBoolean("includeLockInfo", mcp.Description("是否包含锁定信息")),
	), s.handleGetModuleTaskStatus)
}

// registerErrorTools 注册错误处理类工具
func (s *MCPServer) registerErrorTools() {
	// get_project_errors - 获取项目错误列表
	s.server.AddTool(mcp.NewTool("get_project_errors",
		mcp.WithDescription("返回整个项目的错误列表，遍历所有任务检查issue_details和bug_log"),
		mcp.WithString("pathName", mcp.Description("项目路径名称"), mcp.Required()),
		mcp.WithString("severity", mcp.Description("严重级别过滤：error/warning/info")),
		mcp.WithBoolean("includeDetails", mcp.Description("是否包含详细信息")),
	), s.handleGetProjectErrors)

	// get_module_errors - 获取模块错误列表
	s.server.AddTool(mcp.NewTool("get_module_errors",
		mcp.WithDescription("返回某模块的所有子任务错误列表"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithString("severity", mcp.Description("严重级别过滤")),
		mcp.WithBoolean("includeDetails", mcp.Description("是否包含详细信息")),
	), s.handleGetModuleErrors)
}

// registerCheckTools 注册检查类工具
func (s *MCPServer) registerCheckTools() {
	// check_module - 检查模块并生成报告
	s.server.AddTool(mcp.NewTool("check_module",
		mcp.WithDescription("根据 .aitdd/rule.json 中的规则检查模块并生成报告"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
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
		mcp.WithString("pathName", mcp.Description("目标任务路径名称"), mcp.Required()),
		mcp.WithString("type", mcp.Description("通知类型"), mcp.Required()),
		mcp.WithString("title", mcp.Description("通知标题"), mcp.Required()),
		mcp.WithString("message", mcp.Description("通知内容"), mcp.Required()),
	), s.handleSendNotification)

	s.server.AddTool(mcp.NewTool("read_notifications",
		mcp.WithDescription("读取未读通知"),
		mcp.WithString("pathName", mcp.Description("任务路径名称")),
	), s.handleReadNotifications)

	// 锁相关
	s.server.AddTool(mcp.NewTool("lock_resource",
		mcp.WithDescription("锁定资源"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("资源路径名称"), mcp.Required()),
	), s.handleLockResource)

	s.server.AddTool(mcp.NewTool("unlock_resource",
		mcp.WithDescription("解锁资源"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("资源路径名称"), mcp.Required()),
	), s.handleUnlockResource)

	s.server.AddTool(mcp.NewTool("get_lock_status",
		mcp.WithDescription("查询锁定状态"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("资源路径名称"), mcp.Required()),
	), s.handleGetLockStatus)

	// 模块依赖相关
	s.server.AddTool(mcp.NewTool("create_module_dependency",
		mcp.WithDescription("创建模块依赖"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
		mcp.WithString("dependsOnPathName", mcp.Description("被依赖模块的路径名称"), mcp.Required()),
		mcp.WithString("dependencyType", mcp.Description("依赖类型")),
		mcp.WithString("contractSummary", mcp.Description("契约摘要")),
	), s.handleCreateModuleDependency)

	s.server.AddTool(mcp.NewTool("get_module_dependencies",
		mcp.WithDescription("获取模块依赖"),
		mcp.WithString("pathName", mcp.Description("模块路径名称"), mcp.Required()),
	), s.handleGetModuleDependencies)

	// 任务依赖相关
	s.server.AddTool(mcp.NewTool("create_task_dependency",
		mcp.WithDescription("创建任务依赖"),
		mcp.WithString("upstreamPathName", mcp.Description("上游任务路径名称"), mcp.Required()),
		mcp.WithString("downstreamPathName", mcp.Description("下游任务路径名称"), mcp.Required()),
		mcp.WithString("contractSummary", mcp.Description("契约摘要")),
	), s.handleCreateTaskDependency)

	s.server.AddTool(mcp.NewTool("get_task_dependencies",
		mcp.WithDescription("获取任务依赖"),
		mcp.WithString("pathName", mcp.Description("任务路径名称"), mcp.Required()),
	), s.handleGetTaskDependencies)
}

// RunStdio 运行 MCP 服务器 (Stdio 模式)
func (s *MCPServer) RunStdio() error {
	log.Println("Starting AITDD MCP Server (Stdio mode)...")
	return server.ServeStdio(s.server)
}

// NewSSEServer 创建 SSE 服务器
// 返回一个可以集成到 HTTP 路由中的 SSE 服务器
func (s *MCPServer) NewSSEServer(basePath string) *server.SSEServer {
	return server.NewSSEServer(s.server,
		server.WithBasePath(basePath),
	)
}

// GetMCPServer 获取底层的 mcp-go 服务器实例
// 用于更高级的自定义配置
func (s *MCPServer) GetMCPServer() *server.MCPServer {
	return s.server
}

// RegisterSSERoutes 将 SSE 路由注册到现有的 http.ServeMux
// basePath 是 SSE 端点的基础路径，例如 "/mcp"
func (s *MCPServer) RegisterSSERoutes(mux *http.ServeMux, basePath string) {
	sseServer := s.NewSSEServer(basePath)
	// 使用 SSEHandler() 和 MessageHandler() 方法注册端点
	// 这样可以避免 SSEServer.ServeHTTP 内部的路径匹配问题
	mux.Handle(basePath+"/sse", sseServer.SSEHandler())
	mux.Handle(basePath+"/message", sseServer.MessageHandler())
}

// RegisterSSERoutesGin 将 SSE 路由直接注册到 Gin 路由器
// 这种方式更可靠，避免了 http.ServeMux 与 Gin 路由的冲突
func (s *MCPServer) RegisterSSERoutesGin(router *gin.Engine) {
	sseServer := s.NewSSEServer("/mcp")
	log.Println("正在注册 MCP SSE 路由...")
	// 注册 SSE 端点 - 使用 wrapHandler 将 http.Handler 转换为 Gin 处理器
	router.GET("/mcp/sse", wrapHandler(sseServer.SSEHandler()))
	router.POST("/mcp/message", wrapHandler(sseServer.MessageHandler()))
	log.Println("MCP SSE 路由注册完成: GET /mcp/sse, POST /mcp/message")
}

// wrapHandler 将 http.Handler 包装为 Gin 处理函数
func wrapHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// ==================== 辅助函数 ====================

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

// ==================== 工具处理函数存根 ====================
// 这些函数将在 tools_*.go 文件中实现

func (s *MCPServer) handleInitProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleInitProjectImpl(ctx, request)
}

func (s *MCPServer) handleGetConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetConfigImpl(ctx, request)
}

func (s *MCPServer) handleSetProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleSetProjectImpl(ctx, request)
}

func (s *MCPServer) handleGetProjectInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetProjectInfoImpl(ctx, request)
}

func (s *MCPServer) handleGetAllTaskCodePaths(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetAllTaskCodePathsImpl(ctx, request)
}

func (s *MCPServer) handleGetAllModules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetAllModulesImpl(ctx, request)
}

func (s *MCPServer) handleGetModuleTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetModuleTasksImpl(ctx, request)
}

func (s *MCPServer) handleGetTaskDetail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetTaskDetailImpl(ctx, request)
}

func (s *MCPServer) handleGetTaskContracts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetTaskContractsImpl(ctx, request)
}

func (s *MCPServer) handleGetConstitution(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetConstitutionImpl(ctx, request)
}

func (s *MCPServer) handleGetModuleOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetModuleOverviewImpl(ctx, request)
}

func (s *MCPServer) handleGetModuleTaskPathNames(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetModuleTaskPathNamesImpl(ctx, request)
}

func (s *MCPServer) handleDeleteModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteModuleImpl(ctx, request)
}

func (s *MCPServer) handleCreateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateModuleImpl(ctx, request)
}

func (s *MCPServer) handleUpdateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateModuleImpl(ctx, request)
}

func (s *MCPServer) handleDeleteModuleTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteModuleTasksImpl(ctx, request)
}

func (s *MCPServer) handleCreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateTaskImpl(ctx, request)
}

func (s *MCPServer) handleUpdateModuleFull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateModuleFullImpl(ctx, request)
}

func (s *MCPServer) handleUpdateTaskFull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateTaskFullImpl(ctx, request)
}

func (s *MCPServer) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateTaskImpl(ctx, request)
}

func (s *MCPServer) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteTaskImpl(ctx, request)
}

func (s *MCPServer) handleGetAllTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetAllTaskStatusImpl(ctx, request)
}

func (s *MCPServer) handleGetModuleTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetModuleTaskStatusImpl(ctx, request)
}

func (s *MCPServer) handleGetProjectErrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetProjectErrorsImpl(ctx, request)
}

func (s *MCPServer) handleGetModuleErrors(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetModuleErrorsImpl(ctx, request)
}

func (s *MCPServer) handleCheckModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckModuleImpl(ctx, request)
}

func (s *MCPServer) handleOpenFrontend(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleOpenFrontendImpl(ctx, request)
}

func (s *MCPServer) handleSendNotification(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleSendNotificationImpl(ctx, request)
}

func (s *MCPServer) handleReadNotifications(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleReadNotificationsImpl(ctx, request)
}

func (s *MCPServer) handleLockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleLockResourceImpl(ctx, request)
}

func (s *MCPServer) handleUnlockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUnlockResourceImpl(ctx, request)
}

func (s *MCPServer) handleGetLockStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetLockStatusImpl(ctx, request)
}

func (s *MCPServer) handleCreateModuleDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateModuleDependencyImpl(ctx, request)
}

func (s *MCPServer) handleGetModuleDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetModuleDependenciesImpl(ctx, request)
}

func (s *MCPServer) handleCreateTaskDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateTaskDependencyImpl(ctx, request)
}

func (s *MCPServer) handleGetTaskDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetTaskDependenciesImpl(ctx, request)
}
