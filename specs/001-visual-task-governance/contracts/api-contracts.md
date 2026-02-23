# API设计文档

## 概述

本地服务提供RESTful API，供第三方AI编程软件通过MCP调用。所有请求和响应均使用JSON格式，路径前缀为 `/api/v1`。

**基础信息**:
- 默认端口: 34567
- 基础URL: `http://localhost:34567/api/v1`
- 认证: 默认无（仅监听localhost），可配置简单令牌

## 通用响应格式

### 成功响应

```json
{
  "success": true,
  "data": { ... },
  "timestamp": 1645564800000
}
```

### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": { ... }
  },
  "timestamp": 1645564800000
}
```

### 错误码

| 错误码 | HTTP状态码 | 说明 |
|--------|-----------|------|
| NOT_FOUND | 404 | 资源不存在 |
| VERSION_CONFLICT | 409 | 版本冲突（乐观锁） |
| LOCKED | 423 | 资源被锁定 |
| VALIDATION_ERROR | 400 | 请求参数验证失败 |
| CIRCULAR_DEPENDENCY | 400 | 循环依赖 |
| INTERNAL_ERROR | 500 | 内部错误 |

---

## API接口清单

### 项目相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/project | 获取项目简述 |
| GET | /api/v1/project/constitution | 获取项目公约 |
| PUT | /api/v1/project/constitution | 更新项目公约 |

### 模块相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/modules | 获取所有模块（树形） |
| GET | /api/v1/modules/:id | 获取模块概览 |
| POST | /api/v1/modules | 创建模块 |
| PUT | /api/v1/modules/:id | 更新模块 |
| DELETE | /api/v1/modules/:id | 删除模块 |
| GET | /api/v1/modules/:id/tasks | 获取模块中的任务ID列表 |

### 任务相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/tasks | 获取任务列表 |
| GET | /api/v1/tasks/:id | 获取任务详情 |
| POST | /api/v1/tasks | 创建任务 |
| PUT | /api/v1/tasks/:id | 更新任务 |
| DELETE | /api/v1/tasks/:id | 删除任务 |

### 依赖相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/dependencies | 获取依赖列表 |
| POST | /api/v1/dependencies | 创建依赖 |
| DELETE | /api/v1/dependencies/:id | 删除依赖 |

### 通知相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/notifications | 获取通知列表 |
| POST | /api/v1/notifications | 发送通知 |
| POST | /api/v1/notifications/:id/read | 标记通知已读 |

### 锁相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/lock | 锁定资源 |
| POST | /api/v1/unlock | 解锁资源 |
| GET | /api/v1/lock/status | 查询锁定状态 |

### 工具相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/tools/open-browser | 打开前端页面 |

---

## 接口详细定义

### 1. 项目相关

#### 1.1 获取项目简述

```
GET /api/v1/project
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": "proj-123",
    "name": "My Project",
    "constitution": "# 项目宪法\n\n1. 所有公共函数必须有文档...",
    "version": 5,
    "createdAt": 1645564800000,
    "updatedAt": 1645564900000
  }
}
```

#### 1.2 获取项目公约

```
GET /api/v1/project/constitution
```

**响应**:
```json
{
  "success": true,
  "data": {
    "constitution": "# 项目宪法\n\n1. 所有公共函数必须有文档...",
    "version": 5,
    "updatedAt": 1645564900000
  }
}
```

#### 1.3 更新项目公约

```
PUT /api/v1/project/constitution
```

**请求体**:
```json
{
  "constitution": "# 更新后的项目宪法\n\n1. 新规则...",
  "version": 5
}
```

**响应**: 返回更新后的项目信息

---

### 2. 模块相关

#### 2.1 获取所有模块（树形）

```
GET /api/v1/modules
```

**查询参数**:
- `project_id`: 项目ID（可选，默认当前项目）
- `flat`: 是否返回扁平列表（可选，默认false）

**响应（树形）**:
```json
{
  "success": true,
  "data": [
    {
      "id": "mod-1",
      "name": "前端",
      "status": "developing",
      "children": [
        {
          "id": "mod-2",
          "name": "组件库",
          "status": "completed",
          "children": []
        }
      ]
    }
  ]
}
```

#### 2.2 获取模块概览

```
GET /api/v1/modules/:id
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": "mod-1",
    "parentId": null,
    "projectId": "proj-123",
    "name": "前端",
    "description": "前端模块",
    "prompt": "使用React实现...",
    "status": "developing",
    "testCoverage": 75.5,
    "upstreamContractSummary": "需要API模块提供REST接口",
    "downstreamContractSummary": "提供UI组件",
    "locked": false,
    "taskCount": 10,
    "completedTaskCount": 7,
    "version": 3,
    "createdAt": 1645564800000,
    "updatedAt": 1645564900000
  }
}
```

#### 2.3 创建模块

```
POST /api/v1/modules
```

**请求体**:
```json
{
  "parentId": "mod-parent",  // 可选，null表示根模块
  "name": "新模块",
  "description": "模块描述",
  "prompt": "模块提示词"
}
```

**响应**: 201 Created，返回创建的模块对象

#### 2.4 更新模块

```
PUT /api/v1/modules/:id
```

**请求体**:
```json
{
  "name": "更新后的名称",
  "description": "更新后的描述",
  "status": "completed",
  "testCoverage": 90.0,
  "version": 3
}
```

**响应**: 返回更新后的模块对象

#### 2.5 删除模块

```
DELETE /api/v1/modules/:id
```

**查询参数**:
- `force`: 是否强制删除（包含子模块和任务）

**响应**: 204 No Content

#### 2.6 获取模块中的任务ID列表

```
GET /api/v1/modules/:id/tasks
```

**查询参数**:
- `fields`: 返回字段（默认id，可指定name,status等）

**响应**:
```json
{
  "success": true,
  "data": [
    { "id": "task-1", "name": "任务1", "status": "completed" },
    { "id": "task-2", "name": "任务2", "status": "in_progress" }
  ]
}
```

---

### 3. 任务相关

#### 3.1 获取任务列表

```
GET /api/v1/tasks
```

**查询参数**:
- `module_id`: 按模块筛选
- `status`: 按状态筛选
- `assignee`: 按负责人筛选

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": "task-1",
      "moduleId": "mod-1",
      "name": "实现登录按钮",
      "status": "completed",
      "assignee": "ai-agent-1"
    }
  ],
  "pagination": {
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

#### 3.2 获取任务详情

```
GET /api/v1/tasks/:id
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": "task-456",
    "moduleId": "mod-2",
    "name": "实现登录按钮",
    "description": "使用React实现一个登录按钮组件",
    "status": "ready",
    "assignee": null,
    "upstreamContractDetail": {
      "interfaces": [
        {
          "name": "AuthService",
          "type": "service",
          "signature": "login(username: string, password: string): Promise<User>"
        }
      ]
    },
    "downstreamContractDetail": {
      "interfaces": [
        {
          "name": "LoginButton",
          "type": "component",
          "signature": "LoginButton(props: { onLogin: Function })"
        }
      ]
    },
    "prompt": "使用React实现一个登录按钮，需要调用AuthService.login方法",
    "tests": [
      {
        "id": "test-1",
        "name": "点击触发登录事件",
        "command": "npm test login.test.js",
        "status": "pending"
      }
    ],
    "logs": [
      {
        "id": "log-1",
        "level": "info",
        "message": "任务开始执行",
        "timestamp": 1645564800000
      }
    ],
    "codePaths": ["src/components/LoginButton.tsx"],
    "humanAssistance": {
      "items": [],
      "allApproved": true
    },
    "locked": false,
    "lockedBy": null,
    "version": 3,
    "createdAt": 1645564800000,
    "updatedAt": 1645564900000
  }
}
```

#### 3.3 创建任务

```
POST /api/v1/tasks
```

**请求体**:
```json
{
  "moduleId": "mod-2",
  "name": "实现注册表单",
  "description": "实现用户注册表单组件",
  "upstreamContractDetail": {
    "interfaces": [...]
  },
  "downstreamContractDetail": {
    "interfaces": [...]
  },
  "prompt": "使用React实现注册表单",
  "tests": [
    {
      "name": "表单验证",
      "command": "npm test register.test.js"
    }
  ],
  "codePaths": ["src/components/RegisterForm.tsx"]
}
```

**响应**: 201 Created，返回创建的任务对象

#### 3.4 更新任务

```
PUT /api/v1/tasks/:id
```

**请求体**:
```json
{
  "status": "completed",
  "tests": [
    {
      "id": "test-1",
      "status": "passed"
    }
  ],
  "version": 3
}
```

**响应**: 返回更新后的任务对象

**错误响应（版本冲突）**:
```json
{
  "success": false,
  "error": {
    "code": "VERSION_CONFLICT",
    "message": "任务已被其他操作修改，请刷新后重试",
    "details": {
      "currentVersion": 4,
      "providedVersion": 3
    }
  }
}
```

#### 3.5 删除任务

```
DELETE /api/v1/tasks/:id
```

**响应**: 204 No Content

**错误响应（存在依赖）**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "无法删除任务，存在依赖关系",
    "details": {
      "dependencies": ["task-789", "task-790"]
    }
  }
}
```

---

### 4. 依赖相关

#### 4.1 获取依赖列表

```
GET /api/v1/dependencies
```

**查询参数**:
- `task_id`: 按任务筛选
- `direction`: upstream | downstream

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": "dep-1",
      "upstreamTaskId": "task-1",
      "downstreamTaskId": "task-2",
      "contract": {
        "interfaceVersion": "1.0.0"
      },
      "status": "active"
    }
  ]
}
```

#### 4.2 创建依赖

```
POST /api/v1/dependencies
```

**请求体**:
```json
{
  "upstreamTaskId": "task-1",
  "downstreamTaskId": "task-2",
  "contract": {
    "interfaceVersion": "1.0.0",
    "description": "task-2依赖task-1的输出"
  }
}
```

**响应**: 201 Created

**错误响应（循环依赖）**:
```json
{
  "success": false,
  "error": {
    "code": "CIRCULAR_DEPENDENCY",
    "message": "创建此依赖将导致循环依赖",
    "details": {
      "path": ["task-2", "task-3", "task-1", "task-2"]
    }
  }
}
```

#### 4.3 删除依赖

```
DELETE /api/v1/dependencies/:id
```

**响应**: 204 No Content

---

### 5. 通知相关

#### 5.1 获取通知列表

```
GET /api/v1/notifications
```

**查询参数**:
- `to_task_id`: 接收任务ID
- `from_task_id`: 发送任务ID
- `read`: true | false
- `type`: 通知类型

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": "notif-1",
      "fromTaskId": "task-123",
      "toTaskId": "task-456",
      "type": "task_completed",
      "title": "上游任务已完成",
      "message": "task-123已完成，可以开始task-456",
      "read": false,
      "createdAt": 1645564800000
    }
  ]
}
```

#### 5.2 发送通知

```
POST /api/v1/notifications
```

**请求体**:
```json
{
  "fromTaskId": "task-123",
  "toTaskId": "task-456",
  "type": "task_completed",
  "title": "上游任务已完成",
  "message": "task-123已完成，可以开始task-456"
}
```

**响应**: 201 Created

#### 5.3 标记通知已读

```
POST /api/v1/notifications/:id/read
```

**响应**: 返回更新后的通知对象

---

### 6. 锁相关

#### 6.1 锁定资源

```
POST /api/v1/lock
```

**请求体**:
```json
{
  "targetType": "task",  // task | module
  "targetId": "task-456",
  "locker": "ai-agent-1"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "locked": true,
    "lockedBy": "ai-agent-1",
    "expiresAt": 1645568400000
  }
}
```

**错误响应（已被锁定）**:
```json
{
  "success": false,
  "error": {
    "code": "LOCKED",
    "message": "资源已被锁定",
    "details": {
      "lockedBy": "ai-agent-2",
      "expiresAt": 1645568400000
    }
  }
}
```

#### 6.2 解锁资源

```
POST /api/v1/unlock
```

**请求体**:
```json
{
  "targetType": "task",
  "targetId": "task-456",
  "locker": "ai-agent-1"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "locked": false
  }
}
```

#### 6.3 查询锁定状态

```
GET /api/v1/lock/status?target_type=task&target_id=task-456
```

**响应**:
```json
{
  "success": true,
  "data": {
    "locked": true,
    "lockedBy": "ai-agent-1",
    "lockedAt": 1645564800000,
    "expiresAt": 1645568400000
  }
}
```

---

### 7. 工具相关

#### 7.1 打开前端页面

```
POST /api/v1/tools/open-browser
```

**请求体**:
```json
{
  "path": "/modules/mod-1"  // 可选，打开后跳转的路径
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "url": "http://localhost:34567/modules/mod-1"
  }
}
```

---

## WebSocket事件

服务支持WebSocket连接，用于实时推送状态变更。

**连接地址**: `ws://localhost:34567/ws`

### 事件类型

| 事件 | 说明 | 数据 |
|------|------|------|
| task.created | 任务创建 | 任务对象 |
| task.updated | 任务更新 | 任务对象 |
| task.deleted | 任务删除 | { id } |
| module.updated | 模块更新 | 模块对象 |
| notification | 新通知 | 通知对象 |

### 示例

```json
{
  "event": "task.updated",
  "data": {
    "id": "task-456",
    "status": "completed"
  },
  "timestamp": 1645564800000
}
```
