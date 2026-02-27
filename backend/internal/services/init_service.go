package services

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitService 初始化服务
type InitService struct {
	projectDir string
	projectName string
	pluginType string
}

// NewInitService 创建初始化服务
func NewInitService(projectDir, projectName, pluginType string) *InitService {
	return &InitService{
		projectDir:  projectDir,
		projectName: projectName,
		pluginType:  pluginType,
	}
}

// Initialize 执行初始化
func (s *InitService) Initialize() error {
	// 创建目录结构
	if err := s.createDirectories(); err != nil {
		return fmt.Errorf("创建目录结构失败: %w", err)
	}

	// 创建配置文件
	if err := s.createConfigFiles(); err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}

	// 创建workflow模板
	if err := s.createWorkflowTemplate(); err != nil {
		return fmt.Errorf("创建workflow模板失败: %w", err)
	}

	// 创建AITDD_GUIDE.md
	if err := s.createGuideFile(); err != nil {
		return fmt.Errorf("创建指南文件失败: %w", err)
	}

	// 初始化数据库
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	return nil
}

// createDirectories 创建目录结构
func (s *InitService) createDirectories() error {
	dirs := []string{
		".aitdd",
		".aitdd/data",
		".aitdd/templates",
		".aitdd/logs",
		".kilocode",
		".kilocode/workflows",
		".kilocode/sessions",
		"specs",
		"specs/001-first-feature",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(s.projectDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	return nil
}

// createConfigFiles 创建配置文件
func (s *InitService) createConfigFiles() error {
	// 创建 .aitdd/project.json
	configContent := fmt.Sprintf(`{
  "projectName": "%s",
  "pluginType": "%s",
  "serverPort": 34567,
  "database": {
    "type": "sqlite",
    "path": ".aitdd/data/aitdd.db"
  },
  "sync": {
    "enabled": false,
    "remoteUrl": ""
  }
}`, s.projectName, s.pluginType)

	configPath := filepath.Join(s.projectDir, ".aitdd", "project.json")
	return os.WriteFile(configPath, []byte(configContent), 0644)
}

// createWorkflowTemplate 创建workflow模板
func (s *InitService) createWorkflowTemplate() error {
	templateService := NewTemplateService(s.projectDir, s.pluginType)
	return templateService.CreateWorkflowTemplate()
}

// createGuideFile 创建指南文件
func (s *InitService) createGuideFile() error {
	templateService := NewTemplateService(s.projectDir, s.pluginType)
	return templateService.CreateGuideFile()
}

// initDatabase 初始化数据库
func (s *InitService) initDatabase() error {
	dbService := NewDatabaseService(filepath.Join(s.projectDir, ".aitdd", "data", "aitdd.db"))
	return dbService.Initialize()
}
