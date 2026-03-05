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

// GetModules 获取模块列表
func GetModules(c *gin.Context) {
	var modules []models.Module

	query := database.DB.Model(&models.Module{})

	// 过滤条件 - 支持多个项目ID（逗号分隔）
	if projectIDs := c.Query("projectIds"); projectIDs != "" {
		// 分割逗号分隔的项目ID列表
		query = query.Where("project_id IN ?", strings.Split(projectIDs, ","))
	} else if projectID := c.Query("projectId"); projectID != "" {
		// 兼容单个项目ID
		query = query.Where("project_id = ?", projectID)
	} else if projectPathName := c.Query("projectPathName"); projectPathName != "" {
		// 通过项目 pathName 查询：先获取项目ID，再过滤模块
		var project models.Project
		if err := database.DB.Where("path_name = ?", projectPathName).First(&project).Error; err == nil {
			query = query.Where("project_id = ?", project.ID)
		}
		// 如果项目不存在，query 不会添加过滤条件，返回空列表
	}
	if parentID := c.Query("parentId"); parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&modules).Error; err != nil {
		InternalError(c, "查询模块失败")
		return
	}

	Success(c, gin.H{
		"modules": modules,
		"total":   len(modules),
	})
}

// GetModule 获取单个模块
func GetModule(c *gin.Context) {
	id := c.Param("id")

	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		NotFound(c, "模块不存在")
		return
	}

	Success(c, gin.H{
		"module": module,
	})
}

// CreateModule 创建模块
func CreateModule(c *gin.Context) {
	var req struct {
		ProjectID   string  `json:"projectId" binding:"required"`
		ParentID    *string `json:"parentId"`
		Name        string  `json:"name" binding:"required"`
		PathName    string  `json:"pathName"`
		Description string  `json:"description"`
		Prompt      string  `json:"prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 获取项目信息以生成正确的 pathName
	var project models.Project
	if err := database.DB.First(&project, "id = ?", req.ProjectID).Error; err != nil {
		NotFound(c, "项目不存在")
		return
	}

	// 生成唯一的 pathName，格式：项目pathName/模块名称
	basePathName := fmt.Sprintf("%s/%s", project.PathName, req.Name)
	pathName := basePathName
	// 检查pathName是否已存在，如果存在则添加数字后缀
	suffix := 1
	for {
		var count int64
		database.DB.Model(&models.Module{}).Where("path_name = ?", pathName).Count(&count)
		if count == 0 {
			break
		}
		pathName = fmt.Sprintf("%s-%d", basePathName, suffix)
		suffix++
	}

	module := models.Module{
		ID:          uuid.New().String(),
		ParentID:    req.ParentID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		PathName:    pathName,
		Description: req.Description,
		Prompt:      req.Prompt,
		Status:      models.ModuleStatusDesigning,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		Version:     1,
		SyncStatus:  models.SyncStatusSynced,
	}

	if err := database.DB.Create(&module).Error; err != nil {
		InternalError(c, "创建模块失败")
		return
	}

	Success(c, gin.H{
		"module": module,
	})
}

// UpdateModule 更新模块
func UpdateModule(c *gin.Context) {
	id := c.Param("id")

	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		NotFound(c, "模块不存在")
		return
	}

	var req struct {
		Name                      *string  `json:"name"`
		Description               *string  `json:"description"`
		Prompt                    *string  `json:"prompt"`
		Status                    *string  `json:"status"`
		TestCoverage              *float64 `json:"testCoverage"`
		UpstreamContractSummary   *string  `json:"upstreamContractSummary"`
		DownstreamContractSummary *string  `json:"downstreamContractSummary"`
		Version                   int      `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查
	if module.Version != req.Version {
		VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 更新字段
	nameChanged := false
	if req.Name != nil && *req.Name != module.Name {
		module.Name = *req.Name
		nameChanged = true
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

	// 如果名称改变，需要更新 pathName 和所有子任务的 pathName
	if nameChanged {
		// 获取项目信息以生成新的 pathName
		var project models.Project
		if err := database.DB.First(&project, "id = ?", module.ProjectID).Error; err != nil {
			InternalError(c, "获取项目失败")
			return
		}

		// 计算新的 pathName
		oldPathName := module.PathName
		basePathName := fmt.Sprintf("%s/%s", project.PathName, module.Name)
		newPathName := basePathName
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Module{}).Where("path_name = ? AND id != ?", newPathName, module.ID).Count(&count)
			if count == 0 {
				break
			}
			newPathName = fmt.Sprintf("%s-%d", basePathName, suffix)
			suffix++
		}
		module.PathName = newPathName

		// 级联更新所有子任务的 pathName
		database.DB.Model(&models.Task{}).
			Where("module_id = ?", module.ID).
			Update("path_name", fmt.Sprintf("REPLACE(path_name, '%s', '%s')", oldPathName, newPathName))
	}

	module.Version++
	module.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&module).Error; err != nil {
		InternalError(c, "更新模块失败")
		return
	}

	Success(c, gin.H{
		"module": module,
	})
}

// DeleteModule 删除模块
func DeleteModule(c *gin.Context) {
	id := c.Param("id")

	var module models.Module
	if err := database.DB.First(&module, "id = ?", id).Error; err != nil {
		NotFound(c, "模块不存在")
		return
	}

	// 检查是否锁定
	if module.Locked {
		Locked(c, "模块已被锁定，无法删除", gin.H{
			"lockedBy": module.LockedBy,
		})
		return
	}

	if err := database.DB.Delete(&module).Error; err != nil {
		InternalError(c, "删除模块失败")
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}

// GetModuleTasks 获取模块下的任务
func GetModuleTasks(c *gin.Context) {
	moduleID := c.Param("id")

	var tasks []models.Task
	if err := database.DB.Where("module_id = ?", moduleID).Find(&tasks).Error; err != nil {
		InternalError(c, "查询任务失败")
		return
	}

	Success(c, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// GetModuleDependencies 获取模块的依赖列表
func GetModuleDependencies(c *gin.Context) {
	moduleID := c.Param("id")

	var dependencies []models.ModuleDependency
	if err := database.DB.Where("module_id = ?", moduleID).Find(&dependencies).Error; err != nil {
		InternalError(c, "查询模块依赖失败")
		return
	}

	// 获取被依赖模块的详细信息
	type DependencyWithModule struct {
		models.ModuleDependency
		DependsOnModule *models.Module `json:"dependsOnModule"`
	}

	var result []DependencyWithModule
	for _, dep := range dependencies {
		item := DependencyWithModule{
			ModuleDependency: dep,
		}

		var module models.Module
		if err := database.DB.First(&module, "id = ?", dep.DependsOnModuleID).Error; err == nil {
			item.DependsOnModule = &module
		}

		result = append(result, item)
	}

	Success(c, gin.H{
		"dependencies": result,
		"total":        len(result),
	})
}

// GetModuleDependents 获取依赖此模块的模块列表
func GetModuleDependents(c *gin.Context) {
	moduleID := c.Param("id")

	var dependencies []models.ModuleDependency
	if err := database.DB.Where("depends_on_module_id = ?", moduleID).Find(&dependencies).Error; err != nil {
		InternalError(c, "查询模块被依赖列表失败")
		return
	}

	// 获取依赖方模块的详细信息
	type DependentWithModule struct {
		models.ModuleDependency
		Module *models.Module `json:"module"`
	}

	var result []DependentWithModule
	for _, dep := range dependencies {
		item := DependentWithModule{
			ModuleDependency: dep,
		}

		var module models.Module
		if err := database.DB.First(&module, "id = ?", dep.ModuleID).Error; err == nil {
			item.Module = &module
		}

		result = append(result, item)
	}

	Success(c, gin.H{
		"dependents": result,
		"total":      len(result),
	})
}

// CreateModuleDependency 创建模块依赖
func CreateModuleDependency(c *gin.Context) {
	moduleID := c.Param("id")

	var req struct {
		DependsOnModuleID string `json:"dependsOnModuleId" binding:"required"`
		DependencyType    string `json:"dependencyType"`
		ContractSummary   string `json:"contractSummary"`
	}

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

	// 检查被依赖模块是否存在
	var dependsOnModule models.Module
	if err := database.DB.First(&dependsOnModule, "id = ?", req.DependsOnModuleID).Error; err != nil {
		NotFound(c, "被依赖的模块不存在")
		return
	}

	// 不能依赖自己
	if moduleID == req.DependsOnModuleID {
		BadRequest(c, "模块不能依赖自己")
		return
	}

	// 检查是否已存在依赖关系
	var existingCount int64
	database.DB.Model(&models.ModuleDependency{}).Where("module_id = ? AND depends_on_module_id = ?", moduleID, req.DependsOnModuleID).Count(&existingCount)
	if existingCount > 0 {
		BadRequest(c, "依赖关系已存在")
		return
	}

	// 检查是否会形成循环依赖
	if wouldCreateCycle(moduleID, req.DependsOnModuleID) {
		BadRequest(c, "会形成循环依赖")
		return
	}

	// 设置默认依赖类型
	dependencyType := req.DependencyType
	if dependencyType == "" {
		dependencyType = models.ModuleDependencyRequired
	}

	dependency := models.ModuleDependency{
		ID:                uuid.New().String(),
		ModuleID:          moduleID,
		DependsOnModuleID: req.DependsOnModuleID,
		DependencyType:    dependencyType,
		ContractSummary:   req.ContractSummary,
		CreatedAt:         time.Now().UnixMilli(),
		UpdatedAt:         time.Now().UnixMilli(),
		Version:          1,
		SyncStatus:       models.SyncStatusSynced,
	}

	if err := database.DB.Create(&dependency).Error; err != nil {
		InternalError(c, "创建模块依赖失败")
		return
	}

	Success(c, gin.H{
		"dependency": dependency,
	})
}

// wouldCreateCycle 检查是否会形成循环依赖
func wouldCreateCycle(moduleID, dependsOnModuleID string) bool {
	visited := make(map[string]bool)
	return hasDependencyPath(dependsOnModuleID, moduleID, visited)
}

// hasDependencyPath 检查是否存在依赖路径
func hasDependencyPath(fromModuleID, toModuleID string, visited map[string]bool) bool {
	if fromModuleID == toModuleID {
		return true
	}

	if visited[fromModuleID] {
		return false
	}
	visited[fromModuleID] = true

	var dependencies []models.ModuleDependency
	database.DB.Where("module_id = ?", fromModuleID).Find(&dependencies)

	for _, dep := range dependencies {
		if hasDependencyPath(dep.DependsOnModuleID, toModuleID, visited) {
			return true
		}
	}

	return false
}

// DeleteModuleDependency 删除模块依赖
func DeleteModuleDependency(c *gin.Context) {
	moduleID := c.Param("id")
	dependencyID := c.Param("depId")

	var dependency models.ModuleDependency
	if err := database.DB.First(&dependency, "id = ? AND module_id = ?", dependencyID, moduleID).Error; err != nil {
		NotFound(c, "依赖关系不存在")
		return
	}

	if err := database.DB.Delete(&dependency).Error; err != nil {
		InternalError(c, "删除模块依赖失败")
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}

// UpdateModuleDependency 更新模块依赖
func UpdateModuleDependency(c *gin.Context) {
	moduleID := c.Param("id")
	dependencyID := c.Param("depId")

	var dependency models.ModuleDependency
	if err := database.DB.First(&dependency, "id = ? AND module_id = ?", dependencyID, moduleID).Error; err != nil {
		NotFound(c, "依赖关系不存在")
		return
	}

	var req struct {
		DependencyType  *string `json:"dependencyType"`
		ContractSummary *string `json:"contractSummary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	if req.DependencyType != nil {
		dependency.DependencyType = *req.DependencyType
	}
	if req.ContractSummary != nil {
		dependency.ContractSummary = *req.ContractSummary
	}

	dependency.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&dependency).Error; err != nil {
		InternalError(c, "更新模块依赖失败")
		return
	}

	Success(c, gin.H{
		"dependency": dependency,
	})
}

// GetModuleByPathName 通过 pathName 获取模块
func GetModuleByPathName(c *gin.Context) {
	// 获取 pathName，移除前导斜杠
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var module models.Module
	if err := database.DB.First(&module, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "模块不存在: "+pathName)
		return
	}

	// 检查 action 参数
	action := c.Query("action")
	switch action {
	case "tasks":
		// 返回模块的任务列表
		var tasks []models.Task
		if err := database.DB.Where("module_id = ?", module.ID).Find(&tasks).Error; err != nil {
			InternalError(c, "查询任务失败")
			return
		}
		Success(c, gin.H{
			"tasks": tasks,
			"total": len(tasks),
		})
		return
	case "dependencies":
		// 返回模块的依赖列表
		var dependencies []models.ModuleDependency
		if err := database.DB.Where("module_id = ?", module.ID).Find(&dependencies).Error; err != nil {
			InternalError(c, "查询模块依赖失败")
			return
		}
		// 获取被依赖模块的详细信息
		type DependencyWithModule struct {
			models.ModuleDependency
			DependsOnModule *models.Module `json:"dependsOnModule"`
		}
		var result []DependencyWithModule
		for _, dep := range dependencies {
			item := DependencyWithModule{
				ModuleDependency: dep,
			}
			var mod models.Module
			if err := database.DB.First(&mod, "id = ?", dep.DependsOnModuleID).Error; err == nil {
				item.DependsOnModule = &mod
			}
			result = append(result, item)
		}
		Success(c, gin.H{
			"dependencies": result,
			"total":        len(result),
		})
		return
	}

	Success(c, gin.H{
		"module": module,
	})
}

// UpdateModuleByPathName 通过 pathName 更新模块
func UpdateModuleByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var module models.Module
	if err := database.DB.First(&module, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "模块不存在: "+pathName)
		return
	}

	var req struct {
		Name                      *string  `json:"name"`
		Description               *string  `json:"description"`
		Prompt                    *string  `json:"prompt"`
		Status                    *string  `json:"status"`
		TestCoverage              *float64 `json:"testCoverage"`
		UpstreamContractSummary   *string  `json:"upstreamContractSummary"`
		DownstreamContractSummary *string  `json:"downstreamContractSummary"`
		Version                   int      `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 版本检查
	if module.Version != req.Version {
		VersionConflict(c, "数据已被其他请求修改，请刷新后重试")
		return
	}

	// 更新字段
	nameChanged := false
	if req.Name != nil && *req.Name != module.Name {
		module.Name = *req.Name
		nameChanged = true
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

	// 如果名称改变，需要更新 pathName 和所有子任务的 pathName
	if nameChanged {
		// 获取项目信息以生成新的 pathName
		var project models.Project
		if err := database.DB.First(&project, "id = ?", module.ProjectID).Error; err != nil {
			InternalError(c, "获取项目失败")
			return
		}

		// 计算新的 pathName
		oldPathName := module.PathName
		basePathName := fmt.Sprintf("%s/%s", project.PathName, module.Name)
		newPathName := basePathName
		suffix := 1
		for {
			var count int64
			database.DB.Model(&models.Module{}).Where("path_name = ? AND id != ?", newPathName, module.ID).Count(&count)
			if count == 0 {
				break
			}
			newPathName = fmt.Sprintf("%s-%d", basePathName, suffix)
			suffix++
		}
		module.PathName = newPathName

		// 级联更新所有子任务的 pathName
		// 使用 Exec 执行原生 SQL 以确保 REPLACE 函数正确执行
		database.DB.Exec(
			"UPDATE tasks SET path_name = REPLACE(path_name, ?, ?) WHERE module_id = ?",
			oldPathName, newPathName, module.ID,
		)
	}

	module.Version++
	module.UpdatedAt = time.Now().UnixMilli()

	if err := database.DB.Save(&module).Error; err != nil {
		InternalError(c, "更新模块失败")
		return
	}

	Success(c, gin.H{
		"module": module,
	})
}

// DeleteModuleByPathName 通过 pathName 删除模块
func DeleteModuleByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var module models.Module
	if err := database.DB.First(&module, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "模块不存在: "+pathName)
		return
	}

	// 检查是否锁定
	if module.Locked {
		Locked(c, "模块已被锁定，无法删除", gin.H{
			"lockedBy": module.LockedBy,
		})
		return
	}

	if err := database.DB.Delete(&module).Error; err != nil {
		InternalError(c, "删除模块失败")
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}

// GetModuleTasksByPathName 通过 pathName 获取模块下的任务
func GetModuleTasksByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var module models.Module
	if err := database.DB.First(&module, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "模块不存在: "+pathName)
		return
	}

	var tasks []models.Task
	if err := database.DB.Where("module_id = ?", module.ID).Find(&tasks).Error; err != nil {
		InternalError(c, "查询任务失败")
		return
	}

	Success(c, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// GetModuleDependenciesByPathName 通过 pathName 获取模块的依赖列表
func GetModuleDependenciesByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")

	var module models.Module
	if err := database.DB.First(&module, "path_name = ?", pathName).Error; err != nil {
		NotFound(c, "模块不存在: "+pathName)
		return
	}

	var dependencies []models.ModuleDependency
	if err := database.DB.Where("module_id = ?", module.ID).Find(&dependencies).Error; err != nil {
		InternalError(c, "查询模块依赖失败")
		return
	}

	// 获取被依赖模块的详细信息
	type DependencyWithModule struct {
		models.ModuleDependency
		DependsOnModule *models.Module `json:"dependsOnModule"`
	}

	var result []DependencyWithModule
	for _, dep := range dependencies {
		item := DependencyWithModule{
			ModuleDependency: dep,
		}

		var depModule models.Module
		if err := database.DB.First(&depModule, "id = ?", dep.DependsOnModuleID).Error; err == nil {
			item.DependsOnModule = &depModule
		}

		result = append(result, item)
	}

	Success(c, gin.H{
		"dependencies": result,
		"total":        len(result),
	})
}
