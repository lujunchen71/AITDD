package handlers

import (
	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// PortData 端口数据
type PortData struct {
	ID               string   `json:"id"`
	TaskID           string   `json:"taskId"`
	TaskName         string   `json:"taskName"`
	Direction        string   `json:"direction"` // "input" or "output"
	ConnectedModules []string `json:"connectedModules"`
}

// GetModulePorts 获取模块的输入输出端口
func GetModulePorts(c *gin.Context) {
	moduleID := c.Param("id")

	// 获取模块下的所有任务
	var tasks []models.Task
	if err := database.DB.Where("module_id = ?", moduleID).Find(&tasks).Error; err != nil {
		InternalError(c, "查询任务失败")
		return
	}

	if len(tasks) == 0 {
		Success(c, gin.H{
			"inputPorts":  []PortData{},
			"outputPorts": []PortData{},
		})
		return
	}

	// 获取任务 ID 列表
	taskIDs := make([]string, len(tasks))
	taskMap := make(map[string]models.Task)
	for i, task := range tasks {
		taskIDs[i] = task.ID
		taskMap[task.ID] = task
	}

	// 获取模块 ID 列表（用于判断依赖是否在模块内）
	moduleIDSet := make(map[string]bool)
	for _, task := range tasks {
		moduleIDSet[task.ModuleID] = true
	}

	// 获取任务依赖
	var dependencies []models.Dependency
	if err := database.DB.Where("upstream_task_id IN ? OR downstream_task_id IN ?", taskIDs, taskIDs).
		Find(&dependencies).Error; err != nil {
		InternalError(c, "查询依赖失败")
		return
	}

	// 计算输入输出端口
	inputPorts := []PortData{}
	outputPorts := []PortData{}

	for _, dep := range dependencies {
		upstreamTask, upstreamExists := taskMap[dep.UpstreamTaskID]
		downstreamTask, downstreamExists := taskMap[dep.DownstreamTaskID]

		if !upstreamExists || !downstreamExists {
			// 获取缺失的任务信息
			if !upstreamExists {
				if err := database.DB.First(&upstreamTask, "id = ?", dep.UpstreamTaskID).Error; err != nil {
					continue
				}
			}
			if !downstreamExists {
				if err := database.DB.First(&downstreamTask, "id = ?", dep.DownstreamTaskID).Error; err != nil {
					continue
				}
			}
		}

		upstreamInModule := moduleIDSet[upstreamTask.ModuleID]
		downstreamInModule := moduleIDSet[downstreamTask.ModuleID]

		if upstreamInModule && !downstreamInModule {
			// 输出端口：上游在当前模块，下游在其他模块
			outputPorts = append(outputPorts, PortData{
				ID:               "output-" + dep.ID,
				TaskID:           upstreamTask.ID,
				TaskName:         upstreamTask.Name,
				Direction:        "output",
				ConnectedModules: []string{downstreamTask.ModuleID},
			})
		} else if !upstreamInModule && downstreamInModule {
			// 输入端口：上游在其他模块，下游在当前模块
			inputPorts = append(inputPorts, PortData{
				ID:               "input-" + dep.ID,
				TaskID:           downstreamTask.ID,
				TaskName:         downstreamTask.Name,
				Direction:        "input",
				ConnectedModules: []string{upstreamTask.ModuleID},
			})
		}
	}

	Success(c, gin.H{
		"inputPorts":  inputPorts,
		"outputPorts": outputPorts,
	})
}

// GetProjectGraph 获取完整的项目图表数据
func GetProjectGraph(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		BadRequest(c, "缺少 projectId 参数")
		return
	}

	// 获取所有模块
	var modules []models.Module
	if err := database.DB.Where("project_id = ?", projectID).Find(&modules).Error; err != nil {
		InternalError(c, "查询模块失败")
		return
	}

	// 获取模块 ID 列表
	moduleIDs := make([]string, len(modules))
	for i, module := range modules {
		moduleIDs[i] = module.ID
	}

	// 获取所有任务
	var tasks []models.Task
	if err := database.DB.Where("module_id IN ?", moduleIDs).Find(&tasks).Error; err != nil {
		InternalError(c, "查询任务失败")
		return
	}

	// 获取任务 ID 列表
	taskIDs := make([]string, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}

	// 获取所有任务依赖
	var taskDependencies []models.Dependency
	if err := database.DB.Where("upstream_task_id IN ? OR downstream_task_id IN ?", taskIDs, taskIDs).
		Find(&taskDependencies).Error; err != nil {
		InternalError(c, "查询任务依赖失败")
		return
	}

	// 获取所有模块依赖
	var moduleDependencies []models.ModuleDependency
	if err := database.DB.Where("module_id IN ?", moduleIDs).
		Find(&moduleDependencies).Error; err != nil {
		InternalError(c, "查询模块依赖失败")
		return
	}

	Success(c, gin.H{
		"modules":            modules,
		"tasks":              tasks,
		"taskDependencies":   taskDependencies,
		"moduleDependencies": moduleDependencies,
	})
}
