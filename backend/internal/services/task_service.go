package services

import (
	"errors"
	"time"

	"github.com/aitdd/backend/internal/database"
	"github.com/aitdd/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct{}

// NewTaskService 创建任务服务
func NewTaskService() *TaskService {
	return &TaskService{}
}

// GetTasks 获取任务列表
func (s *TaskService) GetTasks(moduleID string, status string, assignee string, page, pageSize int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := database.DB.Model(&models.Task{})

	if moduleID != "" {
		query = query.Where("module_id = ?", moduleID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if assignee != "" {
		query = query.Where("assignee = ?", assignee)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(id string) (*models.Task, error) {
	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(moduleID, name, description, prompt, upstreamContract, downstreamContract string) (*models.Task, error) {
	task := models.Task{
		ID:                     uuid.New().String(),
		ModuleID:               moduleID,
		Name:                   name,
		Description:            description,
		Prompt:                 prompt,
		UpstreamContractDetail: upstreamContract,
		DownstreamContractDetail: downstreamContract,
		Status:                 models.TaskStatusReady,
		CreatedAt:              time.Now().UnixMilli(),
		UpdatedAt:              time.Now().UnixMilli(),
		Version:                1,
		SyncStatus:             models.SyncStatusSynced,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(id string, updates map[string]interface{}, version int) (*models.Task, error) {
	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务不存在")
		}
		return nil, err
	}

	// 版本检查
	if task.Version != version {
		return nil, errors.New("版本冲突")
	}

	// 检查锁定
	if task.Locked {
		return nil, errors.New("任务已被锁定")
	}

	// 更新字段
	updates["version"] = task.Version + 1
	updates["updated_at"] = time.Now().UnixMilli()

	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取
	return s.GetTask(id)
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(id string) error {
	var task models.Task
	if err := database.DB.First(&task, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("任务不存在")
		}
		return err
	}

	// 检查锁定
	if task.Locked {
		return errors.New("任务已被锁定")
	}

	// 删除关联的依赖
	database.DB.Where("task_id = ? OR depends_on = ?", id, id).Delete(&models.Dependency{})

	return database.DB.Delete(&task).Error
}

// GetTasksByModule 获取模块下的所有任务
func (s *TaskService) GetTasksByModule(moduleID string) ([]models.Task, error) {
	var tasks []models.Task
	if err := database.DB.Where("module_id = ?", moduleID).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// AssignTask 分配任务
func (s *TaskService) AssignTask(id, assignee string, version int) (*models.Task, error) {
	updates := map[string]interface{}{
		"assignee": assignee,
		"status":   models.TaskStatusInProgress,
	}
	return s.UpdateTask(id, updates, version)
}

// CompleteTask 完成任务
func (s *TaskService) CompleteTask(id string, version int) (*models.Task, error) {
	updates := map[string]interface{}{
		"status": models.TaskStatusDone,
	}
	return s.UpdateTask(id, updates, version)
}
