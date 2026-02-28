package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTasks 获取任务列表
func GetTasks(c *gin.Context) {
	var tasks []models.Task

	query := database.DB.Model(&models.Task{})

	// 过滤条件 - 支持多个项目ID（逗号分隔）
	if projectIDs := c.Query("projectIds"); projectIDs != "" {
		// 通过模块关联查询任务
		query = query.Joins("JOIN modules ON modules.id = tasks.module_id").
			Where("modules.project_id IN ?", strings.Split(projectIDs, ","))
	} else if projectID := c.Query("projectId"); projectID != "" {
		// 兼容单个项目ID
		query = query.Joins("JOIN modules ON modules.id = tasks.module_id").
			Where("modules.project_id = ?", projectID)
	}
	if moduleID := c.Query("moduleId"); moduleID != "" {
		query = query.Where("tasks.module_id = ?", moduleID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("tasks.status = ?", status)
	}
	if assignee := c.Query("assignee"); assignee != "" {
		query = query.Where("tasks.assignee = ?", assignee)
	}

	// 分页
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		page = parseInt(p)
	}
	if ps := c.Query("pageSize"); ps != "" {
		pageSize = parseInt(ps)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		InternalError(c, "查询任务失败")
		return
	}

	Success(c, gin.H{
		"tasks":    tasks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetTask 获取单个任务
func GetTask(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	Success(c, gin.H{
		"task": task,
	})
}

// CreateTask 创建任务
func CreateTask(c *gin.Context) {
	var req struct {
		ModuleID                 string `json:"moduleId" binding:"required"`
		Name                     string `json:"name" binding:"required"`
		PathName                 string `json:"pathName"`
		Description              string `json:"description"`
		Prompt                   string `json:"prompt"`
		UpstreamContractDetail   string `json:"upstreamContractDetail"`
		DownstreamContractDetail string `json:"downstreamContractDetail"`
		CodePaths                string `json:"codePaths"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 获取模块信息以生成正确的 pathName
	var module models.Module
	if err := database.DB.First(&module, "id = ?", req.ModuleID).Error; err != nil {
		NotFound(c, "模块不存在")
		return
	}

	// 生成唯一的 pathName，格式：模块pathName/任务名称
	basePathName := fmt.Sprintf("%s/%s", module.PathName, req.Name)
	pathName := basePathName
	// 检查pathName是否已存在，如果存在则添加数字后缀
	suffix := 1
	for {
		var count int64
		database.DB.Model(&models.Task{}).Where("path_name = ?", pathName).Count(&count)
		if count == 0 {
			break
		}
		pathName = fmt.Sprintf("%s-%d", basePathName, suffix)
		suffix++
	}

	task := models.Task{
		ID:                       uuid.New().String(),
		ModuleID:                 req.ModuleID,
		Name:                     req.Name,
		PathName:                 pathName,
		Description:              req.Description,
		Prompt:                   req.Prompt,
		UpstreamContractDetail:   req.UpstreamContractDetail,
		DownstreamContractDetail: req.DownstreamContractDetail,
		CodePaths:                req.CodePaths,
		Status:                   models.TaskStatusReady,
		CreatedAt:                time.Now().UnixMilli(),
		UpdatedAt:                time.Now().UnixMilli(),
		Version:                  1,
		SyncStatus:               models.SyncStatusSynced,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		InternalError(c, "创建任务失败")
		return
	}

	Success(c, gin.H{
		"task": task,
	})
}

// UpdateTask 更新任务
func UpdateTask(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	var req struct {
		Name                     *string `json:"name"`
		Description              *string `json:"description"`
		Status                   *string `json:"status"`
		Assignee                 *string `json:"assignee"`
		Prompt                   *string `json:"prompt"`
		ModuleID                 *string `json:"moduleId"`
		UpstreamContractDetail   *string `json:"upstreamContractDetail"`
		DownstreamContractDetail *string `json:"downstreamContractDetail"`
		Tests                    *string `json:"tests"`
		Logs                     *string `json:"logs"`
		CodePaths                *string `json:"codePaths"`
		HumanAssistance          *string `json:"humanAssistance"`
		Version                  int     `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查
	if task.Version != req.Version {
		VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 检查是否锁定
	if task.Locked {
		Locked(c, "任务已被锁定，无法修改", gin.H{
			"lockedBy": task.LockedBy,
		})
		return
	}

	// 更新字段
	nameChanged := false
	moduleChanged := false
	if req.Name != nil && *req.Name != task.Name {
		task.Name = *req.Name
		nameChanged = true
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Assignee != nil {
		task.Assignee = req.Assignee
	}
	if req.Prompt != nil {
		task.Prompt = *req.Prompt
	}
	if req.ModuleID != nil && *req.ModuleID != task.ModuleID {
		task.ModuleID = *req.ModuleID
		moduleChanged = true
	}
	if req.UpstreamContractDetail != nil {
		task.UpstreamContractDetail = *req.UpstreamContractDetail
	}
	if req.DownstreamContractDetail != nil {
		task.DownstreamContractDetail = *req.DownstreamContractDetail
	}
	if req.Tests != nil {
		task.Tests = *req.Tests
	}
	if req.Logs != nil {
		task.BugLog = *req.Logs
	}
	if req.CodePaths != nil {
		task.CodePaths = *req.CodePaths
	}
	if req.HumanAssistance != nil {
		task.HumanAssistance = *req.HumanAssistance
	}

	// 如果名称改变或模块改变，需要更新 pathName
	if nameChanged || moduleChanged {
		// 获取模块信息以生成新的 pathName
		var module models.Module
		if err := database.DB.First(&module, "id = ?", task.ModuleID).Error; err != nil {
			BadRequest(c, "目标模块不存在")
			return
		}

		// 计算新的 pathName
		basePathName := fmt.Sprintf("%s/%s", module.PathName, task.Name)
		newPathName := basePathName
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Task{}).Where("path_name = ? AND id != ?", newPathName, task.ID).Count(&count)
			if count == 0 {
				break
			}
			newPathName = fmt.Sprintf("%s-%d", basePathName, suffix)
			suffix++
		}
		task.PathName = newPathName
	}

	task.Version++
	task.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&task).Error; err != nil {
		InternalError(c, "更新任务失败")
		return
	}

	Success(c, gin.H{
		"task": task,
	})
}

// DeleteTask 删除任务
func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	// 检查是否锁定
	if task.Locked {
		Locked(c, "任务已被锁定，无法删除", gin.H{
			"lockedBy": task.LockedBy,
		})
		return
	}

	if err := database.DB.Delete(&task).Error; err != nil {
		InternalError(c, "删除任务失败")
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}

// CheckTask 检查任务
func CheckTask(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	// TODO: 实现实际的检查逻辑
	// 这里可以检查任务的完整性、依赖关系等
	hasIssues := false
	var issues []string

	if task.Name == "" {
		hasIssues = true
		issues = append(issues, "任务名称不能为空")
	}
	if task.Prompt == "" {
		hasIssues = true
		issues = append(issues, "任务提示词不能为空")
	}

	if hasIssues {
		Success(c, gin.H{
			"success": false,
			"message": "任务检查发现问题",
			"issues":  issues,
		})
		return
	}

	Success(c, gin.H{
		"success": true,
		"message": "任务检查通过",
	})
}

// RefactorTask 重构任务
func RefactorTask(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	// 检查是否锁定
	if task.Locked {
		Locked(c, "任务已被锁定，无法重构", gin.H{
			"lockedBy": task.LockedBy,
		})
		return
	}

	// TODO: 实现实际的重构逻辑
	// 这里可以重新分析任务的结构和依赖

	task.Version++
	task.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&task).Error; err != nil {
		InternalError(c, "重构任务失败")
		return
	}

	Success(c, gin.H{
		"task":    task,
		"message": "任务重构成功",
	})
}

// DuplicateTask 复制任务
func DuplicateTask(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	// 创建新任务
	newTask := models.Task{
		ID:                       uuid.New().String(),
		ModuleID:                 task.ModuleID,
		Name:                     task.Name + " (副本)",
		Description:              task.Description,
		Prompt:                   task.Prompt,
		UpstreamContractDetail:   task.UpstreamContractDetail,
		DownstreamContractDetail: task.DownstreamContractDetail,
		Status:                   models.TaskStatusReady,
		CreatedAt:                time.Now().UnixMilli(),
		UpdatedAt:                time.Now().UnixMilli(),
		Version:                  1,
		SyncStatus:               models.SyncStatusSynced,
	}

	if err := database.DB.Create(&newTask).Error; err != nil {
		InternalError(c, "复制任务失败")
		return
	}

	Success(c, gin.H{
		"task":    newTask,
		"message": "任务复制成功",
	})
}

// ToggleTaskLock 切换任务锁定状态
func ToggleTaskLock(c *gin.Context) {
	id := c.Param("id")

	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		NotFound(c, "任务不存在")
		return
	}

	// 切换锁定状态
	task.Locked = !task.Locked
	task.UpdatedAt = time.Now().UnixMilli()

	// TODO: 从请求中获取当前用户
	// if task.Locked {
	//     task.LockedBy = currentUser
	// } else {
	//     task.LockedBy = ""
	// }

	if err := database.DB.Save(&task).Error; err != nil {
		InternalError(c, "切换锁定状态失败")
		return
	}

	Success(c, gin.H{
		"locked":  task.Locked,
		"message": func() string {
			if task.Locked {
				return "任务已锁定"
			}
			return "任务已解锁"
		}(),
	})
}

// GetTaskByPathName 通过 pathName 获取任务
func GetTaskByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var task models.Task
	if err := database.DB.First(&task, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "任务不存在: "+pathName)
		return
	}

	Success(c, gin.H{
		"task": task,
	})
}

// UpdateTaskByPathName 通过 pathName 更新任务
func UpdateTaskByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var task models.Task
	if err := database.DB.First(&task, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "任务不存在: "+pathName)
		return
	}

	var req struct {
		Name                     *string `json:"name"`
		Description              *string `json:"description"`
		Status                   *string `json:"status"`
		Assignee                 *string `json:"assignee"`
		Prompt                   *string `json:"prompt"`
		ModuleID                 *string `json:"moduleId"`
		UpstreamContractDetail   *string `json:"upstreamContractDetail"`
		DownstreamContractDetail *string `json:"downstreamContractDetail"`
		Tests                    *string `json:"tests"`
		Logs                     *string `json:"logs"`
		CodePaths                *string `json:"codePaths"`
		HumanAssistance          *string `json:"humanAssistance"`
		Version                  int     `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查
	if task.Version != req.Version {
		VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 检查是否锁定
	if task.Locked {
		Locked(c, "任务已被锁定，无法修改", gin.H{
			"lockedBy": task.LockedBy,
		})
		return
	}

	// 更新字段
	nameChanged := false
	moduleChanged := false
	if req.Name != nil && *req.Name != task.Name {
		task.Name = *req.Name
		nameChanged = true
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Assignee != nil {
		task.Assignee = req.Assignee
	}
	if req.Prompt != nil {
		task.Prompt = *req.Prompt
	}
	if req.ModuleID != nil && *req.ModuleID != task.ModuleID {
		task.ModuleID = *req.ModuleID
		moduleChanged = true
	}
	if req.UpstreamContractDetail != nil {
		task.UpstreamContractDetail = *req.UpstreamContractDetail
	}
	if req.DownstreamContractDetail != nil {
		task.DownstreamContractDetail = *req.DownstreamContractDetail
	}
	if req.Tests != nil {
		task.Tests = *req.Tests
	}
	if req.Logs != nil {
		task.BugLog = *req.Logs
	}
	if req.CodePaths != nil {
		task.CodePaths = *req.CodePaths
	}
	if req.HumanAssistance != nil {
		task.HumanAssistance = *req.HumanAssistance
	}

	// 如果名称改变或模块改变，需要更新 pathName
	if nameChanged || moduleChanged {
		// 获取模块信息以生成新的 pathName
		var module models.Module
		if err := database.DB.First(&module, "id = ?", task.ModuleID).Error; err != nil {
			BadRequest(c, "目标模块不存在")
			return
		}

		// 计算新的 pathName
		basePathName := fmt.Sprintf("%s/%s", module.PathName, task.Name)
		newPathName := basePathName
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Task{}).Where("path_name = ? AND id != ?", newPathName, task.ID).Count(&count)
			if count == 0 {
				break
			}
			newPathName = fmt.Sprintf("%s-%d", basePathName, suffix)
			suffix++
		}
		task.PathName = newPathName
	}

	task.Version++
	task.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&task).Error; err != nil {
		InternalError(c, "更新任务失败")
		return
	}

	Success(c, gin.H{
		"task": task,
	})
}

// DeleteTaskByPathName 通过 pathName 删除任务
func DeleteTaskByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var task models.Task
	if err := database.DB.First(&task, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "任务不存在: "+pathName)
		return
	}

	// 检查是否锁定
	if task.Locked {
		Locked(c, "任务已被锁定，无法删除", gin.H{
			"lockedBy": task.LockedBy,
		})
		return
	}

	if err := database.DB.Delete(&task).Error; err != nil {
		InternalError(c, "删除任务失败")
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}
