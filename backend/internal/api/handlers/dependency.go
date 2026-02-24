package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetDependencies 获取依赖列表
func GetDependencies(c *gin.Context) {
	var dependencies []models.Dependency

	query := database.DB.Model(&models.Dependency{})

	// 过滤条件
	if upstreamTaskID := c.Query("upstreamTaskId"); upstreamTaskID != "" {
		query = query.Where("upstream_task_id = ?", upstreamTaskID)
	}
	if downstreamTaskID := c.Query("downstreamTaskId"); downstreamTaskID != "" {
		query = query.Where("downstream_task_id = ?", downstreamTaskID)
	}

	if err := query.Find(&dependencies).Error; err != nil {
		InternalError(c, "查询依赖失败")
		return
	}

	Success(c, gin.H{
		"dependencies": dependencies,
		"total":        len(dependencies),
	})
}

// CreateDependency 创建依赖
func CreateDependency(c *gin.Context) {
	var req struct {
		UpstreamTaskID   string `json:"upstreamTaskId" binding:"required"`
		DownstreamTaskID string `json:"downstreamTaskId" binding:"required"`
		ContractSummary  string `json:"contractSummary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 检查是否已存在
	var existing models.Dependency
	if err := database.DB.Where("upstream_task_id = ? AND downstream_task_id = ?", req.UpstreamTaskID, req.DownstreamTaskID).First(&existing).Error; err == nil {
		ValidationError(c, "依赖关系已存在", nil)
		return
	}

	dependency := models.Dependency{
		ID:               uuid.New().String(),
		UpstreamTaskID:   req.UpstreamTaskID,
		DownstreamTaskID: req.DownstreamTaskID,
		ContractSummary:  req.ContractSummary,
		CreatedAt:        time.Now().UnixMilli(),
	}

	if err := database.DB.Create(&dependency).Error; err != nil {
		InternalError(c, "创建依赖失败")
		return
	}

	Success(c, gin.H{
		"dependency": dependency,
	})
}

// DeleteDependency 删除依赖
func DeleteDependency(c *gin.Context) {
	id := c.Param("id")

	var dependency models.Dependency
	if err := database.DB.First(&dependency, "id = ?", id).Error; err != nil {
		NotFound(c, "依赖不存在")
		return
	}

	if err := database.DB.Delete(&dependency).Error; err != nil {
		InternalError(c, "删除依赖失败")
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}
