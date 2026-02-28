package mcp

import (
	"context"
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
	cache         *MCPCache
	logger        *MCPLogger
}

// NewMCPServer 创建 MCP 服务器
func NewMCPServer() *MCPServer {
	s := &MCPServer{
		configManager: GetConfigManager(),
		ruleEngine:    GetRuleEngine(),
		cache:         GetMCPCache(),
		logger:        GetMCPLogger(),
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
	// ==================== 1. 项目上下文 (2个) ====================
	s.registerContextTools()

	// ==================== 2. 信息查询 (3个) ====================
	s.registerQueryTools()

	// ==================== 3. 验证检查 (3个) ====================
	s.registerCheckTools()

	// ==================== 4. 节点操作 (3个) ====================
	s.registerNodeTools()

	// ==================== 5. 依赖管理 (3个) ====================
	s.registerDependencyTools()

	// ==================== 6. 问答系统 (4个) ====================
	s.registerIssueTools()

	// ==================== 7. 锁管理 (2个) ====================
	s.registerLockTools()

	// ==================== 8. 编译接口 (2个) ====================
	s.registerCompileTools()

	// ==================== 9. 配置和规则管理 (4个) ====================
	s.registerConfigRuleTools()

	// ==================== 10. 状态管理 (1个) ====================
	s.registerStatusTools()
}

// ==================== 1. 项目上下文工具 ====================
func (s *MCPServer) registerContextTools() {
	// init_project - 初始化/设置当前项目
	s.server.AddTool(mcp.NewTool("init_project",
		mcp.WithDescription("初始化/设置当前项目。不传 pathName 则列出所有项目供选择；传入 pathName 则设置为当前项目。"),
		mcp.WithString("pathName", mcp.Description("项目路径名称（可选，不传则返回项目列表供选择）")),
	), s.handleInitProject)

	// get_context - 获取当前上下文
	s.server.AddTool(mcp.NewTool("get_context",
		mcp.WithDescription("获取当前项目上下文信息，包括项目配置、当前状态、最近操作的模块/任务等。"),
	), s.handleGetContext)
}

// ==================== 2. 信息查询工具 ====================
func (s *MCPServer) registerQueryTools() {
	// query_module - 查询模块
	s.server.AddTool(mcp.NewTool("query_module",
		mcp.WithDescription(`查询模块信息。

可查询字段: name, description, status, prompt, upstreamContractSummary, downstreamContractSummary, testCoverage, locked, version

fields 不填返回所有字段。`),
		mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
		mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"status\"]")),
	), s.handleQueryModule)

	// query_task - 查询任务
	s.server.AddTool(mcp.NewTool("query_task",
		mcp.WithDescription(`查询任务信息。

可查询字段: name, description, status, prompt, upstreamContractDetail, downstreamContractDetail, tests, testResult, codePaths, bugLog, humanAssistance, issueDetails, locked, version

fields 不填返回所有字段。`),
		mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
		mcp.WithArray("fields", mcp.Description("指定返回字段，如 [\"name\",\"tests\"]")),
	), s.handleQueryTask)

	// query_project_index_tree - 查询项目计划索引树
	s.server.AddTool(mcp.NewTool("query_project_index_tree",
		mcp.WithDescription("查询项目计划索引树，返回项目和模块、任务的树形结构，包含名称、类型和状态。返回格式为简洁的树形文本。"),
		mcp.WithString("pathName", mcp.Description("起始 pathName，不传则返回整个项目树")),
	), s.handleQueryProjectIndexTree)

	// query_file_code_path_tree - 查询代码文件路径树
	s.server.AddTool(mcp.NewTool("query_file_code_path_tree",
		mcp.WithDescription("查询任务或模块关联的代码文件路径树，显示代码文件与任务的对应关系。"),
		mcp.WithString("pathName", mcp.Description("任务或模块的 pathName，不传则返回整个项目的代码文件列表")),
	), s.handleQueryFileCodePathTree)
}

// ==================== 3. 验证检查工具 ====================
func (s *MCPServer) registerCheckTools() {
	// check_contract_alignment - 检查上下游任务契约是否对齐
	s.server.AddTool(mcp.NewTool("check_contract_alignment",
		mcp.WithDescription("检查上下游任务契约是否对齐。比较上游任务的 downstreamContractDetail 与下游任务的 upstreamContractDetail，返回差异列表。"),
		mcp.WithString("upstreamPathName", mcp.Description("上游任务 pathName"), mcp.Required()),
		mcp.WithString("downstreamPathName", mcp.Description("下游任务 pathName"), mcp.Required()),
	), s.handleCheckContractAlignment)

	// check_task_readiness - 检查任务是否准备好开始开发
	s.server.AddTool(mcp.NewTool("check_task_readiness",
		mcp.WithDescription("检查任务是否准备好开始开发。检查项包括：prompt 是否为空(error)、tests 是否为空(warning)、上游契约是否存在(error)、是否被锁定(info)、上游任务是否完成(warning)。"),
		mcp.WithString("pathName", mcp.Description("任务 pathName"), mcp.Required()),
	), s.handleCheckTaskReadiness)

	// check_dependencies - 检查依赖关系和阻塞状态
	s.server.AddTool(mcp.NewTool("check_dependencies",
		mcp.WithDescription("检查任务或模块的依赖关系和阻塞状态。返回是否有循环依赖、是否被阻塞、阻塞方列表、上下游依赖列表。"),
		mcp.WithString("pathName", mcp.Description("任务或模块 pathName"), mcp.Required()),
		mcp.WithString("type", mcp.Description("类型：task 或 module"), mcp.Required()),
	), s.handleCheckDependencies)
}

// ==================== 4. 节点操作工具 ====================
func (s *MCPServer) registerNodeTools() {
	// create_node - 统一创建节点
	s.server.AddTool(mcp.NewTool("create_node",
		mcp.WithDescription("统一创建节点接口。根据 type 创建模块或任务。type='module' 时创建模块，type='task' 时创建任务。"),
		mcp.WithString("type", mcp.Description("节点类型：module 或 task"), mcp.Required()),
		mcp.WithString("parentPath", mcp.Description("父节点 pathName。创建模块时为项目或父模块 pathName，创建任务时为所属模块 pathName"), mcp.Required()),
		mcp.WithString("name", mcp.Description("节点名称"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("节点 pathName（可选，不传则根据 parentPath 和 name 自动生成）")),
		mcp.WithObject("data", mcp.Description("类型相关的具体字段。module: description, prompt, upstreamContractSummary, downstreamContractSummary；task: description, prompt, upstreamContractDetail, downstreamContractDetail, tests, codePaths")),
	), s.handleCreateNode)

	// modify_module - 修改模块
	s.server.AddTool(mcp.NewTool("modify_module",
		mcp.WithDescription(`修改模块。只更新 data 中传入的字段。

可修改字段: name(string), description(string), status(designing|developing|completed|deprecated), prompt(string), upstreamContractSummary(string), downstreamContractSummary(string), testCoverage(0-100)`),
		mcp.WithString("pathName", mcp.Description("模块路径，格式: 项目名/模块名"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
		mcp.WithObject("data", mcp.Description("要修改的字段，如 {\"description\":\"新描述\",\"status\":\"developing\"}"), mcp.Required()),
	), s.handleModifyModule)

	// modify_task - 修改任务
	s.server.AddTool(mcp.NewTool("modify_task",
		mcp.WithDescription(`修改任务。只更新 data 中传入的字段。

可修改字段:
- name, description, status, prompt: string
- upstreamContractDetail/downstreamContractDetail: {title, list:[{label,contract_api,from}]}
- tests: [{target:string, api:string}]
- testResult/codePaths/bugLog: string[]
- humanAssistance: object
- issueDetails: string`),
		mcp.WithString("pathName", mcp.Description("任务路径，格式: 项目名/模块名/任务名"), mcp.Required()),
		mcp.WithNumber("version", mcp.Description("当前版本号，用于乐观锁"), mcp.Required()),
		mcp.WithObject("data", mcp.Description("要修改的字段，如 {\"status\":\"in_progress\",\"tests\":[{\"target\":\"验证XX\",\"api\":\"test_xx()\"}]}"), mcp.Required()),
	), s.handleModifyTask)

	// delete_node - 统一删除节点
	s.server.AddTool(mcp.NewTool("delete_node",
		mcp.WithDescription("统一删除节点接口。根据 path 自动识别类型（module/task）。删除模块时级联删除其下所有任务。"),
		mcp.WithString("path", mcp.Description("节点 pathName"), mcp.Required()),
		mcp.WithBoolean("force", mcp.Description("强制删除（忽略依赖）")),
	), s.handleDeleteNode)
}

// ==================== 5. 依赖管理工具 ====================
func (s *MCPServer) registerDependencyTools() {
	// create_dependency - 统一创建依赖
	s.server.AddTool(mcp.NewTool("create_dependency",
		mcp.WithDescription("统一创建依赖接口。根据 type 创建模块或任务依赖。type='module' 时创建模块依赖，type='task' 时创建任务依赖。"),
		mcp.WithString("type", mcp.Description("依赖类型：module 或 task"), mcp.Required()),
		mcp.WithString("upstreamPath", mcp.Description("上游资源 pathName（被依赖方）"), mcp.Required()),
		mcp.WithString("downstreamPath", mcp.Description("下游资源 pathName（依赖方）"), mcp.Required()),
		mcp.WithString("dependencyType", mcp.Description("依赖类型：required / optional / conditional")),
		mcp.WithString("contractSummary", mcp.Description("契约摘要")),
	), s.handleCreateDependency)

	// delete_dependency - 统一删除依赖
	s.server.AddTool(mcp.NewTool("delete_dependency",
		mcp.WithDescription("统一删除依赖接口。根据 type 删除模块或任务依赖。"),
		mcp.WithString("type", mcp.Description("依赖类型：module 或 task"), mcp.Required()),
		mcp.WithString("upstreamPath", mcp.Description("上游资源 pathName（被依赖方）"), mcp.Required()),
		mcp.WithString("downstreamPath", mcp.Description("下游资源 pathName（依赖方）"), mcp.Required()),
	), s.handleDeleteDependency)

	// query_dependencies - 统一查询依赖
	s.server.AddTool(mcp.NewTool("query_dependencies",
		mcp.WithDescription("统一查询依赖接口。根据 type 查询模块或任务的依赖关系，支持指定方向（上游/下游/双向）。"),
		mcp.WithString("type", mcp.Description("依赖类型：module 或 task"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("资源 pathName"), mcp.Required()),
		mcp.WithString("direction", mcp.Description("查询方向：upstream（上游）/ downstream（下游）/ both（双向，默认）")),
	), s.handleQueryDependencies)
}

// ==================== 6. 问答系统工具 ====================
func (s *MCPServer) registerIssueTools() {
	// create_issue - 创建问题
	s.server.AddTool(mcp.NewTool("create_issue",
		mcp.WithDescription("创建问题（下游任务发现上游有问题时发起）"),
		mcp.WithString("fromTaskPathName", mcp.Description("发起方任务 pathName"), mcp.Required()),
		mcp.WithString("toTaskPathName", mcp.Description("接收方任务 pathName"), mcp.Required()),
		mcp.WithString("type", mcp.Description("问题类型: contract/test/other"), mcp.Required()),
		mcp.WithString("title", mcp.Description("问题标题"), mcp.Required()),
		mcp.WithString("content", mcp.Description("问题内容"), mcp.Required()),
	), s.handleCreateIssue)

	// reply_issue - 回复问题
	s.server.AddTool(mcp.NewTool("reply_issue",
		mcp.WithDescription("回复问题"),
		mcp.WithString("fromTaskPathName", mcp.Description("发起方任务 pathName"), mcp.Required()),
		mcp.WithString("toTaskPathName", mcp.Description("接收方任务 pathName"), mcp.Required()),
		mcp.WithString("title", mcp.Description("问题标题（定位用）"), mcp.Required()),
		mcp.WithString("replyContent", mcp.Description("回复内容"), mcp.Required()),
	), s.handleReplyIssue)

	// resolve_issue - 解决问题
	s.server.AddTool(mcp.NewTool("resolve_issue",
		mcp.WithDescription("解决问题"),
		mcp.WithString("fromTaskPathName", mcp.Description("发起方任务 pathName"), mcp.Required()),
		mcp.WithString("toTaskPathName", mcp.Description("接收方任务 pathName"), mcp.Required()),
		mcp.WithString("title", mcp.Description("问题标题（定位用）"), mcp.Required()),
	), s.handleResolveIssue)

	// query_issues - 查询问题列表
	s.server.AddTool(mcp.NewTool("query_issues",
		mcp.WithDescription("查询问题列表"),
		mcp.WithString("taskPathName", mcp.Description("任务 pathName 过滤")),
		mcp.WithString("direction", mcp.Description("方向: from/to/both（默认 both）")),
		mcp.WithString("status", mcp.Description("状态过滤: pending/replied/resolved")),
		mcp.WithString("type", mcp.Description("类型过滤: contract/test/other")),
	), s.handleQueryIssues)
}

// ==================== 7. 锁管理工具 ====================
func (s *MCPServer) registerLockTools() {
	// acquire_lock - 获取锁
	s.server.AddTool(mcp.NewTool("acquire_lock",
		mcp.WithDescription("锁定资源（模块或任务），防止并发修改冲突。"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("资源路径名称"), mcp.Required()),
	), s.handleLockResource)

	// release_lock - 释放锁
	s.server.AddTool(mcp.NewTool("release_lock",
		mcp.WithDescription("释放资源锁。"),
		mcp.WithString("resourceType", mcp.Description("资源类型：module 或 task"), mcp.Required()),
		mcp.WithString("pathName", mcp.Description("资源路径名称"), mcp.Required()),
	), s.handleUnlockResource)
}

// ==================== 8. 编译接口工具 ====================
func (s *MCPServer) registerCompileTools() {
	// compile_static - 静态编译
	s.server.AddTool(mcp.NewTool("compile_static",
		mcp.WithDescription("静态编译：检查项目结构完整性、契约对齐、依赖关系等，生成编译报告。"),
		mcp.WithString("pathName", mcp.Description("项目或模块 pathName"), mcp.Required()),
		mcp.WithBoolean("includeWarnings", mcp.Description("是否包含警告")),
	), s.handleCompileStatic)

	// compile_dynamic - 动态编译
	s.server.AddTool(mcp.NewTool("compile_dynamic",
		mcp.WithDescription("动态编译：执行实际代码生成、测试运行等，生成执行报告。"),
		mcp.WithString("pathName", mcp.Description("项目或模块 pathName"), mcp.Required()),
		mcp.WithBoolean("runTests", mcp.Description("是否运行测试")),
	), s.handleCompileDynamic)
}

// ==================== 9. 配置和规则管理工具 ====================
func (s *MCPServer) registerConfigRuleTools() {
	// get_config - 获取配置
	s.server.AddTool(mcp.NewTool("get_config",
		mcp.WithDescription("获取当前 .aitdd/project.json 配置文件的内容"),
	), s.handleGetConfig)

	// update_config - 更新配置
	s.server.AddTool(mcp.NewTool("update_config",
		mcp.WithDescription("更新 .aitdd/project.json 配置文件"),
		mcp.WithObject("config", mcp.Description("配置对象"), mcp.Required()),
	), s.handleUpdateConfig)

	// get_rule - 获取规则
	s.server.AddTool(mcp.NewTool("get_rule",
		mcp.WithDescription("获取 .aitdd/rule.json 规则文件的内容"),
	), s.handleGetRule)

	// update_rule - 更新规则
	s.server.AddTool(mcp.NewTool("update_rule",
		mcp.WithDescription("更新 .aitdd/rule.json 规则文件"),
		mcp.WithObject("rule", mcp.Description("规则对象"), mcp.Required()),
	), s.handleUpdateRule)
}

// ==================== 10. 状态管理工具 ====================
func (s *MCPServer) registerStatusTools() {
	// get_status - 获取状态
	s.server.AddTool(mcp.NewTool("get_status",
		mcp.WithDescription("获取项目、模块或任务的状态信息，包括进度、错误、警告等。"),
		mcp.WithString("pathName", mcp.Description("节点 pathName，不传则返回整个项目状态")),
		mcp.WithString("type", mcp.Description("类型：project/module/task，不传则根据 pathName 自动识别")),
		mcp.WithBoolean("includeErrors", mcp.Description("是否包含错误信息")),
		mcp.WithBoolean("includeLockInfo", mcp.Description("是否包含锁定信息")),
	), s.handleGetStatus)
}

// RunStdio 运行 MCP 服务器 (Stdio 模式)
func (s *MCPServer) RunStdio() error {
	s.logger.Info("Starting AITDD MCP Server (Stdio mode)...")
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
	s.logger.Info("Registering MCP SSE routes...")
	// 注册 SSE 端点 - 使用 wrapHandler 将 http.Handler 转换为 Gin 处理器
	router.GET("/mcp/sse", wrapHandler(sseServer.SSEHandler()))
	router.POST("/mcp/message", wrapHandler(sseServer.MessageHandler()))
	s.logger.Info("MCP SSE routes registered: GET /mcp/sse, POST /mcp/message")
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

// 1. 项目上下文
func (s *MCPServer) handleInitProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleInitProjectImpl(ctx, request)
}

func (s *MCPServer) handleGetContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetContextImpl(ctx, request)
}

// 2. 信息查询
func (s *MCPServer) handleQueryModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQueryModuleImpl(ctx, request)
}

func (s *MCPServer) handleQueryTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQueryTaskImpl(ctx, request)
}

func (s *MCPServer) handleQueryProjectIndexTree(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQueryProjectIndexTreeImpl(ctx, request)
}

func (s *MCPServer) handleQueryFileCodePathTree(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQueryFileCodePathTreeImpl(ctx, request)
}

// 3. 验证检查
func (s *MCPServer) handleCheckContractAlignment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckContractAlignmentImpl(ctx, request)
}

func (s *MCPServer) handleCheckTaskReadiness(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckTaskReadinessImpl(ctx, request)
}

func (s *MCPServer) handleCheckDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckDependenciesImpl(ctx, request)
}

// 4. 节点操作
func (s *MCPServer) handleCreateNode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateNodeImpl(ctx, request)
}

func (s *MCPServer) handleModifyModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleModifyModuleImpl(ctx, request)
}

func (s *MCPServer) handleModifyTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleModifyTaskImpl(ctx, request)
}

func (s *MCPServer) handleDeleteNode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteNodeImpl(ctx, request)
}

// 5. 依赖管理
func (s *MCPServer) handleCreateDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateDependencyImpl(ctx, request)
}

func (s *MCPServer) handleDeleteDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleDeleteDependencyImpl(ctx, request)
}

func (s *MCPServer) handleQueryDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQueryDependenciesImpl(ctx, request)
}

// 6. 问答系统
func (s *MCPServer) handleCreateIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreateIssueImpl(ctx, request)
}

func (s *MCPServer) handleReplyIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleReplyIssueImpl(ctx, request)
}

func (s *MCPServer) handleResolveIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleResolveIssueImpl(ctx, request)
}

func (s *MCPServer) handleQueryIssues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleQueryIssuesImpl(ctx, request)
}

// 7. 锁管理
func (s *MCPServer) handleLockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleLockResourceImpl(ctx, request)
}

func (s *MCPServer) handleUnlockResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUnlockResourceImpl(ctx, request)
}

// 8. 编译接口
func (s *MCPServer) handleCompileStatic(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCompileStaticImpl(ctx, request)
}

func (s *MCPServer) handleCompileDynamic(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCompileDynamicImpl(ctx, request)
}

// 9. 配置和规则管理
func (s *MCPServer) handleGetConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetConfigImpl(ctx, request)
}

func (s *MCPServer) handleUpdateConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateConfigImpl(ctx, request)
}

func (s *MCPServer) handleGetRule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetRuleImpl(ctx, request)
}

func (s *MCPServer) handleUpdateRule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdateRuleImpl(ctx, request)
}

// 10. 状态管理
func (s *MCPServer) handleGetStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleGetStatusImpl(ctx, request)
}
