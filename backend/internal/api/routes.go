package api

import (
	"net/http"

	"github.com/aitdd/backend/internal/api/handlers"
	"github.com/aitdd/backend/internal/api/middleware"
	"github.com/aitdd/backend/internal/mcp"
	"github.com/gin-gonic/gin"
)

// 全局 MCP 服务器实例
var mcpServer = mcp.NewMCPServer()

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORS())

	// 注册 MCP SSE 路由
	registerMCPRoutes(r)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 项目相关
		project := v1.Group("/project")
		{
			project.GET("", handlers.GetProject)
			project.GET("/constitution", handlers.GetConstitution)
			project.PUT("/constitution", handlers.UpdateConstitution)
		}

		// 多项目相关
		projects := v1.Group("/projects")
		{
			projects.GET("", handlers.GetProjects)
			projects.GET("/by-path/*pathName", handlers.GetProjectByPathName)
			projects.POST("", handlers.CreateProject)
			projects.DELETE("/:id", handlers.DeleteProject)
		}

		// 模块相关
		modules := v1.Group("/modules")
		{
			modules.GET("", handlers.GetModules)
			modules.GET("/by-path/*pathName", handlers.GetModuleByPathName)
			modules.PUT("/by-path/*pathName", handlers.UpdateModuleByPathName)
			modules.DELETE("/by-path/*pathName", handlers.DeleteModuleByPathName)
			// 使用查询参数区分操作: /by-path/path?action=tasks 或 ?action=dependencies
			modules.GET("/:id", handlers.GetModule)
			modules.POST("", handlers.CreateModule)
			modules.PUT("/:id", handlers.UpdateModule)
			modules.DELETE("/:id", handlers.DeleteModule)
			modules.GET("/:id/tasks", handlers.GetModuleTasks)
			// 模块位置相关
			modules.GET("/positions", handlers.GetModulePositions)
			modules.PUT("/positions", handlers.BatchUpdateModulePositions)
			modules.PUT("/:id/position", handlers.UpdateModulePosition)
			// 模块依赖相关
			modules.GET("/:id/dependencies", handlers.GetModuleDependencies)
			modules.GET("/:id/dependents", handlers.GetModuleDependents)
			modules.POST("/:id/dependencies", handlers.CreateModuleDependency)
			modules.PUT("/:id/dependencies/:depId", handlers.UpdateModuleDependency)
			modules.DELETE("/:id/dependencies/:depId", handlers.DeleteModuleDependency)
		}

		// 任务相关
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)
			tasks.GET("/by-path/*pathName", handlers.GetTaskByPathName)
			tasks.PUT("/by-path/*pathName", handlers.UpdateTaskByPathName)
			tasks.DELETE("/by-path/*pathName", handlers.DeleteTaskByPathName)
			tasks.GET("/:id", handlers.GetTask)
			tasks.POST("", handlers.CreateTask)
			tasks.PUT("/:id", handlers.UpdateTask)
			tasks.DELETE("/:id", handlers.DeleteTask)
			tasks.POST("/:id/check", handlers.CheckTask)
			tasks.POST("/:id/refactor", handlers.RefactorTask)
			tasks.POST("/:id/duplicate", handlers.DuplicateTask)
			tasks.POST("/:id/toggle-lock", handlers.ToggleTaskLock)
		}

		// 依赖相关
		dependencies := v1.Group("/dependencies")
		{
			dependencies.GET("", handlers.GetDependencies)
			dependencies.POST("", handlers.CreateDependency)
			dependencies.DELETE("/:id", handlers.DeleteDependency)
		}

		// 通知相关
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", handlers.GetNotifications)
			notifications.POST("", handlers.CreateNotification)
			notifications.POST("/:id/read", handlers.MarkNotificationRead)
		}

		// 锁相关
		lock := v1.Group("/lock")
		{
			lock.POST("", handlers.LockResource)
			lock.POST("/unlock", handlers.UnlockResource)
			lock.GET("/status", handlers.GetLockStatus)
			// 通过 pathName 锁定/解锁
			lock.POST("/by-path/*pathName", handlers.LockResourceByPathName)
			lock.POST("/unlock-by-path/*pathName", handlers.UnlockResourceByPathName)
			lock.GET("/status-by-path/*pathName", handlers.GetLockStatusByPathName)
		}

		// 工具相关
		tools := v1.Group("/tools")
		{
			tools.POST("/open-browser", handlers.OpenBrowser)
		}

		// 提示词版本相关
		promptVersions := v1.Group("/prompt-versions")
		{
			promptVersions.GET("/:type/:id", handlers.GetPromptVersions)
			promptVersions.GET("/:type/:id/:version", handlers.GetPromptVersion)
			promptVersions.GET("/:type/:id/latest", handlers.GetLatestPromptVersion)
			promptVersions.POST("", handlers.CreatePromptVersion)
			promptVersions.POST("/:type/:id/compare", handlers.ComparePromptVersions)
		}

		// 问题相关
		issues := v1.Group("/issues")
		{
			issues.GET("", handlers.QueryIssues)
			issues.POST("", handlers.CreateIssue)
			issues.POST("/reply", handlers.ReplyIssue)
			issues.POST("/resolve", handlers.ResolveIssue)
			issues.GET("/id/:id", handlers.GetIssueByID)
			issues.DELETE("/:id", handlers.DeleteIssue)
			issues.GET("/by-task/*pathName", handlers.GetIssuesByTask)
			// 通过关键字段获取问题（需要在最后，避免与其他路由冲突）
			issues.GET("/:fromTaskPathName/:toTaskPathName/:title", handlers.GetIssue)
		}

		// 图表相关
		graph := v1.Group("/graph")
		{
			graph.GET("/project", handlers.GetProjectGraph)
			graph.GET("/modules/:id/ports", handlers.GetModulePorts)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

// registerMCPRoutes 注册 MCP SSE 路由
func registerMCPRoutes(r *gin.Engine) {
	sseServer := mcpServer.NewSSEServer("/mcp")

	// 注册 SSE 端点
	r.GET("/mcp/sse", func(c *gin.Context) {
		sseServer.SSEHandler().ServeHTTP(c.Writer, c.Request)
	})
	r.POST("/mcp/message", func(c *gin.Context) {
		sseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
	})
}

// wrapHandler 将 http.Handler 包装为 Gin 处理函数
func wrapHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
