# Tasks: AITDD可视化任务治理系统

**Input**: 设计文档来自 `/specs/001-visual-task-governance/`
**Prerequisites**: plan.md (required), spec.md (required), data-model.md, contracts/api-contracts.md, frontend-architecture.md

**组织方式**: 任务按用户故事分组，支持独立实现和测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行执行（不同文件，无依赖）
- **[Story]**: 所属用户故事 (US1, US2, US3...)
- 包含精确文件路径

---

## Phase 1: Setup (项目初始化)

**Purpose**: 项目结构和基础配置

- [ ] T001 创建后端项目结构 backend/
- [ ] T002 初始化Go模块 backend/go.mod
- [ ] T003 [P] 创建前端项目结构 frontend/
- [ ] T004 [P] 初始化前端项目 frontend/package.json
- [ ] T005 [P] 配置TypeScript frontend/tsconfig.json
- [ ] T006 [P] 配置Vite frontend/vite.config.ts
- [ ] T007 [P] 配置TailwindCSS frontend/tailwind.config.js
- [ ] T008 创建CLI入口文件 cmd/aitdd/main.go

---

## Phase 2: Foundational (阻塞性前置条件)

**Purpose**: 所有用户故事依赖的核心基础设施

**⚠️ CRITICAL**: 此阶段必须完成后才能开始用户故事实现

### 数据库层

- [ ] T009 创建Project模型 backend/internal/models/project.go
- [ ] T010 [P] 创建Module模型 backend/internal/models/module.go
- [ ] T011 [P] 创建Task模型 backend/internal/models/task.go
- [ ] T012 [P] 创建Dependency模型 backend/internal/models/dependency.go
- [ ] T013 [P] 创建Notification模型 backend/internal/models/notification.go
- [ ] T014 [P] 创建ChangeHistory模型 backend/internal/models/change_history.go
- [ ] T015 [P] 创建Config模型 backend/internal/models/config.go
- [ ] T016 创建数据库初始化脚本 backend/internal/database/init.go
- [ ] T017 创建数据库迁移SQL backend/migrations/001_init.sql

### API基础设施

- [ ] T018 创建Gin路由器 backend/internal/api/routes.go
- [ ] T019 [P] 创建CORS中间件 backend/internal/api/middleware/cors.go
- [ ] T020 [P] 创建日志中间件 backend/internal/api/middleware/logger.go
- [ ] T021 [P] 创建错误处理中间件 backend/internal/api/middleware/error.go
- [ ] T022 创建统一响应结构 backend/internal/api/response.go

### 前端基础设施

- [ ] T023 创建API客户端 frontend/src/services/api.ts
- [ ] T024 [P] 创建UI状态Store frontend/src/stores/useUIStore.ts
- [ ] T025 [P] 创建通用类型定义 frontend/src/types/common.ts
- [ ] T026 [P] 创建API类型定义 frontend/src/types/api.ts
- [ ] T027 创建主布局组件 frontend/src/components/layout/MainLayout/index.tsx
- [ ] T028 [P] 创建Header组件 frontend/src/components/layout/Header/index.tsx
- [ ] T029 [P] 创建Sidebar组件 frontend/src/components/layout/Sidebar/index.tsx

**Checkpoint**: 基础设施就绪 - 可以开始用户故事实现

---

## Phase 3: User Story 1 - 项目初始化与插件集成 (Priority: P1) 🎯 MVP

**Goal**: 用户执行`aitdd init`命令初始化项目并选择AI插件类型

**Independent Test**: 在空目录执行`aitdd init`，选择KiloCode，验证生成正确的目录结构和workflow文件

### CLI实现

- [ ] T030 [US1] 实现init命令逻辑 cmd/init.go
- [ ] T031 [US1] 创建目录结构生成器 backend/internal/services/init_service.go
- [ ] T032 [US1] 创建数据库初始化服务 backend/internal/services/database_service.go
- [ ] T033 [US1] 创建配置文件生成器 backend/internal/services/config_service.go

### Workflow模板

- [ ] T034 [P] [US1] 创建KiloCode workflow模板 backend/templates/kilocode/
- [ ] T035 [P] [US1] 创建OpenCode workflow模板 backend/templates/opencode/
- [ ] T036 [P] [US1] 创建ClaudeCode workflow模板 backend/templates/claudecode/
- [ ] T037 [US1] 创建AITDD_GUIDE.md模板 backend/templates/AITDD_GUIDE.md

### 前端初始化页面

- [ ] T038 [US1] 创建欢迎/初始化页面 frontend/src/features/setup/index.tsx
- [ ] T039 [US1] 创建插件选择组件 frontend/src/features/setup/components/PluginSelector.tsx

**Checkpoint**: 用户故事1完成 - `aitdd init`命令可用

---

## Phase 4: User Story 2 - 服务启动与MCP注册 (Priority: P1)

**Goal**: 用户执行`aitdd serve`启动服务，AI通过`/aitdd.start`注册MCP

**Independent Test**: 执行`aitdd serve`，在浏览器访问localhost:34567验证前端加载

### CLI实现

- [ ] T040 [US2] 实现serve命令逻辑 cmd/serve.go
- [ ] T041 [US2] 创建HTTP服务器 backend/internal/server/http.go
- [ ] T042 [US2] 创建静态文件服务 backend/internal/server/static.go

### 项目API

- [ ] T043 [US2] 创建项目Handler backend/internal/api/handlers/project.go
- [ ] T044 [US2] 实现GET /api/v1/project backend/internal/api/handlers/project.go
- [ ] T045 [US2] 实现GET /api/v1/project/constitution backend/internal/api/handlers/project.go
- [ ] T046 [US2] 实现PUT /api/v1/project/constitution backend/internal/api/handlers/project.go

### 前端仪表盘

- [ ] T047 [US2] 创建仪表盘页面 frontend/src/features/dashboard/index.tsx
- [ ] T048 [US2] 创建项目信息组件 frontend/src/features/dashboard/components/ProjectInfo.tsx
- [ ] T049 [US2] 创建宪法展示组件 frontend/src/features/dashboard/components/ConstitutionView.tsx

**Checkpoint**: 用户故事2完成 - 服务启动和MCP注册可用

---

## Phase 5: User Story 3 - 模块与任务的增删改查 (Priority: P1)

**Goal**: AI通过API管理模块和任务

**Independent Test**: 通过API创建模块、创建任务、查询、更新、删除

### 模块API

- [ ] T050 [US3] 创建模块Handler backend/internal/api/handlers/module.go
- [ ] T051 [US3] 实现GET /api/v1/modules backend/internal/api/handlers/module.go
- [ ] T052 [US3] 实现GET /api/v1/modules/:id backend/internal/api/handlers/module.go
- [ ] T053 [US3] 实现POST /api/v1/modules backend/internal/api/handlers/module.go
- [ ] T054 [US3] 实现PUT /api/v1/modules/:id backend/internal/api/handlers/module.go
- [ ] T055 [US3] 实现DELETE /api/v1/modules/:id backend/internal/api/handlers/module.go
- [ ] T056 [US3] 实现GET /api/v1/modules/:id/tasks backend/internal/api/handlers/module.go

### 任务API

- [ ] T057 [US3] 创建任务Handler backend/internal/api/handlers/task.go
- [ ] T058 [US3] 实现GET /api/v1/tasks backend/internal/api/handlers/task.go
- [ ] T059 [US3] 实现GET /api/v1/tasks/:id backend/internal/api/handlers/task.go
- [ ] T060 [US3] 实现POST /api/v1/tasks backend/internal/api/handlers/task.go
- [ ] T061 [US3] 实现PUT /api/v1/tasks/:id backend/internal/api/handlers/task.go
- [ ] T062 [US3] 实现DELETE /api/v1/tasks/:id backend/internal/api/handlers/task.go

### 业务服务

- [ ] T063 [US3] 创建模块服务 backend/internal/services/module_service.go
- [ ] T064 [US3] 创建任务服务 backend/internal/services/task_service.go

### 前端模块管理

- [ ] T065 [US3] 创建模块树组件 frontend/src/features/modules/components/ModuleTree/index.tsx
- [ ] T066 [US3] 创建模块树节点组件 frontend/src/features/modules/components/ModuleTree/ModuleTreeNode.tsx
- [ ] T067 [US3] 创建模块详情组件 frontend/src/features/modules/components/ModuleDetail/index.tsx
- [ ] T068 [US3] 创建模块表单组件 frontend/src/features/modules/components/ModuleForm/index.tsx

### 前端任务管理

- [ ] T069 [US3] 创建任务列表组件 frontend/src/features/tasks/components/TaskList/index.tsx
- [ ] T070 [US3] 创建任务表单组件 frontend/src/features/tasks/components/TaskForm/index.tsx

**Checkpoint**: 用户故事3完成 - 模块和任务CRUD可用

---

## Phase 6: User Story 4 - 任务依赖与契约管理 (Priority: P2)

**Goal**: 定义任务依赖关系和契约接口

**Independent Test**: 创建两个任务，设置依赖，验证依赖关系正确存储

### 依赖API

- [ ] T071 [US4] 创建依赖Handler backend/internal/api/handlers/dependency.go
- [ ] T072 [US4] 实现GET /api/v1/dependencies backend/internal/api/handlers/dependency.go
- [ ] T073 [US4] 实现POST /api/v1/dependencies backend/internal/api/handlers/dependency.go
- [ ] T074 [US4] 实现DELETE /api/v1/dependencies/:id backend/internal/api/handlers/dependency.go
- [ ] T075 [US4] 实现循环依赖检测 backend/internal/services/dependency_service.go

### 业务服务

- [ ] T076 [US4] 创建依赖服务 backend/internal/services/dependency_service.go

### 前端依赖管理

- [ ] T077 [US4] 创建契约编辑器组件 frontend/src/features/tasks/components/ContractEditor/index.tsx
- [ ] T078 [US4] 创建依赖关系组件 frontend/src/features/tasks/components/DependencyList/index.tsx

**Checkpoint**: 用户故事4完成 - 依赖和契约管理可用

---

## Phase 7: User Story 5 - 可视化任务网络 (Priority: P2)

**Goal**: 浏览器中查看任务依赖图

**Independent Test**: 打开模块页面，查看任务网络图，验证节点和连线正确渲染

### 前端任务图

- [ ] T079 [US5] 创建任务图组件 frontend/src/features/tasks/components/TaskGraph/index.tsx
- [ ] T080 [US5] 创建自定义任务节点 frontend/src/features/tasks/components/TaskGraph/TaskNode.tsx
- [ ] T081 [US5] 创建依赖边组件 frontend/src/features/tasks/components/TaskGraph/DependencyEdge.tsx
- [ ] T082 [US5] 创建图工具栏 frontend/src/features/tasks/components/TaskGraph/GraphToolbar.tsx
- [ ] T083 [US5] 创建图迷你地图 frontend/src/features/tasks/components/TaskGraph/MiniMap.tsx

### 任务详情面板

- [ ] T084 [US5] 创建任务详情面板 frontend/src/features/tasks/components/TaskDetailPanel/index.tsx
- [ ] T085 [US5] 创建基本信息标签页 frontend/src/features/tasks/components/TaskDetailPanel/BasicInfoTab.tsx
- [ ] T086 [US5] 创建契约标签页 frontend/src/features/tasks/components/TaskDetailPanel/ContractTab.tsx
- [ ] T087 [US5] 创建测试标签页 frontend/src/features/tasks/components/TaskDetailPanel/TestsTab.tsx
- [ ] T088 [US5] 创建日志标签页 frontend/src/features/tasks/components/TaskDetailPanel/LogsTab.tsx
- [ ] T089 [US5] 创建人类协助标签页 frontend/src/features/tasks/components/TaskDetailPanel/AssistanceTab.tsx

**Checkpoint**: 用户故事5完成 - 任务网络可视化可用

---

## Phase 8: User Story 6 - 任务锁定与协作 (Priority: P3)

**Goal**: 锁定任务防止并发冲突

**Independent Test**: 锁定任务，验证其他请求无法修改

### 锁API

- [ ] T090 [US6] 创建锁Handler backend/internal/api/handlers/lock.go
- [ ] T091 [US6] 实现POST /api/v1/lock backend/internal/api/handlers/lock.go
- [ ] T092 [US6] 实现POST /api/v1/unlock backend/internal/api/handlers/lock.go
- [ ] T093 [US6] 实现GET /api/v1/lock/status backend/internal/api/handlers/lock.go

### 业务服务

- [ ] T094 [US6] 创建锁服务 backend/internal/services/lock_service.go

### 前端锁功能

- [ ] T095 [US6] 创建锁状态组件 frontend/src/features/tasks/components/LockStatus/index.tsx
- [ ] T096 [US6] 创建锁操作按钮 frontend/src/features/tasks/components/LockButton/index.tsx

**Checkpoint**: 用户故事6完成 - 锁机制可用

---

## Phase 9: User Story 7 - 通知系统 (Priority: P3)

**Goal**: 接收和发送任务间通知

**Independent Test**: 从任务A发送通知到任务B，任务B查询未读通知

### 通知API

- [ ] T097 [US7] 创建通知Handler backend/internal/api/handlers/notification.go
- [ ] T098 [US7] 实现GET /api/v1/notifications backend/internal/api/handlers/notification.go
- [ ] T099 [US7] 实现POST /api/v1/notifications backend/internal/api/handlers/notification.go
- [ ] T100 [US7] 实现POST /api/v1/notifications/:id/read backend/internal/api/handlers/notification.go

### WebSocket

- [ ] T101 [US7] 创建WebSocket服务 backend/internal/websocket/hub.go
- [ ] T102 [US7] 创建WebSocket连接处理 backend/internal/websocket/connection.go

### 前端通知

- [ ] T103 [US7] 创建通知下拉组件 frontend/src/components/feedback/NotificationDropdown/index.tsx
- [ ] T104 [US7] 创建通知列表组件 frontend/src/features/notifications/components/NotificationList/index.tsx
- [ ] T105 [US7] 创建WebSocket Hook frontend/src/hooks/useWebSocket.ts

**Checkpoint**: 用户故事7完成 - 通知系统可用

---

## Phase 10: User Story 8 - 数据同步与版本控制 (Priority: P3)

**Goal**: 本地数据同步到远程数据库

**Independent Test**: 配置远程同步，修改本地数据，验证同步状态正确

### 同步服务

- [ ] T106 [US8] 创建同步服务 backend/internal/services/sync_service.go
- [ ] T107 [US8] 实现变更追踪 backend/internal/services/change_tracker.go
- [ ] T108 [US8] 实现冲突检测 backend/internal/services/conflict_resolver.go

### 前端同步

- [ ] T109 [US8] 创建同步设置页面 frontend/src/features/settings/components/SyncSettings.tsx
- [ ] T110 [US8] 创建冲突解决组件 frontend/src/features/settings/components/ConflictResolver.tsx

**Checkpoint**: 用户故事8完成 - 数据同步可用

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: 跨用户故事的改进

### 工具功能

- [ ] T111 [P] 实现POST /api/v1/tools/open-browser backend/internal/api/handlers/tools.go

### 文档

- [ ] T112 [P] 创建API文档 backend/docs/api.md
- [ ] T113 [P] 创建用户手册 docs/user-guide.md
- [ ] T114 [P] 创建开发文档 docs/development.md

### 测试

- [ ] T115 [P] 后端单元测试 backend/internal/services/*_test.go
- [ ] T116 [P] 前端单元测试 frontend/src/**/*.test.ts

### 构建

- [ ] T117 创建构建脚本 scripts/build.sh
- [ ] T118 创建发布脚本 scripts/release.sh

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 无依赖 - 立即开始
- **Foundational (Phase 2)**: 依赖Setup完成 - **阻塞所有用户故事**
- **User Stories (Phase 3-10)**: 全部依赖Foundational完成
  - US1-US3 (P1) 可并行
  - US4-US5 (P2) 可并行（依赖US3）
  - US6-US8 (P3) 可并行
- **Polish (Phase 11)**: 依赖所有用户故事完成

### User Story Dependencies

```
US1 (P1) ──┐
US2 (P1) ──┼──► US4 (P2) ──► US5 (P2)
US3 (P1) ──┘        │
                    │
                    ▼
              US6 (P3)
              US7 (P3)
              US8 (P3)
```

### Parallel Opportunities

- Phase 1: T003-T008 可并行
- Phase 2: T010-T015 可并行，T019-T021 可并行，T024-T026 可并行，T028-T029 可并行
- Phase 3: T034-T036 可并行
- Phase 11: T111-T116 可并行

---

## Implementation Strategy

### MVP First (仅User Story 1-3)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational (**CRITICAL - 阻塞所有故事**)
3. 完成 Phase 3: User Story 1 (初始化)
4. 完成 Phase 4: User Story 2 (服务启动)
5. 完成 Phase 5: User Story 3 (CRUD)
6. **STOP and VALIDATE**: 测试MVP功能
7. 可选部署/演示

### Incremental Delivery

1. Setup + Foundational → 基础就绪
2. Add US1 → 初始化可用 → 部署/演示
3. Add US2 → 服务可用 → 部署/演示
4. Add US3 → CRUD可用 → **MVP完成!**
5. Add US4-US5 → 可视化可用
6. Add US6-US8 → 协作功能可用

---

## Summary

| 统计项 | 数量 |
|--------|------|
| 总任务数 | 118 |
| Phase 1 (Setup) | 8 |
| Phase 2 (Foundational) | 21 |
| US1 (初始化) | 10 |
| US2 (服务启动) | 10 |
| US3 (CRUD) | 21 |
| US4 (依赖) | 8 |
| US5 (可视化) | 11 |
| US6 (锁定) | 7 |
| US7 (通知) | 9 |
| US8 (同步) | 5 |
| Phase 11 (Polish) | 8 |

**MVP范围**: Phase 1-5 (US1-US3), 共70个任务
