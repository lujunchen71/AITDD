# 前端架构文档

## 技术栈

| 类别 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 框架 | React | 18.x | UI框架 |
| 语言 | TypeScript | 5.x | 类型安全 |
| 构建工具 | Vite | 5.x | 快速构建 |
| UI组件库 | Ant Design | 5.x | 企业级组件 |
| 状态管理 | Zustand | 4.x | 轻量全局状态 |
| 服务端状态 | React Query | 5.x | API缓存和同步 |
| 可视化 | React Flow | 11.x | 任务依赖图 |
| 路由 | React Router | 6.x | 前端路由 |
| HTTP客户端 | Axios | 1.x | API请求 |
| 样式 | TailwindCSS | 3.x | 原子化CSS |
| 图标 | Lucide React | - | 图标库 |

## 项目结构

```
frontend/
├── public/                  # 静态资源
├── src/
│   ├── components/          # 通用组件
│   │   ├── common/          # 基础组件
│   │   │   ├── Button/
│   │   │   ├── Modal/
│   │   │   └── ...
│   │   ├── layout/          # 布局组件
│   │   │   ├── Header/
│   │   │   ├── Sidebar/
│   │   │   └── MainLayout/
│   │   └── feedback/        # 反馈组件
│   │       ├── NotificationDropdown/
│   │       └── Toast/
│   ├── features/            # 功能模块
│   │   ├── dashboard/       # 仪表盘
│   │   │   ├── components/
│   │   │   ├── hooks/
│   │   │   ├── types.ts
│   │   │   └── index.tsx
│   │   ├── modules/         # 模块管理
│   │   │   ├── components/
│   │   │   │   ├── ModuleTree/
│   │   │   │   ├── ModuleDetail/
│   │   │   │   └── ModuleForm/
│   │   │   ├── hooks/
│   │   │   ├── types.ts
│   │   │   └── index.tsx
│   │   ├── tasks/           # 任务管理
│   │   │   ├── components/
│   │   │   │   ├── TaskGraph/
│   │   │   │   ├── TaskDetailPanel/
│   │   │   │   ├── TaskForm/
│   │   │   │   └── TaskList/
│   │   │   ├── hooks/
│   │   │   ├── types.ts
│   │   │   └── index.tsx
│   │   ├── notifications/   # 通知中心
│   │   └── settings/        # 设置页面
│   ├── services/            # API服务
│   │   ├── api.ts           # Axios实例
│   │   ├── project.ts
│   │   ├── module.ts
│   │   ├── task.ts
│   │   ├── dependency.ts
│   │   ├── notification.ts
│   │   └── lock.ts
│   ├── stores/              # Zustand状态
│   │   ├── useUIStore.ts
│   │   ├── useModuleStore.ts
│   │   └── useTaskStore.ts
│   ├── hooks/               # 通用Hooks
│   │   ├── useWebSocket.ts
│   │   ├── useLocalStorage.ts
│   │   └── useDebounce.ts
│   ├── utils/               # 工具函数
│   │   ├── format.ts
│   │   ├── validation.ts
│   │   └── tree.ts
│   ├── types/               # 类型定义
│   │   ├── api.ts
│   │   ├── module.ts
│   │   ├── task.ts
│   │   └── common.ts
│   ├── styles/              # 全局样式
│   │   └── globals.css
│   ├── App.tsx              # 根组件
│   ├── main.tsx             # 入口文件
│   └── vite-env.d.ts
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## 核心组件设计

### 1. MainLayout（主布局）

```
┌─────────────────────────────────────────────────────┐
│Header                                    [通知] [设置]│
├────────────┬────────────────────────────────────────┤
│            │                                        │
│  Sidebar   │           Main Content                │
│  (模块树)   │                                        │
│            │                                        │
│            │                                        │
└────────────┴────────────────────────────────────────┘
```

**职责**:
- 管理全局布局
- 响应式适配
- 侧边栏折叠

### 2. ModuleTree（模块树组件）

**功能**:
- 递归渲染模块层级
- 支持展开/折叠
- 右键菜单（新建、删除、复制ID）
- 拖拽排序
- 状态图标显示

**状态图标**:
- 🟢 completed
- 🔵 developing
- 🟡 designing
- ⚫ deprecated
- 🔒 locked

**组件结构**:
```tsx
<ModuleTree>
  <ModuleTreeNode
    module={module}
    level={0}
    onSelect={handleSelect}
    onContextMenu={handleContextMenu}
  >
    {module.children.map(child => (
      <ModuleTreeNode ... />
    ))}
  </ModuleTreeNode>
</ModuleTree>
```

### 3. TaskGraphView（任务网络视图）

**功能**:
- 基于React Flow渲染有向图
- 节点表示任务，边表示依赖
- 节点颜色区分状态
- 支持拖拽创建依赖
- 缩放和平移
- 节点搜索和高亮

**节点类型**:
```tsx
const nodeTypes = {
  taskNode: TaskNode,
};

// TaskNode 自定义节点
const TaskNode: React.FC<NodeProps<TaskData>> = ({ data }) => {
  return (
    <div className={cn(
      "task-node",
      `task-node--${data.status}`,
      data.locked && "task-node--locked"
    )}>
      <div className="task-node__header">
        <span className="task-node__name">{data.name}</span>
        {data.locked && <LockIcon />}
      </div>
      <div className="task-node__status">{data.status}</div>
    </div>
  );
};
```

**边类型**:
```tsx
const edgeTypes = {
  dependency: DependencyEdge,
};

// DependencyEdge 自定义边
const DependencyEdge: React.FC<EdgeProps> = ({ sourceX, sourceY, targetX, targetY, data }) => {
  return (
    <>
      <path ... />
      <EdgeLabel>
        {data.contract?.interfaceVersion}
      </EdgeLabel>
    </>
  );
};
```

### 4. TaskDetailPanel（任务详情面板）

**功能**:
- 右侧滑出面板
- 分页签展示信息
- 支持编辑和保存
- 版本冲突处理

**页签结构**:
```tsx
<Tabs>
  <TabPane key="basic" tab="基本信息">
    <BasicInfo task={task} />
  </TabPane>
  <TabPane key="contract" tab="契约">
    <ContractInfo 
      upstream={task.upstreamContractDetail}
      downstream={task.downstreamContractDetail}
    />
  </TabPane>
  <TabPane key="tests" tab="测试">
    <TestList tests={task.tests} onRun={handleRunTest} />
  </TabPane>
  <TabPane key="logs" tab="日志">
    <LogViewer logs={task.logs} />
  </TabPane>
  <TabPane key="assistance" tab="人类协助">
    <HumanAssistance items={task.humanAssistance} />
  </TabPane>
</Tabs>
```

### 5. NotificationDropdown（通知下拉）

**功能**:
- 显示未读通知数量
- 下拉列表展示通知
- 点击跳转到相关任务
- 一键全部已读

**组件结构**:
```tsx
<Dropdown>
  <Badge count={unreadCount}>
    <BellIcon />
  </Badge>
  <DropdownMenu>
    {notifications.map(notif => (
      <NotificationItem
        key={notif.id}
        notification={notif}
        onClick={() => handleNavigate(notif)}
      />
    ))}
    <Button onClick={handleMarkAllRead}>全部已读</Button>
  </DropdownMenu>
</Dropdown>
```

## 状态管理

### Zustand Store

#### useUIStore
```typescript
interface UIState {
  sidebarCollapsed: boolean;
  selectedModuleId: string | null;
  selectedTaskId: string | null;
  detailPanelOpen: boolean;
  
  toggleSidebar: () => void;
  selectModule: (id: string | null) => void;
  selectTask: (id: string | null) => void;
  openDetailPanel: () => void;
  closeDetailPanel: () => void;
}
```

#### useModuleStore
```typescript
interface ModuleState {
  expandedKeys: string[];
  searchKeyword: string;
  
  setExpandedKeys: (keys: string[]) => void;
  toggleExpand: (key: string) => void;
  setSearchKeyword: (keyword: string) => void;
}
```

### React Query

#### 查询Hooks
```typescript
// 获取模块树
const useModules = () => {
  return useQuery({
    queryKey: ['modules'],
    queryFn: () => moduleApi.getAll(),
    staleTime: 30000,
  });
};

// 获取任务详情
const useTask = (id: string) => {
  return useQuery({
    queryKey: ['tasks', id],
    queryFn: () => taskApi.getById(id),
    enabled: !!id,
  });
};

// 获取模块中的任务
const useModuleTasks = (moduleId: string) => {
  return useQuery({
    queryKey: ['modules', moduleId, 'tasks'],
    queryFn: () => taskApi.getByModule(moduleId),
    enabled: !!moduleId,
  });
};
```

#### 变更Hooks
```typescript
// 创建任务
const useCreateTask = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: taskApi.create,
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries(['modules', variables.moduleId, 'tasks']);
      queryClient.invalidateQueries(['tasks']);
    },
  });
};

// 更新任务
const useUpdateTask = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTaskDto }) => 
      taskApi.update(id, data),
    onSuccess: (data) => {
      queryClient.invalidateQueries(['tasks', data.id]);
      queryClient.invalidateQueries(['tasks']);
    },
    onError: (error) => {
      if (error.code === 'VERSION_CONFLICT') {
        // 处理版本冲突
        showConflictModal(error.details);
      }
    },
  });
};
```

## 数据流

### 1. 初始化流程

```
AppMount
    │
    ▼
┌─────────────────┐
│  useModules()   │────► API: GET /api/v1/modules
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  setModulesData │────► Zustand Store
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  ModuleTree     │────► Render
└─────────────────┘
```

### 2. 任务选择流程

```
Click TaskNode
    │
    ▼
┌─────────────────┐
│ selectTask(id)  │────► Zustand Store
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  useTask(id)    │────► API: GET /api/v1/tasks/:id
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ openDetailPanel │────► Zustand Store
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ TaskDetailPanel │────► Render
└─────────────────┘
```

### 3. 实时更新流程

```
WebSocket Message
    │
    ▼
┌─────────────────┐
│ useWebSocket()  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────┐
│ queryClient.invalidateQueries│
└────────┬────────────────────┘
         │
         ▼
┌─────────────────┐
│ Auto Refetch    │────► UI Update
└─────────────────┘
```

## WebSocket集成

```typescript
// hooks/useWebSocket.ts
export const useWebSocket = () => {
  const queryClient = useQueryClient();
  const [connected, setConnected] = useState(false);
  
  useEffect(() => {
    const ws = new WebSocket('ws://localhost:34567/ws');
    
    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    
    ws.onmessage = (event) => {
      const { event: eventType, data } = JSON.parse(event.data);
      
      switch (eventType) {
        case 'task.updated':
          queryClient.invalidateQueries(['tasks', data.id]);
          break;
        case 'task.created':
        case 'task.deleted':
          queryClient.invalidateQueries(['tasks']);
          break;
        case 'notification':
          queryClient.invalidateQueries(['notifications']);
          // 显示toast通知
          toast.info(data.title);
          break;
      }
    };
    
    return () => ws.close();
  }, [queryClient]);
  
  return { connected };
};
```

## 性能优化

### 1. 虚拟列表
- 模块树节点过多时使用虚拟滚动
- 任务列表使用虚拟列表

### 2. 图渲染优化
- React Flow节点使用`memo`避免不必要的重渲染
- 大图（500+节点）使用聚类和分层渲染

### 3. 懒加载
- 任务详情面板懒加载
- 代码分割按路由

```typescript
// 路由懒加载
const Dashboard = lazy(() => import('./features/dashboard'));
const Modules = lazy(() => import('./features/modules'));
const Tasks = lazy(() => import('./features/tasks'));
```

### 4. 缓存策略
- React Query配置合理的`staleTime`
- 模块树数据缓存5分钟
- 任务详情缓存1分钟

## 响应式设计

### 断点

```css
/* tailwind.config.js */
module.exports = {
  theme: {
    screens: {
      'sm': '640px',
      'md': '768px',
      'lg': '1024px',
      'xl': '1280px',
      '2xl': '1536px',
    },
  },
}
```

### 移动端适配

- 侧边栏在移动端变为抽屉
- 任务详情面板在移动端变为全屏模态框
- 任务图在移动端支持手势缩放
