package mcp

import (
	"sync"
	"time"

	"github.com/aitdd/backend/internal/models"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      interface{}
	ExpiresAt int64 // 过期时间戳（毫秒）
}

// IsExpired 检查缓存项是否过期
func (item *CacheItem) IsExpired() bool {
	return time.Now().UnixMilli() > item.ExpiresAt
}

// MCPCache MCP 缓存管理器
type MCPCache struct {
	// 项目缓存：pathName -> *Project
	projects sync.Map

	// 模块缓存：pathName -> *Module
	modules sync.Map

	// 任务缓存：pathName -> *Task
	tasks sync.Map

	// 模块列表缓存：projectPathName -> []Module
	moduleLists sync.Map

	// 任务列表缓存：modulePathName -> []Task
	taskLists sync.Map

	// 默认过期时间（毫秒）
	defaultExpiry int64

	// 清理间隔
	cleanupInterval time.Duration

	// 停止清理信号
	stopCleanup chan struct{}
}

// 默认缓存实例
var defaultCache *MCPCache
var once sync.Once

// DefaultCacheExpiry 默认缓存过期时间（5分钟）
const DefaultCacheExpiry = 5 * 60 * 1000

// GetMCPCache 获取默认缓存实例
func GetMCPCache() *MCPCache {
	once.Do(func() {
		defaultCache = NewMCPCache(DefaultCacheExpiry)
	})
	return defaultCache
}

// NewMCPCache 创建新的缓存实例
func NewMCPCache(expiryMs int64) *MCPCache {
	c := &MCPCache{
		defaultExpiry:   expiryMs,
		cleanupInterval: time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// 启动后台清理协程
	go c.cleanup()

	return c
}

// Stop 停止缓存清理
func (c *MCPCache) Stop() {
	close(c.stopCleanup)
}

// cleanup 定期清理过期缓存
func (c *MCPCache) cleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanExpired 清理所有过期的缓存项
func (c *MCPCache) cleanExpired() {
	// 清理项目缓存
	c.projects.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok && item.IsExpired() {
			c.projects.Delete(key)
		}
		return true
	})

	// 清理模块缓存
	c.modules.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok && item.IsExpired() {
			c.modules.Delete(key)
		}
		return true
	})

	// 清理任务缓存
	c.tasks.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok && item.IsExpired() {
			c.tasks.Delete(key)
		}
		return true
	})

	// 清理模块列表缓存
	c.moduleLists.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok && item.IsExpired() {
			c.moduleLists.Delete(key)
		}
		return true
	})

	// 清理任务列表缓存
	c.taskLists.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok && item.IsExpired() {
			c.taskLists.Delete(key)
		}
		return true
	})
}

// ==================== 项目缓存 ====================

// GetProject 获取项目缓存
func (c *MCPCache) GetProject(pathName string) (*models.Project, bool) {
	if value, ok := c.projects.Load(pathName); ok {
		if item, ok := value.(*CacheItem); ok && !item.IsExpired() {
			if project, ok := item.Data.(*models.Project); ok {
				return project, true
			}
		}
		// 过期或类型错误，删除缓存
		c.projects.Delete(pathName)
	}
	return nil, false
}

// SetProject 设置项目缓存
func (c *MCPCache) SetProject(pathName string, p *models.Project) {
	c.projects.Store(pathName, &CacheItem{
		Data:      p,
		ExpiresAt: time.Now().UnixMilli() + c.defaultExpiry,
	})
}

// DeleteProject 删除项目缓存
func (c *MCPCache) DeleteProject(pathName string) {
	c.projects.Delete(pathName)
	// 同时删除相关的模块列表缓存
	c.moduleLists.Delete(pathName)
}

// ==================== 模块缓存 ====================

// GetModule 获取模块缓存
func (c *MCPCache) GetModule(pathName string) (*models.Module, bool) {
	if value, ok := c.modules.Load(pathName); ok {
		if item, ok := value.(*CacheItem); ok && !item.IsExpired() {
			if module, ok := item.Data.(*models.Module); ok {
				return module, true
			}
		}
		c.modules.Delete(pathName)
	}
	return nil, false
}

// SetModule 设置模块缓存
func (c *MCPCache) SetModule(pathName string, m *models.Module) {
	c.modules.Store(pathName, &CacheItem{
		Data:      m,
		ExpiresAt: time.Now().UnixMilli() + c.defaultExpiry,
	})
}

// DeleteModule 删除模块缓存
func (c *MCPCache) DeleteModule(pathName string) {
	c.modules.Delete(pathName)
	// 同时删除相关的任务列表缓存
	c.taskLists.Delete(pathName)
}

// GetModuleList 获取模块列表缓存
func (c *MCPCache) GetModuleList(projectPathName string) ([]models.Module, bool) {
	if value, ok := c.moduleLists.Load(projectPathName); ok {
		if item, ok := value.(*CacheItem); ok && !item.IsExpired() {
			if modules, ok := item.Data.([]models.Module); ok {
				return modules, true
			}
		}
		c.moduleLists.Delete(projectPathName)
	}
	return nil, false
}

// SetModuleList 设置模块列表缓存
func (c *MCPCache) SetModuleList(projectPathName string, modules []models.Module) {
	c.moduleLists.Store(projectPathName, &CacheItem{
		Data:      modules,
		ExpiresAt: time.Now().UnixMilli() + c.defaultExpiry,
	})
}

// ==================== 任务缓存 ====================

// GetTask 获取任务缓存
func (c *MCPCache) GetTask(pathName string) (*models.Task, bool) {
	if value, ok := c.tasks.Load(pathName); ok {
		if item, ok := value.(*CacheItem); ok && !item.IsExpired() {
			if task, ok := item.Data.(*models.Task); ok {
				return task, true
			}
		}
		c.tasks.Delete(pathName)
	}
	return nil, false
}

// SetTask 设置任务缓存
func (c *MCPCache) SetTask(pathName string, t *models.Task) {
	c.tasks.Store(pathName, &CacheItem{
		Data:      t,
		ExpiresAt: time.Now().UnixMilli() + c.defaultExpiry,
	})
}

// DeleteTask 删除任务缓存
func (c *MCPCache) DeleteTask(pathName string) {
	c.tasks.Delete(pathName)
}

// GetTaskList 获取任务列表缓存
func (c *MCPCache) GetTaskList(modulePathName string) ([]models.Task, bool) {
	if value, ok := c.taskLists.Load(modulePathName); ok {
		if item, ok := value.(*CacheItem); ok && !item.IsExpired() {
			if tasks, ok := item.Data.([]models.Task); ok {
				return tasks, true
			}
		}
		c.taskLists.Delete(modulePathName)
	}
	return nil, false
}

// SetTaskList 设置任务列表缓存
func (c *MCPCache) SetTaskList(modulePathName string, tasks []models.Task) {
	c.taskLists.Store(modulePathName, &CacheItem{
		Data:      tasks,
		ExpiresAt: time.Now().UnixMilli() + c.defaultExpiry,
	})
}

// ==================== 批量操作 ====================

// InvalidateByPath 根据路径前缀失效相关缓存
func (c *MCPCache) InvalidateByPath(pathPrefix string) {
	// 失效模块缓存
	c.modules.Range(func(key, value interface{}) bool {
		if pathName, ok := key.(string); ok {
			if len(pathName) > len(pathPrefix) && pathName[:len(pathPrefix)] == pathPrefix {
				c.modules.Delete(key)
			}
		}
		return true
	})

	// 失效任务缓存
	c.tasks.Range(func(key, value interface{}) bool {
		if pathName, ok := key.(string); ok {
			if len(pathName) > len(pathPrefix) && pathName[:len(pathPrefix)] == pathPrefix {
				c.tasks.Delete(key)
			}
		}
		return true
	})
}

// InvalidateAll 清空所有缓存
func (c *MCPCache) InvalidateAll() {
	c.projects = sync.Map{}
	c.modules = sync.Map{}
	c.tasks = sync.Map{}
	c.moduleLists = sync.Map{}
	c.taskLists = sync.Map{}
}

// CacheStats 缓存统计
type CacheStats struct {
	Projects    int `json:"projects"`
	Modules     int `json:"modules"`
	Tasks       int `json:"tasks"`
	ModuleLists int `json:"moduleLists"`
	TaskLists   int `json:"taskLists"`
}

// GetStats 获取缓存统计信息
func (c *MCPCache) GetStats() CacheStats {
	stats := CacheStats{}

	c.projects.Range(func(_, _ interface{}) bool {
		stats.Projects++
		return true
	})

	c.modules.Range(func(_, _ interface{}) bool {
		stats.Modules++
		return true
	})

	c.tasks.Range(func(_, _ interface{}) bool {
		stats.Tasks++
		return true
	})

	c.moduleLists.Range(func(_, _ interface{}) bool {
		stats.ModuleLists++
		return true
	})

	c.taskLists.Range(func(_, _ interface{}) bool {
		stats.TaskLists++
		return true
	})

	return stats
}

// ==================== 缓存预热 ====================

// WarmupProject 预热项目缓存
func (c *MCPCache) WarmupProject(project *models.Project, modules []models.Module) {
	// 缓存项目
	c.SetProject(project.PathName, project)

	// 缓存模块列表
	c.SetModuleList(project.PathName, modules)

	// 缓存单个模块
	for i := range modules {
		c.SetModule(modules[i].PathName, &modules[i])
	}
}
