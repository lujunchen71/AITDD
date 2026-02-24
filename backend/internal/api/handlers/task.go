package handlers

import (
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

	// 过滤条件
	if moduleID := c.Query("moduleId"); moduleID != "" {
		query = query.Where("module_id = ?", moduleID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if assignee := c.Query("assignee"); assignee != "" {
		query = query.Where("assignee = ?", assignee)
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
		Description              string `json:"description"`
		Prompt                   string `json:"prompt"`
		UpstreamContractDetail   string `json:"upstreamContractDetail"`
		DownstreamContractDetail string `json:"downstreamContractDetail"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	task := models.Task{
		ID:                       uuid.New().String(),
		ModuleID:                 req.ModuleID,
		Name:                     req.Name,
		Description:              req.Description,
		Prompt:                   req.Prompt,
		UpstreamContractDetail:   req.UpstreamContractDetail,
		DownstreamContractDetail: req.DownstreamContractDetail,
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
	if req.Name != nil {
		task.Name = *req.Name
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
		task.Logs = *req.Logs
	}
	if req.CodePaths != nil {
		task.CodePaths = *req.CodePaths
	}
	if req.HumanAssistance != nil {
		task.HumanAssistance = *req.HumanAssistance
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
