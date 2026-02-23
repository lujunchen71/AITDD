package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/api"
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
	if moduleID := c.Query("moduleId"); moduleID != "" {
		query = query.Where("module_id = ?", moduleID)
	}
	if dependsOn := c.Query("dependsOn"); dependsOn != "" {
		query = query.Where("depends_on = ?", dependsOn)
	}

	if err := query.Find(&dependencies).Error; err != nil {
		api.InternalError(c, "查询依赖失败")
		return
	}

	api.Success(c, gin.H{
		"dependencies": dependencies,
		"total":        len(dependencies),
	})
}

// CreateDependency 创建依赖
func CreateDependency(c *gin.Context) {
	var req struct {
		ModuleID   string `json:"moduleId" binding:"required"`
		DependsOn  string `json:"dependsOn" binding:"required"`
		Dependency string `json:"dependency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 检查是否已存在
	var existing models.Dependency
	if err := database.DB.Where("module_id = ? AND depends_on = ?", req.ModuleID, req.DependsOn).First(&existing).Error; err == nil {
		api.ValidationError(c, "依赖关系已存在", nil)
		return
	}

	dependency := models.Dependency{
		ID:         uuid.New().String(),
		ModuleID:   req.ModuleID,
		DependsOn:  req.DependsOn,
		Dependency: req.Dependency,
		CreatedAt:  time.Now().UnixMilli(),
	}

	if err := database.DB.Create(&dependency).Error; err != nil {
		api.InternalError(c, "创建依赖失败")
		return
	}

	api.Success(c, gin.H{
		"dependency": dependency,
	})
}

// DeleteDependency 删除依赖
func DeleteDependency(c *gin.Context) {
	id := c.Param("id")

	var dependency models.Dependency
	if err := database.DB.First(&dependency, "id = ?", id).Error; err != nil {
		api.NotFound(c, "依赖不存在")
		return
	}

	if err := database.DB.Delete(&dependency).Error; err != nil {
		api.InternalError(c, "删除依赖失败")
		return
	}

	api.Success(c, gin.H{
		"deleted": true,
	})
}
