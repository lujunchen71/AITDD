
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

### 技术特性

- **暗色主题** - 现代化暗色 UI 设计
- **响应式布局** - 适配各种屏幕尺寸
- **乐观锁** - 基于版本的并发控制
- **MCP 接口** - 标准化的 AI 编程工具接口

## 📋 系统要求

### 后端
- Go 1.21+
- SQLite 3

### 前端
- Node.js 18+
- npm 9+

## 🛠️ 安装与运行

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

## 📁 项目结构

```
AITDD/
├── backend/                    # Go 后端
│   ├── cmd/                    # CLI 命令
│   │   └── aitdd/             # 主入口
│   ├── internal/
│   │   ├── api/               # API 层
│   │   │   ├── handlers/      # 请求处理器
│   │   │   ├── middleware/    # 中间件
│   │   │   └── response.go    # 统一响应
│   │   ├── database/          # 数据库初始化
│   │   ├── models/            # 数据模型
│   │   ├── server/            # HTTP 服务器
│   │   └── services/          # 业务逻辑
│   ├── migrations/            # 数据库迁移
│   └── go.mod                 # Go 依赖
│
├── frontend/                   # React 前端
│   ├── src/
│   │   ├── components/        # 通用组件
│   │   │   └── layout/        # 布局组件
│   │   ├── features/          # 功能模块
│   │   │   ├── modules/       # 模块管理
│   │   │   └── tasks/         # 任务管理
│   │   ├── services/          # API 服务
│   │   ├── stores/            # 状态管理
│   │   ├── types/             # TypeScript 类型
│   │   ├── App.tsx            # 主应用
│   │   ├── main.tsx           # 入口
│   │   └── index.css          # 全局样式
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
│
├── specs/                      # 规格文档
│   └── 001-visual-task-governance/
│       ├── spec.md            # 功能规格
│       ├── plan.md            # 技术计划
│       ├── data-model.md      # 数据模型
│       ├── tasks.md           # 任务列表
│       ├── contracts/         # API 契约
│       └── checklists/        # 检查清单
│
├── start-backend.bat          # Windows 后端启动脚本
├── start-frontend.bat         # Windows 前端启动脚本
└── README.md
```

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

---

**Built with ❤️ using Go, React, and SQLite**
