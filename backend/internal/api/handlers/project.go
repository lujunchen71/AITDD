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

// GetProjects 获取所有项目列表
func GetProjects(c *gin.Context) {
	var projects []models.Project

	result := database.DB.Order("created_at DESC").Find(&projects)
	if result.Error != nil {
		InternalError(c, "获取项目列表失败")
		return
	}

	// 如果没有项目，返回空数组
	if projects == nil {
		projects = []models.Project{}
	}

	Success(c, gin.H{
		"projects": projects,
		"total":    len(projects),
	})
}

// CreateProject 创建新项目
func CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 生成唯一的 path_name
	pathName := req.Name
	suffix := 1
	for {
		var count int64
		database.DB.Model(&models.Project{}).Where("path_name = ?", pathName).Count(&count)
		if count == 0 {
			break
		}
		pathName = fmt.Sprintf("%s-%d", req.Name, suffix)
		suffix++
	}

	// 创建新项目
	project := models.Project{
		ID:           uuid.New().String(),
		Name:         req.Name,
		PathName:     pathName,
		Constitution: req.Description,
		CreatedAt:    time.Now().UnixMilli(),
		UpdatedAt:    time.Now().UnixMilli(),
		Version:      1,
		SyncStatus:   models.SyncStatusSynced,
	}

	result := database.DB.Create(&project)
	if result.Error != nil {
		InternalError(c, "创建项目失败")
		return
	}

	Success(c, gin.H{
		"project": project,
	})
}

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

	// 获取项目（单项目模式 - 获取第一个项目）
	result := database.DB.First(&project)
	if result.Error != nil {
		// 项目不存在，返回404错误，不自动创建空项目
		// 前端应该使用 POST /projects 接口创建项目
		NotFound(c, "项目不存在，请先创建项目")
		return
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

	// 版本检查（乐观锁）
	// 如果前端传入 version: 0，说明是初始化请求，跳过版本检查
	if req.Version != 0 && project.Version != req.Version {
		VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 更新宪法
	project.Constitution = req.Constitution
	project.Version++
	project.UpdatedAt = time.Now().UnixMilli()

	// 保存
	result = database.DB.Save(&project)

	if result.Error != nil {
		InternalError(c, "保存失败")
		return
	}

	Success(c, gin.H{
		"project": project,
	})
}

// UpdateProject 更新项目
func UpdateProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ValidationError(c, "项目ID不能为空", nil)
		return
	}

	var project models.Project
	if err := database.DB.First(&project, "id = ?", id).Error; err != nil {
		NotFound(c, "项目不存在")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now().UnixMilli()

	if req.Description != "" {
		updates["constitution"] = req.Description
	}

	nameChanged := false
	if req.Name != "" && req.Name != project.Name {
		updates["name"] = req.Name
		
		// 生成新的唯一 path_name
		pathName := req.Name
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Project{}).Where("path_name = ? AND id != ?", pathName, id).Count(&count)
			if count == 0 {
				break
			}
			pathName = fmt.Sprintf("%s-%d", req.Name, suffix)
			suffix++
		}
		updates["path_name"] = pathName
		nameChanged = true
	}

	if err := database.DB.Model(&project).Updates(updates).Error; err != nil {
		InternalError(c, "更新项目失败")
		return
	}

	// 如果名称改变，级联更新模块和任务的 path_name（同步执行确保数据一致性）
	if nameChanged {
		newPathName := updates["path_name"].(string)
		cascadeUpdateProjectPathName(id, newPathName)
	}

	// 重新获取
	database.DB.First(&project, "id = ?", id)

	Success(c, gin.H{
		"project": project,
	})
}

func cascadeUpdateProjectPathName(projectID string, projectPathName string) {
	var modules []models.Module
	database.DB.Where("project_id = ?", projectID).Find(&modules)

	for _, mod := range modules {
		basePathName := fmt.Sprintf("%s/%s", projectPathName, mod.Name)
		pathName := basePathName
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Module{}).Where("path_name = ? AND id != ?", pathName, mod.ID).Count(&count)
			if count == 0 {
				break
			}
			pathName = fmt.Sprintf("%s-%d", basePathName, suffix)
			suffix++
		}
		
		database.DB.Model(&mod).Update("path_name", pathName)
		
		// 级联更新任务
		cascadeUpdateModulePathName(mod.ID, pathName)
	}
}

func cascadeUpdateModulePathName(moduleID string, modulePathName string) {
	var tasks []models.Task
	database.DB.Where("module_id = ?", moduleID).Find(&tasks)

	for _, task := range tasks {
		basePathName := fmt.Sprintf("%s/%s", modulePathName, task.Name)
		pathName := basePathName
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Task{}).Where("path_name = ? AND id != ?", pathName, task.ID).Count(&count)
			if count == 0 {
				break
			}
			pathName = fmt.Sprintf("%s-%d", basePathName, suffix)
			suffix++
		}
		
		database.DB.Model(&task).Update("path_name", pathName)
	}
}

// DeleteProject 删除项目
func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ValidationError(c, "项目ID不能为空", nil)
		return
	}

	// 检查项目是否存在
	var project models.Project
	result := database.DB.First(&project, "id = ?", id)
	if result.Error != nil {
		NotFound(c, "项目不存在")
		return
	}

	// 删除项目相关的模块、任务等（级联删除）
	// 先删除任务依赖
	database.DB.Where("task_id IN (SELECT id FROM tasks WHERE project_id = ?)", id).Delete(&models.Dependency{})
	database.DB.Where("dependent_task_id IN (SELECT id FROM tasks WHERE project_id = ?)", id).Delete(&models.Dependency{})

	// 删除任务
	database.DB.Where("project_id = ?", id).Delete(&models.Task{})

	// 删除模块依赖
	database.DB.Where("module_id IN (SELECT id FROM modules WHERE project_id = ?)", id).Delete(&models.ModuleDependency{})
	database.DB.Where("depends_on_module_id IN (SELECT id FROM modules WHERE project_id = ?)", id).Delete(&models.ModuleDependency{})

	// 删除模块
	database.DB.Where("project_id = ?", id).Delete(&models.Module{})

	// 删除项目
	result = database.DB.Delete(&project)
	if result.Error != nil {
		InternalError(c, "删除项目失败")
		return
	}

	Success(c, gin.H{
		"message": "项目已删除",
	})
}

// GetProjectByPathName 通过 pathName 获取项目
func GetProjectByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var project models.Project
	if err := database.DB.First(&project, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "项目不存在: "+pathName)
		return
	}

	Success(c, gin.H{
		"project": project,
	})
}
