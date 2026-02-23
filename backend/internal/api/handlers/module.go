package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/api"
	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetModules 获取模块列表
func GetModules(c *gin.Context) {
	var modules []models.Module

	query := database.DB.Model(&models.Module{})

	// 过滤条件
	if projectID := c.Query("projectId"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if parentID := c.Query("parentId"); parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&modules).Error; err != nil {
		api.InternalError(c, "查询模块失败")
		return
	}

	api.Success(c, gin.H{
		"modules": modules,
		"total":   len(modules),
	})
}

// GetModule 获取单个模块
func GetModule(c *gin.Context) {
	id := c.Param("id")

	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		api.NotFound(c, "模块不存在")
		return
	}

	api.Success(c, gin.H{
		"module": module,
	})
}

// CreateModule 创建模块
func CreateModule(c *gin.Context) {
	var req struct {
		ProjectID   string  `json:"projectId" binding:"required"`
		ParentID    *string `json:"parentId"`
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Prompt      string  `json:"prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	module := models.Module{
		ID:          uuid.New().String(),
		ParentID:    req.ParentID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		Prompt:      req.Prompt,
		Status:      models.ModuleStatusDesigning,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		Version:     1,
		SyncStatus:  models.SyncStatusSynced,
	}

	if err := database.DB.Create(&module).Error; err != nil {
		api.InternalError(c, "创建模块失败")
		return
	}

	api.Success(c, gin.H{
		"module": module,
	})
}

// UpdateModule 更新模块
func UpdateModule(c *gin.Context) {
	id := c.Param("id")

	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		api.NotFound(c, "模块不存在")
		return
	}

	var req struct {
		Name                     *string  `json:"name"`
		Description              *string  `json:"description"`
		Prompt                   *string  `json:"prompt"`
		Status                   *string  `json:"status"`
		TestCoverage             *float64 `json:"testCoverage"`
		UpstreamContractSummary  *string  `json:"upstreamContractSummary"`
		DownstreamContractSummary *string `json:"downstreamContractSummary"`
		Version                  int      `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查
	if module.Version != req.Version {
		api.VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 更新字段
	if req.Name != nil {
		module.Name = *req.Name
	}
	if req.Description != nil {
		module.Description = *req.Description
	}
	if req.Prompt != nil {
		module.Prompt = *req.Prompt
	}
	if req.Status != nil {
		module.Status = *req.Status
	}
	if req.TestCoverage != nil {
		module.TestCoverage = *req.TestCoverage
	}
	if req.UpstreamContractSummary != nil {
		module.UpstreamContractSummary = *req.UpstreamContractSummary
	}
	if req.DownstreamContractSummary != nil {
		module.DownstreamContractSummary = *req.DownstreamContractSummary
	}

	module.Version++
	module.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&module).Error; err != nil {
		api.InternalError(c, "更新模块失败")
		return
	}

	api.Success(c, gin.H{
		"module": module,
	})
}

// DeleteModule 删除模块
func DeleteModule(c *gin.Context) {
	id := c.Param("id")

	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		api.NotFound(c, "模块不存在")
		return
	}

	// 检查是否锁定
	if module.Locked {
		api.Locked(c, "模块已被锁定，无法删除", gin.H{
			"lockedBy": module.LockedBy,
		})
		return
	}

	if err := database.DB.Delete(&module).Error; err != nil {
		api.InternalError(c, "删除模块失败")
		return
	}

	api.Success(c, gin.H{
		"deleted": true,
	})
}

// GetModuleTasks 获取模块下的任务
func GetModuleTasks(c *gin.Context) {
	moduleID := c.Param("id")

	var tasks []models.Task
	if err := database.DB.Where("module_id = ?", moduleID).Find(&tasks).Error; err != nil {
		api.InternalError(c, "查询任务失败")
		return
	}

	api.Success(c, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}
