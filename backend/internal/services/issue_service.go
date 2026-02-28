package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IssueService 问题服务
type IssueService struct{}

// NewIssueService 创建问题服务
func NewIssueService() *IssueService {
	return &IssueService{}
}

// Issue 状态常量
const (
	IssueStatusPending  = "pending"
	IssueStatusReplied  = "replied"
	IssueStatusResolved = "resolved"
)

// Issue 类型常量
const (
	IssueTypeContract = "contract"
	IssueTypeTest     = "test"
	IssueTypeOther    = "other"
)

// CreateIssue 创建问题
func (s *IssueService) CreateIssue(fromTaskPathName, toTaskPathName, issueType, title, content string) (*models.Issue, error) {
	// 验证任务是否存在
	if !s.taskExists(fromTaskPathName) {
		return nil, errors.New("来源任务不存在")
	}
	if !s.taskExists(toTaskPathName) {
		return nil, errors.New("目标任务不存在")
	}

	// 检查是否已存在相同的问题
	var existingCount int64
	database.DB.Model(&models.Issue{}).
		Where("from_task_path_name = ? AND to_task_path_name = ? AND title = ?", fromTaskPathName, toTaskPathName, title).
		Count(&existingCount)
	if existingCount > 0 {
		return nil, errors.New("相同的问题已存在")
	}

	now := time.Now().UnixMilli()
	messages := []models.Message{
		{
			Sender:    fromTaskPathName,
			Timestamp: now,
			Content:   content,
			Read:      false,
		},
	}
	messagesJSON, _ := json.Marshal(messages)

	issue := models.Issue{
		ID:               uuid.New().String(),
		FromTaskPathName: fromTaskPathName,
		ToTaskPathName:   toTaskPathName,
		Type:             issueType,
		Title:            title,
		Status:           IssueStatusPending,
		Messages:         string(messagesJSON),
		CreatedAt:        now,
		UpdatedAt:        now,
		Version:          1,
		SyncStatus:       models.SyncStatusSynced,
	}

	if err := database.DB.Create(&issue).Error; err != nil {
		return nil, err
	}

	return &issue, nil
}

// ReplyIssue 回复问题
func (s *IssueService) ReplyIssue(fromTaskPathName, toTaskPathName, title, replyContent string) (*models.Issue, error) {
	issue, err := s.GetIssueByKey(fromTaskPathName, toTaskPathName, title)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("问题不存在")
	}

	// 检查问题是否已解决
	if issue.Status == IssueStatusResolved {
		return nil, errors.New("问题已解决，无法回复")
	}

	// 解析现有消息
	var messages []models.Message
	if err := json.Unmarshal([]byte(issue.Messages), &messages); err != nil {
		return nil, fmt.Errorf("解析消息失败: %w", err)
	}

	// 添加新消息
	now := time.Now().UnixMilli()
	messages = append(messages, models.Message{
		Sender:    toTaskPathName, // 回复者是接收方
		Timestamp: now,
		Content:   replyContent,
		Read:      false,
	})
	messagesJSON, _ := json.Marshal(messages)

	// 更新问题
	updates := map[string]interface{}{
		"messages":   string(messagesJSON),
		"status":     IssueStatusReplied,
		"updated_at": now,
		"version":    issue.Version + 1,
	}

	if err := database.DB.Model(&issue).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取
	return s.GetIssueByKey(fromTaskPathName, toTaskPathName, title)
}

// ResolveIssue 解决问题
func (s *IssueService) ResolveIssue(fromTaskPathName, toTaskPathName, title string) (*models.Issue, error) {
	issue, err := s.GetIssueByKey(fromTaskPathName, toTaskPathName, title)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, errors.New("问题不存在")
	}

	// 检查问题是否已解决
	if issue.Status == IssueStatusResolved {
		return nil, errors.New("问题已解决")
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{
		"status":     IssueStatusResolved,
		"updated_at": now,
		"version":    issue.Version + 1,
	}

	if err := database.DB.Model(&issue).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取
	return s.GetIssueByKey(fromTaskPathName, toTaskPathName, title)
}

// QueryIssues 查询问题列表
func (s *IssueService) QueryIssues(taskPathName, direction, status, issueType string) ([]models.Issue, int64, error) {
	var issues []models.Issue
	var total int64

	query := database.DB.Model(&models.Issue{})

	// 按任务路径过滤
	if taskPathName != "" {
		if direction == "from" {
			query = query.Where("from_task_path_name = ?", taskPathName)
		} else if direction == "to" {
			query = query.Where("to_task_path_name = ?", taskPathName)
		} else {
			// 默认查询两个方向
			query = query.Where("from_task_path_name = ? OR to_task_path_name = ?", taskPathName, taskPathName)
		}
	}

	// 按状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 按类型过滤
	if issueType != "" {
		query = query.Where("type = ?", issueType)
	}

	query.Count(&total)

	if err := query.Order("created_at DESC").Find(&issues).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// GetIssueByKey 通过关键字段获取问题
func (s *IssueService) GetIssueByKey(fromTaskPathName, toTaskPathName, title string) (*models.Issue, error) {
	var issue models.Issue
	if err := database.DB.First(&issue, "from_task_path_name = ? AND to_task_path_name = ? AND title = ?", fromTaskPathName, toTaskPathName, title).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &issue, nil
}

// GetIssueByID 通过ID获取问题
func (s *IssueService) GetIssueByID(id string) (*models.Issue, error) {
	var issue models.Issue
	if err := database.DB.First(&issue, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &issue, nil
}

// DeleteIssue 删除问题
func (s *IssueService) DeleteIssue(id string) error {
	var issue models.Issue
	if err := database.DB.First(&issue, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("问题不存在")
		}
		return err
	}

	return database.DB.Delete(&issue).Error
}

// taskExists 检查任务是否存在
func (s *IssueService) taskExists(pathName string) bool {
	var count int64
	database.DB.Model(&models.Task{}).Where("path_name = ?", pathName).Count(&count)
	return count > 0
}
