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
		pluginType: pluginType,
	}
}

// CreateWorkflowTemplate 创建workflow模板
func (s *TemplateService) CreateWorkflowTemplate() error {
	var templateContent string
	var templatePath string

	switch s.pluginType {
	case "kilocode":
		templateContent = s.getKiloCodeTemplate()
		templatePath = ".kilocode/workflows/default.md"
	case "opencode":
		templateContent = s.getOpenCodeTemplate()
		templatePath = ".opencode/workflows/default.md"
	case "claudecode":
		templateContent = s.getClaudeCodeTemplate()
		templatePath = ".claude/workflows/default.md"
	default:
		templateContent = s.getKiloCodeTemplate()
		templatePath = ".kilocode/workflows/default.md"
	}

	fullPath := filepath.Join(s.projectDir, templatePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("创建模板目录失败: %w", err)
	}

	return os.WriteFile(fullPath, []byte(templateContent), 0644)
}

// CreateGuideFile 创建指南文件
func (s *TemplateService) CreateGuideFile() error {
	guideContent := s.getGuideContent()
	guidePath := filepath.Join(s.projectDir, "AITDD_GUIDE.md")
	return os.WriteFile(guidePath, []byte(guideContent), 0644)
}

func (s *TemplateService) getKiloCodeTemplate() string {
	return `# KiloCode Workflow Template

## 项目信息
- 项目名称: {{.ProjectName}}
- 创建时间: {{.CreatedAt}}

## 工作流程

### 1. 任务分析
- 分析需求文档
- 识别关键功能点
- 评估技术复杂度

### 2. 模块设计
- 划分功能模块
- 定义模块边界
- 设计模块接口

### 3. 任务分解
- 创建任务列表
- 设置任务依赖
- 估算工作量

### 4. 实现开发
- 按任务顺序开发
- 编写单元测试
- 代码审查

### 5. 测试验证
- 集成测试
- 性能测试
- 用户验收

## MCP工具使用

### 获取项目上下文
\`\`\`
aitdd.get_context
\`\`\`

### 创建模块
\`\`\`
aitdd.create_module
\`\`\`

### 创建任务
\`\`\`
aitdd.create_task
\`\`\`

### 更新任务状态
\`\`\`
aitdd.update_task_status
\`\`\`
`
}

func (s *TemplateService) getOpenCodeTemplate() string {
	return `# OpenCode Workflow Template

## 项目信息
- 项目名称: {{.ProjectName}}
- 创建时间: {{.CreatedAt}}

## 工作流程

### 1. 需求理解
- 阅读需求文档
- 确认功能范围
- 识别风险点

### 2. 架构设计
- 设计系统架构
- 选择技术栈
- 规划数据模型

### 3. 任务规划
- 分解开发任务
- 排列优先级
- 分配资源

### 4. 编码实现
- 编写代码
- 单元测试
- 代码重构

### 5. 质量保证
- 代码审查
- 测试覆盖
- 性能优化

## AITDD集成

### 获取任务列表
\`\`\`
GET /api/v1/tasks
\`\`\`

### 创建新任务
\`\`\`
POST /api/v1/tasks
\`\`\`

### 更新任务
\`\`\`
PUT /api/v1/tasks/:id
\`\`\`
`
}

func (s *TemplateService) getClaudeCodeTemplate() string {
	return `# ClaudeCode Workflow Template

## 项目信息
- 项目名称: {{.ProjectName}}
- 创建时间: {{.CreatedAt}}

## 工作流程

### 1. 需求分析
- 理解业务需求
- 识别用户场景
- 定义验收标准

### 2. 设计阶段
- 概要设计
- 详细设计
- 接口设计

### 3. 开发阶段
- 功能开发
- 代码测试
- 文档编写

### 4. 测试阶段
- 功能测试
- 集成测试
- 回归测试

### 5. 发布阶段
- 部署准备
- 上线发布
- 监控运维

## AITDD工具

### 项目信息
\`\`\`
GET /api/v1/project
\`\`\`

### 模块管理
\`\`\`
GET/POST /api/v1/modules
\`\`\`

### 任务管理
\`\`\`
GET/POST /api/v1/tasks
\`\`\`
`
}

func (s *TemplateService) getGuideContent() string {
	return `# AITDD 使用指南

## 简介

AITDD是一个AI辅助可视化任务治理系统，帮助您更好地管理开发任务和模块。

## 快速开始

### 1. 启动服务

\`\`\`bash
aitdd serve
\`\`\`

服务将在 http://localhost:34567 启动

### 2. 访问界面

在浏览器中打开 http://localhost:34567 即可访问可视化界面。

### 3. API端点

- 项目信息: \`GET /api/v1/project\`
- 模块列表: \`GET /api/v1/modules\`
- 任务列表: \`GET /api/v1/tasks\`
- 依赖关系: \`GET /api/v1/dependencies\`
- 通知列表: \`GET /api/v1/notifications\`

## MCP工具

AITDD提供以下MCP工具供AI编程助手使用：

### 项目管理
- \`aitdd.get_context\` - 获取项目上下文
- \`aitdd.get_constitution\` - 获取项目宪法

### 模块管理
- \`aitdd.list_modules\` - 列出所有模块
- \`aitdd.get_module\` - 获取模块详情
- \`aitdd.create_module\` - 创建模块
- \`aitdd.update_module\` - 更新模块
- \`aitdd.delete_module\` - 删除模块

### 任务管理
- \`aitdd.list_tasks\` - 列出所有任务
- \`aitdd.get_task\` - 获取任务详情
- \`aitdd.create_task\` - 创建任务
- \`aitdd.update_task\` - 更新任务
- \`aitdd.delete_task\` - 删除任务
- \`aitdd.claim_task\` - 认领任务

### 依赖管理
- \`aitdd.list_dependencies\` - 列出依赖关系
- \`aitdd.create_dependency\` - 创建依赖
- \`aitdd.delete_dependency\` - 删除依赖

## 目录结构

\`\`\`
.
├── .aitdd/              # AITDD配置目录
│   ├── config.json      # 配置文件
│   ├── data/            # 数据目录
│   │   └── aitdd.db     # SQLite数据库
│   └── logs/            # 日志目录
├── .kilocode/           # KiloCode配置
│   └── workflows/       # 工作流模板
├── specs/               # 规格文档目录
└── AITDD_GUIDE.md       # 本指南文件
\`\`\`

## 常见问题

### Q: 如何修改服务端口？
A: 编辑 \`.aitdd/config.json\` 文件中的 \`serverPort\` 配置。

### Q: 如何启用远程同步？
A: 在配置文件中设置 \`sync.enabled\` 为 \`true\` 并配置 \`sync.remoteUrl\`。

### Q: 数据存储在哪里？
A: 默认使用SQLite数据库，存储在 \`.aitdd/data/aitdd.db\`。

## 更多信息

- GitHub: https://github.com/aitdd/aitdd
- 文档: https://docs.aitdd.io
- 问题反馈: https://github.com/aitdd/aitdd/issues
`
}
