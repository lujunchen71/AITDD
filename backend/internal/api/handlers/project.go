package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/api"
	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetProject 获取项目信息
func GetProject(c *gin.Context) {
	var project models.Project

	// 获取第一个项目（单项目模式）
	result := database.DB.First(&project)
	if result.Error != nil {
		// 如果没有项目，返回空
		api.Success(c, gin.H{
			"project": nil,
		})
		return
	}

	api.Success(c, gin.H{
		"project": project,
	})
}

// GetConstitution 获取项目宪法
func GetConstitution(c *gin.Context) {
	var project models.Project

	result := database.DB.First(&project)
	if result.Error != nil {
		api.NotFound(c, "项目不存在")
		return
	}

	api.Success(c, gin.H{
		"constitution": project.Constitution,
	})
}

// UpdateConstitution 更新项目宪法
func UpdateConstitution(c *gin.Context) {
	var project models.Project

	// 获取或创建项目
	result := database.DB.First(&project)
	if result.Error != nil {
		// 创建默认项目
		project = models.Project{
			ID:           uuid.New().String(),
			Name:         "AITDD Project",
			Constitution: "",
			CreatedAt:    time.Now().UnixMilli(),
			UpdatedAt:    time.Now().UnixMilli(),
			Version:      1,
			SyncStatus:   models.SyncStatusSynced,
		}
	}

	// 解析请求
	var req struct {
		Constitution string `json:"constitution"`
		Version      int    `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查（乐观锁）
	if project.Version != req.Version {
		api.VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 更新宪法
	project.Constitution = req.Constitution
	project.Version++
	project.UpdatedAt = time.Now().UnixMilli()

	// 保存
	if result.Error == nil {
		result = database.DB.Save(&project)
	} else {
		result = database.DB.Create(&project)
	}

	if result.Error != nil {
		api.InternalError(c, "保存失败")
		return
	}

	api.Success(c, gin.H{
		"project": project,
	})
}
