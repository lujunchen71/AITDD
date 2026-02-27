package handlers

import (
	"strings"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/gin-gonic/gin"
)

// LockRequest 锁定请求
type LockRequest struct {
	ResourceType string `json:"resourceType" binding:"required"` // module or task
	ResourceID   string `json:"resourceId" binding:"required"`
	LockedBy     string `json:"lockedBy" binding:"required"` // AI代理标识
}

// LockResource 锁定资源
func LockResource(c *gin.Context) {
	var req LockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	now := time.Now().UnixMilli()

	switch req.ResourceType {
	case "module":
		var module struct {
			ID       string
			Locked   bool
			LockedBy *string
			LockedAt *int64
			Version  int
		}
		if err := database.DB.Table("modules").Where("id = ?", req.ResourceID).First(&module).Error; err != nil {
			NotFound(c, "模块不存在")
			return
		}
		if module.Locked {
			Locked(c, "资源已被锁定", gin.H{
				"lockedBy": module.LockedBy,
				"lockedAt": module.LockedAt,
			})
			return
		}
		if err := database.DB.Table("modules").Where("id = ?", req.ResourceID).Updates(map[string]interface{}{
			"locked":    true,
			"locked_by": req.LockedBy,
			"locked_at": now,
		}).Error; err != nil {
			InternalError(c, "锁定失败")
			return
		}

	case "task":
		var task struct {
			ID       string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("tasks").Where("id = ?", req.ResourceID).First(&task).Error; err != nil {
			NotFound(c, "任务不存在")
			return
		}
		if task.Locked {
			Locked(c, "资源已被锁定", gin.H{
				"lockedBy": task.LockedBy,
				"lockedAt": task.LockedAt,
			})
			return
		}
		if err := database.DB.Table("tasks").Where("id = ?", req.ResourceID).Updates(map[string]interface{}{
			"locked":    true,
			"locked_by": req.LockedBy,
			"locked_at": now,
		}).Error; err != nil {
			InternalError(c, "锁定失败")
			return
		}

	default:
		ValidationError(c, "不支持的资源类型", nil)
		return
	}

	Success(c, gin.H{
		"locked":    true,
		"lockedBy":  req.LockedBy,
		"lockedAt":  now,
	})
}

// UnlockResource 解锁资源
func UnlockResource(c *gin.Context) {
	var req LockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	switch req.ResourceType {
	case "module":
		result := database.DB.Table("modules").Where("id = ? AND locked_by = ?", req.ResourceID, req.LockedBy).Updates(map[string]interface{}{
			"locked":    false,
			"locked_by": nil,
			"locked_at": nil,
		})
		if result.Error != nil {
			InternalError(c, "解锁失败")
			return
		}
		if result.RowsAffected == 0 {
			NotFound(c, "未找到锁定的资源或无权解锁")
			return
		}

	case "task":
		result := database.DB.Table("tasks").Where("id = ? AND locked_by = ?", req.ResourceID, req.LockedBy).Updates(map[string]interface{}{
			"locked":    false,
			"locked_by": nil,
			"locked_at": nil,
		})
		if result.Error != nil {
			InternalError(c, "解锁失败")
			return
		}
		if result.RowsAffected == 0 {
			NotFound(c, "未找到锁定的资源或无权解锁")
			return
		}

	default:
		ValidationError(c, "不支持的资源类型", nil)
		return
	}

	Success(c, gin.H{
		"unlocked": true,
	})
}

// GetLockStatus 获取锁定状态
func GetLockStatus(c *gin.Context) {
	resourceType := c.Query("resourceType")
	resourceID := c.Query("resourceId")

	if resourceType == "" || resourceID == "" {
		ValidationError(c, "缺少必要参数", nil)
		return
	}

	switch resourceType {
	case "module":
		var module struct {
			ID       string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("modules").Where("id = ?", resourceID).First(&module).Error; err != nil {
			NotFound(c, "模块不存在")
			return
		}
		Success(c, gin.H{
			"locked":   module.Locked,
			"lockedBy": module.LockedBy,
			"lockedAt": module.LockedAt,
		})

	case "task":
		var task struct {
			ID       string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("tasks").Where("id = ?", resourceID).First(&task).Error; err != nil {
			NotFound(c, "任务不存在")
			return
		}
		Success(c, gin.H{
			"locked":   task.Locked,
			"lockedBy": task.LockedBy,
			"lockedAt": task.LockedAt,
		})

	default:
		ValidationError(c, "不支持的资源类型", nil)
	}
}

// LockResourceByPathName 通过 pathName 锁定资源
func LockResourceByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")
	resourceType := c.Query("resourceType")
	lockedBy := c.Query("lockedBy")

	if resourceType == "" {
		resourceType = "task" // 默认为任务
	}
	if lockedBy == "" {
		ValidationError(c, "缺少 lockedBy 参数", nil)
		return
	}

	now := time.Now().UnixMilli()

	switch resourceType {
	case "module":
		var module struct {
			ID       string
			PathName string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("modules").Where("path_name = ?", pathName).First(&module).Error; err != nil {
			NotFound(c, "模块不存在: "+pathName)
			return
		}
		if module.Locked {
			Locked(c, "资源已被锁定", gin.H{
				"lockedBy": module.LockedBy,
				"lockedAt": module.LockedAt,
			})
			return
		}
		if err := database.DB.Table("modules").Where("id = ?", module.ID).Updates(map[string]interface{}{
			"locked":    true,
			"locked_by": lockedBy,
			"locked_at": now,
		}).Error; err != nil {
			InternalError(c, "锁定失败")
			return
		}

	case "task":
		var task struct {
			ID       string
			PathName string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("tasks").Where("path_name = ?", pathName).First(&task).Error; err != nil {
			NotFound(c, "任务不存在: "+pathName)
			return
		}
		if task.Locked {
			Locked(c, "资源已被锁定", gin.H{
				"lockedBy": task.LockedBy,
				"lockedAt": task.LockedAt,
			})
			return
		}
		if err := database.DB.Table("tasks").Where("id = ?", task.ID).Updates(map[string]interface{}{
			"locked":    true,
			"locked_by": lockedBy,
			"locked_at": now,
		}).Error; err != nil {
			InternalError(c, "锁定失败")
			return
		}

	default:
		ValidationError(c, "不支持的资源类型", nil)
		return
	}

	Success(c, gin.H{
		"locked":    true,
		"lockedBy":  lockedBy,
		"lockedAt":  now,
	})
}

// UnlockResourceByPathName 通过 pathName 解锁资源
func UnlockResourceByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")
	resourceType := c.Query("resourceType")
	lockedBy := c.Query("lockedBy")

	if resourceType == "" {
		resourceType = "task"
	}
	if lockedBy == "" {
		ValidationError(c, "缺少 lockedBy 参数", nil)
		return
	}

	switch resourceType {
	case "module":
		var module struct {
			ID       string
			LockedBy *string
		}
		if err := database.DB.Table("modules").Where("path_name = ?", pathName).First(&module).Error; err != nil {
			NotFound(c, "模块不存在: "+pathName)
			return
		}
		result := database.DB.Table("modules").Where("id = ? AND locked_by = ?", module.ID, lockedBy).Updates(map[string]interface{}{
			"locked":    false,
			"locked_by": nil,
			"locked_at": nil,
		})
		if result.Error != nil {
			InternalError(c, "解锁失败")
			return
		}
		if result.RowsAffected == 0 {
			NotFound(c, "未找到锁定的资源或无权解锁")
			return
		}

	case "task":
		var task struct {
			ID       string
			LockedBy *string
		}
		if err := database.DB.Table("tasks").Where("path_name = ?", pathName).First(&task).Error; err != nil {
			NotFound(c, "任务不存在: "+pathName)
			return
		}
		result := database.DB.Table("tasks").Where("id = ? AND locked_by = ?", task.ID, lockedBy).Updates(map[string]interface{}{
			"locked":    false,
			"locked_by": nil,
			"locked_at": nil,
		})
		if result.Error != nil {
			InternalError(c, "解锁失败")
			return
		}
		if result.RowsAffected == 0 {
			NotFound(c, "未找到锁定的资源或无权解锁")
			return
		}

	default:
		ValidationError(c, "不支持的资源类型", nil)
		return
	}

	Success(c, gin.H{
		"unlocked": true,
	})
}

// GetLockStatusByPathName 通过 pathName 获取锁定状态
func GetLockStatusByPathName(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")
	resourceType := c.Query("resourceType")

	if resourceType == "" {
		resourceType = "task"
	}

	switch resourceType {
	case "module":
		var module struct {
			ID       string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("modules").Where("path_name = ?", pathName).First(&module).Error; err != nil {
			NotFound(c, "模块不存在: "+pathName)
			return
		}
		Success(c, gin.H{
			"locked":   module.Locked,
			"lockedBy": module.LockedBy,
			"lockedAt": module.LockedAt,
		})

	case "task":
		var task struct {
			ID       string
			Locked   bool
			LockedBy *string
			LockedAt *int64
		}
		if err := database.DB.Table("tasks").Where("path_name = ?", pathName).First(&task).Error; err != nil {
			NotFound(c, "任务不存在: "+pathName)
			return
		}
		Success(c, gin.H{
			"locked":   task.Locked,
			"lockedBy": task.LockedBy,
			"lockedAt": task.LockedAt,
		})

	default:
		ValidationError(c, "不支持的资源类型", nil)
	}
}
