# AITDD API 文档

## 概述

AITDD 提供 RESTful API 接口，默认端口为 34567。

**Base URL**: `http://localhost:34567/api/v1`

## 通用响应格式

所有 API 返回统一的 JSON 格式：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "错误描述",
  "data": null
}
```

## 项目 API

### 获取项目列表

```
GET /projects
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| page | int | 页码 (默认1) |
| pageSize | int | 每页数量 (默认20) |

**响应**:
```json
{
  "code": 200,
  "data": {
    "projects": [...],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### 创建项目

```
POST /projects
```

**请求体**:
```json
{
  "name": "项目名称",
  "description": "项目描述",
  "path": "/path/to/project"
}
```

### 获取项目详情

```
GET /projects/:id
```

### 更新项目

```
PUT /projects/:id
```

### 删除项目

```
DELETE /projects/:id
```

## 模块 API

### 获取模块列表

```
GET /modules
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| projectId | string | 项目ID (必填) |
| parentId | string | 父模块ID (可选) |

### 创建模块

```
POST /modules
```

**请求体**:
```json
{
  "projectId": "项目ID",
  "parentId": "父模块ID",
  "name": "模块名称",
  "description": "模块描述",
  "order": 0
}
```

### 获取模块树

```
GET /modules/tree
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| projectId | string | 项目ID (必填) |

## 任务 API

### 获取任务列表

```
GET /tasks
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| moduleId | string | 模块ID |
| status | string | 状态过滤 |
| page | int | 页码 |
| pageSize | int | 每页数量 |

### 创建任务

```
POST /tasks
```

**请求体**:
```json
{
  "moduleId": "模块ID",
  "title": "任务标题",
  "description": "任务描述",
  "status": "pending",
  "priority": "medium",
  "inputContract": {
    "requiredFiles": ["file1.ts"],
    "requiredData": ["data1"],
    "description": "输入描述"
  },
  "outputContract": {
    "producedFiles": ["output.ts"],
    "producedData": ["result"],
    "description": "输出描述"
  }
}
```

### 更新任务状态

```
PUT /tasks/:id/status
```

**请求体**:
```json
{
  "status": "in_progress",
  "version": 1
}
```

## 依赖 API

### 获取依赖列表

```
GET /dependencies
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| taskId | string | 任务ID |

### 创建依赖

```
POST /dependencies
```

**请求体**:
```json
{
  "taskId": "任务ID",
  "dependsOnTaskId": "依赖的任务ID",
  "type": "finish_to_start"
}
```

### 验证依赖 (循环检测)

```
POST /dependencies/validate
```

**请求体**:
```json
{
  "taskId": "任务ID",
  "dependsOnTaskId": "依赖的任务ID"
}
```

## 锁定 API

### 锁定资源

```
POST /lock
```

**请求体**:
```json
{
  "resourceType": "task",
  "resourceId": "资源ID",
  "lockedBy": "代理标识"
}
```

### 解锁资源

```
DELETE /lock
```

**请求体**:
```json
{
  "resourceType": "task",
  "resourceId": "资源ID",
  "lockedBy": "代理标识"
}
```

### 获取锁定状态

```
GET /lock/status
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| resourceType | string | 资源类型 |
| resourceId | string | 资源ID |

## 通知 API

### 获取通知列表

```
GET /notifications
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| unread | boolean | 仅未读 |
| type | string | 类型过滤 |
| page | int | 页码 |
| pageSize | int | 每页数量 |

### 创建通知

```
POST /notifications
```

**请求体**:
```json
{
  "type": "system",
  "title": "通知标题",
  "content": "通知内容",
  "link": "/tasks/123"
}
```

### 标记通知已读

```
POST /notifications/:id/read
```

## WebSocket

### 连接

```
ws://localhost:34567/ws
```

**查询参数**:
| 参数 | 类型 | 描述 |
|------|------|------|
| clientId | string | 客户端ID (可选) |

### 消息格式

```json
{
  "type": "notification",
  "payload": { ... }
}
```

### 消息类型

| 类型 | 描述 |
|------|------|
| connected | 连接成功 |
| notification | 新通知 |
| task_updated | 任务更新 |
| module_updated | 模块更新 |
| ping/pong | 心跳 |

## 工具 API

### 打开浏览器

```
POST /tools/open-browser
```

**请求体**:
```json
{
  "url": "https://example.com"
}
```

## HTTP 状态码

| 状态码 | 描述 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 423 | 资源被锁定 |
| 500 | 服务器内部错误 |
