package services

import (
	"fmt"
	"os"
	"path/filepath"
)

// TemplateService 模板服务
type TemplateService struct {
	projectDir string
	pluginType string
}

// NewTemplateService 创建模板服务
func NewTemplateService(projectDir, pluginType string) *TemplateService {
	return &TemplateService{
		projectDir: projectDir,
		pluginType:  pluginType,
	}
}

// CreateWorkflowTemplate 创建workflow模板
func (s *TemplateService) CreateWorkflowTemplate() error {
	workflowContent := `# AITDD Workflow

## 开发流程

1. **需求分析** - 分析需求，创建规格文档
2. **设计阶段** - 设计系统架构和接口
3. **实现阶段** - 编写代码实现功能
4. **测试阶段** - 编写测试用例并执行测试
5. **部署阶段** - 部署到生产环境

## 模块依赖规则

- 模块之间的依赖关系需要明确定义
- 避免循环依赖
- 保持单向数据流
`

	workflowPath := filepath.Join(s.projectDir, ".kilocode", "workflows", "aitdd.start.md")
	return os.WriteFile(workflowPath, []byte(workflowContent), 0644)
}

// CreateGuideFile 创建指南文件
func (s *TemplateService) CreateGuideFile() error {
	guideContent := fmt.Sprintf(`# AITDD 使用指南

## 项目信息

- **项目名称**: %s
- **插件类型**: %s

## 快速开始

1. 创建模块定义功能边界
2. 为每个模块创建任务
3. 定义任务之间的依赖关系
4. 按依赖顺序执行任务

## 最佳实践

- 保持模块职责单一
- 任务粒度适中
- 及时更新任务状态
- 定期同步数据
`, s.projectDir, s.pluginType)

	guidePath := filepath.Join(s.projectDir, "AITDD_GUIDE.md")
	return os.WriteFile(guidePath, []byte(guideContent), 0644)
}
