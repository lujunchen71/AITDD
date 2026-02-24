import React, { useCallback, useMemo, useState } from 'react';
import ReactFlow, {
  Node,
  Edge,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  MarkerType,
  Panel,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Module, Task, TaskDependency, ModuleDependency } from '../../../../types';
import { AppstoreOutlined } from '@ant-design/icons';

interface ModuleGraphViewProps {
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
  moduleDependencies: ModuleDependency[];
  onModuleClick?: (moduleId: string) => void;
  onTaskClick?: (taskId: string) => void;
}

// 模块节点数据
interface ModuleNodeData {
  label: string;
  status: string;
  collapsed: boolean;
  onToggle: () => void;
  color: string;
}

// 任务节点数据
interface TaskNodeData {
  label: string;
  status: string;
  moduleId: string;
  color: string;
}

const ModuleGraphView: React.FC<ModuleGraphViewProps> = ({
  modules,
  tasks,
  taskDependencies,
  onModuleClick,
  onTaskClick,
}) => {
  const [collapsedModules, setCollapsedModules] = useState<Set<string>>(new Set());

  // 生成节点和边
  const { nodes: initialNodes, edges: initialEdges } = useMemo(() => {
    const nodes: Node<ModuleNodeData | TaskNodeData>[] = [];
    const edges: Edge[] = [];

    // 模块布局位置计算
    const modulePositions = new Map<string, { x: number; y: number }>();
    const columns = Math.ceil(Math.sqrt(modules.length));
    const columnWidth = 350;
    const rowHeight = 400;

    modules.forEach((module, index) => {
      const col = index % columns;
      const row = Math.floor(index / columns);
      modulePositions.set(module.id, {
        x: col * columnWidth + 50,
        y: row * rowHeight + 50,
      });
    });

    // 创建模块节点
    modules.forEach((module) => {
      const position = modulePositions.get(module.id) || { x: 0, y: 0 };
      const isCollapsed = collapsedModules.has(module.id);
      const moduleTasks = tasks.filter(t => t.moduleId === module.id);

      // 模块节点
      nodes.push({
        id: `module-${module.id}`,
        type: 'module',
        position,
        data: {
          label: module.name,
          status: module.status,
          collapsed: isCollapsed,
          onToggle: () => {
            setCollapsedModules(prev => {
              const next = new Set(prev);
              if (next.has(module.id)) {
                next.delete(module.id);
              } else {
                next.add(module.id);
              }
              return next;
            });
          },
          color: getModuleColor(module.status),
        },
        style: {
          width: isCollapsed ? 200 : 300,
          height: isCollapsed ? 60 : Math.max(200, 60 + moduleTasks.length * 40),
          background: getModuleColor(module.status),
          border: '2px solid #3d3d5c',
          borderRadius: '12px',
          padding: '0',
          color: '#ffffff',
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.4)',
        },
        zIndex: 1,
      });

      // 创建任务节点（如果模块未折叠）
      if (!isCollapsed) {
        moduleTasks.forEach((task, taskIndex) => {
          nodes.push({
            id: `task-${task.id}`,
            type: 'task',
            position: { x: 20, y: 70 + taskIndex * 40 },
            parentId: `module-${module.id}`,
            data: {
              label: task.name,
              status: task.status,
              moduleId: module.id,
              color: getTaskColor(task.status),
            },
            style: {
              width: 260,
              height: 32,
              background: getTaskColor(task.status),
              border: '1px solid #2d2d44',
              borderRadius: '6px',
              padding: '4px 8px',
              color: '#ffffff',
              fontSize: '12px',
            },
          });
        });
      }
    });

    // 创建任务依赖边
    taskDependencies.forEach(dep => {
      const upstreamTask = tasks.find(t => t.id === dep.upstreamTaskId);
      const downstreamTask = tasks.find(t => t.id === dep.downstreamTaskId);

      if (upstreamTask && downstreamTask) {
        edges.push({
          id: `edge-${dep.id}`,
          source: `task-${upstreamTask.id}`,
          target: `task-${downstreamTask.id}`,
          type: 'bezier',
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: '#00d9ff',
          },
          style: {
            stroke: '#00d9ff',
            strokeWidth: 2,
          },
          animated: true,
        });
      }
    });

    return { nodes, edges };
  }, [modules, tasks, taskDependencies, collapsedModules]);

  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, , onEdgesChange] = useEdgesState(initialEdges);

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      if (node.id.startsWith('module-')) {
        const moduleId = node.id.replace('module-', '');
        onModuleClick?.(moduleId);
      } else if (node.id.startsWith('task-')) {
        const taskId = node.id.replace('task-', '');
        onTaskClick?.(taskId);
      }
    },
    [onModuleClick, onTaskClick]
  );

  // 自定义节点类型
  const nodeTypes = useMemo(() => ({
    module: ModuleNodeComponent,
    task: TaskNodeComponent,
  }), []);

  return (
    <div style={{ width: '100%', height: '100%', background: '#0f0f23' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        fitView
        attributionPosition="bottom-left"
        style={{ background: '#0f0f23' }}
        nodeTypes={nodeTypes}
        defaultEdgeOptions={{
          type: 'bezier',
          style: { stroke: '#00d9ff', strokeWidth: 2 },
        }}
      >
        <Background color="#2d2d44" gap={20} />
        <Controls
          style={{
            background: '#1a1a2e',
            border: '1px solid #3d3d5c',
            borderRadius: '4px',
          }}
        />
        <Panel position="top-left" style={{ background: 'transparent' }}>
          <div style={{
            background: '#1a1a2e',
            border: '1px solid #2d2d44',
            borderRadius: '6px',
            padding: '8px 12px',
            color: '#a0a0a0',
            fontSize: '12px',
          }}>
            <AppstoreOutlined style={{ marginRight: 8 }} />
            模块：{modules.length} | 任务：{tasks.length} | 依赖：{taskDependencies.length}
          </div>
        </Panel>
      </ReactFlow>
    </div>
  );
};

// 模块节点组件
const ModuleNodeComponent = ({ data }: { data: ModuleNodeData }) => {
  const { label, status, collapsed, onToggle, color } = data;

  return (
    <div style={{
      width: '100%',
      height: '100%',
      background: color,
      border: '2px solid #3d3d5c',
      borderRadius: '12px',
      padding: '0',
      color: '#ffffff',
      boxShadow: '0 4px 12px rgba(0, 0, 0, 0.4)',
      display: 'flex',
      flexDirection: 'column',
    }}>
      {/* 模块头部 */}
      <div
        onClick={onToggle}
        style={{
          padding: '8px 12px',
          borderBottom: '1px solid #3d3d5c',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
        }}
      >
        <AppstoreOutlined style={{ color: '#00d9ff' }} />
        <span style={{ fontWeight: 500, fontSize: '14px' }}>{label}</span>
        <span style={{
          fontSize: '10px',
          padding: '2px 6px',
          background: 'rgba(255,255,255,0.2)',
          borderRadius: '4px',
        }}>
          {status}
        </span>
        <span style={{ marginLeft: 'auto', fontSize: '10px', color: '#a0a0a0' }}>
          {collapsed ? '▶' : '▼'}
        </span>
      </div>
    </div>
  );
};

// 任务节点组件
const TaskNodeComponent = ({ data }: { data: TaskNodeData }) => {
  const { label, status, color } = data;

  return (
    <div style={{
      width: '100%',
      height: '100%',
      background: color,
      border: '1px solid #2d2d44',
      borderRadius: '6px',
      padding: '4px 8px',
      color: '#ffffff',
      fontSize: '12px',
      display: 'flex',
      alignItems: 'center',
      gap: '6px',
    }}>
      <div style={{
        width: '6px',
        height: '6px',
        background: '#ffffff',
        borderRadius: '50%',
      }} />
      <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
        {label}
      </span>
      <span style={{
        fontSize: '10px',
        padding: '1px 4px',
        background: 'rgba(255,255,255,0.2)',
        borderRadius: '3px',
      }}>
        {status}
      </span>
    </div>
  );
};

// 获取模块颜色
function getModuleColor(status: string): string {
  switch (status) {
    case 'designing':
      return 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)';
    case 'developing':
      return 'linear-gradient(135deg, #7b2cbf 0%, #a855f7 100%)';
    case 'completed':
      return 'linear-gradient(135deg, #065f46 0%, #10b981 100%)';
    case 'deprecated':
      return 'linear-gradient(135deg, #374151 0%, #6b7280 100%)';
    default:
      return 'linear-gradient(135deg, #1e3a8a 0%, #3b82f6 100%)';
  }
}

// 获取任务颜色
function getTaskColor(status: string): string {
  switch (status) {
    case 'ready':
      return 'linear-gradient(90deg, #374151 0%, #4b5563 100%)';
    case 'claimed':
      return 'linear-gradient(90deg, #1e3a8a 0%, #3b82f6 100%)';
    case 'in_progress':
      return 'linear-gradient(90deg, #7b2cbf 0%, #a855f7 100%)';
    case 'pending_review':
      return 'linear-gradient(90deg, #b45309 0%, #f59e0b 100%)';
    case 'completed':
      return 'linear-gradient(90deg, #065f46 0%, #10b981 100%)';
    case 'failed':
      return 'linear-gradient(90deg, #991b1b 0%, #ef4444 100%)';
    case 'blocked':
      return 'linear-gradient(90deg, #374151 0%, #6b7280 100%)';
    default:
      return 'linear-gradient(90deg, #374151 0%, #4b5563 100%)';
  }
}

export default ModuleGraphView;
