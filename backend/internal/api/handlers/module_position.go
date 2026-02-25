package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ModulePositionRequest 单个模块位置请求
type ModulePositionRequest struct {
	PositionX float64 `json:"positionX"`
	PositionY float64 `json:"positionY"`
}

// BatchPositionItem 批量位置更新项
type BatchPositionItem struct {
	ModuleID  string  `json:"moduleId" binding:"required"`
	PositionX float64 `json:"positionX"`
	PositionY float64 `json:"positionY"`
}

// BatchPositionRequest 批量位置更新请求
type BatchPositionRequest struct {
	Positions []BatchPositionItem `json:"positions" binding:"required"`
}

// ModulePositionResponse 模块位置响应
type ModulePositionResponse struct {
	ModuleID  string   `json:"moduleId"`
	PositionX *float64 `json:"positionX"`
	PositionY *float64 `json:"positionY"`
}

// GetModulePositions 获取所有模块位置
func GetModulePositions(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		ValidationError(c, "projectId 是必需的", nil)
		return
	}

	var modules []models.Module
	if err := database.DB.Select("id, position_x, position_y").
		Where("project_id = ?", projectID).
		Find(&modules).Error; err != nil {
		InternalError(c, "查询模块位置失败")
		return
	}

	positions := make([]ModulePositionResponse, 0, len(modules))
	for _, m := range modules {
		positions = append(positions, ModulePositionResponse{
			ModuleID:  m.ID,
			PositionX: m.PositionX,
			PositionY: m.PositionY,
		})
	}

	Success(c, gin.H{
		"positions": positions,
	})
}

// UpdateModulePosition 更新单个模块位置
func UpdateModulePosition(c *gin.Context) {
	moduleID := c.Param("id")

	var req ModulePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 检查模块是否存在
	var module models.Module
	if err := database.DB.First(&module, "id = ?", moduleID).Error; err != nil {
		NotFound(c, "模块不存在")
		return
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{
		"position_x":          req.PositionX,
		"position_y":          req.PositionY,
		"position_updated_at": now,
	}

	if err := database.DB.Model(&models.Module{}).
		Where("id = ?", moduleID).
		Updates(updates).Error; err != nil {
		InternalError(c, "更新模块位置失败")
		return
	}

	// 重新获取更新后的模块
	if err := database.DB.First(&module, "id = ?", moduleID).Error; err != nil {
		InternalError(c, "获取模块失败")
		return
	}

	Success(c, gin.H{
		"module": module,
	})
}

// BatchUpdateModulePositions 批量更新模块位置
func BatchUpdateModulePositions(c *gin.Context) {
	var req BatchPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	if len(req.Positions) == 0 {
		Success(c, gin.H{
			"updated": 0,
		})
		return
	}

	now := time.Now().UnixMilli()
	updated := 0

	tx := database.DB.Begin()
	for _, pos := range req.Positions {
		updates := map[string]interface{}{
			"position_x":          pos.PositionX,
			"position_y":          pos.PositionY,
			"position_updated_at": now,
		}
		result := tx.Model(&models.Module{}).
			Where("id = ?", pos.ModuleID).
			Updates(updates)
		if result.Error != nil {
			tx.Rollback()
			InternalError(c, "批量更新模块位置失败")
			return
		}
		if result.RowsAffected > 0 {
			updated++
		}
	}

	if err := tx.Commit().Error; err != nil {
		InternalError(c, "提交事务失败")
		return
	}

	Success(c, gin.H{
		"updated": updated,
	})
}
