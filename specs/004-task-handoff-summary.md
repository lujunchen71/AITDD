# AITDD 系统重构 - 任务交接总结

**创建时间**: 2026-02-24  
**状态**: 第一阶段完成

---

## 一、上下文总结

### 1.1 项目定位
AITDD 是一个**提示词依赖管理系统**，通过可视化界面和 MCP 协议，让 AI 编程软件（KiloCode、OpenCode、ClaudeCode 等）能够协同工作，管理大型软件项目的模块、任务、依赖关系和提示词。

### 1.2 已完成功能

#### 前端组件 ✅
| 组件 | 文件路径 | 功能 |
|------|----------|------|
| ViewSwitcher | `frontend/src/features/modules/components/ViewSwitcher/index.tsx` | 视图切换器（列表/图表） |
| ModuleGraphView | `frontend/src/features/modules/components/ModuleGraphView/index.tsx` | 节点图表视图（React Flow） |
| ModuleTree | `frontend/src/features/modules/components/ModuleTree/index.tsx` | 模块树（支持展开/折叠） |
| ModuleDetail | `frontend/src/features/modules/components/ModuleDetail/index.tsx` | 模块详情（含任务列表） |
| DependencyGraph | `frontend/src/features/modules/components/DependencyGraph/index.tsx` | 依赖关系图 |
| modules/index.tsx | `frontend/src/features/modules/index.tsx` | 主页面集成 |

#### 后端 API ✅
| API | 文件路径 | 功能 |
|-----|----------|------|
| graph.go | `backend/internal/api/handlers/graph.go` | GetProjectGraph, GetModulePorts |
| prompt_version.go | `backend/internal/api/handlers/prompt_version.go` | 提示词版本管理 |
| routes.go | `backend/internal/api/routes.go` | 路由配置（已添加 graph 路由） |

#### 数据模型 ✅
| 模型 | 文件路径 | 功能 |
|------|----------|------|
| prompt_version.go | `backend/internal/models/prompt_version.go` | 提示词版本模型 |
| 003_prompt_versions.sql | `backend/migrations/003_prompt_versions.sql` | 数据库迁移 |

#### 类型定义 ✅
- `frontend/src/types/index.ts` - 添加了 ViewMode, ModuleNodeData, TaskNodeData, PortData 等类型

#### 架构文档 ✅
- `specs/003-visual-dependency-graph/architecture.md` - 可视化依赖图架构设计

### 1.3 已修复问题
1. **API 响应格式处理** - 兼容多种响应格式
2. **Empty 组件警告** - 使用新的 `styles` 属性
3. **模块列表显示** - 修复数据加载逻辑
4. **模块树展开/折叠** - 优化交互体验

### 1.4 Git 提交记录
```
feat: 完成可视化依赖图重构 (f6b6395)
fix: 修复 API 响应格式处理和 Empty 组件警告 (84cafac)
feat: 优化模块树组件 (962c9c5)
```

---

## 二、待实现功能

### 2.1 高优先级（P0）

#### 1. 模块内创建子模块功能
**需求**: 在模块树中右键菜单或按钮创建子模块
**实现位置**: 
- `frontend/src/features/modules/components/ModuleTree/index.tsx`
- 需要添加右键菜单组件

**步骤**:
1. 添加右键菜单组件（可使用 antd Dropdown）
2. 实现"创建子模块"菜单项
3. 调用现有的 ModuleForm 组件
4. 传递 parentId 参数

#### 2. 模块内创建任务功能
**需求**: 在模块详情中创建任务（已有按钮，需确认功能完整）
**实现位置**: 
- `frontend/src/features/modules/components/ModuleDetail/index.tsx`

**步骤**:
1. 确认"新建任务"按钮功能正常
2. 测试任务创建后列表刷新
3. 添加任务创建成功提示

#### 3. 任务间手动链接功能
**需求**: 拖拽创建任务依赖关系
**实现位置**: 
- `frontend/src/features/modules/components/ModuleGraphView/index.tsx`

**步骤**:
1. 启用 React Flow 的连线功能
2. 实现 onConnect 回调
3. 调用后端 API 创建依赖
4. 添加视觉反馈

### 2.2 中优先级（P1）

#### 4. 依赖关系自动聚合
**需求**: 任务依赖自动汇总到模块端口显示
**实现位置**: 
- `frontend/src/features/modules/components/ModuleGraphView/index.tsx`

**步骤**:
1. 实现 computeModulePorts 函数
2. 在模块节点左右侧显示端口
3. 端口与任务连线
4. 模块间依赖自动连线

#### 5. 节点面板右键创建任务
**需求**: 在图形视图右键菜单创建任务
**实现位置**: 
- `frontend/src/features/modules/components/ModuleGraphView/index.tsx`

**步骤**:
1. 添加右键菜单事件监听
2. 实现上下文菜单
3. 添加"创建任务"选项
4. 调用 TaskForm 组件

#### 6. 后端 graph API 编译
**需求**: 重新编译后端包含 graph API
**实现位置**: 
- 需要修复后端编译错误
- 重新构建 aitdd.exe

**步骤**:
1. 修复 `change_tracker.go` 编译错误
2. 修复 `sync_service.go` 编译错误
3. 运行 `go build` 编译
4. 测试 API 端点

### 2.3 低优先级（P2）

#### 7. 任务契约信息显示
**需求**: 在任务节点显示上下游契约
**实现位置**: 
- `frontend/src/features/modules/components/ModuleGraphView/index.tsx`

#### 8. 模块节点自定义样式
**需求**: 优化模块节点视觉设计
**实现位置**: 
- `frontend/src/features/modules/components/ModuleGraphView/index.tsx`

---

## 三、技术栈

### 前端
- **框架**: React 18 + TypeScript
- **UI 库**: Ant Design 5.x
- **图表**: React Flow 11.x
- **状态管理**: Zustand
- **数据请求**: Axios + TanStack Query

### 后端
- **语言**: Go 1.21+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **API 风格**: RESTful

---

## 四、开发环境

### 启动命令
```bash
# 前端（终端 1）
cd frontend && npm run dev

# 后端（终端 2）
cd backend && go run cmd/aitdd/main.go serve
# 或使用批处理
start-backend.bat
```

### 访问地址
- 前端：http://localhost:5173
- 后端 API：http://localhost:34567/api/v1

---

## 五、关键文件路径

### 核心组件
```
frontend/src/features/modules/
├── index.tsx                      # 主页面
└── components/
    ├── ViewSwitcher/              # 视图切换器
    ├── ModuleTree/                # 模块树
    ├── ModuleDetail/              # 模块详情
    ├── ModuleGraphView/           # 图表视图
    └── DependencyGraph/           # 依赖图
```

### 后端 API
```
backend/internal/api/handlers/
├── graph.go                       # 图表 API
├── module.go                      # 模块 API
├── task.go                        # 任务 API
└── prompt_version.go              # 提示词版本 API
```

### 类型定义
```
frontend/src/types/index.ts        # 所有类型定义
```

---

## 六、API 端点

### 已有端点
```
GET    /api/v1/modules?projectId=xxx     # 获取模块列表
GET    /api/v1/modules/:id               # 获取模块详情
GET    /api/v1/modules/:id/tasks         # 获取模块任务
GET    /api/v1/modules/:id/dependencies  # 获取模块依赖
POST   /api/v1/modules                   # 创建模块
POST   /api/v1/tasks                     # 创建任务
POST   /api/v1/modules/:id/dependencies  # 创建模块依赖
```

### 待实现端点
```
POST   /api/v1/tasks/:id/dependencies    # 创建任务依赖（已有 dependency API）
GET    /api/v1/graph/project             # 获取项目图表（已实现，需编译）
GET    /api/v1/graph/modules/:id/ports   # 获取模块端口（已实现，需编译）
```

---

## 七、注意事项

### 7.1 API 响应格式
所有 API 返回统一格式：
```typescript
{
  success: boolean;
  data?: {
    data?: T;  // 注意：有时数据在 data.data 中
    modules?: Module[];
    tasks?: Task[];
    total?: number;
  };
  timestamp?: number;
}
```

### 7.2 组件通信
- 使用 props 传递回调函数
- 复杂状态使用 Zustand store
- API 数据使用 TanStack Query 缓存

### 7.3 样式规范
- 使用内联样式（style 属性）
- 深色主题配色：
  - 背景：`#0f0f23`, `#1a1a2e`, `#16213e`
  - 强调色：`#e94560`, `#00d9ff`
  - 文字：`#ffffff`, `#e0e0e0`, `#a0a0a0`

---

## 八、下一步建议

### 立即可做
1. **测试现有功能** - 刷新浏览器，确认模块列表显示、创建模块、创建任务功能正常
2. **实现子模块创建** - 在 ModuleTree 添加右键菜单
3. **实现任务链接** - 在 ModuleGraphView 启用连线功能

### 后续优化
1. 修复后端编译错误
2. 实现依赖自动聚合
3. 优化视觉设计
4. 添加更多交互反馈

---

## 九、相关文件

- 架构设计：`specs/003-visual-dependency-graph/architecture.md`
- 重构计划：`specs/002-ai-integration-architecture/refactor-plan.md`
- API 文档：`backend/docs/api.md`

---

**祝开发顺利！** 🚀
