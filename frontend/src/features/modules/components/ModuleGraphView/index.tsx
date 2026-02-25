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
  Handle,
  Position,
  EdgeProps,
  getBezierPath,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Module, Task, TaskDependency, ModuleDependency } from '../../../../types';
import { AppstoreOutlined, SettingOutlined } from '@ant-design/icons';
import { Checkbox, Popover, Button, Tooltip } from 'antd';

interface ModuleGraphViewProps {
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
  moduleDependencies: ModuleDependency[];
  onModuleClick?: (moduleId: string) => void;
  onTaskClick?: (taskId: string) => void;
}

// 全局显示设置接口
interface DisplaySettings {
  showInternalEdges: boolean;      // 显示模块内部连线
  showCrossModuleEdges: boolean;   // 显示跨模块连线
  showPrompt: boolean;             // 显示提示词标识
  showError: boolean;              // 显示错误信息标识
  showNotification: boolean;       // 显示通知标识
  showStatus: boolean;             // 显示状态
  fontSize: number;                // 全局字体大小
  requireAltForTooltip: boolean;   // 需要按Alt键才显示Tooltip
  tooltipScale: number;            // Tooltip整体缩放比例
}

// 默认显示设置
const defaultDisplaySettings: DisplaySettings = {
  showInternalEdges: true,
  showCrossModuleEdges: true,
  showPrompt: true,
  showError: true,
  showNotification: true,
  showStatus: true,
  fontSize: 12,
  requireAltForTooltip: true,  // 默认开启Alt键显示
  tooltipScale: 1,             // 默认缩放比例为1
};

// 模块节点数据 - 包含任务列表
interface ModuleNodeData {
  label: string;
  status: string;
  collapsed: boolean;
  onToggle: () => void;
  color: string;
  tasks: Task[];
  taskDependencies: TaskDependency[];
  onTaskClick?: (taskId: string) => void;
  displaySettings: DisplaySettings;
}

const ModuleGraphView: React.FC<ModuleGraphViewProps> = ({
  modules,
  tasks,
  taskDependencies,
  onModuleClick,
  onTaskClick,
}) => {
  const [collapsedModules, setCollapsedModules] = useState<Set<string>>(new Set());
  const [displaySettings, setDisplaySettings] = useState<DisplaySettings>(defaultDisplaySettings);

  // 更新显示设置
  const updateDisplaySettings = (key: keyof DisplaySettings, value: boolean) => {
    setDisplaySettings(prev => ({ ...prev, [key]: value }));
  };

  // 生成节点和边
  const { nodes: initialNodes, edges: initialEdges } = useMemo(() => {
    const nodes: Node<ModuleNodeData>[] = [];
    const edges: Edge[] = [];

    // 模块布局位置计算
    const modulePositions = new Map<string, { x: number; y: number }>();
    const columns = Math.ceil(Math.sqrt(modules.length));
    const columnWidth = 400;
    const rowHeight = 500;

    modules.forEach((module, index) => {
      const col = index % columns;
      const row = Math.floor(index / columns);
      modulePositions.set(module.id, {
        x: col * columnWidth + 50,
        y: row * rowHeight + 50,
      });
    });

    // 创建模块节点（包含任务）
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
          tasks: moduleTasks,
          taskDependencies,
          onTaskClick,
          displaySettings,
        },
        style: {
          width: isCollapsed ? 200 : 320,
          height: isCollapsed ? 50 : Math.max(150, 50 + moduleTasks.length * 36),
          background: 'transparent',
          border: 'none',
          borderRadius: '12px',
          padding: '0',
        },
        zIndex: 1,
      });
    });

    // 创建任务依赖边
    taskDependencies.forEach(dep => {
      const upstreamTask = tasks.find(t => t.id === dep.upstreamTaskId);
      const downstreamTask = tasks.find(t => t.id === dep.downstreamTaskId);

      if (upstreamTask && downstreamTask) {
        const isCrossModule = upstreamTask.moduleId !== downstreamTask.moduleId;

        // 根据显示设置过滤边
        if (isCrossModule && !displaySettings.showCrossModuleEdges) return;
        if (!isCrossModule && !displaySettings.showInternalEdges) return;

        edges.push({
          id: `edge-${dep.id}`,
          source: `module-${upstreamTask.moduleId}`,
          target: `module-${downstreamTask.moduleId}`,
          sourceHandle: `task-${upstreamTask.id}`,
          targetHandle: `task-${downstreamTask.id}`,
          type: 'flowEdge',
          data: {
            isCrossModule,
            upstreamTaskName: upstreamTask.name,
            downstreamTaskName: downstreamTask.name,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: isCrossModule ? '#ff6b6b' : '#00d9ff',
            width: 20,
            height: 20,
          },
          style: {
            stroke: isCrossModule ? '#ff6b6b' : '#00d9ff',
            strokeWidth: isCrossModule ? 4 : 2,  // 外部粗，内部细
          },
          zIndex: 1000,
        });
      }
    });

    return { nodes, edges };
  }, [modules, tasks, taskDependencies, collapsedModules, onTaskClick, displaySettings]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // 当 initialNodes/initialEdges 变化时更新节点和边
  React.useEffect(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      if (node.id.startsWith('module-')) {
        const moduleId = node.id.replace('module-', '');
        onModuleClick?.(moduleId);
      }
    },
    [onModuleClick]
  );

  // 自定义节点类型
  const nodeTypes = useMemo(() => ({
    module: ModuleNodeComponent,
  }), []);

  // 自定义边类型 - 带流动箭头
  const edgeTypes = useMemo(() => ({
    flowEdge: FlowEdge,
  }), []);

  // 全局控制面板内容
  const settingsContent = (
    <div style={{ width: 280, padding: '8px 0' }}>
      <div style={{ marginBottom: 16 }}>
        <div style={{ fontWeight: 600, marginBottom: 8, color: '#fff' }}>连线显示</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Checkbox
            checked={displaySettings.showInternalEdges}
            onChange={(e) => updateDisplaySettings('showInternalEdges', e.target.checked)}
          >
            <span style={{ color: '#00d9ff' }}>●模块内部连线（细蓝线）</span>
          </Checkbox>
          <Checkbox
            checked={displaySettings.showCrossModuleEdges}
            onChange={(e) => updateDisplaySettings('showCrossModuleEdges', e.target.checked)}
          >
            <span style={{ color: '#ff6b6b' }}>●跨模块连线（粗红线）</span>
          </Checkbox>
        </div>
      </div>
      <div style={{ borderTop: '1px solid #3d3d5c', paddingTop: 12, marginBottom: 16 }}>
        <div style={{ fontWeight: 600, marginBottom: 8, color: '#fff' }}>任务节点提示信息标识</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Checkbox
            checked={displaySettings.showPrompt}
            onChange={(e) => updateDisplaySettings('showPrompt', e.target.checked)}
          >
            📝 提示词
          </Checkbox>
          <Checkbox
            checked={displaySettings.showError}
            onChange={(e) => updateDisplaySettings('showError', e.target.checked)}
          >
            ❌ 错误信息
          </Checkbox>
          <Checkbox
            checked={displaySettings.showNotification}
            onChange={(e) => updateDisplaySettings('showNotification', e.target.checked)}
          >
            🔔 通知
          </Checkbox>
          <Checkbox
            checked={displaySettings.showStatus}
            onChange={(e) => updateDisplaySettings('showStatus', e.target.checked)}
          >
            📊 状态
          </Checkbox>
        </div>
      </div>
      <div style={{ borderTop: '1px solid #3d3d5c', paddingTop: 12, marginBottom: 16 }}>
        <div style={{ fontWeight: 600, marginBottom: 8, color: '#fff' }}>字体大小</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Button
            size="small"
            onClick={() => updateDisplaySettings('fontSize', Math.max(10, displaySettings.fontSize - 1))}
            disabled={displaySettings.fontSize <= 10}
          >
            A-
          </Button>
          <span style={{ color: '#fff', minWidth: 40, textAlign: 'center' }}>{displaySettings.fontSize}px</span>
          <Button
            size="small"
            onClick={() => updateDisplaySettings('fontSize', Math.min(18, displaySettings.fontSize + 1))}
            disabled={displaySettings.fontSize >= 18}
          >
            A+
          </Button>
        </div>
      </div>
      <div style={{ borderTop: '1px solid #3d3d5c', paddingTop: 12, marginBottom: 16 }}>
        <div style={{ fontWeight: 600, marginBottom: 8, color: '#fff' }}>Tooltip设置</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Checkbox
            checked={displaySettings.requireAltForTooltip}
            onChange={(e) => updateDisplaySettings('requireAltForTooltip', e.target.checked)}
          >
            按住Alt键显示Tooltip
          </Checkbox>
        </div>
        <div style={{ marginTop: 8 }}>
          <div style={{ fontSize: 12, color: '#888', marginBottom: 4 }}>Tooltip缩放比例</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <Button
              size="small"
              onClick={() => updateDisplaySettings('tooltipScale', Math.max(0.5, displaySettings.tooltipScale - 0.1))}
              disabled={displaySettings.tooltipScale <= 0.5}
            >
              -
            </Button>
            <span style={{ color: '#fff', minWidth: 50, textAlign: 'center' }}>{(displaySettings.tooltipScale * 100).toFixed(0)}%</span>
            <Button
              size="small"
              onClick={() => updateDisplaySettings('tooltipScale', Math.min(2, displaySettings.tooltipScale + 0.1))}
              disabled={displaySettings.tooltipScale >= 2}
            >
              +
            </Button>
          </div>
        </div>
      </div>
      <div style={{ borderTop: '1px solid #3d3d5c', paddingTop: 12 }}>
        <Button
          size="small"
          onClick={() => setDisplaySettings(defaultDisplaySettings)}
          style={{ width: '100%' }}
        >
          重置为默认
        </Button>
      </div>
    </div>
  );

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
        edgeTypes={edgeTypes}
        defaultEdgeOptions={{
          type: 'flowEdge',
          style: { stroke: '#00d9ff', strokeWidth: 2 },
        }}
        elevateEdgesOnSelect={true}
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
            display: 'flex',
            alignItems: 'center',
            gap: 12,
          }}>
            <AppstoreOutlined style={{ marginRight: 4 }} />
            <span>模块：{modules.length}</span>
            <span>|</span>
            <span>任务：{tasks.length}</span>
            <span>|</span>
            <span>依赖：{taskDependencies.length}</span>
          </div>
        </Panel>
        <Panel position="top-right" style={{ background: 'transparent' }}>
          <Popover
            content={settingsContent}
            title={
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <SettingOutlined />
                <span>全局显示设置</span>
              </div>
            }
            trigger="click"
            placement="bottomRight"
            overlayStyle={{
              background: '#1a1a2e',
              border: '1px solid #3d3d5c',
            }}
          >
            <Button
              type="primary"
              icon={<SettingOutlined />}
              style={{
                background: '#1a1a2e',
                border: '1px solid #3d3d5c',
                color: '#fff',
              }}
            >
              显示设置
            </Button>
          </Popover>
        </Panel>
      </ReactFlow>
    </div>
  );
};

// 模块节点组件 - 包含任务列表
const ModuleNodeComponent = ({ data }: { data: ModuleNodeData }) => {
  const { label, status, collapsed, onToggle, color, tasks, taskDependencies, onTaskClick, displaySettings } = data;

  // 获取任务的依赖信息
  const getTaskDependencyInfo = (taskId: string) => {
    const upstreamDeps = taskDependencies.filter(d => d.downstreamTaskId === taskId);
    const downstreamDeps = taskDependencies.filter(d => d.upstreamTaskId === taskId);
    return { upstreamDeps, downstreamDeps };
  };

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
      overflow: 'hidden',
    }}>
      {/* 模块头部 */}
      <div
        onClick={onToggle}
        style={{
          padding: '10px 14px',
          borderBottom: collapsed ? 'none' : '1px solid rgba(255,255,255,0.1)',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          flexShrink: 0,
          background: 'rgba(0,0,0,0.2)',
        }}
      >
        <AppstoreOutlined style={{ color: '#00d9ff' }} />
        <span style={{ fontWeight: 600, fontSize: '14px', flex: 1 }}>{label}</span>
        <span style={{
          fontSize: '10px',
          padding: '2px 8px',
          background: 'rgba(255,255,255,0.2)',
          borderRadius: '4px',
        }}>
          {status}
        </span>
        <span style={{ fontSize: '10px', color: '#a0a0a0' }}>
          {collapsed ? '▶' : '▼'}
        </span>
      </div>
      
      {/* 任务列表 */}
      {!collapsed && (
        <div style={{
          flex: 1,
          padding: '8px',
          overflow: 'auto',
        }}>
          {tasks.length === 0 ? (
            <div style={{
              textAlign: 'center',
              color: '#666',
              fontSize: '12px',
              padding: '20px',
            }}>
              暂无任务
            </div>
          ) : (
            tasks.map((task) => {
              const { upstreamDeps, downstreamDeps } = getTaskDependencyInfo(task.id);
              const hasUpstream = upstreamDeps.length > 0;
              const hasDownstream = downstreamDeps.length > 0;
              const hasDependency = hasUpstream || hasDownstream;
              
              return (
                <TaskNode
                  key={task.id}
                  task={task}
                  upstreamDeps={upstreamDeps}
                  downstreamDeps={downstreamDeps}
                  displaySettings={displaySettings}
                  onTaskClick={onTaskClick}
                  tasks={tasks}
                />
              );
            })
          )}
        </div>
      )}
    </div>
  );
};

// 任务节点组件 - 支持悬停不同位置显示不同信息
const TaskNode: React.FC<{
  task: Task;
  upstreamDeps: TaskDependency[];
  downstreamDeps: TaskDependency[];
  displaySettings: DisplaySettings;
  onTaskClick?: (taskId: string) => void;
  tasks: Task[];
}> = ({ task, upstreamDeps, downstreamDeps, displaySettings, onTaskClick, tasks }) => {
  const hasUpstream = upstreamDeps.length > 0;
  const hasDownstream = downstreamDeps.length > 0;
  const hasDependency = hasUpstream || hasDownstream;

  // 获取上游任务信息
  const getUpstreamTaskInfo = () => {
    return upstreamDeps.map(dep => {
      const upstreamTask = tasks.find(t => t.id === dep.upstreamTaskId);
      return {
        task: upstreamTask,
        contract: dep.interfaceContract || '未定义接口',
      };
    });
  };

  // 获取下游任务信息
  const getDownstreamTaskInfo = () => {
    return downstreamDeps.map(dep => {
      const downstreamTask = tasks.find(t => t.id === dep.downstreamTaskId);
      return {
        task: downstreamTask,
        contract: dep.interfaceContract || '未定义接口',
      };
    });
  };

  return (
    <div
      onClick={(e) => {
        e.stopPropagation();
        onTaskClick?.(task.id);
      }}
      style={{
        width: '100%',
        height: '32px',
        background: getTaskColor(task.status),
        border: hasDependency ? '1px solid #00d9ff' : '1px solid rgba(255,255,255,0.1)',
        borderRadius: '6px',
        padding: '4px 10px',
        marginBottom: '4px',
        color: '#ffffff',
        fontSize: `${displaySettings.fontSize}px`,
        display: 'flex',
        alignItems: 'center',
        gap: '6px',
        cursor: 'pointer',
        transition: 'transform 0.1s, box-shadow 0.1s',
        boxShadow: hasDependency ? '0 0 8px rgba(0, 217, 255, 0.3)' : 'none',
        position: 'relative',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.transform = 'translateX(4px)';
        e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.3)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.transform = 'translateX(0)';
        e.currentTarget.style.boxShadow = hasDependency ? '0 0 8px rgba(0, 217, 255, 0.3)' : 'none';
      }}
    >
      {/* 左侧 Handle - 用于接收上游依赖 */}
      <Handle
        type="target"
        id={`task-${task.id}`}
        position={Position.Left}
        style={{
          background: hasUpstream ? '#f59e0b' : '#555',
          width: '8px',
          height: '8px',
          border: '2px solid #fff',
          left: '-4px',
          top: '50%',
          transform: 'translateY(-50%)',
        }}
      />
      <div style={{
        width: '6px',
        height: '6px',
        background: '#ffffff',
        borderRadius: '50%',
        flexShrink: 0,
      }} />
      <span style={{
        flex: 1,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap'
      }}>
        {task.name}
      </span>
      {/* 提示标识 - 每个都有独立的Tooltip */}
      {displaySettings.showPrompt && task.prompt && (
        <InfoBadge
          icon="📝"
          title="提示词"
          content={task.prompt}
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
        />
      )}
      {displaySettings.showError && task.logs && (
        <InfoBadge
          icon="❌"
          title="错误信息"
          content={task.logs}
          type="error"
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
        />
      )}
      {/* 上游依赖箭头 - 黄色，带Tooltip显示上游接口信息 */}
      {hasUpstream && (
        <DependencyBadge
          type="upstream"
          count={upstreamDeps.length}
          taskInfo={getUpstreamTaskInfo()}
          contractDetail={task.upstreamContractDetail}
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
        />
      )}
      {/* 下游依赖箭头 - 蓝色，带Tooltip显示下游接口信息 */}
      {hasDownstream && (
        <DependencyBadge
          type="downstream"
          count={downstreamDeps.length}
          taskInfo={getDownstreamTaskInfo()}
          contractDetail={task.downstreamContractDetail}
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
        />
      )}
      {displaySettings.showStatus && (
        <span style={{
          fontSize: `${displaySettings.fontSize - 3}px`,
          padding: '1px 6px',
          background: 'rgba(255,255,255,0.2)',
          borderRadius: '3px',
          flexShrink: 0,
        }}>
          {task.status}
        </span>
      )}
      {/* 右侧 Handle - 用于连接下游依赖 */}
      <Handle
        type="source"
        id={`task-${task.id}`}
        position={Position.Right}
        style={{
          background: hasDownstream ? '#00d9ff' : '#555',
          width: '8px',
          height: '8px',
          border: '2px solid #fff',
          right: '-4px',
          top: '50%',
          transform: 'translateY(-50%)',
        }}
      />
    </div>
  );
};

// 信息标识组件 - 悬停显示详细信息（使用Ant Design Tooltip）
const InfoBadge: React.FC<{
  icon: string;
  title: string;
  content: string;
  type?: 'default' | 'error' | 'warning' | 'success';
  subItems?: { label: string; value: string }[];
  fontSize?: number;
  requireAltForTooltip?: boolean;
  tooltipScale?: number;
}> = ({ icon, title, content, type = 'default', subItems, fontSize = 12, requireAltForTooltip = true, tooltipScale = 1 }) => {
  const [isAltPressed, setIsAltPressed] = React.useState(false);
  const [showTooltip, setShowTooltip] = React.useState(false);

  React.useEffect(() => {
    if (!requireAltForTooltip) return;
    
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.altKey) setIsAltPressed(true);
    };
    const handleKeyUp = (e: KeyboardEvent) => {
      if (!e.altKey) setIsAltPressed(false);
    };
    
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  }, [requireAltForTooltip]);

  const getColor = () => {
    switch (type) {
      case 'error': return '#ff6b6b';
      case 'warning': return '#f59e0b';
      case 'success': return '#10b981';
      default: return '#00d9ff';
    }
  };

  const color = getColor();
  const scaledFontSize = Math.round(fontSize * tooltipScale);
  const scaledWidth = Math.round(450 * tooltipScale);

  const tooltipContent = (
    <div style={{ minWidth: scaledWidth, maxWidth: scaledWidth + 100 }}>
      <div style={{
        fontWeight: 600,
        fontSize: scaledFontSize + 2,
        marginBottom: 8,
        color: color,
        borderBottom: `1px solid ${color}33`,
        paddingBottom: 4,
      }}>
        {title}
      </div>
      <div style={{
        background: `${color}15`,
        padding: 10,
        borderRadius: 4,
        fontSize: scaledFontSize,
        color: '#ccc',
        maxHeight: 200 * tooltipScale,
        overflow: 'auto',
        whiteSpace: 'pre-wrap',
        lineHeight: 1.5,
      }}>
        {content}
      </div>
      {subItems && subItems.length > 0 && (
        <div style={{ marginTop: 8 }}>
          <div style={{ fontSize: scaledFontSize, color: '#888', marginBottom: 4 }}>关联任务：</div>
          {subItems.map((item, idx) => (
            <div key={idx} style={{
              background: '#2d2d44',
              padding: 8,
              borderRadius: 4,
              marginBottom: 4,
              fontSize: scaledFontSize - 1,
            }}>
              <div style={{ color: '#fff', marginBottom: 2 }}>📋 {item.label}</div>
              <div style={{ color: color, fontSize: scaledFontSize - 2 }}>{item.value}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  // 如果需要Alt键但Alt未按下，不显示Tooltip
  if (requireAltForTooltip && !isAltPressed) {
    return (
      <span style={{
        fontSize: `${fontSize - 2}px`,
        cursor: 'pointer',
        opacity: 0.7,
      }}>
        {icon}
      </span>
    );
  }

  return (
    <Tooltip
      title={tooltipContent}
      color="#1a1a2e"
      overlayInnerStyle={{ padding: 16 * tooltipScale, maxWidth: (520 + 100) * tooltipScale }}
      mouseEnterDelay={0}
      mouseLeaveDelay={0.1}
      open={showTooltip}
      onOpenChange={(visible) => {
        if (!requireAltForTooltip || isAltPressed) {
          setShowTooltip(visible);
        }
      }}
    >
      <span style={{
        fontSize: `${fontSize - 2}px`,
        cursor: 'pointer',
      }}>
        {icon}
      </span>
    </Tooltip>
  );
};

// 依赖箭头组件 - 带Tooltip显示上下游接口信息
const DependencyBadge: React.FC<{
  type: 'upstream' | 'downstream';
  count: number;
  taskInfo: { task?: Task; contract: string }[];
  contractDetail?: string;
  fontSize?: number;
  requireAltForTooltip?: boolean;
  tooltipScale?: number;
}> = ({ type, count, taskInfo, contractDetail, fontSize = 12, requireAltForTooltip = true, tooltipScale = 1 }) => {
  const [isAltPressed, setIsAltPressed] = React.useState(false);
  const [showTooltip, setShowTooltip] = React.useState(false);

  React.useEffect(() => {
    if (!requireAltForTooltip) return;
    
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.altKey) setIsAltPressed(true);
    };
    const handleKeyUp = (e: KeyboardEvent) => {
      if (!e.altKey) setIsAltPressed(false);
    };
    
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  }, [requireAltForTooltip]);

  const isUpstream = type === 'upstream';
  const color = isUpstream ? '#f59e0b' : '#00d9ff';
  const bgColor = isUpstream ? '#f59e0b' : '#00d9ff';
  const textColor = isUpstream ? '#000' : '#000';
  const title = isUpstream ? '上游依赖接口' : '下游依赖接口';
  const arrow = isUpstream ? '↑' : '↓';

  const scaledFontSize = Math.round(fontSize * tooltipScale);
  const scaledWidth = Math.round(450 * tooltipScale);

  const tooltipContent = (
    <div style={{ minWidth: scaledWidth, maxWidth: scaledWidth + 100 }}>
      <div style={{
        fontWeight: 600,
        fontSize: scaledFontSize + 2,
        marginBottom: 8,
        color: color,
        borderBottom: `1px solid ${color}33`,
        paddingBottom: 4,
      }}>
        {title}（共{count}个依赖）
      </div>
      {contractDetail && (
        <div style={{
          background: `${color}15`,
          padding: 10,
          borderRadius: 4,
          fontSize: scaledFontSize,
          color: '#ccc',
          maxHeight: 150 * tooltipScale,
          overflow: 'auto',
          whiteSpace: 'pre-wrap',
          lineHeight: 1.5,
          marginBottom: 8,
        }}>
          {contractDetail}
        </div>
      )}
      <div style={{ marginTop: 8 }}>
        <div style={{ fontSize: scaledFontSize, color: '#888', marginBottom: 6 }}>关联任务：</div>
        {taskInfo.map((info, idx) => (
          <div key={idx} style={{
            background: '#2d2d44',
            padding: 10,
            borderRadius: 4,
            marginBottom: 6,
            fontSize: scaledFontSize,
          }}>
            <div style={{ color: '#fff', marginBottom: 4, fontWeight: 500 }}>📋 {info.task?.name || '未知任务'}</div>
            <div style={{ color: color, fontSize: fontSize - 1, background: '#1a1a2e', padding: 6, borderRadius: 3 }}>
              {info.contract}
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  // 如果需要Alt键但Alt未按下，显示提示
  if (requireAltForTooltip && !isAltPressed) {
    return (
      <Tooltip
        title={<span style={{ fontSize: scaledFontSize }}>按住 Alt 键查看详情</span>}
        color="#1a1a2e"
        overlayInnerStyle={{ padding: 8 * tooltipScale }}
        mouseEnterDelay={0}
        mouseLeaveDelay={0.1}
      >
        <span style={{
          fontSize: `${fontSize - 3}px`,
          padding: '1px 5px',
          background: bgColor,
          color: textColor,
          borderRadius: '3px',
          flexShrink: 0,
          cursor: 'pointer',
          fontWeight: 600,
          fontFamily: 'monospace',
          opacity: 0.7,
        }}>
          {arrow}[{count}]
        </span>
      </Tooltip>
    );
  }

  return (
    <Tooltip
      title={tooltipContent}
      color="#1a1a2e"
      overlayInnerStyle={{ padding: 16 * tooltipScale, maxWidth: (580) * tooltipScale }}
      mouseEnterDelay={0}
      mouseLeaveDelay={0.1}
    >
      <span style={{
        fontSize: `${fontSize - 3}px`,
        padding: '1px 5px',
        background: bgColor,
        color: textColor,
        borderRadius: '3px',
        flexShrink: 0,
        cursor: 'pointer',
        fontWeight: 600,
        fontFamily: 'monospace',
      }}>
        {arrow}[{count}]
      </span>
    </Tooltip>
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

// 自定义边组件 - 带流动箭头动画
const FlowEdge: React.FC<EdgeProps> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  data,
}) => {
  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const isCrossModule = data?.isCrossModule || false;
  const edgeColor = isCrossModule ? '#ff6b6b' : '#00d9ff';
  const strokeWidth = isCrossModule ? 4 : 2;  // 外部粗，内部细

  return (
    <>
      {/* 定义流动箭头动画的 SVG */}
      <defs>
        <marker
          id={`arrow-${id}`}
          viewBox="0 0 10 10"
          refX="8"
          refY="5"
          markerWidth="6"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 0 0 L 10 5 L 0 10 z" fill={edgeColor} />
        </marker>
        {/* 流动粒子效果 */}
        <filter id={`glow-${id}`} x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="2" result="coloredBlur" />
          <feMerge>
            <feMergeNode in="coloredBlur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        {/* 动态渐变 - 根据是否跨模块使用不同颜色 */}
        <linearGradient id={`flowGradient-${id}`} gradientUnits="userSpaceOnUse">
          <stop offset="0%" stopColor="transparent" />
          <stop offset="50%" stopColor={edgeColor} />
          <stop offset="100%" stopColor="transparent" />
        </linearGradient>
      </defs>
      
      {/* 主边线 */}
      <path
        id={id}
        style={{
          ...style,
          stroke: edgeColor,
          strokeWidth: strokeWidth,
          fill: 'none',
        }}
        className="react-flow__edge-path"
        d={edgePath}
        markerEnd={markerEnd}
      />
      
      {/* 流动动画层 */}
      <path
        d={edgePath}
        fill="none"
        stroke={`url(#flowGradient-${id})`}
        strokeWidth={strokeWidth + 2}
        strokeLinecap="round"
        strokeDasharray={isCrossModule ? "12 24" : "8 16"}
        filter={`url(#glow-${id})`}
        style={{
          animation: isCrossModule ? 'flowAnimationFast 1s linear infinite' : 'flowAnimation 1.5s linear infinite',
        }}
      />
    </>
  );
};

// 添加 CSS 动画样式（注入到页面）
if (typeof document !== 'undefined') {
  const styleId = 'flow-edge-animation';
  if (!document.getElementById(styleId)) {
    const style = document.createElement('style');
    style.id = styleId;
    style.textContent = `
      @keyframes flowAnimation {
        0% {
          stroke-dashoffset: 24;
        }
        100% {
          stroke-dashoffset: 0;
        }
      }
      
      @keyframes flowAnimationFast {
        0% {
          stroke-dashoffset: 36;
        }
        100% {
          stroke-dashoffset: 0;
        }
      }
      
      .react-flow__edge {
        z-index: 1000 !important;
      }
      
      .react-flow__edge-path {
        filter: drop-shadow(0 0 3px rgba(0, 217, 255, 0.5));
      }
    `;
    document.head.appendChild(style);
  }
}

export default ModuleGraphView;
