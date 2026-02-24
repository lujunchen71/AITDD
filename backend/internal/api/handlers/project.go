package handlers

import (
	"time"

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
		Success(c, gin.H{
			"project": nil,
		})
		return
	}

	Success(c, gin.H{
		"project": project,
	})
}

// GetConstitution 获取项目宪法
func GetConstitution(c *gin.Context) {
	var project models.Project

	result := database.DB.First(&project)
	if result.Error != nil {
		NotFound(c, "项目不存在")
		return
	}

	Success(c, gin.H{
		"constitution": project.Constitution,
	})
}

// UpdateConstitution 更新项目宪法
func UpdateConstitution(c *gin.Context) {
	var project models.Project
	isNewProject := false

	// 获取或创建项目
	result := database.DB.First(&project)
	if result.Error != nil {
		// 项目不存在，标记为新项目
		isNewProject = true
		// 创建默认项目，版本从0开始（与前端传入的version一致）
		project = models.Project{
			ID:           uuid.New().String(),
			Name:         "AITDD Project",
			Constitution: "",
			CreatedAt:    time.Now().UnixMilli(),
			UpdatedAt:    time.Now().UnixMilli(),
			Version:      0,
			SyncStatus:   models.SyncStatusSynced,
		}
	}

	// 解析请求
	var req struct {
		Constitution string `json:"constitution"`
		Version      int    `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查（乐观锁）- 仅对已存在的项目进行检查
	// 如果前端传入 version: 0，说明是初始化请求，跳过版本检查
	if !isNewProject && req.Version != 0 && project.Version != req.Version {
		VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 更新宪法
	project.Constitution = req.Constitution
	project.Version++
	project.UpdatedAt = time.Now().UnixMilli()

	// 保存
	if isNewProject {
		result = database.DB.Create(&project)
	} else {
		result = database.DB.Save(&project)
	}

	if result.Error != nil {
		InternalError(c, "保存失败")
		return
	}

	Success(c, gin.H{
		"project": project,
	})
}
