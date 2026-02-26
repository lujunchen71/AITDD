package api

import (
	"github.com/aitdd/backend/internal/api/handlers"
	"github.com/aitdd/backend/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORS())

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
			projects.POST("", handlers.CreateProject)
		}

		// 模块相关
		modules := v1.Group("/modules")
		{
			modules.GET("", handlers.GetModules)
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
