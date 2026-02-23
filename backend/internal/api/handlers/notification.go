package handlers

import (
	"time"

	"github.com/aitdd/backend/internal/api"
	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetNotifications 获取通知列表
func GetNotifications(c *gin.Context) {
	var notifications []models.Notification

	query := database.DB.Model(&models.Notification{})

	// 过滤条件
	if unread := c.Query("unread"); unread == "true" {
		query = query.Where("read_at IS NULL")
	}
	if notificationType := c.Query("type"); notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// 分页
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		page = parseInt(p)
	}
	if ps := c.Query("pageSize"); ps != "" {
		pageSize = parseInt(ps)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
		api.InternalError(c, "查询通知失败")
		return
	}

	api.Success(c, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"pageSize":      pageSize,
	})
}

// CreateNotification 创建通知
func CreateNotification(c *gin.Context) {
	var req struct {
		Type    string `json:"type" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
		Link    string `json:"link"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		api.ValidationError(c, "无效的请求数据", nil)
		return
	}

	notification := models.Notification{
		ID:        uuid.New().String(),
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Link:      req.Link,
		CreatedAt: time.Now().UnixMilli(),
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		api.InternalError(c, "创建通知失败")
		return
	}

	api.Success(c, gin.H{
		"notification": notification,
	})
}

// MarkNotificationRead 标记通知为已读
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")

	var notification models.Notification
	if err := database.DB.First(&notification, "id = ?", id).Error; err != nil {
		api.NotFound(c, "通知不存在")
		return
	}

	now := time.Now().UnixMilli()
	notification.ReadAt = &now

	if err := database.DB.Save(&notification).Error; err != nil {
		api.InternalError(c, "更新通知失败")
		return
	}

	api.Success(c, gin.H{
		"notification": notification,
	})
}
