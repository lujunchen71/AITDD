package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/api"
	"github.com/aitdd/backend/internal/database"
	"github.com/gin-gonic/gin"
)

// LockRequest 锁定请求
type LockRequest struct {
	ResourceType string `json:"resourceType" binding:"required"` // module 或 task
	ResourceID   string `json:"resourceId" binding:"required"`
	LockedBy     string `json:"lockedBy" binding:"required"` // AI代理标识
}

// LockResource 锁定资源
func LockResource(c *gin.Context) {
	var req LockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	now := time.Now().UnixMilli()

	switch req.ResourceType {
	case "module":
		var module struct {
			ID        string
			Locked    bool
			LockedBy  *string
			LockedAt  *int64
			Version   int
		}
		if err := database.DB.Table("modules").Where("id = ?", req.ResourceID).First(&module).Error; err != nil {
			api.NotFound(c, "模块不存在")
			return
		}
		if module.Locked {
			api.Locked(c, "资源已被锁定", gin.H{
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
			api.InternalError(c, "锁定失败")
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
			api.NotFound(c, "任务不存在")
			return
		}
		if task.Locked {
			api.Locked(c, "资源已被锁定", gin.H{
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
			api.InternalError(c, "锁定失败")
			return
		}

	default:
		api.ValidationError(c, "不支持的资源类型", nil)
		return
	}

	api.Success(c, gin.H{
		"locked":    true,
		"lockedBy":  req.LockedBy,
		"lockedAt":  now,
	})
}

// UnlockResource 解锁资源
func UnlockResource(c *gin.Context) {
	var req LockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
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
			api.InternalError(c, "解锁失败")
			return
		}
		if result.RowsAffected == 0 {
			api.NotFound(c, "未找到锁定的资源或无权解锁")
			return
		}

	case "task":
		result := database.DB.Table("tasks").Where("id = ? AND locked_by = ?", req.ResourceID, req.LockedBy).Updates(map[string]interface{}{
			"locked":    false,
			"locked_by": nil,
			"locked_at": nil,
		})
		if result.Error != nil {
			api.InternalError(c, "解锁失败")
			return
		}
		if result.RowsAffected == 0 {
			api.NotFound(c, "未找到锁定的资源或无权解锁")
			return
		}

	default:
		api.ValidationError(c, "不支持的资源类型", nil)
		return
	}

	api.Success(c, gin.H{
		"unlocked": true,
	})
}

// GetLockStatus 获取锁定状态
func GetLockStatus(c *gin.Context) {
	resourceType := c.Query("resourceType")
	resourceID := c.Query("resourceId")

	if resourceType == "" || resourceID == "" {
		api.ValidationError(c, "缺少必要参数", nil)
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
			api.NotFound(c, "模块不存在")
			return
		}
		api.Success(c, gin.H{
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
			api.NotFound(c, "任务不存在")
			return
		}
		api.Success(c, gin.H{
			"locked":   task.Locked,
			"lockedBy": task.LockedBy,
			"lockedAt": task.LockedAt,
		})

	default:
		api.ValidationError(c, "不支持的资源类型", nil)
	}
}
