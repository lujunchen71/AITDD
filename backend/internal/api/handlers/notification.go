package handlers

import (
	"time"

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
		query = query.Where("read = ?", false)
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
		InternalError(c, "查询通知失败")
		return
	}

	Success(c, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"pageSize":      pageSize,
	})
}

// CreateNotification 创建通知
func CreateNotification(c *gin.Context) {
	var req struct {
		FromTaskID string `json:"fromTaskId" binding:"required"`
		ToTaskID   string `json:"toTaskId" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	notification := models.Notification{
		ID:         uuid.New().String(),
		FromTaskID: req.FromTaskID,
		ToTaskID:   req.ToTaskID,
		Type:       req.Type,
		Title:      req.Title,
		Content:    req.Content,
		Read:       false,
		CreatedAt:  time.Now().UnixMilli(),
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		InternalError(c, "创建通知失败")
		return
	}

	Success(c, gin.H{
		"notification": notification,
	})
}

// MarkNotificationRead 标记通知为已读
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")

	var notification models.Notification
	if err := database.DB.First(&notification, "id = ?", id).Error; err != nil {
		NotFound(c, "通知不存在")
		return
	}

	notification.Read = true

	if err := database.DB.Save(&notification).Error; err != nil {
		InternalError(c, "更新通知失败")
		return
	}

	Success(c, gin.H{
		"notification": notification,
	})
}
