# AITDD - 提示词编译器

AITDD (AI-Assisted Task-Driven Development) 是一个**提示词编译器**，用于管理和检查 AI 编程助手的任务提示词。通过静态规则检查和动态分析，确保提示词的结构完整性和契约一致性。

## 核心定位

```
提示词 → 编译检查 → AI执行 → 结果验证
```

AITDD 将 AI 编程视为一个"编译"过程：
- **提示词 (Prompt)** = 模块/任务的描述和契约
- **编译 (Compile)** = 静态规则检查 + 动态依赖分析
- **执行 (Execute)** = AI 根据提示词生成代码
- **验证 (Verify)** = 测试用例检查 + 契约一致性验证

## 编译检查规则

### 静态检查

| 规则ID | 检查项 | 严重级别 |
|--------|--------|----------|
| S-01 | 根对象结构完整性 | error |
| S-02 | 项目字段完整性 | error |
| S-03~07 | 模块字段格式检查 | error |
| T-01~05 | 任务字段完整性 | error |
| T-06~10 | 任务可选字段验证 | warning |
| T-11~14 | 任务契约格式检查 | error/warning |
| T-15~16 | 任务依赖引用检查 | error |
| D-01~08 | 依赖关系完整性 | error |
| D-09~11 | 跨模块依赖警告 | warning/info |

### 动态检查

| 规则ID | 检查项 | 说明 |
|--------|--------|------|
| DYN-01 | 契约一致性 | 下游输入应匹配上游输出 |
| DYN-02 | 模块设计合理性 | 任务数量建议 1-10个 |
| DYN-05 | 阻塞任务检测 | 长期阻塞的任务告警 |
| DYN-06 | 依赖链完整性 | 检查断裂的依赖链 |
| DYN-07 | 关键路径分析 | 识别项目关键路径 |
| DYN-09 | 模块耦合度 | 检查模块间耦合程度 |
| DYN-12 | 测试覆盖率 | 检查测试覆盖情况 |
| DYN-13 | 问题跟踪 | 检查任务问题状态 |

### 规则配置

规则定义在 [`.aitdd/rule.json`](.aitdd/rule.json) 中，支持：
- 自定义规则启用/禁用
- 调整严重级别 (error/warning/info)
- 添加新的检查规则

## 快速开始

### 系统要求

- Go 1.21+
- Node.js 18+
- SQLite 3

### 安装步骤

```bash
# 克隆项目
git clone https://github.com/lujunchen71/AITDD.git
cd AITDD

# 安装前端依赖
cd frontend && npm install && cd ..

# 启动后端
cd backend && go run ./cmd/aitdd serve

# 启动前端 (新终端)
cd frontend && npm run dev
```

### 访问地址

| 服务 | 地址 |
|------|------|
| 前端界面 | http://localhost:5173 |
| REST API | http://localhost:34567/api/v1 |
| MCP SSE | http://localhost:34567/mcp/sse |

## 数据模型

### 层级结构

```
Project (项目)
├── pathName: 唯一标识
├── constitution: 项目公约
├── architecture: 架构说明
└── codingStandards: 编码规范

Module (模块)
├── pathName: 层级路径 (如: 项目名/模块名)
├── prompt: 模块提示词 ⭐
├── parentId: 父模块引用
└── status: designing/developing/completed/deprecated

Task (任务)
├── pathName: 层级路径 (如: 项目名/模块名/任务名)
├── prompt: 任务提示词 ⭐
├── upstreamContract: 上游契约
├── downstreamContract: 下游契约
├── testCases: 测试用例
├── codePaths: 代码路径
└── status: pending/in_progress/completed/blocked
```

### 契约系统

契约定义了任务间的"接口约定"：

```json
{
  "upstreamContract": {
    "requiredData": ["用户输入", "配置参数"],
    "description": "上游模块提供的数据"
  },
  "downstreamContract": {
    "producedData": ["计算结果", "日志记录"],
    "description": "本任务产出的数据"
  }
}
```

契约一致性检查 (DYN-01) 会验证：
- 下游任务的 `requiredData` 是否在上游任务的 `producedData` 中定义
- 依赖链上的数据流转是否完整

### PathName 标识符

使用层级路径作为唯一标识符：

```
{projectPathName}/{moduleName}/{taskName}
```

示例：
- 项目: `PYQT6Calculator`
- 模块: `PYQT6Calculator/历史记录模块`
- 任务: `PYQT6Calculator/历史记录模块/历史存储服务`

父级名称变更时，子级 pathName 自动级联更新。

## MCP 集成

通过 MCP (Model Context Protocol) 协议，AI 编程工具可以：

- 读取项目结构和提示词
- 获取任务的上下游契约
- 更新任务状态和进度
- 执行编译检查

### 配置示例

```json
{
  "mcpServers": {
    "aitdd": {
      "url": "http://localhost:34567/mcp/sse",
      "transport": "sse",
      "enabled": true
    }
  }
}
```

### 核心 MCP 工具

| 工具 | 说明 |
|------|------|
| `init_project` | 初始化项目 |
| `get_project_info` | 获取项目信息、架构、编码规范 |
| `get_module_overview` | 获取模块概览和提示词 |
| `get_task_detail` | 获取任务详情和契约 |
| `get_task_contracts` | 获取上下游契约接口 |
| `check_module` | 执行模块编译检查 ⭐ |
| `update_task` | 更新任务状态 |
| `lock_resource` | 锁定资源避免冲突 |

> 完整 MCP 工具文档: [`docs/mcp-tools-reference.md`](docs/mcp-tools-reference.md)

## 编译检查 API

### MCP 工具: check_module

```
参数:
- pathName: 模块路径名称
- rules: 指定检查规则ID列表 (可选)
- includeDynamic: 是否包含动态检查 (默认 true)

返回:
- summary: 检查摘要 (error/warning/info 数量)
- errors: 错误列表
- warnings: 警告列表
- suggestions: 建议列表
```

### REST API

```
GET /api/v1/modules/by-path/:pathName    # 按路径获取模块
GET /api/v1/tasks/by-path/:pathName      # 按路径获取任务
POST /api/v1/dependencies/validate       # 验证依赖循环
```

## 项目结构

```
AITDD/
├── backend/
│   └── internal/
│       ├── mcp/
│       │   ├── rule_engine.go    # 规则引擎 ⭐
│       │   ├── server.go         # MCP 服务器
│       │   └── tools_*.go        # MCP 工具定义
│       ├── api/handlers/         # REST API
│       ├── models/               # 数据模型
│       └── services/             # 业务逻辑
│
├── frontend/
│   └── src/features/
│       ├── modules/              # 模块树/图视图
│       └── tasks/                # 任务管理
│
├── .aitdd/
│   ├── rule.json                 # 规则配置 ⭐
│   └── project.json              # 项目配置
│
└── docs/
    ├── mcp-tools-reference.md    # MCP 工具文档
    └── database-models.md        # 数据模型文档
```

## 开发

```bash
# 后端
cd backend
go run ./cmd/aitdd serve  # 启动服务
go test ./...             # 运行测试

# 前端
cd frontend
npm run dev               # 开发模式
npm run build             # 构建
```

## 许可证

MIT License
