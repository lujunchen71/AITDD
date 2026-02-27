
# AITDD - AI-Assisted Visual Task Governance System

AITDD 是一个 AI 辅助的可视化任务治理系统，为第三方 AI 编程软件提供 MCP (Model Context Protocol) 接口，使用本地 SQLite 数据库存储，并提供现代化的 React 前端界面。

## 🚀 功能特性

### 核心功能

- **项目管理** - 创建和管理多个项目
- **模块树结构** - 支持无限层级的模块组织
- **任务管理** - 完整的任务 CRUD 操作
- **依赖关系** - 任务间依赖关系管理，支持循环检测
- **可视化任务图** - 基于 React Flow 的交互式任务网络图
- **契约管理** - 任务输入/输出契约定义
- **资源锁定** - 支持 AI 代理协作的资源锁机制
- **通知系统** - WebSocket 实时通知
- **变更历史** - 完整的操作审计日志
- **模块位置持久化** - 图表视图中的节点位置保存

### 技术特性

- **暗色主题** - 现代化暗色 UI 设计
- **响应式布局** - 适配各种屏幕尺寸
- **乐观锁** - 基于版本的并发控制
- **MCP 接口** - 标准化的 AI 编程工具接口

## 📋 系统要求

### 后端
- Go 1.21+
- SQLite 3
- Delve (调试器，可选)

### 前端
- Node.js 18+
- npm 9+

## 🛠️ 安装与运行

### 环境安装指南

#### 1. Go 语言环境安装

如果你还没有安装 Go 语言环境，需要先访问 Go 官方下载页面下载对应操作系统的安装包进行安装：

**下载地址**: https://go.dev/dl/

**安装后验证**:
```bash
go version
```

#### 2. 设置 Go 调试器 (Delve)

Delve 是 Go 语言的调试器，用于开发调试：

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

**验证安装**:
```bash
dlv version
```

> 💡 **提示**: 确保 Go 的 bin 目录已添加到系统 PATH 环境变量中（通常为 `$GOPATH/bin` 或 `$HOME/go/bin`）

#### 3. Node.js 环境安装

访问 Node.js 官网下载并安装 LTS 版本：https://nodejs.org/

**验证安装**:
```bash
node --version
npm --version
```

### 前端依赖安装

前端项目基于 React + TypeScript，需要安装以下依赖：

#### 核心框架依赖
```bash
cd frontend
npm install react react-dom
```

#### TypeScript 类型支持
```bash
npm install -D typescript @types/react @types/react-dom
```

> 💡 **说明**: TypeScript 配合 React 使用时，必须安装 `@types/react` 和 `@types/react-dom` 来提供 React 的类型定义。

#### 路由依赖
项目中有多个功能模块（dashboard、modules、tasks等），需要路由来组织：

```bash
npm install react-router-dom
```

> 💡 **说明**: `react-router-dom` 的 TypeScript 类型已内置，无需额外安装。

#### 状态管理 (Zustand)
项目使用 Zustand 进行状态管理：

```bash
npm install zustand
```

如果需要 Immer 支持（不可变数据更新）：

```bash
npm install immer
# 或使用 Zustand 的 immer 中间件
npm install zustand-middleware
```

#### 开发工具和构建相关

项目使用 Vite 作为构建工具：

```bash
# Vite 核心和 React 插件
npm install -D vite @vitejs/plugin-react

# TypeScript 相关
npm install -D typescript @types/node

# 如果需要路径别名支持（如 @/ 指向 src/）
npm install -D vite-tsconfig-paths
```

#### TypeScript 类型定义配置

项目的 `src/types/` 目录用于存放全局类型定义，需要在 `tsconfig.json` 中配置：

```json
{
  "compilerOptions": {
    "typeRoots": ["./node_modules/@types", "./src/types"]
  }
}
```

#### 一键安装所有前端依赖

```bash
cd frontend
npm install
```

> 💡 **说明**: `package.json` 中已配置好所有依赖，执行 `npm install` 即可自动安装所有必需的包。

### 快速启动

**Windows 用户:**

```batch
# 启动后端
start-backend.bat

# 启动前端 (新终端窗口)
start-frontend.bat
```

**手动启动:**

```bash
# 后端
cd backend
go mod download
go build -o aitdd ./cmd/aitdd
./aitdd serve

# 前端
cd frontend
npm install
npm run dev
```

### 访问应用

- **前端界面**: http://localhost:5173
- **API 接口**: http://localhost:34567/api/v1
- **WebSocket**: ws://localhost:34567/ws

## 🔌 API 接口

### 项目 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/projects` | 获取项目列表 |
| POST | `/api/v1/projects` | 创建项目 |
| GET | `/api/v1/projects/:id` | 获取项目详情 |
| PUT | `/api/v1/projects/:id` | 更新项目 |
| DELETE | `/api/v1/projects/:id` | 删除项目 |

### 模块 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/modules` | 获取模块列表 |
| POST | `/api/v1/modules` | 创建模块 |
| GET | `/api/v1/modules/:id` | 获取模块详情 |
| PUT | `/api/v1/modules/:id` | 更新模块 |
| DELETE | `/api/v1/modules/:id` | 删除模块 |
| GET | `/api/v1/modules/tree` | 获取模块树 |
| PUT | `/api/v1/modules/:id/position` | 更新模块位置 |

### 任务 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/tasks` | 获取任务列表 |
| POST | `/api/v1/tasks` | 创建任务 |
| GET | `/api/v1/tasks/:id` | 获取任务详情 |
| PUT | `/api/v1/tasks/:id` | 更新任务 |
| DELETE | `/api/v1/tasks/:id` | 删除任务 |
| PUT | `/api/v1/tasks/:id/status` | 更新任务状态 |

### 依赖 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/dependencies` | 获取依赖列表 |
| POST | `/api/v1/dependencies` | 创建依赖 |
| DELETE | `/api/v1/dependencies/:id` | 删除依赖 |
| POST | `/api/v1/dependencies/validate` | 验证依赖 (循环检测) |

### 锁定 API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/lock` | 锁定资源 |
| DELETE | `/api/v1/lock` | 解锁资源 |
| GET | `/api/v1/lock/status` | 获取锁定状态 |

### 通知 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/notifications` | 获取通知列表 |
| PUT | `/api/v1/notifications/:id/read` | 标记已读 |
| WebSocket | `/ws` | 实时通知连接 |

## 🎨 UI 设计

### 颜色主题 (暗色)

- **背景色**: `#1a1a2e`
- **卡片色**: `#16213e`
- **边框色**: `#0f3460`
- **强调色**: `#e94560`
- **文字色**: `#eaeaea`

### 任务状态颜色

| 状态 | 颜色 | 描述 |
|------|------|------|
| pending | 灰色 | 待处理 |
| in_progress | 蓝色 | 进行中 |
| completed | 绿色 | 已完成 |
| blocked | 红色 | 已阻塞 |
| cancelled | 暗灰色 | 已取消 |

## 📊 数据模型

### 核心实体

```
Project (项目)
├── Module (模块) [树形结构]
│   └── Task (任务)
│       ├── Dependency (依赖关系)
│       └── ChangeHistory (变更历史)
└── Notification (通知)
```

### 任务契约结构

```json
{
  "inputContract": {
    "requiredFiles": ["string"],
    "requiredData": ["string"],
    "description": "string"
  },
  "outputContract": {
    "producedFiles": ["string"],
    "producedData": ["string"],
    "description": "string"
  }
}
```

## 🔧 配置

### MCP Server 安装

AITDD 提供 MCP (Model Context Protocol) 服务器，可与支持 MCP 的 AI 编程工具（如 Kilo Code、Cursor等）集成。

#### 安装步骤

1. **编译后端**
   ```bash
   cd backend
   go build -o aitdd ./cmd/aitdd
   ```

2. **启动后端服务**
   ```bash
   cd backend
   go run ./cmd/aitdd serve
   ```

3. **配置 MCP 客户端**
   
   在你的 AI 编程工具的 MCP 配置文件中添加（使用 SSE 传输方式）：
   ```json
   {
     "mcpServers": {
       "aitdd": {
         "url": "http://localhost:34567/mcp/sse",
         "transport": "sse",
         "enabled": true,
         "alwaysAllow": [
           "get_config",
           "get_project_info",
           "get_constitution",
           "get_module_overview",
           "get_module_tasks",
           "get_task_detail",
           "get_all_modules",
           "get_all_task_status",
           "update_module",
           "update_task"
         ]
       }
     }
   }
   ```

4. **初始化项目**
   
   首次使用时，通过 MCP 工具 `init_project` 选择或创建项目。

#### 核心 MCP 工具

| 工具 | 说明 |
|------|------|
| `init_project` | 初始化项目配置 |
| `get_project_info` | 获取项目信息 |
| `get_all_modules` | 获取所有模块 |
| `get_module_tasks` | 获取模块任务列表 |
| `get_task_detail` | 获取任务详情 |
| `create_task` | 创建任务 |
| `update_task` | 更新任务 |
| `lock_resource` | 锁定资源 |

> 详细 MCP 工具文档请参考 [`docs/mcp-tools-reference.md`](docs/mcp-tools-reference.md)

### 环境变量

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `AITDD_PORT` | 34567 | 服务端口 |
| `AITDD_DB_PATH` | ./aitdd.db | 数据库路径 |
| `AITDD_FRONTEND_DIR` | ./frontend/dist | 前端静态文件目录 |

### CLI 命令

```bash
# 初始化项目
aitdd init [path]

# 启动服务
aitdd serve [flags]

# 查看版本
aitdd version
```

## 🧪 开发

### 前端开发

```bash
cd frontend
npm run dev      # 开发模式
npm run build    # 构建生产版本
npm run preview  # 预览生产版本
```

### 后端开发

```bash
cd backend
go run ./cmd/aitdd serve  # 开发模式
go test ./...             # 运行测试
```

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📁 项目结构

```
AITDD/
├── backend/                # Go 后端服务
│   ├── cmd/aitdd/         # 主入口
│   ├── internal/
│   │   ├── api/           # REST API 处理器
│   │   ├── mcp/           # MCP 服务器实现
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 业务逻辑层
│   │   └── database/      # 数据库初始化
│   └── migrations/        # SQL 迁移脚本
│
├── frontend/               # React 前端
│   └── src/
│       ├── components/    # 通用组件
│       ├── features/      # 功能模块
│       ├── stores/        # Zustand 状态管理
│       └── services/      # API 服务
│
├── docs/                   # 项目文档
├── scripts/                # 工具脚本
└── specs/                  # 规格文档
```

### 重点目录说明

| 目录 | 说明 |
|------|------|
| `backend/internal/mcp/` | MCP 服务器核心实现，包含所有工具定义 |
| `backend/internal/api/handlers/` | REST API 处理器 |
| `backend/internal/services/` | 业务逻辑层，处理 pathName 级联更新等 |
| `frontend/src/features/` | 前端功能模块（模块树、任务图等） |

---
