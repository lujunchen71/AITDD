# AITDD 系统重构计划

**Created**: 2026-02-24  
**Status**: Draft

## 一、当前系统分析

### 1.1 已完成的功能

#### 后端 (Go + Gin + GORM)
- ✅ 基础数据模型 (Project, Module, Task, Dependency, Notification, ChangeHistory, Config)
- ✅ 模块依赖关系表 (module_dependencies)
- ✅ RESTful API 路由系统
- ✅ 模块 CRUD 操作
- ✅ 任务 CRUD 操作
- ✅ 模块依赖管理 (创建、查询、删除、循环依赖检测)
- ✅ 任务依赖管理
- ✅ 锁机制 (锁定/解锁)
- ✅ 通知系统
- ✅ 变更追踪服务
- ✅ 冲突解决服务
- ✅ 同步服务
- ✅ WebSocket 支持

#### 前端 (React + TypeScript + Ant Design)
- ✅ 主布局框架 (Header, Sidebar, MainLayout)
- ✅ 模块树形展示
- ✅ 模块详情面板
- ✅ 任务列表
- ✅ 任务详情面板
- ✅ 模块表单
- ✅ 任务表单
- ✅ 通知中心
- ✅ 设置页面 (冲突解决、同步设置)
- ✅ 项目状态管理 (Zustand)
- ✅ API 客户端封装

### 1.2 存在的问题

#### UI/UX 问题
1. **页面布局单调**: 当前使用简单的左右两栏布局，信息密度低
2. **缺少依赖可视化**: 没有图形化展示模块和任务的依赖关系
3. **视觉设计欠缺**: 使用 Ant Design 默认主题，缺乏个性化
4. **缺少全局视图**: 无法一眼看到整个项目的依赖网络
5. **提示词编辑体验差**: 没有专门的提示词编辑器和版本对比

#### 数据模型问题
1. **类型定义不一致**: 前端 TypeScript 类型与后端 Go 结构体字段不完全匹配
2. **缺少提示词版本表**: 无法追溯提示词的变更历史
3. **任务契约字段弱类型**: upstream_contract_detail 和 downstream_contract_detail 使用 JSON 字符串，缺少结构化

#### API 问题
1. **MCP 工具集不完整**: 缺少专门的 MCP 服务器实现
2. **响应格式不统一**: 部分 API 返回格式不一致
3. **缺少批量操作**: 无法批量创建/更新任务

#### 工作流问题
1. **缺少 CLI 工具**: 没有 aitdd 命令行工具
2. **缺少 Workflow 文件**: 没有为 KiloCode 等插件生成命令文件
3. **缺少初始化流程**: 没有 aitdd init 命令

---

## 二、重构目标

### 2.1 核心定位

AITDD 是一个**提示词依赖管理系统**，核心功能：
1. 管理模块和任务的层级结构
2. 可视化展示依赖关系网络
3. 管理提示词及其版本
4. 通过 MCP 协议供 AI 编程软件调用
5. 提供 CLI 工具和工作流文件

### 2.2 优先级矩阵

| 优先级 | 功能模块 | 重要性 | 紧急性 |
|--------|----------|--------|--------|
| P0 | UI 重构 - 依赖可视化 | 高 | 高 |
| P0 | 类型定义统一 | 高 | 高 |
| P1 | MCP 工具集实现 | 高 | 中 |
| P1 | 提示词版本管理 | 高 | 中 |
| P2 | CLI 工具开发 | 中 | 中 |
| P2 | Workflow 文件生成 | 中 | 中 |
| P3 | 远程同步功能 | 低 | 低 |

---

## 三、重构计划

### 阶段一：UI 重构与依赖可视化 (P0)

#### 1.1 页面布局重构

```
新布局结构:
┌─────────────────────────────────────────────────────────────┐
│ Header (Logo, 项目名称，全局操作，通知)                        │
├──────────┬────────────────────────────┬──────────────────────┤
│          │                            │                      │
│ Sidebar  │   Main Content Area        │   Right Panel        │
│ 导航菜单 │                            │   - 依赖图           │
│          │   - 模块树 / 任务列表       │   - 属性面板         │
│          │   - 详情面板               │   - 快速操作         │
│          │                            │                      │
├──────────┴────────────────────────────┴──────────────────────┤
│ Footer (状态栏，同步状态，版本信息)                            │
└─────────────────────────────────────────────────────────────┘
```

#### 1.2 依赖图组件

使用 React Flow 实现：
- 节点类型：模块节点、任务节点
- 连线类型：实线 (强依赖)、虚线 (弱依赖)
- 交互功能：
  - 拖拽节点
  - 缩放画布
  - 点击节点显示详情
  - 拖拽连线创建依赖
  - 右键菜单 (删除依赖、锁定等)
- 布局算法：DAG 自动布局
- 筛选功能：按状态筛选、按模块筛选

#### 1.3 视觉设计

```
配色方案 (深色主题):
- 背景色：#0f0f23, #1a1a2e, #16213e
- 强调色：#e94560 (粉), #00d9ff (青), #7b2cbf (紫)
- 文字色：#ffffff, #e0e0e0, #a0a0a0
- 边框色：#2d2d44, #3d3d5c

节点颜色:
- 模块节点：蓝色渐变 (#1e3a8a → #3b82f6)
- 任务节点：绿色渐变 (#065f46 → #10b981)
- 选中状态：金色边框 (#fbbf24) + 发光效果
- 锁定状态：红色边框 (#ef4444)
- 完成状态：绿色对勾标识
```

### 阶段二：数据模型与类型统一 (P0)

#### 2.1 前端类型定义更新

```typescript
// 后端 Go 模型与前端 TypeScript 类型映射

// 模块类型 - 与后端 models.Module 对齐
export interface Module {
  id: string;
  parentId?: string | null;
  projectId: string;
  name: string;
  description?: string;
  prompt?: string;
  status: 'designing' | 'developing' | 'completed' | 'deprecated';
  testCoverage: number;
  upstreamContractSummary?: string;
  downstreamContractSummary?: string;
  locked: boolean;
  lockedBy?: string | null;
  lockedAt?: number | null;
  lockExpiresAt?: number | null;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: 'SYNCED' | 'PENDING_UPLOAD' | 'PENDING_DOWNLOAD' | 'CONFLICT';
  children?: Module[];
}

// 任务类型 - 与后端 models.Task 对齐
export interface Task {
  id: string;
  moduleId: string;
  name: string;
  description?: string;
  status: 'ready' | 'claimed' | 'in_progress' | 'pending_review' | 'completed' | 'failed' | 'blocked';
  assignee?: string | null;
  upstreamContractDetail?: string; // JSON string
  downstreamContractDetail?: string; // JSON string
  prompt?: string;
  tests?: string; // JSON string
  logs?: string; // JSON string
  codePaths?: string; // JSON string
  humanAssistance?: string; // JSON string
  locked: boolean;
  lockedBy?: string | null;
  lockedAt?: number | null;
  lockExpiresAt?: number | null;
  createdAt: number;
  updatedAt: number;
  version: number;
  syncStatus: 'SYNCED' | 'PENDING_UPLOAD' | 'PENDING_DOWNLOAD' | 'CONFLICT';
}

// 契约接口 - 结构化数据
export interface ContractInterface {
  name: string;
  type: 'function' | 'class' | 'api' | 'cli';
  signature: string;
  description: string;
  inputs: unknown[];
  outputs: unknown[];
}

export interface ContractDetail {
  interfaces: ContractInterface[];
  dataStructures: unknown[];
  version: string;
}
```

#### 2.2 新增数据表

```sql
-- 提示词版本表
CREATE TABLE IF NOT EXISTS prompt_versions (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('module', 'task')),
    entity_id TEXT NOT NULL,
    version INTEGER NOT NULL,
    prompt TEXT NOT NULL,
    change_summary TEXT,
    created_by TEXT,
    created_at INTEGER NOT NULL
);

-- 创建索引
CREATE INDEX idx_prompt_versions_entity ON prompt_versions(entity_type, entity_id);
CREATE INDEX idx_prompt_versions_created ON prompt_versions(created_at);
```

### 阶段三：MCP 工具集实现 (P1)

#### 3.1 MCP 服务器架构

```
backend/cmd/aitdd/main.go
└── MCP Server (puppeteer-core based)
    ├── Tools
    │   ├── get_project_summary
    │   ├── get_constitution
    │   ├── get_all_modules
    │   ├── get_module_overview
    │   ├── get_module_task_ids
    │   ├── get_task_details
    │   ├── update_task
    │   ├── create_task
    │   ├── delete_task
    │   ├── delete_module
    │   ├── open_frontend
    │   ├── send_notification
    │   ├── read_notifications
    │   ├── lock_resource
    │   ├── unlock_resource
    │   ├── get_lock_status
    │   ├── create_module_dependency
    │   ├── get_module_dependencies
    │   ├── create_task_dependency
    │   └── get_task_dependencies
    └── API Client
        └── HTTP Client → REST API
```

#### 3.2 MCP 工具定义

```go
// MCP Tool 定义示例
type MCPTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
    Handler     func(context.Context, map[string]interface{}) (interface{}, error)
}

// 工具注册
func RegisterTools() []MCPTool {
    return []MCPTool{
        {
            Name:        "get_project_summary",
            Description: "获取项目简述信息",
            InputSchema: map[string]interface{}{},
            Handler:     GetProjectSummary,
        },
        // ... 其他工具
    }
}
```

### 阶段四：CLI 工具与工作流 (P2)

#### 4.1 CLI 工具结构

```
aitdd-cli/
├── cmd/
│   ├── init.go          # aitdd init
│   ├── serve.go         # aitdd serve
│   ├── version.go       # aitdd version
│   └── config.go        # aitdd config
├── pkg/
│   ├── initializer/     # 初始化逻辑
│   ├── plugins/         # 插件配置
│   │   ├── kilocode.go
│   │   ├── opencode.go
│   │   └── claudecode.go
│   └── templates/       // 模板文件
└── main.go
```

#### 4.2 Workflow 文件模板

```markdown
---
name: aitdd.constitution
description: 制定或更新项目管理原则和开发指南
---

# AITDD 项目公约命令

## 功能说明
此命令用于制定或更新项目的管理原则和开发指南。

## MCP 工具
- `get_constitution`: 获取当前项目公约
- `update_constitution`: 更新项目公约

## 使用示例
用户：我想添加一条新的开发规范
AI: 好的，我将调用 MCP 工具更新项目公约...
```

### 阶段五：提示词版本管理 (P1)

#### 5.1 版本追踪

```go
// PromptVersion 模型
type PromptVersion struct {
    ID            string `json:"id" gorm:"primaryKey;type:text"`
    EntityType    string `json:"entityType" gorm:"not null;type:text"` // 'module' or 'task'
    EntityID      string `json:"entityId" gorm:"not null;type:text;index"`
    Version       int    `json:"version" gorm:"not null"`
    Prompt        string `json:"prompt" gorm:"not null;type:text"`
    ChangeSummary string `json:"changeSummary" gorm:"type:text"`
    CreatedBy     string `json:"createdBy" gorm:"type:text"`
    CreatedAt     int64  `json:"createdAt" gorm:"not null"`
}

// 创建提示词版本
func CreatePromptVersion(entityType, entityID, prompt, changeSummary, createdBy string) (*PromptVersion, error) {
    // 获取当前最新版本
    var lastVersion PromptVersion
    result := database.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
        Order("version DESC").First(&lastVersion)
    
    newVersion := 1
    if result.Error == nil {
        newVersion = lastVersion.Version + 1
    }
    
    version := PromptVersion{
        ID:            uuid.New().String(),
        EntityType:    entityType,
        EntityID:      entityID,
        Version:       newVersion,
        Prompt:        prompt,
        ChangeSummary: changeSummary,
        CreatedBy:     createdBy,
        CreatedAt:     time.Now().UnixMilli(),
    }
    
    return &version, database.DB.Create(&version).Error
}
```

---

## 四、实施时间表

| 阶段 | 任务 | 预计工时 | 依赖 |
|------|------|----------|------|
| 阶段一 | UI 重构 - 布局 | 2 天 | 无 |
| 阶段一 | UI 重构 - 依赖图 | 3 天 | 布局完成 |
| 阶段二 | 类型定义统一 | 1 天 | 无 |
| 阶段二 | 数据模型扩展 | 1 天 | 无 |
| 阶段三 | MCP 工具集 | 3 天 | 类型统一 |
| 阶段四 | CLI 工具 | 3 天 | MCP 完成 |
| 阶段四 | Workflow 文件 | 1 天 | CLI 完成 |
| 阶段五 | 提示词版本管理 | 2 天 | 数据模型 |

**总计**: 约 16 个工作日

---

## 五、成功标准

| 指标 | 目标值 | 测量方式 |
|------|--------|----------|
| 依赖图渲染性能 | 支持 500+ 节点 60fps | Chrome DevTools Performance |
| MCP 工具响应时间 | P95 < 50ms | API 监控 |
| 初始化时间 | < 3 秒 | 计时器 |
| 前端加载时间 | < 1.5 秒 | Lighthouse |
| 类型覆盖率 | 100% TypeScript | tsc --noEmit |

---

## 六、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| React Flow 性能问题 | 高 | 中 | 使用虚拟滚动、懒加载 |
| MCP 协议兼容性 | 高 | 中 | 遵循官方 MCP 规范 |
| CLI 跨平台兼容性 | 中 | 中 | 使用 Go 编译多平台二进制 |
| 数据迁移复杂性 | 中 | 低 | 提供回滚脚本 |
