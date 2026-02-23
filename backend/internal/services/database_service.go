package services

import (
	"fmt"

	"github.com/aitdd/backend/internal/database"
)

// DatabaseService 数据库服务
type DatabaseService struct {
	dbPath string
}

// NewDatabaseService 创建数据库服务
func NewDatabaseService(dbPath string) *DatabaseService {
	return &DatabaseService{
		dbPath: dbPath,
	}
}

// Initialize 初始化数据库
func (s *DatabaseService) Initialize() error {
	cfg := &database.Config{
		Path: s.dbPath,
	}

	if err := database.Initialize(cfg); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (s *DatabaseService) Close() error {
	return database.Close()
}
