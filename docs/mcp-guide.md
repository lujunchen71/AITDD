# AITDD MCP 使用方法文档

## 目录

1. [概述](#1-概述)
2. [安装配置](#2-安装配置)
3. [在不同IDE中集成](#3-在不同ide中集成)
4. [工具列表和使用说明](#4-工具列表和使用说明)
5. [使用示例](#5-使用示例)
6. [检查规则说明](#6-检查规则说明)
7. [故障排除](#7-故障排除)

---

## 1. 概述

### 1.1 AITDD MCP 简介

AITDD MCP（Model Context Protocol）是一个为 AI 编码代理（如 Claude、Cursor 等）提供的工具服务，用于与 AITDD 项目管理系统进行交互。通过 MCP 协议，AI 代理可以：

- 获取项目和模块信息
- 管理任务和模块
- 检查代码质量和规则合规性
- 追踪错误和问题

### 1.2 主要特性

- **配置管理**：初始化项目配置，管理项目信息
- **数据获取**：获取项目、模块、任务的详细信息
- **数据修改**：创建、更新、删除模块和任务
- **状态追踪**：获取任务状态和进度信息
- **错误处理**：获取项目和模块级别的错误列表
- **规则检查**：根据预定义规则检查模块质量

### 1.3 架构图

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   AI Agent      │────▶│   MCP Server    │────▶│   Backend API   │
│ (Claude/Cursor) │     │   (aitdd-mcp)   │     │   (:34567)      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │                        │
                               ▼                        ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │ .aitdd/config   │     │    Database     │
                        │ .aitdd/rule     │     │    (SQLite)     │
                        └─────────────────┘     └─────────────────┘
```

---

## 2. 安装配置

### 2.1 编译 MCP 服务器

AITDD MCP 服务器使用 Go 语言编写，需要先编译后使用。

```bash
# 进入后端目录
cd backend

# 编译 MCP 服务器
go build -o aitdd-mcp ./cmd/mcp

# 将可执行文件移动到 PATH 中或项目根目录
mv aitdd-mcp ../
```

或者使用提供的构建脚本：

```bash
# 使用构建脚本
bash scripts/build.sh
```

### 2.2 配置文件说明（`.aitdd/config.json`）

配置文件位于项目根目录的 `.aitdd/config.json`，包含以下字段：

```json
{
  "projectId": "proj-uuid-xxxx",
  "projectName": "项目名称",
  "apiBaseUrl": "http://localhost:34567/api/v1",
  "mcp": {
    "serverName": "aitdd-mcp",
    "version": "1.0.0",
    "description": "AITDD数据库增删改查MCP服务"
  },
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**字段说明：**

| 字段路径 | 类型 | 必填 | 说明 |
|---------|------|------|------|
| `projectId` | string | 是 | 项目唯一标识符 |
| `projectName` | string | 是 | 项目名称 |
| `apiBaseUrl` | string | 是 | 后端 API 基础地址 |
| `mcp.serverName` | string | 是 | MCP 服务器名称 |
| `mcp.version` | string | 是 | MCP 服务器版本 |
| `mcp.description` | string | 否 | MCP 服务描述 |
| `createdAt` | string | 否 | 创建时间（RFC3339格式） |
| `updatedAt` | string | 否 | 更新时间（RFC3339格式） |

### 2.3 规则文件说明（`.aitdd/rule.json`）

规则文件定义了模块检查时使用的规则，包含静态检查和动态检查两类。

```json
{
  "version": "1.0.0",
  "description": "提示词编译检查规则配置文件",
  "rules": {
    "static": {
      "description": "静态检查规则",
      "rules": [...]
    },
    "task": {
      "description": "任务检查规则",
      "rules": [...]
    },
    "dependency": {
      "description": "依赖关系检查规则",
      "rules": [...]
    },
    "dynamic": {
      "description": "动态检查规则",
      "rules": [...]
    }
  }
}
```

**规则结构：**

```json
{
  "id": "S-01",
  "name": "规则名称",
  "description": "规则描述",
  "severity": "error",
  "enabled": true
}
```

- `id`: 规则唯一标识符
- `name`: 规则名称
- `description`: 规则描述
- `severity`: 严重级别（`error`/`warning`/`info`）
- `enabled`: 是否启用

---

## 3. 在不同IDE中集成

### 3.1 VS Code / Cursor 配置

在 VS Code 或 Cursor 中，需要编辑 `settings.json` 文件来配置 MCP 服务器。

**方式一：使用本地可执行文件**

打开 VS Code 设置（`Ctrl + Shift + P` -> "Open User Settings (JSON)"），添加以下配置：

```json
{
  "mcp.servers": {
    "aitdd": {
      "command": "/path/to/aitdd-mcp",
      "args": [],
      "cwd": "${workspaceFolder}"
    }
  }
}
```

**方式二：使用 npx 运行**

```json
{
  "mcp.servers": {
    "aitdd": {
      "command": "npx",
      "args": ["-y", "aitdd-mcp"],
      "cwd": "${workspaceFolder}"
    }
  }
}
```

**完整配置示例：**

```json
{
  "mcp.servers": {
    "aitdd": {
      "command": "X:/AITDD/aitdd-mcp.exe",
      "args": [],
      "cwd": "X:/AITDD",
      "env": {
        "AITDD_API_URL": "http://localhost:34567/api/v1"
      }
    }
  }
}
```

### 3.2 Claude Desktop 配置

在 Claude Desktop 中配置 MCP 服务器，需要编辑 `claude_desktop_config.json` 文件。

**Windows 路径：**
```
%APPDATA%\Claude\claude_desktop_config.json
```

**macOS 路径：**
```
~/Library/Application Support/Claude/claude_desktop_config.json
```

**配置示例：**

```json
{
  "mcpServers": {
    "aitdd": {
      "command": "X:/AITDD/aitdd-mcp.exe",
      "args": [],
      "cwd": "X:/AITDD"
    }
  }
}
```

**使用 npx 运行：**

```json
{
  "mcpServers": {
    "aitdd": {
      "command": "npx",
      "args": ["-y", "aitdd-mcp"],
      "cwd": "X:/AITDD"
    }
  }
}
```

### 3.3 其他支持MCP的IDE

对于其他支持 MCP 协议的 IDE 或工具，配置方式类似：

1. **找到 MCP 服务器配置文件位置**
2. **添加 AITDD MCP 服务器配置**
3. **指定可执行文件路径和工作目录**

**通用配置格式：**

```json
{
  "mcpServers": {
    "aitdd": {
      "command": "/path/to/aitdd-mcp",
      "args": [],
      "cwd": "/path/to/project"
    }
  }
}
```

---

## 4. 工具列表和使用说明

### 4.1 配置管理工具

#### `init_project` - 初始化项目配置

**描述：** 初始化项目配置，从后端获取项目列表供用户选择，保存到 `.aitdd/config.json`

**参数：** 无

**返回示例：**
```json
{
  "projects": [
    {"id": "proj-001", "name": "项目A", "description": "项目A描述"},
    {"id": "proj-002", "name": "项目B", "description": "项目B描述"}
  ],
  "total": 2
}
```

---

#### `get_config` - 获取当前配置

**描述：** 获取当前 `.aitdd/config.json` 配置文件的内容

**参数：** 无

**返回示例：**
```json
{
  "projectId": "proj-001",
  "projectName": "项目A",
  "apiBaseUrl": "http://localhost:34567/api/v1",
  "mcp": {
    "serverName": "aitdd-mcp",
    "version": "1.0.0"
  }
}
```

---

#### `set_project` - 设置当前项目

**描述：** 设置当前项目信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| projectId | string | 是 | 项目 ID |
| projectName | string | 是 | 项目名称 |

**返回示例：**
```json
{
  "success": true,
  "message": "项目设置成功"
}
```

---

### 4.2 获取类工具

#### `get_project_info` - 获取项目简介

**描述：** 获取项目简介、架构信息、编码规范

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| projectId | string | 否 | 项目ID，不传则使用配置文件中的项目ID |

**返回示例：**
```json
{
  "id": "proj-001",
  "name": "项目名称",
  "description": "项目描述",
  "constitution": "项目公约内容",
  "statistics": {
    "totalModules": 10,
    "totalTasks": 50,
    "completedTasks": 20,
    "inProgressTasks": 15,
    "pendingTasks": 15
  }
}
```

---

#### `get_all_task_code_paths` - 获取所有任务代码路径

**描述：** 获取项目中所有任务的代码结构路径

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| projectId | string | 否 | 项目ID，不传则使用配置文件中的项目ID |

**返回示例：**
```json
{
  "paths": [
    {
      "taskId": "task-001",
      "taskName": "任务A",
      "codePaths": ["src/moduleA/file1.ts", "src/moduleA/file2.ts"]
    }
  ]
}
```

---

#### `get_all_modules` - 获取所有模块概要

**描述：** 获取所有模块的 id、名称、介绍

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| projectId | string | 否 | 项目ID，不传则使用配置文件中的项目ID |
| includeStats | boolean | 否 | 是否包含统计信息 |

**返回示例：**
```json
{
  "modules": [
    {
      "id": "mod-001",
      "name": "用户模块",
      "description": "用户管理相关功能",
      "status": "developing",
      "taskCount": 5,
      "completedTasks": 2
    }
  ]
}
```

---

#### `get_module_tasks` - 获取模块任务列表

**描述：** 获取某模块的所有任务列表概要

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| includeContracts | boolean | 否 | 是否包含契约信息 |

**返回示例：**
```json
{
  "moduleId": "mod-001",
  "tasks": [
    {
      "id": "task-001",
      "name": "用户登录",
      "status": "completed",
      "description": "实现用户登录功能"
    }
  ]
}
```

---

#### `get_task_detail` - 获取任务详情

**描述：** 获取某任务的详细信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| taskId | string | 是 | 任务 ID |

**返回示例：**
```json
{
  "id": "task-001",
  "name": "用户登录",
  "description": "实现用户登录功能",
  "status": "completed",
  "prompt": "实现用户登录功能...",
  "upstreamContractDetail": {...},
  "downstreamContractDetail": {...},
  "tests": [...],
  "codePaths": [...],
  "testResult": "passed",
  "issueDetails": "",
  "bugLog": ""
}
```

---

#### `get_task_contracts` - 获取任务上下游契约

**描述：** 获取某任务的上下游契约接口信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| taskId | string | 是 | 任务 ID |
| direction | string | 否 | 方向：upstream/downstream/both，默认 both |

**返回示例：**
```json
{
  "taskId": "task-001",
  "upstream": [
    {
      "name": "用户数据",
      "type": "object",
      "description": "用户信息对象"
    }
  ],
  "downstream": [
    {
      "name": "登录结果",
      "type": "boolean",
      "description": "登录是否成功"
    }
  ]
}
```

---

### 4.3 修改类工具

#### `delete_module` - 删除模块

**描述：** 删除指定模块及其所有子任务

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| force | boolean | 否 | 是否强制删除（即使有依赖） |

**返回示例：**
```json
{
  "success": true,
  "message": "模块已删除"
}
```

---

#### `create_module` - 创建模块

**描述：** 添加新模块

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| name | string | 是 | 模块名称 |
| description | string | 否 | 模块描述 |
| prompt | string | 否 | 模块提示词 |
| parentId | string | 否 | 父模块ID |
| projectId | string | 否 | 项目ID，不传则使用配置文件中的项目ID |

**返回示例：**
```json
{
  "id": "mod-new",
  "name": "新模块",
  "description": "模块描述",
  "status": "designing"
}
```

---

#### `update_module` - 更新模块

**描述：** 修改模块信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| name | string | 否 | 模块名称 |
| description | string | 否 | 模块描述 |
| prompt | string | 否 | 模块提示词 |
| status | string | 否 | 模块状态 |
| version | number | 是 | 当前版本号（用于乐观锁） |

**返回示例：**
```json
{
  "success": true,
  "version": 2,
  "message": "模块已更新"
}
```

---

#### `delete_module_tasks` - 删除模块所有任务

**描述：** 删除模块的所有任务

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| force | boolean | 否 | 是否强制删除（即使有依赖） |

**返回示例：**
```json
{
  "success": true,
  "deletedCount": 5,
  "message": "已删除5个任务"
}
```

---

#### `create_task` - 创建任务

**描述：** 在模块中添加新任务

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| name | string | 是 | 任务名称 |
| description | string | 否 | 任务描述 |
| prompt | string | 否 | 任务提示词 |
| upstreamContractDetail | string | 否 | 上游契约详情JSON |
| downstreamContractDetail | string | 否 | 下游契约详情JSON |
| tests | string | 否 | 测试用例JSON数组 |
| codePaths | string | 否 | 代码路径JSON数组 |

**返回示例：**
```json
{
  "id": "task-new",
  "name": "新任务",
  "status": "pending"
}
```

---

#### `update_module_full` - 完整更新模块

**描述：** 重新定义模块的所有信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| moduleJson | string | 是 | 完整的模块JSON数据 |

**返回示例：**
```json
{
  "success": true,
  "message": "模块已完整更新"
}
```

---

#### `update_task_full` - 完整更新任务

**描述：** 重新定义任务的完整信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| taskId | string | 是 | 任务 ID |
| taskJson | string | 是 | 完整的任务JSON数据 |

**返回示例：**
```json
{
  "success": true,
  "message": "任务已完整更新"
}
```

---

### 4.4 状态类工具

#### `get_all_task_status` - 获取所有任务状态

**描述：** 获取项目中所有任务的状态信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| projectId | string | 否 | 项目ID，不传则使用配置文件中的项目ID |
| status | string | 否 | 状态过滤 |
| includeLockInfo | boolean | 否 | 是否包含锁定信息 |

**返回示例：**
```json
{
  "tasks": [
    {
      "id": "task-001",
      "name": "用户登录",
      "status": "completed",
      "isLocked": false
    }
  ],
  "summary": {
    "total": 50,
    "completed": 20,
    "inProgress": 15,
    "pending": 15
  }
}
```

---

#### `get_module_task_status` - 获取模块任务状态

**描述：** 获取当前模块的所有任务状态信息

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| includeLockInfo | boolean | 否 | 是否包含锁定信息 |

**返回示例：**
```json
{
  "moduleId": "mod-001",
  "tasks": [
    {
      "id": "task-001",
      "name": "用户登录",
      "status": "completed",
      "isLocked": false
    }
  ],
  "summary": {
    "total": 5,
    "completed": 2,
    "inProgress": 2,
    "pending": 1
  }
}
```

---

### 4.5 错误处理类工具

#### `get_project_errors` - 获取项目错误列表

**描述：** 返回整个项目的错误列表，遍历所有任务检查 issue_details 和 bug_log

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| projectId | string | 否 | 项目ID，不传则使用配置文件中的项目ID |
| severity | string | 否 | 严重级别过滤：error/warning/info |
| includeDetails | boolean | 否 | 是否包含详细信息 |

**返回示例：**
```json
{
  "errors": [
    {
      "taskId": "task-001",
      "taskName": "用户登录",
      "type": "bug",
      "severity": "error",
      "message": "登录验证失败",
      "details": "..."
    }
  ],
  "summary": {
    "total": 5,
    "errors": 2,
    "warnings": 3
  }
}
```

---

#### `get_module_errors` - 获取模块错误列表

**描述：** 返回某模块的所有子任务错误列表

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| severity | string | 否 | 严重级别过滤 |
| includeDetails | boolean | 否 | 是否包含详细信息 |

**返回示例：**
```json
{
  "moduleId": "mod-001",
  "errors": [
    {
      "taskId": "task-001",
      "taskName": "用户登录",
      "type": "issue",
      "severity": "warning",
      "message": "需要优化"
    }
  ]
}
```

---

### 4.6 检查类工具

#### `check_module` - 检查模块并生成报告

**描述：** 根据 `.aitdd/rule.json` 中的规则检查模块并生成报告

**参数：**

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| moduleId | string | 是 | 模块 ID |
| rules | array | 否 | 指定检查规则ID列表，不传则检查所有 |
| includeDynamic | boolean | 否 | 是否包含动态检查，默认 true |
| format | string | 否 | 输出格式：json/markdown，默认 json |

**返回示例：**
```json
{
  "moduleId": "mod-001",
  "moduleName": "用户模块",
  "moduleStatus": "developing",
  "checkTime": "2024-01-01T12:00:00Z",
  "summary": {
    "total": 15,
    "errorCount": 2,
    "warningCount": 5,
    "suggestionCount": 8
  },
  "errors": [...],
  "warnings": [...],
  "suggestions": [...]
}
```

---

## 5. 使用示例

### 5.1 初始化项目示例

```
用户: 请帮我初始化 AITDD 项目配置

AI: 我来为您初始化项目配置。
[调用 init_project 工具]

已获取到以下项目列表：
1. 项目A (proj-001) - 项目A描述
2. 项目B (proj-002) - 项目B描述

请选择要使用的项目？
```

### 5.2 查询模块信息示例

```
用户: 请列出当前项目的所有模块

AI: 我来获取项目的模块列表。
[调用 get_all_modules 工具]

当前项目有以下模块：

| ID | 名称 | 状态 | 任务数 |
|-----|------|------|--------|
| mod-001 | 用户模块 | developing | 5 |
| mod-002 | 订单模块 | designing | 3 |
| mod-003 | 支付模块 | completed | 8 |
```

### 5.3 创建任务示例

```
用户: 请在用户模块中创建一个"用户注册"任务

AI: 我来为您创建任务。
[调用 create_task 工具]
参数：
- moduleId: mod-001
- name: 用户注册
- description: 实现用户注册功能

任务已创建成功：
- 任务ID: task-new
- 任务名称: 用户注册
- 状态: pending
```

### 5.4 检查模块示例

```
用户: 请检查用户模块的质量

AI: 我来检查用户模块。
[调用 check_module 工具]
参数：
- moduleId: mod-001
- format: markdown

## 模块检查报告

**模块**: 用户模块
**状态**: developing
**检查时间**: 2024-01-01 12:00:00

### 摘要
- 总检查项: 15
- 错误: 2
- 警告: 5
- 建议: 8

### 错误 (2)
1. [S-01] 提示词完整性检查 - 任务 task-002 的提示词过短
2. [DYN-03] 契约一致性检查 - 上下游契约不匹配

### 警告 (5)
1. [S-02] 上游契约完整性 - 建议补充上游契约
...
```

---

## 6. 检查规则说明

### 6.1 静态检查规则

静态检查规则用于检查模块和任务的数据完整性和格式正确性。

| 规则ID | 名称 | 严重级别 | 描述 |
|--------|------|----------|------|
| S-01 | 根对象结构检查 | error | 根对象必须包含 project、modules、tasks、taskDependencies 四个顶级字段 |
| S-02 | 项目对象检查 | error | project 对象必须包含 name 和 description 字段 |
| S-03 | 模块数组检查 | error | modules 数组至少包含一个模块 |
| S-04 | 模块对象字段检查 | error | 每个模块对象必须包含 id、name、description、status、file_path |
| S-05 | 模块ID格式检查 | error | 模块 id 必须符合格式 mod-{模块名缩写} |
| S-06 | 模块状态检查 | error | 模块 status 必须是 designing/developing/completed/deprecated 之一 |
| S-07 | 模块路径检查 | error | 模块 file_path 不能为空字符串 |

### 6.2 任务检查规则

| 规则ID | 名称 | 严重级别 | 描述 |
|--------|------|----------|------|
| T-01 | 任务对象字段检查 | error | 每个任务对象必须包含 id、title、description、status、moduleId |
| T-02 | 任务ID格式检查 | error | 任务 id 必须符合格式 task-{模块名缩写}-{序号} |
| T-03 | 任务状态检查 | error | 任务 status 必须是 pending/in-progress/completed/blocked/cancelled 之一 |
| T-04 | 任务模块关联检查 | error | 任务 moduleId 必须引用存在的模块 id |
| T-05 | 任务标题检查 | error | 任务 title 不能为空字符串 |
| T-06 | 任务描述检查 | warning | 任务 description 不能为空字符串 |
| T-13 | 任务契约字段检查 | error | 契约项必须包含 name、type、description 字段 |
| T-14 | 任务契约类型检查 | error | 契约项 type 必须是有效类型 |
| T-17 | 任务测试结果检查 | warning | 任务 testResult 必须是有效值 |

### 6.3 依赖关系检查规则

| 规则ID | 名称 | 严重级别 | 描述 |
|--------|------|----------|------|
| D-01 | 依赖对象字段检查 | error | 每个依赖对象必须包含 id、sourceTaskId、targetTaskId、type |
| D-06 | 自依赖检查 | error | 依赖 sourceTaskId 和 targetTaskId 不能相同 |
| D-07 | 重复依赖检查 | error | 不能存在相同的依赖关系 |
| D-08 | 循环依赖检查 | error | 不能存在循环依赖 |

### 6.4 动态检查规则

动态检查规则用于运行时状态和一致性检查。

| 规则ID | 名称 | 严重级别 | 描述 |
|--------|------|----------|------|
| DYN-01 | 契约一致性检查 | warning | 下游任务的输入契约应与上游任务的输出契约匹配 |
| DYN-02 | 模块设计合理性检查 | warning | 模块的任务数量应合理（建议1-10个） |
| DYN-04 | 任务完成度检查 | info | 检查模块内任务的完成进度 |
| DYN-05 | 阻塞任务检查 | warning | 检查是否有任务长期处于阻塞状态 |
| DYN-06 | 依赖链完整性检查 | error | 检查依赖链是否完整，无断裂 |
| DYN-07 | 关键路径分析 | info | 分析项目的关键路径 |
| DYN-13 | 问题跟踪检查 | warning | 检查任务的问题状态 |
| DYN-14 | 文档完整性检查 | info | 检查模块和任务的文档完整性 |

### 6.5 检查报告格式

检查报告支持 JSON 和 Markdown 两种格式。

**JSON 格式：**
```json
{
  "moduleId": "mod-001",
  "moduleName": "用户模块",
  "moduleStatus": "developing",
  "checkTime": "2024-01-01T12:00:00Z",
  "summary": {
    "total": 15,
    "errorCount": 2,
    "warningCount": 5,
    "suggestionCount": 8
  },
  "errors": [
    {
      "ruleId": "S-01",
      "ruleName": "提示词完整性检查",
      "severity": "error",
      "resourceType": "task",
      "resourceId": "task-001",
      "resourceName": "用户登录",
      "message": "任务提示词过短",
      "suggestion": "建议补充详细的任务描述"
    }
  ],
  "warnings": [...],
  "suggestions": [...]
}
```

**Markdown 格式：**
```markdown
## 模块检查报告

**模块**: 用户模块
**状态**: developing
**检查时间**: 2024-01-01 12:00:00

### 摘要
- 总检查项: 15
- 错误: 2
- 警告: 5
- 建议: 8

### 错误 (2)
1. **[S-01] 提示词完整性检查**
   - 资源: task-001 (用户登录)
   - 问题: 任务提示词过短
   - 建议: 建议补充详细的任务描述

### 警告 (5)
...

### 建议 (8)
...
```

---

## 7. 故障排除

### 7.1 常见问题

#### 问题1: MCP 服务器无法启动

**症状：** 配置完成后，MCP 服务器无法启动或连接失败

**解决方案：**
1. 检查可执行文件路径是否正确
2. 确保工作目录（cwd）设置正确
3. 检查后端 API 服务是否正在运行
4. 查看终端错误日志

```bash
# 手动测试 MCP 服务器
cd /path/to/project
./aitdd-mcp
```

#### 问题2: 配置文件未找到

**症状：** 提示 `.aitdd/config.json` 文件不存在

**解决方案：**
1. 确保在项目根目录下存在 `.aitdd` 文件夹
2. 运行 `init_project` 工具创建配置文件
3. 手动创建配置文件：

```bash
mkdir -p .aitdd
echo '{
  "projectId": "",
  "projectName": "",
  "apiBaseUrl": "http://localhost:34567/api/v1",
  "mcp": {
    "serverName": "aitdd-mcp",
    "version": "1.0.0"
  }
}' > .aitdd/config.json
```

#### 问题3: API 连接失败

**症状：** 工具调用返回连接错误

**解决方案：**
1. 检查后端服务是否运行：
   ```bash
   curl http://localhost:34567/api/v1/projects
   ```
2. 检查 `apiBaseUrl` 配置是否正确
3. 检查防火墙设置
4. 检查端口是否被占用

#### 问题4: 项目ID未配置

**症状：** 工具调用提示"项目未配置"

**解决方案：**
1. 运行 `get_config` 查看当前配置
2. 运行 `init_project` 选择项目
3. 或使用 `set_project` 设置项目：

```json
{
  "projectId": "proj-001",
  "projectName": "项目名称"
}
```

#### 问题5: 版本冲突

**症状：** 更新操作返回版本冲突错误

**解决方案：**
1. 先获取当前资源的版本号
2. 使用正确的版本号进行更新
3. 如果使用 `update_module_full` 或 `update_task_full`，版本号会自动处理

### 7.2 日志调试

启用详细日志可以帮助诊断问题：

```bash
# 设置环境变量启用调试日志
export AITDD_DEBUG=true
./aitdd-mcp
```

### 7.3 重置配置

如果配置出现问题，可以重置配置文件：

```bash
# 备份现有配置
cp .aitdd/config.json .aitdd/config.json.bak

# 删除配置文件
rm .aitdd/config.json

# 重新运行 init_project
```

### 7.4 联系支持

如果以上方法无法解决问题，请：

1. 收集错误日志
2. 记录复现步骤
3. 提交 Issue 到项目仓库

---

## 附录

### A. 工具快速参考

| 工具名称 | 类别 | 描述 |
|---------|------|------|
| `init_project` | 配置 | 初始化项目配置 |
| `get_config` | 配置 | 获取当前配置 |
| `set_project` | 配置 | 设置当前项目 |
| `get_project_info` | 获取 | 获取项目简介 |
| `get_all_task_code_paths` | 获取 | 获取所有任务代码路径 |
| `get_all_modules` | 获取 | 获取所有模块概要 |
| `get_module_tasks` | 获取 | 获取模块任务列表 |
| `get_task_detail` | 获取 | 获取任务详情 |
| `get_task_contracts` | 获取 | 获取任务上下游契约 |
| `delete_module` | 修改 | 删除模块 |
| `create_module` | 修改 | 创建模块 |
| `update_module` | 修改 | 更新模块 |
| `delete_module_tasks` | 修改 | 删除模块所有任务 |
| `create_task` | 修改 | 创建任务 |
| `update_module_full` | 修改 | 完整更新模块 |
| `update_task_full` | 修改 | 完整更新任务 |
| `get_all_task_status` | 状态 | 获取所有任务状态 |
| `get_module_task_status` | 状态 | 获取模块任务状态 |
| `get_project_errors` | 错误 | 获取项目错误列表 |
| `get_module_errors` | 错误 | 获取模块错误列表 |
| `check_module` | 检查 | 检查模块并生成报告 |

### B. 状态值参考

**模块状态：**
- `designing` - 设计中
- `developing` - 开发中
- `completed` - 已完成
- `deprecated` - 已废弃

**任务状态：**
- `pending` - 待处理
- `in-progress` - 进行中
- `completed` - 已完成
- `blocked` - 已阻塞
- `cancelled` - 已取消

**测试结果：**
- `pending` - 待测试
- `passed` - 通过
- `failed` - 失败
- `skipped` - 跳过

**严重级别：**
- `error` - 错误（必须修复）
- `warning` - 警告（建议修复）
- `info` - 信息（仅供参考）
