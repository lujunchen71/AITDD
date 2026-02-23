# 用户安装流程文档

## 安装包与依赖

### 支持的平台

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Windows | x64 | `aitdd-windows-x64.exe` |
| macOS | x64/ARM | `aitdd-darwin-x64`, `aitdd-darwin-arm64` |
| Linux | x64 | `aitdd-linux-x64` |

### 依赖

- **无外部依赖**：可执行文件为静态编译，包含所有运行时
- **操作系统要求**：
  - Windows 10+
  - macOS 11+
  - Linux (glibc 2.31+)

### 安装包内容

```
aitdd-v1.0.0/
├── aitdd.exe (或 aitdd)      # 主程序
├── static/                    # 前端静态文件
│   ├── index.html
│   ├── assets/
│   └── ...
├── templates/                 # workflow模板
│   ├── kilocode/
│   ├── opencode/
│   └── claudecode/
└── README.md
```

---

## 命令行接口

### 全局命令

```bash
aitdd [command] [options]

Commands:
  init        初始化项目
  serve       启动服务
  version     显示版本信息
  help        显示帮助信息

Options:
  --config    指定配置文件路径
  --port      指定服务端口（默认34567）
  --verbose   显示详细日志
```

---

## aitdd init（初始化命令）

### 执行流程

```
用户执行 aitdd init
        │
        ▼
┌───────────────────────┐
│ 检查是否已初始化        │
│ (.aitdd目录是否存在)    │
└───────────┬───────────┘
            │
      ┌─────┴─────┐
      │ 已存在？   │
      └─────┬─────┘
       是 │     │ 否
          │     │
          ▼     ▼
    ┌──────┐ ┌───────────────────────┐
    │ 提示  │ │ 创建.aitdd目录         │
    │ 退出  │ └───────────┬───────────┘
    └──────┘             │
                         ▼
            ┌───────────────────────┐
            │ 显示插件选择界面        │
            └───────────┬───────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │ 用户选择插件类型        │
            └───────────┬───────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │ 生成workflow文件        │
            └───────────┬───────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │ 初始化SQLite数据库      │
            └───────────┬───────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │ 生成配置文件            │
            └───────────┬───────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │ 生成说明文档            │
            └───────────┬───────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │ 显示成功信息            │
            └───────────────────────┘
```

### 交互界面

```
$ aitdd init

? Select your AI coding plugin:
  ▸ KiloCode
    OpenCode
    ClaudeCode
    Other (manual setup)

✔ Selected: KiloCode

✅ Creating .aitdd directory...
✅ Generating workflow files for KiloCode...
✅ Initializing database...
✅ Creating configuration file...
✅ Generating AITDD_GUIDE.md...

🎉 AITDD initialized successfully!

Next steps:
  1. Run 'aitdd serve' to start the service
  2. Open your AI coding tool and use /aitdd.start

Documentation: https://aitdd.dev/docs
```

### 生成的目录结构

```
项目根目录/
├── .aitdd/
│   ├── aitdd.db              # SQLite数据库
│   ├── config.json           # 配置文件
│   └── logs/                 # 日志目录
├── .kilocode/
│   └── workflows/
│       ├── aitdd.start.md
│       ├── aitdd.constitution.md
│       ├── aitdd.specify.md
│       ├── aitdd.plan.md
│       ├── aitdd.tasks.md
│       ├── aitdd.implement.md
│       └── aitdd.debug.md
└── AITDD_GUIDE.md            # 使用指南
```

### 配置文件示例

```json
// .aitdd/config.json
{
  "version": "1.0.0",
  "pluginType": "KiloCode",
  "port": 34567,
  "projectId": "proj-a1b2c3d4",
  "lockTimeout": 3600000,
  "remoteSync": null,
  "createdAt": 1645564800000
}
```

---

## aitdd serve（启动服务）

### 执行流程

```
用户执行 aitdd serve
        │
        ▼
┌───────────────────────┐
│ 读取配置文件            │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ 检查端口是否可用        │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ 启动HTTP服务器          │
│ - API路由              │
│ - 静态文件服务          │
│ - WebSocket            │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ 打开浏览器              │
│ http://localhost:34567 │
└───────────┬───────────┘
            │
            ▼
┌───────────────────────┐
│ 等待请求/信号           │
│ Ctrl+C 优雅关闭        │
└───────────────────────┘
```

### 命令输出

```
$ aitdd serve

Starting AITDD server...

✅ Configuration loaded from .aitdd/config.json
✅ Database connected: .aitdd/aitdd.db
✅ HTTP server listening on port 34567
✅ WebSocket server started

🌐 Dashboard: http://localhost:34567
📚 API Docs: http://localhost:34567/api-docs

Press Ctrl+C to stop the server
```

### 命令选项

```bash
# 指定端口
aitdd serve --port 8080

# 后台运行（仅Linux/macOS）
aitdd serve --daemon

# 指定配置文件
aitdd serve --config /path/to/config.json

# 详细日志
aitdd serve --verbose
```

---

## Workflow文件格式

### KiloCode格式

```markdown
# .kilocode/workflows/aitdd.start.md

---
name: AITDD Start
description: Initialize AITDD system and show guidelines
trigger: /aitdd.start
---

## Instructions

When the user triggers this command:

1. Call the AITDD API to get project information:
   ```
   GET http://localhost:34567/api/v1/project
   ```

2. Display the project constitution to the user

3. Confirm that MCP is connected and ready

## Example Response

```
🎯 AITDD System Ready

Project: {project_name}

Constitution:
{constitution}

You can now use the following commands:
- /aitdd.constitution - Update project principles
- /aitdd.specify - Define requirements
- /aitdd.plan - Create implementation plan
- /aitdd.tasks - Manage task list
- /aitdd.implement - Execute tasks
- /aitdd.debug - Debug and analyze
```
```

### aitdd.constitution.md

```markdown
# .kilocode/workflows/aitdd.constitution.md

---
name: AITDD Constitution
description: Create or update project constitution
trigger: /aitdd.constitution
---

## Instructions

When the user triggers this command with content:

1. Parse the user's input as constitution content

2. Call the AITDD API to update:
   ```
   PUT http://localhost:34567/api/v1/project/constitution
   {
     "constitution": "<user_input>",
     "version": <current_version>
   }
   ```

3. Confirm the update to the user

## User Input Format

The user will provide constitution content after the command:
```
/aitdd.constitution
# Project Principles

1. All code must have tests
2. Use TypeScript for all new files
...
```
```

### aitdd.specify.md

```markdown
# .kilocode/workflows/aitdd.specify.md

---
name: AITDD Specify
description: Define requirements and user stories
trigger: /aitdd.specify
---

## Instructions

When the user triggers this command:

1. Parse the user's feature description

2. Create modules and tasks based on the description:
   ```
   POST http://localhost:34567/api/v1/modules
   POST http://localhost:34567/api/v1/tasks
   ```

3. Define dependencies between tasks:
   ```
   POST http://localhost:34567/api/v1/dependencies
   ```

4. Display the created structure to the user

## Example

User input:
```
/aitdd.specify
I want to add user authentication with OAuth2 support
```

AI should:
1. Create "Authentication" module
2. Create tasks: "OAuth2 integration", "Login UI", "Session management"
3. Define dependencies
4. Show the task graph
```

### aitdd.plan.md

```markdown
# .kilocode/workflows/aitdd.plan.md

---
name: AITDD Plan
description: Create technical implementation plan
trigger: /aitdd.plan
---

## Instructions

When the user triggers this command:

1. Get all modules and tasks:
   ```
   GET http://localhost:34567/api/v1/modules
   GET http://localhost:34567/api/v1/tasks
   ```

2. Analyze dependencies and create execution order

3. For each task, enrich with:
   - Technical approach
   - File paths
   - Test requirements

4. Update tasks with the plan:
   ```
   PUT http://localhost:34567/api/v1/tasks/:id
   ```

5. Display the execution plan
```

### aitdd.tasks.md

```markdown
# .kilocode/workflows/aitdd.tasks.md

---
name: AITDD Tasks
description: List or modify tasks
trigger: /aitdd.tasks
---

## Instructions

When the user triggers this command:

1. Get the current task list:
   ```
   GET http://localhost:34567/api/v1/tasks
   ```

2. Display tasks in a formatted table

3. Wait for user instructions to:
   - Create new tasks
   - Update task status
   - Delete tasks
   - View task details

## Display Format

```
📋 Task List

| ID | Name | Status | Module |
|----|------|--------|--------|
| task-1 | Login UI | in_progress | Frontend |
| task-2 | OAuth2 | ready | Backend |

What would you like to do? (list/create/update/delete/view)
```
```

### aitdd.implement.md

```markdown
# .kilocode/workflows/aitdd.implement.md

---
name: AITDD Implement
description: Execute tasks according to plan
trigger: /aitdd.implement
---

## Instructions

When the user triggers this command:

1. Get all ready tasks:
   ```
   GET http://localhost:34567/api/v1/tasks?status=ready
   ```

2. For each task:
   a. Lock the task:
      ```
      POST http://localhost:34567/api/v1/lock
      ```
   b. Read task details and contracts
   c. Implement the code
   d. Run tests
   e. Update task status:
      ```
      PUT http://localhost:34567/api/v1/tasks/:id
      ```
   f. Unlock the task
   g. Notify dependent tasks

3. Report progress to the user
```

### aitdd.debug.md

```markdown
# .kilocode/workflows/aitdd.debug.md

---
name: AITDD Debug
description: Debug and analyze task chain
trigger: /aitdd.debug
---

## Instructions

When the user triggers this command:

1. Get task and dependency information:
   ```
   GET http://localhost:34567/api/v1/tasks
   GET http://localhost:34567/api/v1/dependencies
   ```

2. Analyze:
   - Tasks with errors (status=failed)
   - Broken dependencies
   - Blocked tasks
   - Contract mismatches

3. Display analysis report

4. Optionally fix issues by updating tasks

## Analysis Report Format

```
🔍 Debug Analysis

❌ Failed Tasks:
  - task-5: Database connection error

⚠️ Blocked Tasks:
  - task-8: Waiting for task-5

🔗 Broken Dependencies:
  - task-3 → task-7: Contract version mismatch

Recommendations:
1. Fix database connection in task-5
2. Update contract in task-3
```
```

---

## AITDD_GUIDE.md模板

```markdown
# AITDD 使用指南

本项目已集成 AITDD 可视化任务治理系统。

## 快速开始

### 1. 启动服务

```bash
aitdd serve
```

服务将在 http://localhost:34567 启动。

### 2. 在AI编程工具中使用

在 KiloCode 对话框中输入：

```
/aitdd.start
```

### 3. 可用命令

| 命令 | 说明 |
|------|------|
| /aitdd.start | 初始化并显示项目信息 |
| /aitdd.constitution | 制定/更新项目原则 |
| /aitdd.specify | 定义需求和用户故事 |
| /aitdd.plan | 创建技术实施计划 |
| /aitdd.tasks | 管理任务列表 |
| /aitdd.implement | 执行任务 |
| /aitdd.debug | 调试和分析 |

## 目录结构

```
.aitdd/
├── aitdd.db      # 数据库
├── config.json   # 配置
└── logs/         # 日志
```

## 更多信息

- 官方文档: https://aitdd.dev
- GitHub: https://github.com/aitdd/aitdd
```

---

## 多人协作配置

### 配置远程同步

```json
// .aitdd/config.json
{
  "remoteSync": {
    "enabled": true,
    "type": "postgresql",
    "connectionString": "postgresql://user:pass@host:5432/aitdd",
    "syncInterval": 30000,
    "conflictResolution": "manual"
  }
}
```

### 同步命令

```bash
# 手动同步
aitdd sync

# 查看同步状态
aitdd sync --status

# 解决冲突
aitdd sync --resolve
```
