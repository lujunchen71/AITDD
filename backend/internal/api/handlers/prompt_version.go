package handlers

import (
	"strconv"

	"github.com/aitdd/backend/internal/services"
	"github.com/gin-gonic/gin"
)

var promptVersionService = services.NewPromptVersionService()

// GetPromptVersions 获取实体的所有提示词版本
func GetPromptVersions(c *gin.Context) {
	entityType := c.Param("type")
	entityID := c.Param("id")

	versions, err := promptVersionService.GetPromptVersions(entityType, entityID)
	if err != nil {
		InternalError(c, "获取提示词版本失败")
		return
	}

	Success(c, gin.H{
		"versions": versions,
		"total":    len(versions),
	})
}

// GetPromptVersion 获取指定版本
func GetPromptVersion(c *gin.Context) {
	entityType := c.Param("type")
	entityID := c.Param("id")
	versionStr := c.Param("version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		BadRequest(c, "无效的版本号")
		return
	}

	v, err := promptVersionService.GetPromptVersion(entityType, entityID, version)
	if err != nil {
		InternalError(c, "获取提示词版本失败")
		return
	}

	if v == nil {
		NotFound(c, "版本不存在")
		return
	}

	Success(c, gin.H{
		"version": v,
	})
}

// GetLatestPromptVersion 获取最新版本
func GetLatestPromptVersion(c *gin.Context) {
	entityType := c.Param("type")
	entityID := c.Param("id")

	v, err := promptVersionService.GetLatestPromptVersion(entityType, entityID)
	if err != nil {
		InternalError(c, "获取最新提示词版本失败")
		return
	}

	if v == nil {
		NotFound(c, "未找到提示词版本")
		return
	}

	Success(c, gin.H{
		"version": v,
	})
}

// CreatePromptVersion 创建提示词版本
func CreatePromptVersion(c *gin.Context) {
	var req struct {
		EntityType    string `json:"entityType" binding:"required"`
		EntityID      string `json:"entityId" binding:"required"`
		Prompt        string `json:"prompt" binding:"required"`
		ChangeSummary string `json:"changeSummary"`
		CreatedBy     string `json:"createdBy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 验证 entityType
	if req.EntityType != "module" && req.EntityType != "task" {
		BadRequest(c, "实体类型必须是'module'或'task'")
		return
	}

	v, err := promptVersionService.CreatePromptVersion(
		req.EntityType,
		req.EntityID,
		req.Prompt,
		req.ChangeSummary,
		req.CreatedBy,
	)
	if err != nil {
		InternalError(c, "创建提示词版本失败")
		return
	}

	Success(c, gin.H{
		"version": v,
	})
}

// ComparePromptVersions 比较两个版本
func ComparePromptVersions(c *gin.Context) {
	entityType := c.Param("type")
	entityID := c.Param("id")

	var req struct {
		Version1 int `json:"version1" binding:"required"`
		Version2 int `json:"version2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	result, err := promptVersionService.CompareVersions(entityType, entityID, req.Version1, req.Version2)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, result)
}
