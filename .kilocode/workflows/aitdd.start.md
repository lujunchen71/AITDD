# KiloCode Workflow 文件

## aitdd.start - 系统指南和 MCP 注册

```markdown
---
name: aitdd.start
description: AITDD 系统指南和 MCP 注册
---

# AITDD 系统启动指南

## 系统概述

AITDD 是一个提示词依赖管理系统，通过 MCP 协议供 AI 编程软件调用，管理大型软件项目的模块、任务、依赖关系和提示词。

## MCP 工具列表

### 项目相关
- `get_project_summary` - 获取项目简述
- `get_constitution` - 获取项目公约

### 模块相关
- `get_all_modules` - 获取所有模块和子模块
- `get_module_overview` - 获取模块概览（提示词、状态、测试覆盖率、契约概览）
- `get_module_task_ids` - 获取模块中所有任务 ID 列表
- `delete_module` - 删除模块
- `create_module_dependency` - 创建模块依赖
- `get_module_dependencies` - 获取模块依赖

### 任务相关
- `get_task_details` - 获取任务详细信息（状态、契约、提示词、测试、日志、代码路径、人类协助）
- `update_task` - 修改任务详细信息
- `create_task` - 在模块中创建任务
- `delete_task` - 删除任务
- `create_task_dependency` - 创建任务依赖
- `get_task_dependencies` - 获取任务依赖

### 工具相关
- `open_frontend` - 打开前端网页 (http://localhost:5173)

### 通知相关
- `send_notification` - 发送通知
- `read_notifications` - 读取未读通知

### 锁相关（预留）
- `lock_resource` - 锁定资源
- `unlock_resource` - 解锁资源
- `get_lock_status` - 查询锁定状态

## 使用示例

### 获取项目信息
```
调用 get_project_summary
```

### 获取模块树
```
调用 get_all_modules, projectId="项目 ID"
```

### 创建任务
```
调用 create_task, moduleId="模块 ID", name="任务名称", description="任务描述", prompt="任务提示词"
```

## API 端点

Base URL: http://localhost:34567/api/v1
```

## aitdd.constitution - 制定或更新项目公约

```markdown
---
name: aitdd.constitution
description: 制定或更新项目管理原则和开发指南
---

# AITDD 项目公约命令

## 功能说明
此命令用于制定或更新项目的管理原则和开发指南。

## MCP 工具
- `get_constitution` - 获取当前项目公约
- 调用 PUT /api/v1/project/constitution - 更新项目公约

## 使用方式

1. 首先获取当前公约：调用 `get_constitution`
2. 根据用户需求更新公约内容
3. 调用 API 更新公约

## 示例

用户：我想添加一条新的开发规范，所有 API 必须遵循 RESTful 风格

AI: 
1. 让我先获取当前的项目公约...
   调用 get_constitution
   
2. 现在我将更新公约，添加新的开发规范...
   调用 PUT /api/v1/project/constitution
```

## aitdd.specify - 明确需求和用户故事

```markdown
---
name: aitdd.specify
description: 明确需求和用户故事
---

# AITDD 需求明确命令

## 功能说明
此命令用于明确项目需求，创建需求模块和用户故事。

## MCP 工具
- `get_all_modules` - 获取现有模块结构
- `create_module` - 创建需求模块
- `create_task` - 创建用户故事任务

## 使用方式

1. 了解现有模块结构
2. 创建需求模块（如"用户需求"、"功能规格"）
3. 为每个需求创建任务

## 示例

用户：我想开发一个用户登录功能

AI:
1. 让我先了解现有的模块结构...
   调用 get_all_modules

2. 我将创建一个"用户认证"模块来管理登录功能...
   调用 create_module

3. 现在创建用户故事任务...
   调用 create_task
```

## aitdd.plan - 创建技术实施计划

```markdown
---
name: aitdd.plan
description: 创建技术实施计划
---

# AITDD 技术计划命令

## 功能说明
此命令用于创建技术实施计划，将需求分解为可执行的技术任务。

## MCP 工具
- `get_all_modules` - 获取现有模块结构
- `create_module` - 创建技术模块
- `create_task` - 创建技术任务
- `create_module_dependency` - 创建模块依赖
- `create_task_dependency` - 创建任务依赖

## 使用方式

1. 分析需求模块
2. 创建技术模块（如"前端"、"后端"、"数据库"）
3. 为每个技术模块创建任务
4. 建立模块和任务的依赖关系

## 示例

用户：请为登录功能制定技术实施计划

AI:
1. 创建技术模块结构...
   - 前端模块
     - 登录页面组件
     - 认证 Hook
   - 后端模块
     - 认证 API
     - JWT 服务
   - 数据库模块
     - 用户表

2. 创建任务并建立依赖关系...
```

## aitdd.tasks - 生成或修改任务清单

```markdown
---
name: aitdd.tasks
description: 生成或修改任务清单
---

# AITDD 任务管理命令

## 功能说明
此命令用于生成、查看和修改任务清单。

## MCP 工具
- `get_all_modules` - 获取模块列表
- `get_module_task_ids` - 获取模块任务 ID
- `get_task_details` - 获取任务详情
- `create_task` - 创建任务
- `update_task` - 更新任务
- `delete_task` - 删除任务

## 使用方式

1. 查看现有任务
2. 创建新任务或更新现有任务
3. 建立任务依赖关系

## 示例

用户：显示所有待办任务

AI:
1. 获取所有模块...
   调用 get_all_modules

2. 获取每个模块的任务...
   调用 get_module_task_ids

3. 获取任务详情...
   调用 get_task_details
```

## aitdd.implement - 执行任务

```markdown
---
name: aitdd.implement
description: 按照计划执行任务
---

# AITDD 任务执行命令

## 功能说明
此命令用于执行任务，更新任务状态和代码。

## MCP 工具
- `get_task_details` - 获取任务详情
- `update_task` - 更新任务状态
- `lock_resource` - 锁定任务
- `unlock_resource` - 解锁任务

## 使用方式

1. 获取任务详情和提示词
2. 锁定任务（防止并发修改）
3. 执行任务（编写代码）
4. 更新任务状态和代码路径
5. 解锁任务

## 示例

用户：开始实现登录功能

AI:
1. 获取任务详情...
   调用 get_task_details

2. 锁定任务...
   调用 lock_resource

3. 执行任务...
   （编写代码）

4. 更新任务状态为 in_progress...
   调用 update_task

5. 完成任务后更新状态为 completed...
   调用 update_task
   调用 unlock_resource
```

## aitdd.debug - 查询和修复 bug

```markdown
---
name: aitdd.debug
description: 查询 bug、修复 bug、分析链路完整性
---

# AITDD 调试命令

## 功能说明
此命令用于查询任务 bug、修复 bug、分析任务链路完整性。

## MCP 工具
- `get_task_details` - 获取任务详情
- `get_task_dependencies` - 获取任务依赖
- `get_module_dependencies` - 获取模块依赖
- `update_task` - 更新任务
- `send_notification` - 发送通知
- `read_notifications` - 读取通知

## 使用方式

1. 分析任务链路完整性
2. 查看哪些任务脱节
3. 同步任务链条
4. 修复 bug

## 示例

用户：登录功能有问题，帮我检查

AI:
1. 获取登录任务详情...
   调用 get_task_details

2. 检查任务依赖...
   调用 get_task_dependencies

3. 检查模块依赖...
   调用 get_module_dependencies

4. 分析链路完整性，找出脱节的地方...

5. 修复问题并更新任务...
   调用 update_task

6. 通知相关任务...
   调用 send_notification
```
