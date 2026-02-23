# Feature Specification: AITDD可视化任务治理系统

**Feature Branch**: `001-visual-task-governance`  
**Created**: 2026-02-23  
**Status**: Draft  
**Input**: AI辅助大型软件开发的可视化任务治理提示词系统

## 系统概述

本系统是一个可视化任务治理工具，旨在帮助开发团队在使用AI编程软件（如KiloCode、OpenCode、ClaudeCode等）时，以结构化的方式管理大型软件项目的模块、任务、依赖关系和验收证据。

**核心定位**：
- 系统本身**不具备AI计划能力**，而是作为**数据中枢**
- 通过**MCP（Model Context Protocol）接口**供第三方AI编程软件调用
- 提供**网页前端**供人工手动调整和监控

**核心特性**：
- **模块化**：支持无限层级模块嵌套，每个模块可包含多个任务
- **任务粒度**：每个任务对应一个或多个代码文件，但文件不可跨任务共享（强制依赖倒置和契约测试）
- **契约驱动**：每个任务明确上游/下游契约接口，确保独立开发
- **本地优先**：使用SQLite数据库，支持离线工作，并可配置远程同步
- **锁机制**：防止多AI或人工同时修改同一模块/任务
- **可视化**：网页前端展示模块树、任务依赖图，支持拖拽编辑

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 项目初始化与插件集成 (Priority: P1)

作为开发者，我需要在项目根目录执行初始化命令并选择AI编程插件类型，以便快速集成AITDD系统到我的开发工作流中。

**Why this priority**: 初始化是使用系统的第一步，没有初始化就无法使用其他功能。

**Independent Test**: 可以通过执行`aitdd init`命令，选择KiloCode插件，验证生成了正确的目录结构和workflow文件。

**Acceptance Scenarios**:

1. **Given** 用户在空项目目录中，**When** 用户执行`aitdd init`，**Then** 系统创建`.aitdd`目录并初始化SQLite数据库
2. **Given** 初始化过程中，**When** 系统提示选择AI插件类型，**Then** 显示KiloCode、OpenCode、ClaudeCode、Other选项
3. **Given** 用户选择KiloCode，**When** 初始化完成，**Then** 在`.kilocode/workflows/`目录生成aitdd.*.md命令文件
4. **Given** 初始化完成，**When** 用户查看生成的文件，**Then** 包含AITDD_GUIDE.md说明文档

---

### User Story 2 - 服务启动与MCP注册 (Priority: P1)

作为开发者，我需要启动本地服务并在AI编程软件中注册MCP，以便AI能够通过API管理任务。

**Why this priority**: MCP注册是AI与系统交互的基础，必须优先实现。

**Independent Test**: 可以通过执行`aitdd serve`启动服务，在浏览器访问localhost:34567验证前端页面加载。

**Acceptance Scenarios**:

1. **Given** 项目已初始化，**When** 用户执行`aitdd serve`，**Then** 系统启动HTTP服务（默认端口34567）并打开浏览器
2. **Given** 服务已启动，**When** 用户在AI编程软件中输入`/aitdd.start`，**Then** AI读取项目宪法并确认MCP连接就绪
3. **Given** MCP已注册，**When** AI调用`GET /api/project`，**Then** 返回项目基本信息
4. **Given** 服务运行中，**When** 用户按Ctrl+C，**Then** 服务优雅关闭

---

### User Story 3 - 模块与任务的增删改查 (Priority: P1)

作为AI助手，我需要通过API创建、读取、更新、删除模块和任务，以便管理整个项目的开发计划。

**Why this priority**: CRUD操作是系统的核心功能，所有其他功能都依赖于此。

**Independent Test**: 可以通过API创建一个模块，在该模块下创建任务，查询任务列表，更新任务状态，最后删除任务。

**Acceptance Scenarios**:

1. **Given** AI需要创建模块，**When** 调用`POST /api/modules`，**Then** 创建成功并返回模块ID
2. **Given** AI需要获取模块树，**When** 调用`GET /api/modules`，**Then** 返回嵌套的模块树结构
3. **Given** AI需要在模块下创建任务，**When** 调用`POST /api/tasks`，**Then** 任务创建成功并关联到指定模块
4. **Given** AI需要更新任务状态，**When** 调用`PUT /api/tasks/:id`，**Then** 任务状态更新（需版本号校验）
5. **Given** AI需要删除任务，**When** 调用`DELETE /api/tasks/:id`，**Then** 任务删除（需确认无依赖）

---

### User Story 4 - 任务依赖与契约管理 (Priority: P2)

作为AI助手，我需要定义任务之间的依赖关系和契约接口，以便确保任务可以独立开发。

**Why this priority**: 依赖关系是任务治理的核心，但需要基础CRUD功能先行。

**Independent Test**: 可以创建两个任务A和B，设置A依赖B，定义契约接口，验证依赖关系正确存储和查询。

**Acceptance Scenarios**:

1. **Given** AI需要定义依赖，**When** 调用`POST /api/dependencies`，**Then** 创建依赖关系并记录契约细节
2. **Given** AI尝试创建循环依赖，**When** 检测到循环，**Then** 系统拒绝操作并返回错误
3. **Given** 任务有上游契约，**When** 查询任务详情，**Then** 返回上游/下游契约的完整定义
4. **Given** 上游任务修改契约，**When** 保存修改，**Then** 通知所有下游任务

---

### User Story 5 - 可视化任务网络 (Priority: P2)

作为项目管理者，我需要在浏览器中查看任务依赖图，以便直观理解项目结构和进度。

**Why this priority**: 可视化是人工干预的主要入口，但依赖数据层。

**Independent Test**: 可以在浏览器中打开模块页面，查看任务网络图，验证节点和连线正确渲染。

**Acceptance Scenarios**:

1. **Given** 用户打开模块页面，**When** 模块包含任务，**Then** 以有向图展示任务和依赖关系
2. **Given** 用户查看任务图，**When** 悬停节点，**Then** 显示任务状态和简要信息
3. **Given** 用户点击任务节点，**When** 节点被选中，**Then** 右侧滑出详情面板
4. **Given** 用户拖拽节点创建依赖，**When** 连接两个节点，**Then** 创建新的依赖关系

---

### User Story 6 - 任务锁定与协作 (Priority: P3)

作为AI助手或开发者，我需要锁定正在工作的任务，以便防止并发冲突。

**Why this priority**: 锁机制对多人协作重要，但单用户场景下优先级较低。

**Independent Test**: 可以锁定一个任务，验证其他请求无法修改该任务，解锁后可以修改。

**Acceptance Scenarios**:

1. **Given** AI需要修改任务，**When** 调用`POST /api/lock`，**Then** 任务被锁定并设置过期时间
2. **Given** 任务已锁定，**When** 其他AI尝试修改，**Then** 返回409冲突错误
3. **Given** 锁已过期，**When** 超过过期时间，**Then** 锁自动释放
4. **Given** 用户手动解锁，**When** 调用`POST /api/unlock`，**Then** 锁立即释放

---

### User Story 7 - 通知系统 (Priority: P3)

作为AI助手，我需要接收和发送任务间通知，以便了解上游任务完成等事件。

**Why this priority**: 通知提升协作效率，但不是核心功能。

**Independent Test**: 可以从任务A发送通知到任务B，任务B查询未读通知并标记已读。

**Acceptance Scenarios**:

1. **Given** 任务A完成，**When** 调用`POST /api/notifications`，**Then** 发送通知给下游任务
2. **Given** 任务B有未读通知，**When** 调用`GET /api/notifications?read=false`，**Then** 返回未读通知列表
3. **Given** 用户查看通知，**When** 调用`POST /api/notifications/:id/read`，**Then** 通知标记为已读
4. **Given** 用户点击通知，**When** 通知关联任务，**Then** 跳转到对应任务详情

---

### User Story 8 - 数据同步与版本控制 (Priority: P3)

作为团队成员，我需要将本地数据同步到远程数据库，以便多人协作。

**Why this priority**: 分布式协作需要，但单用户可跳过。

**Independent Test**: 可以配置远程同步，修改本地数据，验证同步状态正确标记和上传。

**Acceptance Scenarios**:

1. **Given** 用户配置远程同步，**When** 保存配置，**Then** 连接远程数据库
2. **Given** 本地有修改，**When** 数据变更，**Then** 记录标记为PENDING_UPLOAD
3. **Given** 触发同步，**When** 上传成功，**Then** 标记更新为SYNCED
4. **Given** 发生冲突，**When** 远程有更新，**Then** 标记为CONFLICT并提示用户

---

### Edge Cases

- 当模块嵌套层级超过10层时，如何保证树形渲染性能？
- 当任务依赖链超过50个节点时，如何优化图渲染？
- 当SQLite数据库文件损坏时，如何恢复数据？
- 当多个AI同时锁定同一任务时，如何处理竞争？
- 当远程同步中断时，如何保证数据不丢失？

## Requirements *(mandatory)*

### 功能需求：给AI的MCP工具

| ID | 功能 | 描述 |
|----|------|------|
| MCP-01 | 获取项目简述 | `GET /api/project` 返回项目基本信息 |
| MCP-02 | 获取项目公约 | `GET /api/project/constitution` 返回宪法文本 |
| MCP-03 | 获取所有模块 | `GET /api/modules` 返回模块树 |
| MCP-04 | 获取模块概览 | `GET /api/modules/:id` 返回模块详情（提示词、状态、测试覆盖率、契约概览） |
| MCP-05 | 获取模块任务ID列表 | `GET /api/modules/:id/tasks?fields=id` |
| MCP-06 | 获取/修改任务详情 | `GET/PUT /api/tasks/:id`（状态、契约、提示词、测试、日志、代码路径、人类协助） |
| MCP-07 | 创建任务 | `POST /api/tasks` 在指定模块下创建 |
| MCP-08 | 删除任务 | `DELETE /api/tasks/:id` |
| MCP-09 | 删除模块 | `DELETE /api/modules/:id` |
| MCP-10 | 打开前端 | 触发浏览器打开`http://localhost:34567` |
| MCP-11 | 发送通知 | `POST /api/notifications` |
| MCP-12 | 读取通知 | `GET /api/notifications?read=false` |
| MCP-13 | 锁定/解锁 | `POST /api/lock`, `POST /api/unlock`（预留） |
| MCP-14 | 查询锁定状态 | `GET /api/lock/status`（预留） |

### 功能需求：命令行命令

| 命令 | 功能 |
|------|------|
| `/aitdd.start` | 显示系统指南，注册MCP |
| `/aitdd.constitution` | 制定或更新项目管理原则 |
| `/aitdd.specify` | 明确需求和用户故事 |
| `/aitdd.plan` | 创建技术实施计划 |
| `/aitdd.tasks` | 生成或修改任务清单 |
| `/aitdd.implement` | 按计划执行任务 |
| `/aitdd.debug` | 查询bug、修复bug、分析链路完整性 |

### 功能需求：安装流程

| ID | 需求 |
|----|------|
| INST-01 | 提供跨平台可执行文件（Windows/Linux/macOS） |
| INST-02 | `aitdd init` 创建`.aitdd`目录和初始化数据库 |
| INST-03 | 支持选择AI插件类型并生成对应workflow文件 |
| INST-04 | `aitdd serve` 启动HTTP服务和前端页面 |

### Key Entities

- **Project（项目）**: 项目容器，包含宪法、配置
- **Module（模块）**: 无限层级嵌套的组织单元
- **Task（任务）**: 最小工作单元，关联代码文件
- **Dependency（依赖）**: 任务间依赖关系和契约
- **Notification（通知）**: 任务间消息传递
- **ChangeHistory（变更历史）**: 审计和同步冲突解决

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: API响应时间P95 < 100ms（本地SQLite）
- **SC-002**: 前端首次加载时间 < 2秒
- **SC-003**: 任务网络图支持500+节点流畅渲染（60fps）
- **SC-004**: 初始化命令执行时间 < 5秒
- **SC-005**: AI通过MCP完成一次CRUD操作 < 200ms
- **SC-006**: 支持10个并发AI连接无阻塞
- **SC-007**: 数据同步延迟 < 5秒（配置远程时）
- **SC-008**: 90%的用户能在5分钟内完成初始化并运行第一个命令

## Assumptions

- 服务只监听localhost，无需认证（可配置简单令牌）
- SQLite数据库适合单用户本地使用
- 远程同步为可选功能，默认关闭
- AI编程软件支持自定义workflow文件格式
- 用户使用现代浏览器

## Out of Scope

- 内置AI计划能力
- 云端托管服务
- 移动端应用
- 复杂的权限管理系统
- 实时协同编辑（类似Google Docs）
