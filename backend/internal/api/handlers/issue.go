package handlers

import (
	"encoding/json"
	"strings"

	"github.com/aitdd/backend/internal/models"
	"github.com/aitdd/backend/internal/services"
	"github.com/gin-gonic/gin"
)

var issueService = services.NewIssueService()

// CreateIssue 创建问题
// POST /api/v1/issues
func CreateIssue(c *gin.Context) {
	var req struct {
		FromTaskPathName string `json:"fromTaskPathName" binding:"required"`
		ToTaskPathName   string `json:"toTaskPathName" binding:"required"`
		Type             string `json:"type" binding:"required"`
		Title            string `json:"title" binding:"required"`
		Content          string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	// 验证类型
	if req.Type != "contract" && req.Type != "test" && req.Type != "other" {
		ValidationError(c, "无效的问题类型，必须是 contract/test/other", nil)
		return
	}

	issue, err := issueService.CreateIssue(req.FromTaskPathName, req.ToTaskPathName, req.Type, req.Title, req.Content)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"issue": issue,
	})
}

// ReplyIssue 回复问题
// POST /api/v1/issues/reply
func ReplyIssue(c *gin.Context) {
	var req struct {
		FromTaskPathName string `json:"fromTaskPathName" binding:"required"`
		ToTaskPathName   string `json:"toTaskPathName" binding:"required"`
		Title            string `json:"title" binding:"required"`
		ReplyContent     string `json:"replyContent" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	issue, err := issueService.ReplyIssue(req.FromTaskPathName, req.ToTaskPathName, req.Title, req.ReplyContent)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"issue": issue,
	})
}

// ResolveIssue 解决问题
// POST /api/v1/issues/resolve
func ResolveIssue(c *gin.Context) {
	var req struct {
		FromTaskPathName string `json:"fromTaskPathName" binding:"required"`
		ToTaskPathName   string `json:"toTaskPathName" binding:"required"`
		Title            string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "无效的请求数据", nil)
		return
	}

	issue, err := issueService.ResolveIssue(req.FromTaskPathName, req.ToTaskPathName, req.Title)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"issue": issue,
	})
}

// QueryIssues 查询问题列表
// GET /api/v1/issues
func QueryIssues(c *gin.Context) {
	taskPathName := c.Query("taskPathName")
	direction := c.Query("direction") // from/to/both(默认)
	status := c.Query("status")
	issueType := c.Query("type")

	issues, total, err := issueService.QueryIssues(taskPathName, direction, status, issueType)
	if err != nil {
		InternalError(c, "查询问题失败")
		return
	}

	Success(c, gin.H{
		"issues": issues,
		"total":  total,
	})
}

// GetIssue 获取单个问题
// GET /api/v1/issues/:fromTaskPathName/:toTaskPathName/:title
func GetIssue(c *gin.Context) {
	fromTaskPathName := strings.TrimPrefix(c.Param("fromTaskPathName"), "/")
	toTaskPathName := strings.TrimPrefix(c.Param("toTaskPathName"), "/")
	title := c.Param("title")

	// 解码路径参数（URL 编码）
	fromTaskPathName = decodePath(fromTaskPathName)
	toTaskPathName = decodePath(toTaskPathName)
	title = decodePath(title)

	issue, err := issueService.GetIssueByKey(fromTaskPathName, toTaskPathName, title)
	if err != nil {
		InternalError(c, "查询问题失败")
		return
	}

	if issue == nil {
		NotFound(c, "问题不存在")
		return
	}

	Success(c, gin.H{
		"issue": issue,
	})
}

// GetIssueByID 通过ID获取问题
// GET /api/v1/issues/id/:id
func GetIssueByID(c *gin.Context) {
	id := c.Param("id")

	issue, err := issueService.GetIssueByID(id)
	if err != nil {
		InternalError(c, "查询问题失败")
		return
	}

	if issue == nil {
		NotFound(c, "问题不存在")
		return
	}

	Success(c, gin.H{
		"issue": issue,
	})
}

// DeleteIssue 删除问题
// DELETE /api/v1/issues/:id
func DeleteIssue(c *gin.Context) {
	id := c.Param("id")

	if err := issueService.DeleteIssue(id); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"deleted": true,
	})
}

// GetIssuesByTask 获取任务相关的所有问题
// GET /api/v1/issues/by-task/*pathName
func GetIssuesByTask(c *gin.Context) {
	pathName := strings.TrimPrefix(c.Param("pathName"), "/")
	direction := c.DefaultQuery("direction", "both") // from/to/both

	issues, total, err := issueService.QueryIssues(pathName, direction, "", "")
	if err != nil {
		InternalError(c, "查询问题失败")
		return
	}

	Success(c, gin.H{
		"issues": issues,
		"total":  total,
	})
}

// decodePath 解码 URL 路径
func decodePath(path string) string {
	// 简单处理，实际可能需要 url.QueryUnescape
	return path
}

// parseMessagesFromIssue 从 Issue 中解析消息列表
func parseMessagesFromIssue(issue *models.Issue) ([]models.Message, error) {
	var messages []models.Message
	if err := json.Unmarshal([]byte(issue.Messages), &messages); err != nil {
		return nil, err
	}
	return messages, nil
}
