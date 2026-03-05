# 属性信息面板（PropertyPanel）实施方案

> 创建时间：2026-03-05
> 状态：规划中

---

## 一、前端代码结构现状分析

### 1.1 技术栈与依赖

- **框架**：React 18 + TypeScript
- **状态管理**：Zustand（`useProjectStore`、`useUIStore`）
- **UI 库**：Ant Design
- **图形视图**：React Flow（`reactflow`）
- **数据请求**：Axios + `@tanstack/react-query`
- **本地持久化**：`localStorageService`（自定义封装）

### 1.2 文件目录结构

```
frontend/src/
├── App.tsx                          # 路由配置，懒加载各页面
├── stores/
│   ├── useProjectStore.ts           # 项目/多项目状态
│   └── useUIStore.ts                # UI 状态（选中模块、任务面板等）
├── services/
│   ├── api.ts                       # Axios 封装 + 模块依赖/位置 API
│   └── localStorageService.ts       # 本地存储（DisplaySettings等）
├── types/
│   ├── index.ts                     # Project/Module/Task/Dependency 等核心类型
│   └── api.ts                       # API 请求/响应类型
├── features/
│   └── modules/
│       ├── index.tsx                # 模块页面（列表视图 + 图形视图切换）
│       └── components/
│           ├── ModuleGraphView/
│           │   └── index.tsx        # 核心节点图视图（ReactFlow）
│           ├── ModuleDetail/
│           │   └── index.tsx        # 模块详情面板（列表视图用）
│           └── ...
└── features/tasks/components/
    └── TaskDetailPanel/
        └── index.tsx                # 任务详情抽屉（现有的 Drawer 组件）
```

### 1.3 节点视图（ModuleGraphView）关键机制

#### 节点选中状态管理
- `selectedModuleIds: Set<string>` — 多选的模块 ID 集合（本地 state）
- `onNodeClick` — 普通单选；Shift 加选；Ctrl/Meta 减选
- `onSelectionChange` — 框选完成后的回调
- `onPaneClick` — 点击空白处清空选中

#### 显示设置面板
- 当前面板位于 ReactFlow 的 `Panel position="top-right"`
- 通过 Ant Design `Popover` 包裹一个"显示设置"按钮
- `DisplaySettings` 存储在 `localStorageService` 中

#### 任务节点按钮（InfoBadge）
- 每个任务节点有多个提示标识按钮：📝提示词、🧪测试用例、❌错误、🔔通知
- 目前通过 `Tooltip` 鼠标悬停/按 Alt 键显示弹出内容
- 将改为点击后触发 PropertyPanel 显示对应详情

### 1.4 现有状态管理

| Store/Service | 关键状态 | 备注 |
|---|---|---|
| `useProjectStore` | `project`、`projects`、`selectedProjectIds` | 项目级别 |
| `useUIStore` | `selectedModuleId`、`selectedTaskId`、`taskDetailPanelVisible` | UI 级别 |
| `localStorageService` | `DisplaySettings`、`selectedModules` | 持久化 |
| `ModuleGraphView` 内部 | `selectedModuleIds: Set<string>` | 图形视图内部 |

**发现**：`useUIStore` 已有 `selectedModuleId` 和 `selectedTaskId`，但 `ModuleGraphView` 内部用的是自己的 `selectedModuleIds: Set<string>`，两者存在割裂。

### 1.5 类型定义摘要

**Project**（`types/index.ts`）：
```ts
id, name, constitution, createdAt, updatedAt, version, syncStatus
```

**Module**（`types/index.ts`）：
```ts
id, parentId, projectId, name, description, prompt, status, testCoverage,
upstreamContractSummary, downstreamContractSummary, locked, ...
```

**Task**（`types/index.ts`）：
```ts
id, moduleId, name, description, status, assignee,
upstreamContractDetail (JSON string),
downstreamContractDetail (JSON string),
prompt, tests (JSON string), testResult, bugLog, issueDetails,
codePaths, humanAssistance, locked, ...
```

**TaskDependency**（`types/index.ts`）：
```ts
id, upstreamTaskId, downstreamTaskId, contractSummary, status, ...
```

---

## 二、需要创建的新文件

### 2.1 核心组件文件

| 文件路径 | 职责 |
|---|---|
| `frontend/src/features/modules/components/PropertyPanel/index.tsx` | PropertyPanel 主容器组件 |
| `frontend/src/features/modules/components/PropertyPanel/views/ProjectView.tsx` | 项目信息视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/ModuleView.tsx` | 模块信息视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/TaskView.tsx` | 任务基本信息视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/PromptView.tsx` | 提示词详情视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/ContractView.tsx` | 上下契约详情视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/ErrorView.tsx` | 错误警告详情视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/TestsView.tsx` | 测试用例详情视图 |
| `frontend/src/features/modules/components/PropertyPanel/views/DependenciesView.tsx` | 任务依赖关系视图 |
| `frontend/src/features/modules/components/PropertyPanel/index.css` | 面板样式 |

### 2.2 Store 扩展（或新建）

| 文件路径 | 职责 |
|---|---|
| `frontend/src/stores/usePropertyPanelStore.ts` | PropertyPanel 专用状态：显示/隐藏、当前内容类型、选中对象 |

---

## 三、需要修改的现有文件

| 文件 | 改动点 |
|---|---|
| `frontend/src/features/modules/components/ModuleGraphView/index.tsx` | 1. 在 `Panel position="top-right"` 前添加"属性面板"Switch 开关 2. 在空白点击/节点点击时同步更新 `usePropertyPanelStore` 3. InfoBadge 点击事件改为触发面板显示对应内容 4. 添加 `P` 键快捷键监听 5. 渲染 `PropertyPanel` 组件（悬浮层） |
| `frontend/src/stores/useUIStore.ts` | 可选：将 PropertyPanel 的 visible 状态移到此 Store 统一管理 |
| `frontend/src/services/localStorageService.ts` | 添加 `propertyPanelVisible` 的持久化键（可选，默认不持久化） |

---

## 四、PropertyPanel 数据流设计

### 4.1 PropertyPanel Store 设计

```typescript
// frontend/src/stores/usePropertyPanelStore.ts

// 面板内容类型
export type PanelContentType =
  | 'project'           // 未选中任何对象时显示项目信息
  | 'module'            // 选中模块节点
  | 'task'              // 选中任务节点（基本信息）
  | 'prompt'            // 点击提示词按钮
  | 'upstream-contract' // 点击上游契约按钮
  | 'downstream-contract' // 点击下游契约按钮
  | 'error'             // 点击错误按钮
  | 'tests'             // 点击测试用例按钮
  | 'dependencies';     // 点击任务名称

// Store 状态
interface PropertyPanelState {
  // 面板可见性
  visible: boolean;
  toggleVisible: () => void;
  setVisible: (visible: boolean) => void;

  // 当前内容类型
  contentType: PanelContentType;
  setContentType: (type: PanelContentType) => void;

  // 选中的对象 ID
  selectedModuleId: string | null;
  selectedTaskId: string | null;
  setSelectedModule: (moduleId: string | null) => void;
  setSelectedTask: (taskId: string | null) => void;

  // 复合操作：选中模块
  showModuleInfo: (moduleId: string) => void;
  // 复合操作：选中任务
  showTaskInfo: (taskId: string) => void;
  // 复合操作：显示特定内容
  showContent: (type: PanelContentType, taskId?: string, moduleId?: string) => void;
  // 复合操作：无选中时显示项目信息
  showProjectInfo: () => void;
}
```

### 4.2 数据流图

```
用户操作
    │
    ├─── 点击模块节点（onNodeClick）
    │         → ModuleGraphView: setSelectedModuleIds(new Set([moduleId]))
    │         → usePropertyPanelStore.showModuleInfo(moduleId)
    │
    ├─── 点击任务节点（onTaskClick）
    │         → usePropertyPanelStore.showTaskInfo(taskId)
    │
    ├─── 点击 InfoBadge（📝提示词）
    │         → usePropertyPanelStore.showContent('prompt', taskId)
    │
    ├─── 点击 InfoBadge（🧪测试用例）
    │         → usePropertyPanelStore.showContent('tests', taskId)
    │
    ├─── 点击 InfoBadge（❌错误）
    │         → usePropertyPanelStore.showContent('error', taskId)
    │
    ├─── 点击 InfoBadge（上游契约）
    │         → usePropertyPanelStore.showContent('upstream-contract', taskId)
    │
    ├─── 点击任务名称文字
    │         → usePropertyPanelStore.showContent('dependencies', taskId)
    │
    ├─── 点击空白处（onPaneClick）
    │         → usePropertyPanelStore.showProjectInfo()
    │
    └─── P 键 / Switch 开关
              → usePropertyPanelStore.toggleVisible()

usePropertyPanelStore (状态)
    │
    ↓
PropertyPanel 组件
    │
    ├─── contentType === 'project'     → <ProjectView>
    ├─── contentType === 'module'      → <ModuleView moduleId={selectedModuleId}>
    ├─── contentType === 'task'        → <TaskView taskId={selectedTaskId}>
    ├─── contentType === 'prompt'      → <PromptView taskId={selectedTaskId}>
    ├─── contentType === 'upstream-contract' → <ContractView taskId type="upstream">
    ├─── contentType === 'downstream-contract' → <ContractView taskId type="downstream">
    ├─── contentType === 'error'       → <ErrorView taskId={selectedTaskId}>
    ├─── contentType === 'tests'       → <TestsView taskId={selectedTaskId}>
    └─── contentType === 'dependencies' → <DependenciesView taskId={selectedTaskId}>
```

### 4.3 Props 设计

```typescript
// PropertyPanel 主容器
interface PropertyPanelProps {
  // 提供模块、任务原始数据（已在 ModuleGraphView 中加载）
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
  project: Project | null;
}
```

各子视图通过 `taskId` / `moduleId` 直接从 props.modules/tasks 中查找数据，**不需要额外 API 调用**（数据已在父层加载），但 `DependenciesView` 需要调用 `/tasks/{id}/dependencies` API（获取上下游依赖关系）。

---

## 五、各信息视图 UI 结构设计

### 5.1 PropertyPanel 主容器

```jsx
// 定位：fixed 定位，悬浮在节点网络面板右侧
// 宽度：500px
// 顶部：与"显示设置"按钮对齐（top: 同 Panel position="top-right" 的位置）
// z-index 高于 ReactFlow 画布

<div className="property-panel" style={{
  position: 'fixed',
  right: 16,
  top: 64,  // 与显示设置按钮对齐
  width: 500,
  maxHeight: 'calc(100vh - 80px)',
  background: '#1a1a2e',
  border: '1px solid #2d2d44',
  borderRadius: 8,
  boxShadow: '0 4px 20px rgba(0,0,0,0.5)',
  zIndex: 10,
  display: 'flex',
  flexDirection: 'column',
  overflow: 'hidden',
}}>
  {/* 面板标题栏 */}
  <div className="panel-header">
    <span className="panel-title">{标题文字}</span>
    <Tag color="...">{内容类型标签}</Tag>
  </div>

  {/* 可滚动内容区 */}
  <div className="panel-body" style={{ flex: 1, overflowY: 'auto' }}>
    {/* 根据 contentType 渲染对应视图 */}
    {renderContent()}
  </div>
</div>
```

### 5.2 ProjectView（项目信息）

```jsx
<div>
  <Descriptions column={1} size="small" title="项目概览">
    <Descriptions.Item label="项目名称">{project.name}</Descriptions.Item>
    <Descriptions.Item label="创建时间">{formatDate(project.createdAt)}</Descriptions.Item>
    <Descriptions.Item label="更新时间">{formatDate(project.updatedAt)}</Descriptions.Item>
  </Descriptions>

  <Divider />

  {/* 统计信息 */}
  <Row gutter={16}>
    <Col span={8}><Statistic title="模块数" value={modules.length} /></Col>
    <Col span={8}><Statistic title="任务数" value={tasks.length} /></Col>
    <Col span={8}><Statistic title="依赖数" value={taskDependencies.length} /></Col>
  </Row>

  <Divider />

  {/* 任务状态分布 */}
  <div>
    <Typography.Text>任务状态分布</Typography.Text>
    {/* 各状态 Tag + 数量 */}
    {statusGroups.map(g => <Tag ...>{g.status}: {g.count}</Tag>)}
  </div>

  {/* 项目章程/Constitution（可折叠） */}
  {project.constitution && (
    <Collapse>
      <Panel header="项目章程">
        <pre style={{ whiteSpace: 'pre-wrap' }}>{project.constitution}</pre>
      </Panel>
    </Collapse>
  )}
</div>
```

### 5.3 ModuleView（模块信息）

```jsx
<div>
  <Descriptions column={1} size="small">
    <Descriptions.Item label="模块名称">{module.name}</Descriptions.Item>
    <Descriptions.Item label="状态">
      <Tag color={statusColor[module.status]}>{statusLabel[module.status]}</Tag>
    </Descriptions.Item>
    <Descriptions.Item label="描述">{module.description || '—'}</Descriptions.Item>
    <Descriptions.Item label="测试覆盖率">{module.testCoverage}%</Descriptions.Item>
    <Descriptions.Item label="锁定状态">
      {module.locked ? <Tag color="red">已锁定 by {module.lockedBy}</Tag> : <Tag color="green">未锁定</Tag>}
    </Descriptions.Item>
  </Descriptions>

  <Divider />

  {/* 上游契约摘要 */}
  {module.upstreamContractSummary && (
    <Card size="small" title="上游契约摘要">
      <Typography.Text>{module.upstreamContractSummary}</Typography.Text>
    </Card>
  )}

  {/* 下游契约摘要 */}
  {module.downstreamContractSummary && (
    <Card size="small" title="下游契约摘要" style={{ marginTop: 8 }}>
      <Typography.Text>{module.downstreamContractSummary}</Typography.Text>
    </Card>
  )}

  <Divider />

  {/* 任务列表（可滚动） */}
  <Typography.Text strong>任务列表（{moduleTasks.length}）</Typography.Text>
  <List
    size="small"
    dataSource={moduleTasks}
    renderItem={task => (
      <List.Item
        extra={<Tag color={statusColor[task.status]}>{statusLabel[task.status]}</Tag>}
        onClick={() => showTaskInfo(task.id)}  // 点击跳转任务视图
        style={{ cursor: 'pointer' }}
      >
        {task.name}
      </List.Item>
    )}
  />
</div>
```

### 5.4 TaskView（任务基本信息）

```jsx
<div>
  <Descriptions column={1} size="small">
    <Descriptions.Item label="任务名称">{task.name}</Descriptions.Item>
    <Descriptions.Item label="所属模块">{moduleName}</Descriptions.Item>
    <Descriptions.Item label="状态">
      <Tag color={statusColor[task.status]}>{statusLabel[task.status]}</Tag>
    </Descriptions.Item>
    <Descriptions.Item label="负责人">{task.assignee || '未分配'}</Descriptions.Item>
    <Descriptions.Item label="描述">{task.description || '—'}</Descriptions.Item>
    <Descriptions.Item label="锁定状态">
      {task.locked ? <Tag color="red">已锁定 by {task.lockedBy}</Tag> : <Tag color="green">未锁定</Tag>}
    </Descriptions.Item>
  </Descriptions>

  <Divider />

  {/* 快捷跳转按钮 */}
  <Space wrap>
    {task.prompt && <Button size="small" onClick={() => showContent('prompt', task.id)}>查看提示词</Button>}
    {task.tests && <Button size="small" onClick={() => showContent('tests', task.id)}>查看测试用例</Button>}
    {task.upstreamContractDetail && <Button size="small" onClick={() => showContent('upstream-contract', task.id)}>上游契约</Button>}
    {task.downstreamContractDetail && <Button size="small" onClick={() => showContent('downstream-contract', task.id)}>下游契约</Button>}
    {task.bugLog && <Button size="small" danger onClick={() => showContent('error', task.id)}>错误信息</Button>}
    <Button size="small" onClick={() => showContent('dependencies', task.id)}>依赖关系</Button>
  </Space>

  {/* 代码文件路径 */}
  {task.codePaths && (
    <div style={{ marginTop: 16 }}>
      <Typography.Text strong>代码文件</Typography.Text>
      <List
        size="small"
        dataSource={parseCodePaths(task.codePaths)}
        renderItem={p => <List.Item><Typography.Text code>{p}</Typography.Text></List.Item>}
      />
    </div>
  )}
</div>
```

### 5.5 PromptView（提示词详情）

```jsx
<div>
  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
    <Typography.Text strong>提示词内容</Typography.Text>
    <Button size="small" icon={<CopyOutlined />} onClick={() => copyToClipboard(task.prompt)}>复制</Button>
  </div>
  <div style={{
    background: '#0f0f23',
    border: '1px solid #3d3d5c',
    borderRadius: 4,
    padding: 12,
    maxHeight: 400,
    overflowY: 'auto',
    fontFamily: 'monospace',
    fontSize: 12,
    whiteSpace: 'pre-wrap',
    color: '#e0e0e0',
    lineHeight: 1.6,
  }}>
    {task.prompt || '（无提示词）'}
  </div>
  <Typography.Text type="secondary" style={{ fontSize: 11, marginTop: 4 }}>
    字符数：{task.prompt?.length || 0}
  </Typography.Text>
</div>
```

### 5.6 ContractView（上下契约详情）

```jsx
// type: 'upstream' | 'downstream'
<div>
  {!contractDetail ? (
    <Empty description="无契约信息" />
  ) : (
    <>
      <Typography.Text strong>{contractDetail.title}</Typography.Text>
      <Divider style={{ margin: '8px 0' }} />
      <List
        size="small"
        dataSource={contractDetail.list}
        renderItem={(item) => (
          <List.Item>
            <List.Item.Meta
              title={<Typography.Text strong>{item.label}</Typography.Text>}
              description={
                <>
                  <Typography.Text code style={{ fontSize: 11, wordBreak: 'break-all' }}>
                    {item.contract_api}
                  </Typography.Text>
                  {item.from && (
                    <div>
                      <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                        来自：{item.from}
                      </Typography.Text>
                    </div>
                  )}
                </>
              }
            />
          </List.Item>
        )}
      />
    </>
  )}
</div>
```

### 5.7 ErrorView（错误警告详情）

```jsx
<div>
  {!bugLog || bugLog.length === 0 ? (
    <Empty description="无错误信息" />
  ) : (
    <List
      size="small"
      dataSource={bugLog}
      renderItem={(entry) => (
        <List.Item>
          <Card
            size="small"
            style={{
              width: '100%',
              borderColor: entry.level === 'error' ? '#ff4d4f' : entry.level === 'warn' ? '#faad14' : '#1890ff',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <Tag color={levelColors[entry.level]}>{entry.level.toUpperCase()}</Tag>
              <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                {formatTimestamp(entry.timestamp)}
              </Typography.Text>
            </div>
            <Typography.Text style={{ whiteSpace: 'pre-wrap', fontSize: 12 }}>
              {entry.message}
            </Typography.Text>
          </Card>
        </List.Item>
      )}
    />
  )}

  {/* issueDetails */}
  {task.issueDetails && (
    <Collapse style={{ marginTop: 8 }}>
      <Panel header="AI 发现的问题详情">
        <Typography.Text style={{ whiteSpace: 'pre-wrap', fontSize: 12 }}>
          {task.issueDetails}
        </Typography.Text>
      </Panel>
    </Collapse>
  )}
</div>
```

### 5.8 TestsView（测试用例详情）

```jsx
<div>
  {tests.length === 0 ? (
    <Empty description="无测试用例" />
  ) : (
    <Table
      size="small"
      dataSource={tests}
      columns={[
        { title: '测试目标', dataIndex: 'target', key: 'target' },
        {
          title: '测试 API',
          dataIndex: 'api',
          key: 'api',
          render: (text) => <Typography.Text code style={{ fontSize: 11 }}>{text}</Typography.Text>
        },
      ]}
      pagination={false}
    />
  )}

  {/* 测试结果（testResult） */}
  {testResult.length > 0 && (
    <div style={{ marginTop: 16 }}>
      <Typography.Text strong>测试结果记录</Typography.Text>
      <List
        size="small"
        dataSource={testResult}
        renderItem={(r, i) => (
          <List.Item>
            <Typography.Text style={{ fontSize: 12 }}>{i + 1}. {r}</Typography.Text>
          </List.Item>
        )}
      />
    </div>
  )}
</div>
```

### 5.9 DependenciesView（任务依赖关系）

```jsx
// 需要调用 API：GET /tasks/{taskId}/dependencies?direction=both
// 或使用已有的 taskDependencies 数组过滤

<div>
  {/* 上游依赖 */}
  <Typography.Text strong>上游依赖（{upstreamDeps.length}）</Typography.Text>
  <List
    size="small"
    dataSource={upstreamDeps}
    renderItem={(dep) => {
      const upstreamTask = tasks.find(t => t.id === dep.upstreamTaskId);
      return (
        <List.Item>
          <div style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography.Text>
                ↑ {upstreamTask?.name || dep.upstreamTaskId}
              </Typography.Text>
              <Tag color={dep.status === 'active' ? 'green' : 'default'}>{dep.status}</Tag>
            </div>
            {dep.contractSummary && (
              <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                契约：{dep.contractSummary}
              </Typography.Text>
            )}
          </div>
        </List.Item>
      );
    }}
  />

  <Divider />

  {/* 下游依赖 */}
  <Typography.Text strong>下游依赖（{downstreamDeps.length}）</Typography.Text>
  <List
    size="small"
    dataSource={downstreamDeps}
    renderItem={(dep) => {
      const downstreamTask = tasks.find(t => t.id === dep.downstreamTaskId);
      return (
        <List.Item>
          <div>
            <Typography.Text>↓ {downstreamTask?.name || dep.downstreamTaskId}</Typography.Text>
          </div>
        </List.Item>
      );
    }}
  />
</div>
```

---

## 六、键盘快捷键和开关按钮的实现方式

### 6.1 P 键快捷键

在 `ModuleGraphView` 中添加 `useEffect` 监听键盘事件：

```typescript
useEffect(() => {
  const handleKeyDown = (e: KeyboardEvent) => {
    // 排除在输入框内的情况
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
    
    if (e.key === 'p' || e.key === 'P') {
      usePropertyPanelStore.getState().toggleVisible();
    }
  };
  
  window.addEventListener('keydown', handleKeyDown);
  return () => window.removeEventListener('keydown', handleKeyDown);
}, []);
```

### 6.2 Switch 开关按钮

在 `ModuleGraphView` 的 `Panel position="top-right"` 中，在"显示设置"按钮**左边**添加：

```jsx
<Panel position="top-right" style={{ background: 'transparent' }}>
  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
    {/* 新增：属性面板开关 */}
    <div style={{
      background: '#1a1a2e',
      border: '1px solid #3d3d5c',
      borderRadius: 6,
      padding: '6px 12px',
      display: 'flex',
      alignItems: 'center',
      gap: 8,
    }}>
      <UnorderedListOutlined style={{ color: '#a0a0a0' }} />
      <span style={{ color: '#a0a0a0', fontSize: 13 }}>属性面板</span>
      <Switch
        size="small"
        checked={propertyPanelVisible}
        onChange={(checked) => setPropertyPanelVisible(checked)}
      />
    </div>

    {/* 现有：显示设置 Popover */}
    <Popover content={settingsContent} ...>
      <Button type="primary" icon={<SettingOutlined />} ...>
        显示设置
      </Button>
    </Popover>
  </div>
</Panel>
```

---

## 七、与现有节点选中事件的集成方式

### 7.1 集成点总览

```
ModuleGraphView (index.tsx)
    │
    ├── onNodeClick(event, node)
    │     if node.id.startsWith('module-'):
    │       普通点击 → setSelectedModuleIds + usePropertyPanelStore.showModuleInfo(moduleId)
    │       Shift/Ctrl 点击 → 只更新 selectedModuleIds，不触发 PropertyPanel
    │     if node.id.startsWith('task-'):
    │       → usePropertyPanelStore.showTaskInfo(taskId)
    │
    ├── onPaneClick()
    │     → setSelectedModuleIds(new Set())
    │     → usePropertyPanelStore.showProjectInfo()
    │
    └── TaskNodeComponent 内的 InfoBadge 点击
          原来：显示 Tooltip
          改为：stopPropagation() + usePropertyPanelStore.showContent(type, taskId)
```

### 7.2 任务节点点击处理

当前 `onTaskClick` 回调在 `ModuleGraphView` 中传入 `ModuleNodeComponent`，再传给 `TaskItemComponent`。需要扩展这个接口：

```typescript
// 扩展 ModuleNodeData 接口
interface ModuleNodeData {
  // 现有字段...
  onTaskClick?: (taskId: string) => void;
  // 新增：各类 InfoBadge 点击
  onTaskInfoClick?: (taskId: string, type: PanelContentType) => void;
}

// 在 ModuleGraphView 构建节点时
data: {
  ...
  onTaskClick: (taskId) => {
    usePropertyPanelStore.getState().showTaskInfo(taskId);
    onTaskClick?.(taskId);  // 保持原有回调
  },
  onTaskInfoClick: (taskId, type) => {
    usePropertyPanelStore.getState().showContent(type, taskId);
  },
}
```

### 7.3 PropertyPanel 渲染位置

`PropertyPanel` 组件渲染在 `ModuleGraphView` 的 `return` JSX 中，**在 `<div style={{ width: '100%', height: '100%', ... }}>` 内，与 ReactFlow 同层**，通过 `position: fixed` 或 `absolute` 悬浮：

```jsx
return (
  <div style={{ width: '100%', height: '100%', background: '#0f0f23', position: 'relative' }}>
    <ReactFlow ...>
      {/* ... */}
    </ReactFlow>

    {/* PropertyPanel 悬浮层 */}
    {propertyPanelVisible && (
      <PropertyPanel
        modules={modules}
        tasks={tasks}
        taskDependencies={taskDependencies}
        project={project}
      />
    )}
  </div>
);
```

**方案选择**：使用 `position: fixed` 比 `absolute` 更稳定，不受 ReactFlow 内部 transform 影响。

---

## 八、实现步骤（推荐顺序）

1. **创建 `usePropertyPanelStore.ts`**
   - 定义 `PanelContentType` 类型
   - 实现 visible、contentType、selectedModuleId、selectedTaskId 状态
   - 实现 showModuleInfo / showTaskInfo / showContent / showProjectInfo 等方法

2. **创建 `PropertyPanel/index.tsx`**
   - 实现容器布局（fixed 定位、500px 宽、标题栏、滚动区）
   - 实现 contentType 路由渲染逻辑

3. **实现各子视图（views/）**
   - 先实现 `TaskView` 和 `ModuleView`（最常用）
   - 再实现 `PromptView`、`ContractView`、`TestsView`、`ErrorView`
   - 最后实现 `DependenciesView`（需要 API 调用）
   - 最后实现 `ProjectView`

4. **修改 `ModuleGraphView/index.tsx`**
   - 导入并使用 `usePropertyPanelStore`
   - 在 `onNodeClick`、`onPaneClick` 中同步更新 Store
   - 在 `Panel position="top-right"` 中添加 Switch 开关
   - 添加 P 键监听
   - 将 InfoBadge 的 Tooltip 点击改为触发面板
   - 渲染 `<PropertyPanel>` 组件

5. **调整 `ModuleGraphView` 的数据传递**
   - 将 `project` 数据从 `useProjectStore` 注入
   - 确保 `modules`、`tasks`、`taskDependencies` 传入 `PropertyPanel`

6. **样式与细节优化**
   - 面板动画（slide-in / fade）
   - 响应式：宽屏时 500px，窄屏时 400px
   - 深色主题适配（与现有 `#0f0f23`、`#1a1a2e` 保持一致）

---

## 九、技术决策说明

| 决策项 | 选择 | 理由 |
|---|---|---|
| 状态管理 | 新建 `usePropertyPanelStore`（Zustand） | 与现有 Store 解耦，专注 Panel 逻辑；避免污染 `useUIStore` |
| 面板定位 | `position: fixed` | 不受 ReactFlow Canvas transform 影响；固定在视口右侧 |
| 数据获取 | 优先使用已加载数据（props 传入），DependenciesView 按需 API 调用 | 减少冗余请求；ModuleGraphView 已加载完整 modules/tasks 数据 |
| InfoBadge 交互 | 改为点击触发 PropertyPanel（取代 Tooltip 弹出） | 统一信息展示位置，避免 Tooltip 遮挡节点；信息更丰富 |
| 多选时行为 | Shift/Ctrl 多选时不触发 PropertyPanel，只有普通单击才触发 | 多选主要用于框选/拖动，不应打扰 Panel 内容 |
| 快捷键实现 | `window.addEventListener('keydown')` in `useEffect` | 简单直接，避免使用第三方快捷键库 |

---

## 十、与现有代码的注意事项

1. **`onTaskClick` 的现有逻辑**：目前 `features/modules/index.tsx` 中 `onTaskClick={(taskId) => handleEditTask(taskId)}` 会打开编辑 Drawer。建议保留此行为（双击或右键打开编辑），单击改为显示 PropertyPanel。

2. **ReactFlow 事件穿透**：PropertyPanel 覆盖在 ReactFlow 上方，需要确保 `pointer-events: none` 对面板外区域，避免阻碍图形交互。面板本身设 `pointer-events: auto`。

3. **`selectedModuleIds` 与 `usePropertyPanelStore.selectedModuleId` 的同步**：`ModuleGraphView` 内部的 `selectedModuleIds: Set<string>` 是多选状态，而 PropertyPanel 只展示第一个（或普通单选）的模块信息。在 `onNodeClick` 的普通点击分支同步即可，无需复杂合并。

4. **`tasks` 数据包含所有 JSON 字段为字符串**：`task.upstreamContractDetail` 是 JSON 字符串，各子视图需要 `JSON.parse`，应做好错误处理（`try/catch`）。

5. **Panel 的 z-index 层级**：ReactFlow 的 Controls 和 Panel 的 z-index 约为 4~5；PropertyPanel 的 z-index 应设为 10，确保在其之上。
