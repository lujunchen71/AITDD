# AITDD - AI 辅助任务治理系统

AITDD (AI-Assisted Task-Driven Development) 是一个为 AI 编程工具设计的任务治理系统。通过 MCP (Model Context Protocol) 接口，让 AI 助手能够理解项目结构、追踪任务进度、管理模块依赖，实现更智能的代码生成和项目协作。

## 核心理念

```
项目 → 模块 → 任务 → 契约
```

- **项目 (Project)**: 顶层容器，定义项目公约和编码规范
- **模块 (Module)**: 树形组织结构，支持无限层级嵌套
- **任务 (Task)**: 具体的开发任务，包含状态、契约、测试用例
- **契约 (Contract)**: 任务的输入/输出约定，明确上下游依赖

## 快速开始

### 系统要求

- Go 1.21+
- Node.js 18+
- SQLite 3

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/lujunchen71/AITDD.git
cd AITDD

# 2. 安装前端依赖
cd frontend && npm install && cd ..

# 3. 启动后端
cd backend && go run ./cmd/aitdd serve

# 4. 启动前端 (新终端)
cd frontend && npm run dev
```

### 访问地址

| 服务 | 地址 |
|------|------|
| 前端界面 | http://localhost:5173 |
| REST API | http://localhost:34567/api/v1 |
| MCP SSE | http://localhost:34567/mcp/sse |

## MCP 集成

AITDD 的核心价值在于与 AI 编程工具的深度集成。通过 MCP 协议，AI 助手可以：

- 读取项目结构和模块组织
- 获取任务详情和上下游契约
- 更新任务状态和进度
- 创建新的模块和任务
- 锁定资源避免冲突

### 配置示例

在你的 AI 编程工具的 MCP 配置文件中添加：

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

> ⚠️ **注意**: `alwaysAllow` 字段默认不开启，AI 每次调用工具都需要用户确认，以确保操作安全。

### 核心 MCP 工具

| 工具 | 说明 |
|------|------|
| `init_project` | 初始化项目配置 |
| `get_project_info` | 获取项目信息、架构、编码规范 |
| `get_constitution` | 获取项目公约 |
| `get_all_modules` | 获取所有模块列表 |
| `get_module_overview` | 获取模块概览 |
| `get_module_tasks` | 获取模块下的任务列表 |
| `get_task_detail` | 获取任务详细信息 |
| `get_task_contracts` | 获取任务的上下游契约 |
| `create_task` | 创建新任务 |
| `update_task` | 更新任务状态或内容 |
| `lock_resource` | 锁定任务或模块 |

> 完整 MCP 工具文档: [`docs/mcp-tools-reference.md`](docs/mcp-tools-reference.md)

## PathName 标识符

AITDD 使用 `pathName` 作为实体的唯一标识符，采用层级路径格式：

```
{projectPathName}/{moduleName}/{taskName}
```

示例：
- 项目: `PYQT6Calculator`
- 模块: `PYQT6Calculator/历史记录模块`
- 任务: `PYQT6Calculator/历史记录模块/历史存储服务`

当父级名称变更时，子级 pathName 会自动级联更新。

## 功能特性

### 可视化界面

- **模块树视图**: 树形展示项目模块结构
- **任务图视图**: 基于 React Flow 的依赖关系图
- **任务详情面板**: 编辑任务信息、契约、测试用例

### 任务管理

- 任务状态: `pending` → `in_progress` → `completed` / `blocked`
- 依赖关系: 支持任务间依赖，自动循环检测
- 资源锁定: 多 AI 协作时的冲突避免

### 契约系统

```json
{
  "upstreamContract": {
    "requiredData": ["用户输入"],
    "description": "上游输入描述"
  },
  "downstreamContract": {
    "producedData": ["计算结果"],
    "description": "下游输出描述"
  }
}
```

## 数据模型

```
Project (项目)
├── pathName: string (唯一标识)
├── constitution: text (项目公约)
├── architecture: text (架构说明)
└── codingStandards: text (编码规范)

Module (模块)
├── pathName: string (层级路径)
├── parentId: number (父模块ID)
├── prompt: text (模块提示词)
└── position: {x, y} (图表位置)

Task (任务)
├── pathName: string
├── moduleId: number
├── status: enum
├── upstreamContract: json
├── downstreamContract: json
├── testCases: json
└── codePaths: string[]
```

## 项目结构

```
AITDD/
├── backend/                # Go 后端服务
│   ├── cmd/aitdd/         # 主入口
│   └── internal/
│       ├── api/           # REST API
│       ├── mcp/           # MCP 服务器
│       ├── models/        # 数据模型
│       └── services/      # 业务逻辑
│
├── frontend/               # React 前端
│   └── src/
│       ├── features/      # 功能模块
│       └── stores/        # 状态管理
│
└── docs/                   # 文档
```

### 关键目录

| 目录 | 说明 |
|------|------|
| `backend/internal/mcp/` | MCP 服务器实现，工具定义 |
| `backend/internal/services/` | 业务逻辑，pathName 级联更新 |
| `frontend/src/features/` | 前端功能模块 |

## REST API

### 按 PathName 查询

```
GET /api/v1/projects/by-path/:pathName
GET /api/v1/modules/by-path/:pathName
GET /api/v1/tasks/by-path/:pathName
```

### 任务操作

```
GET    /api/v1/tasks              # 任务列表
POST   /api/v1/tasks              # 创建任务
PUT    /api/v1/tasks/:id          # 更新任务
DELETE /api/v1/tasks/:id          # 删除任务
PUT    /api/v1/tasks/:id/status   # 更新状态
```

### 资源锁定

```
POST   /api/v1/lock               # 锁定资源
DELETE /api/v1/lock               # 解锁资源
GET    /api/v1/lock/status        # 锁定状态
```

## 开发

```bash
# 后端开发
cd backend
go run ./cmd/aitdd serve  # 启动服务
go test ./...             # 运行测试

# 前端开发
cd frontend
npm run dev               # 开发模式
npm run build             # 构建生产版本
```

## 许可证

MIT License
