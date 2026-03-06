import React, { useCallback, useMemo, useState, useEffect, useRef } from 'react';
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
  useReactFlow,
  ReactFlowProvider,
  SelectionMode,
  OnSelectionChangeParams,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Module, Task, TaskDependency, ModuleDependency } from '../../../../types';
import BugLogDisplay, { parseBugLog } from '../BugLogDisplay';
import { AppstoreOutlined, SettingOutlined, PlayCircleOutlined, ApiOutlined, ReloadOutlined, DeleteOutlined } from '@ant-design/icons';
import { Checkbox, Popover, Button, Tooltip, message, Dropdown } from 'antd';
import { modulePositionApi } from '../../../../services/api';
import {
  localStorageService,
  DisplaySettings,
  defaultDisplaySettings as storageDefaultDisplaySettings
} from '../../../../services/localStorageService';
import { usePropertyPanelStore } from '../../../../stores/usePropertyPanelStore';
import PropertyPanel from '../PropertyPanel';
import BugLogPanel from '../BugLogPanel';
import { useProjectStore } from '../../../../stores/useProjectStore';

interface ModuleGraphViewProps {
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
  moduleDependencies: ModuleDependency[];
  onModuleClick?: (moduleId: string) => void;
  onTaskClick?: (taskId: string) => void;
  onModuleCompile?: (moduleId: string) => void;
  onModuleAnalyze?: (moduleId: string) => void;
  onModuleRefactor?: (moduleId: string) => void;
  onModuleDelete?: (moduleId: string) => void;
}

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
  isSelected?: boolean;
  onContextMenu?: (e: React.MouseEvent, moduleId: string) => void;
}

const ModuleGraphView: React.FC<ModuleGraphViewProps> = ({
  modules,
  tasks,
  taskDependencies,
  onModuleClick,
  onTaskClick,
  onModuleCompile: _onModuleCompile,
  onModuleAnalyze: _onModuleAnalyze,
  onModuleRefactor: _onModuleRefactor,
  onModuleDelete: _onModuleDelete,
}) => {
  // 属性面板 Store
  const propertyPanelStore = usePropertyPanelStore();
  const isPanelVisible = usePropertyPanelStore(s => s.visible);

  // BugLog 面板可见状态（与属性面板互斥）
  const [bugLogPanelVisible, setBugLogPanelVisible] = React.useState(false);

  const handlePanelModeChange = React.useCallback((mode: 0 | 1 | 2) => {
    if (mode === 0) {
      propertyPanelStore.close();
      setBugLogPanelVisible(false);
    } else if (mode === 1) {
      setBugLogPanelVisible(false);
      propertyPanelStore.showProjectInfo();
    } else {
      propertyPanelStore.close();
      setBugLogPanelVisible(true);
    }
  }, [propertyPanelStore]);

  // 从 projectStore 获取当前项目（根据 modules 的 projectId 匹配，而不是默认的第一个项目）
  const { projects, project: storeProject } = useProjectStore();
  const currentProjectId = modules.length > 0 ? modules[0].projectId : null;
  const project = currentProjectId
    ? (projects.find(p => p.id === currentProjectId) || storeProject)
    : storeProject;

  // 用于动态计算属性面板顶部位置的 ref（指向 top-right Panel 的容器 div）
  const topRightPanelRef = React.useRef<HTMLDivElement>(null);
  const [panelAnchorTop, setPanelAnchorTop] = React.useState(160);

  // 当面板可见时，根据 top-right 控件的实际位置计算属性面板的 top 值
  React.useEffect(() => {
    const anyVisible = isPanelVisible || bugLogPanelVisible;
    if (anyVisible && topRightPanelRef.current) {
      const rect = topRightPanelRef.current.getBoundingClientRect();
      setPanelAnchorTop(rect.bottom + 8);
    }
  }, [isPanelVisible, bugLogPanelVisible]);

  // 从本地存储加载折叠的模块ID列表
  const [collapsedModules, setCollapsedModules] = useState<Set<string>>(() => {
    const savedCollapsedIds = localStorageService.getCollapsedModules();
    console.log('[ModuleGraphView] 从本地存储加载折叠模块:', savedCollapsedIds);
    return new Set(savedCollapsedIds);
  });
  
  // 选中的模块ID列表（用于多选和持久化）
  const [selectedModuleIds, setSelectedModuleIds] = useState<Set<string>>(() => {
    const savedSelectedIds = localStorageService.getSelectedModules();
    console.log('[ModuleGraphView] 从本地存储加载选中模块:', savedSelectedIds);
    return new Set(savedSelectedIds);
  });
  // 从本地存储加载显示设置
  const [displaySettings, setDisplaySettings] = useState<DisplaySettings>(() => {
    const savedSettings = localStorageService.getDisplaySettings();
    console.log('[ModuleGraphView] 从本地存储加载显示设置:', savedSettings);
    return savedSettings;
  });
  
  // 防抖保存的 ref
  const saveTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  
  // 获取 ReactFlow 实例用于操作视口
  const { setViewport, getViewport } = useReactFlow();

  // 组件挂载时恢复视口位置
  useEffect(() => {
    const savedViewport = localStorageService.getGraphViewport();
    console.log('[ModuleGraphView] 恢复视口位置:', savedViewport);
    setViewport(savedViewport);
  }, [setViewport]);

  // 保存折叠状态到本地存储
  useEffect(() => {
    const collapsedArray = Array.from(collapsedModules);
    console.log('[ModuleGraphView] 保存折叠模块到本地存储:', collapsedArray);
    localStorageService.setCollapsedModules(collapsedArray);
  }, [collapsedModules]);

  // 清理无效的折叠模块ID（当模块列表变化时）
  useEffect(() => {
    if (modules.length > 0) {
      const validModuleIds = modules.map(m => m.id);
      const savedCollapsedIds = localStorageService.cleanupInvalidCollapsedModules(validModuleIds);
      // 如果清理后的列表与当前状态不同，更新状态
      const currentIds = Array.from(collapsedModules);
      if (JSON.stringify(currentIds.sort()) !== JSON.stringify(savedCollapsedIds.sort())) {
        setCollapsedModules(new Set(savedCollapsedIds));
      }
    }
  }, [modules]);

  // 保存显示设置到本地存储
  useEffect(() => {
    console.log('[ModuleGraphView] 保存显示设置到本地存储:', displaySettings);
    localStorageService.setDisplaySettings(displaySettings);
  }, [displaySettings]);

  // 保存选中状态到本地存储
  useEffect(() => {
    const selectedArray = Array.from(selectedModuleIds);
    console.log('[ModuleGraphView] 保存选中模块到本地存储:', selectedArray);
    localStorageService.setSelectedModules(selectedArray);
  }, [selectedModuleIds]);

  // 清理无效的选中模块ID（当模块列表变化时）
  useEffect(() => {
    if (modules.length > 0) {
      const validModuleIds = modules.map(m => m.id);
      const savedSelectedIds = localStorageService.cleanupInvalidSelectedModules(validModuleIds);
      // 如果清理后的列表与当前状态不同，更新状态
      const currentIds = Array.from(selectedModuleIds);
      if (JSON.stringify(currentIds.sort()) !== JSON.stringify(savedSelectedIds.sort())) {
        setSelectedModuleIds(new Set(savedSelectedIds));
      }
    }
  }, [modules]);

  // 更新显示设置
  const updateDisplaySettings = (key: keyof DisplaySettings, value: boolean | number) => {
    setDisplaySettings(prev => ({ ...prev, [key]: value }));
  };

  // 保存视口位置的防抖函数
  const saveViewportTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  
  const saveViewportToStorage = useCallback(() => {
    if (saveViewportTimeoutRef.current) {
      clearTimeout(saveViewportTimeoutRef.current);
    }
    saveViewportTimeoutRef.current = setTimeout(() => {
      const viewport = getViewport();
      console.log('[ModuleGraphView] 保存视口位置:', viewport);
      localStorageService.setGraphViewport(viewport);
    }, 300);
  }, [getViewport]);

  // 监听视口变化
  const onMoveEnd = useCallback(() => {
    saveViewportToStorage();
  }, [saveViewportToStorage]);

  // 防抖保存单个模块位置
  const debouncedSavePosition = useCallback(async (id: string, x: number, y: number) => {
    if (saveTimeoutRef.current) {
      clearTimeout(saveTimeoutRef.current);
    }
    saveTimeoutRef.current = setTimeout(async () => {
      try {
        // id 格式为 "module-xxx"，需要提取实际的模块 id
        const moduleId = id.replace('module-', '');
        await modulePositionApi.updatePosition(moduleId, x, y);
      } catch (error) {
        console.error('Failed to save position:', error);
      }
    }, 500);
  }, []);

  // 批量保存位置
  const savePositions = useCallback(async (positions: { [key: string]: { x: number; y: number } }) => {
    try {
      await modulePositionApi.batchUpdatePositions(
        Object.entries(positions).map(([moduleId, pos]) => ({
          moduleId,
          positionX: pos.x,
          positionY: pos.y,
        }))
      );
      message.success('布局已保存');
    } catch (error) {
      console.error('Failed to save positions:', error);
      message.error('保存布局失败');
    }
  }, []);

  // 基于依赖关系的层次布局算法
  const calculateAutoLayout = useCallback(() => {
    const newPositions: { [moduleId: string]: { x: number; y: number } } = {};
    
    // 计算每个模块的入度
    const inDegree: { [key: string]: number } = {};
    const dependents: { [key: string]: string[] } = {}; // 被哪些模块依赖
    
    modules.forEach(m => {
      inDegree[m.id] = 0;
      dependents[m.id] = [];
    });
    
    // 根据任务依赖关系计算模块间的依赖
    taskDependencies.forEach(dep => {
      const upstreamTask = tasks.find(t => t.id === dep.upstreamTaskId);
      const downstreamTask = tasks.find(t => t.id === dep.downstreamTaskId);
      
      if (upstreamTask && downstreamTask) {
        const upstreamModuleId = upstreamTask.moduleId;
        const downstreamModuleId = downstreamTask.moduleId;
        
        // 只处理跨模块依赖
        if (upstreamModuleId !== downstreamModuleId) {
          inDegree[downstreamModuleId]++;
          if (!dependents[upstreamModuleId].includes(downstreamModuleId)) {
            dependents[upstreamModuleId].push(downstreamModuleId);
          }
        }
      }
    });
    
    // 拓扑排序分配层级
    const levels: string[][] = [];
    let remaining = [...modules.map(m => m.id)];
    
    while (remaining.length > 0) {
      // 找出入度为0的节点
      const currentLevel = remaining.filter(id => inDegree[id] === 0);
      if (currentLevel.length === 0) break; // 防止循环
      
      levels.push(currentLevel);
      
      // 移除当前层节点，更新入度
      currentLevel.forEach(id => {
        remaining = remaining.filter(r => r !== id);
        dependents[id].forEach(depId => {
          inDegree[depId]--;
        });
      });
    }
    
    // 如果还有剩余节点（循环依赖），添加到最后一层
    if (remaining.length > 0) {
      levels.push(remaining);
    }
    
    // 分配位置 - 左右排列（层级在水平方向，同一层节点在垂直方向）
    const LAYER_WIDTH = 500;   // 层与层之间的水平间距
    const NODE_HEIGHT = 450;   // 同一层内节点之间的垂直间距
    
    // 记录每个模块的层级
    const moduleLevel: { [key: string]: number } = {};
    levels.forEach((level, levelIndex) => {
      level.forEach(moduleId => {
        moduleLevel[moduleId] = levelIndex;
      });
    });
    
    // 迭代优化：检查同一层级的模块，如果有依赖关系则往后移动
    let changed = true;
    let iterations = 0;
    const MAX_ITERATIONS = 30;
    
    while (changed && iterations < MAX_ITERATIONS) {
      changed = false;
      iterations++;
      
      // 遍历每一层
      for (let levelIndex = 0; levelIndex < levels.length; levelIndex++) {
        const currentLevel = [...levels[levelIndex]]; // 复制数组避免迭代时修改
        
        currentLevel.forEach(moduleId => {
          const currentModuleLevel = moduleLevel[moduleId];
          
          // 检查这个模块的所有下游依赖
          (dependents[moduleId] || []).forEach(downstreamId => {
            const downstreamLevel = moduleLevel[downstreamId];
            
            // 如果下游模块的层级 <= 当前模块的层级，需要往后移
            if (downstreamLevel <= currentModuleLevel) {
              const newLevel = currentModuleLevel + 1;
              // 从原层级移除
              levels[downstreamLevel] = levels[downstreamLevel].filter(id => id !== downstreamId);
              // 添加到新层级
              if (!levels[newLevel]) {
                levels[newLevel] = [];
              }
              levels[newLevel].push(downstreamId);
              moduleLevel[downstreamId] = newLevel;
              changed = true;
            }
          });
        });
      }
    }
    
    // 重新计算每个层级的模块索引并分配最终位置
    levels.forEach((level, levelIndex) => {
      const levelHeight = level.length * NODE_HEIGHT;
      const startY = -levelHeight / 2;
      
      level.forEach((moduleId, nodeIndex) => {
        newPositions[moduleId] = {
          x: levelIndex * LAYER_WIDTH,           // 层级在水平方向排列（从左到右）
          y: startY + nodeIndex * NODE_HEIGHT    // 同一层节点在垂直方向排列
        };
      });
    });
    
    return newPositions;
  }, [modules, tasks, taskDependencies]);

  // 右键菜单处理
  const handleContextMenu = useCallback((e: React.MouseEvent, moduleId: string) => {
    e.preventDefault();
    e.stopPropagation();
    // 如果右键的模块未选中，则选中它
    if (!selectedModuleIds.has(moduleId)) {
      console.log('[ModuleGraphView] 右键选中模块:', moduleId);
      setSelectedModuleIds(new Set([moduleId]));
    }
  }, [selectedModuleIds]);

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
      const isSelected = selectedModuleIds.has(module.id);
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
          isSelected,
          onContextMenu: handleContextMenu,
        },
        style: {
          width: isCollapsed ? 200 : 320,
          height: isCollapsed ? 50 : Math.max(150, 50 + moduleTasks.length * 36),
          background: 'transparent',
          border: isSelected ? '3px solid #00d9ff' : 'none',
          borderRadius: '12px',
          padding: '0',
        },
        zIndex: isSelected ? 10 : 1,
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
            width: isCrossModule ? 6 : 15,
            height: isCrossModule ? 6 : 15,
          },
          style: {
            stroke: isCrossModule ? '#ff6b6b' : '#00d9ff',
            strokeWidth: isCrossModule ? 1.5 : 2,  // 外部略粗，内部细
          },
          zIndex: 1000,
        });
      }
    });

    return { nodes, edges };
  }, [modules, tasks, taskDependencies, collapsedModules, selectedModuleIds, onTaskClick, displaySettings, handleContextMenu]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // 当 initialNodes/initialEdges 变化时更新节点和边
  // 注意：更新节点时保留现有位置，避免折叠/展开时位置被重置
  React.useEffect(() => {
    setNodes(prevNodes =>
      initialNodes.map(node => {
        const existingNode = prevNodes.find(n => n.id === node.id);
        return existingNode
          ? { ...node, position: existingNode.position }
          : node;
      })
    );
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  // 加载保存的位置（组件初始化时）
  useEffect(() => {
    const loadPositions = async () => {
      // 如果没有模块数据，不加载位置
      if (!modules || modules.length === 0) return;
      
      // 收集所有不同的 projectId
      const projectIds = [...new Set(modules.map(m => m.projectId).filter(Boolean))];
      if (projectIds.length === 0) return;
      
      try {
        // 为每个项目加载位置
        const allPositions: Array<{ moduleId: string; positionX: number | null; positionY: number | null }> = [];
        
        for (const projectId of projectIds) {
          const response = await modulePositionApi.getPositions(projectId);
          if (response.data?.positions) {
            allPositions.push(...response.data.positions);
          }
        }
        
        if (allPositions.length > 0) {
          setNodes(nodes => nodes.map(node => {
            // node.id 格式为 "module-xxx"
            const moduleId = node.id.replace('module-', '');
            const pos = allPositions.find(p => p.moduleId === moduleId);
            // 只有当位置不为 null 时才更新
            if (pos && pos.positionX != null && pos.positionY != null) {
              return { ...node, position: { x: pos.positionX, y: pos.positionY } };
            }
            return node;
          }));
        }
      } catch (error) {
        console.error('Failed to load positions:', error);
      }
    };
    
    loadPositions();
  }, [modules, setNodes]);

  // L 键自动布局
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'l' || e.key === 'L') {
        // 检查是否在输入框中
        const target = e.target as HTMLElement;
        if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
          return;
        }
        
        e.preventDefault();
        const newPositions = calculateAutoLayout();
        
        // 应用位置到节点
        setNodes(nodes => nodes.map(node => {
          const moduleId = node.id.replace('module-', '');
          const pos = newPositions[moduleId];
          return pos ? { ...node, position: pos } : node;
        }));
        
        // 保存到后端
        savePositions(newPositions);
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [calculateAutoLayout, setNodes, savePositions]);

  // 用 ref 追踪当前面板模式（0=关闭, 1=属性面板, 2=BugLog面板）
  // 避免 useEffect 闭包问题，直接读取 ref 值
  const panelModeRef = React.useRef<0 | 1 | 2>(0);
  // 记录上次打开的是哪个面板，用于 P 键恢复
  const lastActivePanelRef = React.useRef<1 | 2>(1);

  // 同步 panelMode 到 ref（每次渲染时更新）
  const panelMode = bugLogPanelVisible ? 2 : isPanelVisible ? 1 : 0;
  panelModeRef.current = panelMode;
  if (panelMode !== 0) lastActivePanelRef.current = panelMode;

  // P 键：有面板打开时关闭；没有面板时恢复上次打开的面板
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'p' || e.key === 'P') {
        const target = e.target as HTMLElement;
        if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
          return;
        }
        const currentMode = panelModeRef.current;
        if (currentMode === 1) {
          // 属性面板打开 → 关闭
          usePropertyPanelStore.getState().close();
        } else if (currentMode === 2) {
          // BugLog 面板打开 → 关闭
          setBugLogPanelVisible(false);
        } else {
          // 没有面板 → 恢复上次打开的面板
          const last = lastActivePanelRef.current;
          if (last === 2) {
            setBugLogPanelVisible(true);
            usePropertyPanelStore.getState().close();
          } else {
            usePropertyPanelStore.getState().showProjectInfo();
            setBugLogPanelVisible(false);
          }
        }
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // 记录拖拽开始时的位置，用于计算偏移量
  const dragStartPos = useRef<{ [key: string]: { x: number; y: number } }>({});
  
  // 节点拖拽开始时记录初始位置
  const onNodeDragStart = useCallback(
    (_: React.MouseEvent, _node: Node) => {
      // 记录所有选中节点的初始位置
      dragStartPos.current = {};
      nodes.forEach(n => {
        if (selectedModuleIds.has(n.id.replace('module-', ''))) {
          dragStartPos.current[n.id] = { ...n.position };
        }
      });
    },
    [nodes, selectedModuleIds]
  );

  // 节点拖拽过程中，同步移动其他选中的节点
  const onNodeDrag = useCallback(
    (_: React.MouseEvent, draggedNode: Node) => {
      const draggedModuleId = draggedNode.id.replace('module-', '');
      
      // 只有拖拽的节点是选中状态时，才同步移动其他选中节点
      if (!selectedModuleIds.has(draggedModuleId) || selectedModuleIds.size <= 1) {
        return;
      }
      
      const startPos = dragStartPos.current[draggedNode.id];
      if (!startPos) return;
      
      // 计算偏移量
      const deltaX = draggedNode.position.x - startPos.x;
      const deltaY = draggedNode.position.y - startPos.y;
      
      // 更新所有其他选中节点的位置
      setNodes(prevNodes => prevNodes.map(node => {
        if (node.id === draggedNode.id) return node;
        
        const moduleId = node.id.replace('module-', '');
        if (selectedModuleIds.has(moduleId) && dragStartPos.current[node.id]) {
          return {
            ...node,
            position: {
              x: dragStartPos.current[node.id].x + deltaX,
              y: dragStartPos.current[node.id].y + deltaY,
            }
          };
        }
        return node;
      }));
    },
    [selectedModuleIds, setNodes]
  );

  // 节点拖拽结束时的处理
  const onNodeDragStop = useCallback(
    (_: React.MouseEvent, _node: Node) => {
      // 保存所有选中节点的位置到后端
      selectedModuleIds.forEach(moduleId => {
        const nodeId = `module-${moduleId}`;
        const nodeEl = nodes.find(n => n.id === nodeId);
        if (nodeEl) {
          debouncedSavePosition(nodeId, nodeEl.position.x, nodeEl.position.y);
        }
      });
    },
    [debouncedSavePosition, selectedModuleIds, nodes]
  );

  const onNodeClick = useCallback(
    (event: React.MouseEvent, node: Node) => {
      if (node.id.startsWith('module-')) {
        const moduleId = node.id.replace('module-', '');
        
        // 处理多选逻辑
        if (event.shiftKey) {
          // Shift+点击：加选
          setSelectedModuleIds(prev => {
            const next = new Set(prev);
            if (next.has(moduleId)) {
              next.delete(moduleId);
            } else {
              next.add(moduleId);
            }
            console.log('[ModuleGraphView] Shift+点击，切换选中:', moduleId, '当前选中:', Array.from(next));
            return next;
          });
        } else if (event.ctrlKey || event.metaKey) {
          // Ctrl+点击：减选（如果已选中则取消，未选中则选中）
          setSelectedModuleIds(prev => {
            const next = new Set(prev);
            if (next.has(moduleId)) {
              next.delete(moduleId);
              console.log('[ModuleGraphView] Ctrl+点击，取消选中:', moduleId);
            } else {
              next.add(moduleId);
              console.log('[ModuleGraphView] Ctrl+点击，加选:', moduleId);
            }
            return next;
          });
        } else {
          // 普通点击：覆盖选择
          console.log('[ModuleGraphView] 普通点击，覆盖选中:', moduleId);
          setSelectedModuleIds(new Set([moduleId]));
          onModuleClick?.(moduleId);
          // 同步更新属性面板
          if (isPanelVisible) {
            propertyPanelStore.showModuleInfo(moduleId);
          }
        }
      }
    },
    [onModuleClick, isPanelVisible, propertyPanelStore]
  );

  // 处理选择变化（框选完成时）
  const onSelectionChange = useCallback(
    (params: OnSelectionChangeParams) => {
      const selectedNodes = params.nodes.filter(n => n.id.startsWith('module-'));
      const selectedIds = selectedNodes.map(n => n.id.replace('module-', ''));
      
      // 只有当选中节点真正变化时才更新状态
      setSelectedModuleIds(prev => {
        const prevIds = Array.from(prev);
        // 比较新旧选中ID是否相同
        if (prevIds.length === selectedIds.length &&
            prevIds.every(id => selectedIds.includes(id))) {
          return prev; // 没有变化，返回旧状态
        }
        // 有变化才更新
        if (selectedIds.length > 0) {
          console.log('[ModuleGraphView] 框选完成，选中模块:', selectedIds);
          return new Set(selectedIds);
        }
        return prev;
      });
    },
    []
  );

  // 点击空白处清除选中
  const onPaneClick = useCallback(() => {
    setSelectedModuleIds(new Set());
    // 属性面板显示项目信息
    if (isPanelVisible) {
      propertyPanelStore.showProjectInfo();
    }
  }, [isPanelVisible, propertyPanelStore]);

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
            checked={displaySettings.showTests}
            onChange={(e) => updateDisplaySettings('showTests', e.target.checked)}
          >
            🧪 测试用例
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
      <div style={{ borderTop: '1px solid #3d3d5c', paddingTop: 12, marginBottom: 16 }}>
        <div style={{ fontWeight: 600, marginBottom: 8, color: '#fff' }}>布局操作</div>
        <Button
          size="small"
          onClick={() => {
            const newPositions = calculateAutoLayout();
            setNodes(nodes => nodes.map(node => {
              const moduleId = node.id.replace('module-', '');
              const pos = newPositions[moduleId];
              return pos ? { ...node, position: pos } : node;
            }));
            savePositions(newPositions);
          }}
          style={{ width: '100%', marginBottom: 8 }}
        >
          自动布局 (快捷键 L)
        </Button>
      </div>
      <div style={{ borderTop: '1px solid #3d3d5c', paddingTop: 12 }}>
        <Button
          size="small"
          onClick={() => setDisplaySettings(storageDefaultDisplaySettings)}
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
        onNodeDragStart={onNodeDragStart}
        onNodeDrag={onNodeDrag}
        onNodeDragStop={onNodeDragStop}
        onMoveEnd={onMoveEnd}
        onSelectionChange={onSelectionChange}
        onPaneClick={onPaneClick}
        attributionPosition="bottom-left"
        style={{ background: '#0f0f23' }}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        defaultEdgeOptions={{
          type: 'flowEdge',
          style: { stroke: '#00d9ff', strokeWidth: 2 },
        }}
        elevateEdgesOnSelect={true}
        selectionMode={SelectionMode.Partial}
        selectionOnDrag={true}
        panOnDrag={[1]}
        selectNodesOnDrag={false}
        zoomOnScroll={true}
        panOnScroll={false}
        preventScrolling={true}
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
          <div ref={topRightPanelRef} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {/* 面板选择器：圆点指示器 */}
            <Tooltip
              title={
                panelMode === 0 ? '关闭面板' :
                panelMode === 1 ? '属性面板' :
                'BugLog 面板'
              }
              placement="bottom"
            >
              <div style={{
                background: '#12122a',
                border: '1px solid #3d3d5c',
                borderRadius: 12,
                padding: '5px 10px',
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                cursor: 'pointer',
              }}>
                {/* 圆点 0：关闭 */}
                <div
                  title="关闭面板"
                  style={{
                    width: panelMode === 0 ? 10 : 8,
                    height: panelMode === 0 ? 10 : 8,
                    borderRadius: '50%',
                    background: panelMode === 0 ? '#555566' : '#3d3d5c',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    flexShrink: 0,
                  }}
                  onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.width = '10px'; (e.currentTarget as HTMLDivElement).style.height = '10px'; }}
                  onMouseLeave={e => { if (panelMode !== 0) { (e.currentTarget as HTMLDivElement).style.width = '8px'; (e.currentTarget as HTMLDivElement).style.height = '8px'; } }}
                  onClick={() => handlePanelModeChange(0)}
                />
                {/* 圆点 1：属性面板 */}
                <div
                  title="属性面板"
                  style={{
                    width: panelMode === 1 ? 10 : 8,
                    height: panelMode === 1 ? 10 : 8,
                    borderRadius: '50%',
                    background: panelMode === 1 ? '#1890ff' : '#3d3d5c',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    flexShrink: 0,
                  }}
                  onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.width = '10px'; (e.currentTarget as HTMLDivElement).style.height = '10px'; }}
                  onMouseLeave={e => { if (panelMode !== 1) { (e.currentTarget as HTMLDivElement).style.width = '8px'; (e.currentTarget as HTMLDivElement).style.height = '8px'; } }}
                  onClick={() => handlePanelModeChange(panelMode === 1 ? 0 : 1)}
                />
                {/* 圆点 2：BugLog 面板 */}
                <div
                  title="BugLog 面板"
                  style={{
                    width: panelMode === 2 ? 10 : 8,
                    height: panelMode === 2 ? 10 : 8,
                    borderRadius: '50%',
                    background: panelMode === 2 ? '#ff4d4f' : '#3d3d5c',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    flexShrink: 0,
                  }}
                  onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.width = '10px'; (e.currentTarget as HTMLDivElement).style.height = '10px'; }}
                  onMouseLeave={e => { if (panelMode !== 2) { (e.currentTarget as HTMLDivElement).style.width = '8px'; (e.currentTarget as HTMLDivElement).style.height = '8px'; } }}
                  onClick={() => handlePanelModeChange(panelMode === 2 ? 0 : 2)}
                />
              </div>
            </Tooltip>
            {/* 显示设置 Popover */}
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
          </div>
        </Panel>
      </ReactFlow>

      {/* 属性面板悬浮层 - 位于显示设置按钮下方，延伸到底部 */}
      <PropertyPanel
        project={project as any}
        modules={modules}
        tasks={tasks}
        taskDependencies={taskDependencies}
        anchorRight={16}
        anchorTop={panelAnchorTop}
      />

      {/* BugLog 面板悬浮层 - 与属性面板同位置，互斥显示 */}
      {bugLogPanelVisible && (
        <BugLogPanel
          project={project as any}
          modules={modules}
          tasks={tasks}
          anchorRight={16}
          anchorTop={panelAnchorTop}
          onClose={() => setBugLogPanelVisible(false)}
        />
      )}
    </div>
  );
};

// 右键菜单配置 - 在组件外部定义避免重复创建
const moduleContextMenuItems = [
  {
    key: 'compile',
    icon: <PlayCircleOutlined />,
    label: '编译',
  },
  {
    key: 'analyze',
    icon: <ApiOutlined />,
    label: '分析',
  },
  {
    key: 'refactor',
    icon: <ReloadOutlined />,
    label: '重构',
  },
  {
    type: 'divider' as const,
  },
  {
    key: 'delete',
    icon: <DeleteOutlined />,
    label: '删除',
    danger: true,
  },
];

// 模块节点组件 - 包含任务列表
const ModuleNodeComponent: React.FC<{ data: ModuleNodeData }> = ({ data }) => {
  const { label, status, collapsed, onToggle, color, tasks, taskDependencies, onTaskClick, displaySettings, isSelected, onContextMenu } = data;

  // 获取任务的依赖信息
  const getTaskDependencyInfo = (taskId: string) => {
    const upstreamDeps = taskDependencies.filter(d => d.downstreamTaskId === taskId);
    const downstreamDeps = taskDependencies.filter(d => d.upstreamTaskId === taskId);
    return { upstreamDeps, downstreamDeps };
  };

  // 右键菜单处理
  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (onContextMenu) {
      // 从当前元素的最近父元素中找到data-id
      const nodeElement = e.currentTarget.closest('[data-id]');
      const id = nodeElement?.getAttribute('data-id')?.replace('module-', '') || '';
      onContextMenu(e, id);
    }
  };

  return (
    <Dropdown
      menu={{
        items: moduleContextMenuItems,
        onClick: ({ key }) => {
          console.log('[ModuleNodeComponent] 菜单点击:', key);
          switch (key) {
            case 'compile':
              message.info(`编译模块: ${label}`);
              break;
            case 'analyze':
              message.info(`分析模块: ${label}`);
              break;
            case 'refactor':
              message.info(`重构模块: ${label}`);
              break;
            case 'delete':
              message.warning(`删除模块: ${label}`);
              break;
          }
        },
      }}
      trigger={['contextMenu']}
    >
      <div
        onContextMenu={handleContextMenu}
        style={{
          width: '100%',
          height: '100%',
          background: color,
          border: isSelected ? '3px solid #00d9ff' : '2px solid #3d3d5c',
          borderRadius: '12px',
          padding: '0',
          color: '#ffffff',
          boxShadow: isSelected ? '0 0 20px rgba(0, 217, 255, 0.5)' : '0 4px 12px rgba(0, 0, 0, 0.4)',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          cursor: 'pointer',
        }}
      >
      {/* 模块头部 */}
      <div
        style={{
          padding: '10px 14px',
          borderBottom: collapsed ? 'none' : '1px solid rgba(255,255,255,0.1)',
          cursor: 'default',
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
        <span
          onClick={(e) => {
            e.stopPropagation();
            onToggle();
          }}
          style={{
            fontSize: '10px',
            color: '#a0a0a0',
            cursor: 'pointer',
            padding: '2px 4px',
          }}
        >
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
              void (hasUpstream || hasDownstream); // 保留逻辑供未来扩展
              
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
    </Dropdown>
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
        contract: dep.contractSummary || '未定义接口',
      };
    });
  };

  // 获取下游任务信息
  const getDownstreamTaskInfo = () => {
    return downstreamDeps.map(dep => {
      const downstreamTask = tasks.find(t => t.id === dep.downstreamTaskId);
      return {
        task: downstreamTask,
        contract: dep.contractSummary || '未定义接口',
      };
    });
  };

  // 处理任务节点点击：单击→属性面板显示任务信息，shift+点击→打开编辑窗口
  const handleTaskClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (e.shiftKey) {
      // Shift+点击：打开编辑窗口
      onTaskClick?.(task.id);
    } else {
      // 普通点击：在属性面板中显示任务信息
      usePropertyPanelStore.getState().showTaskInfo(task.id);
    }
  };

  return (
    <div
      onClick={handleTaskClick}
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
      {/* 提示标识 - 单击切换属性面板内容，shift+click 打开编辑 */}
      {displaySettings.showPrompt && task.prompt && (
        <InfoBadge
          icon="📝"
          title="提示词"
          content={task.prompt}
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
          onPanelShow={() => usePropertyPanelStore.getState().showContent('prompt', task.id)}
          onEditClick={() => onTaskClick?.(task.id)}
        />
      )}
      {displaySettings.showTests && task.tests && (
        <InfoBadge
          icon="🧪"
          title="测试用例"
          content={task.tests}
          type="tests"
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
          onPanelShow={() => usePropertyPanelStore.getState().showContent('tests', task.id)}
          onEditClick={() => onTaskClick?.(task.id)}
        />
      )}
      {displaySettings.showError && task.bugLog && (
        parseBugLog(task.bugLog) ? (
          <BugLogDisplay
            bugLogStr={task.bugLog}
            fontSize={displaySettings.fontSize}
            requireAltForTooltip={displaySettings.requireAltForTooltip}
            tooltipScale={displaySettings.tooltipScale}
          />
        ) : (
          <InfoBadge
            icon="❌"
            title="错误信息"
            content={task.bugLog}
            type="error"
            fontSize={displaySettings.fontSize}
            requireAltForTooltip={displaySettings.requireAltForTooltip}
            tooltipScale={displaySettings.tooltipScale}
            onPanelShow={() => usePropertyPanelStore.getState().showContent('error', task.id)}
            onEditClick={() => onTaskClick?.(task.id)}
          />
        )
      )}
      {/* 上游依赖箭头 - 黄色，单击切换属性面板上游契约内容 */}
      {hasUpstream && (
        <DependencyBadge
          type="upstream"
          count={upstreamDeps.length}
          taskInfo={getUpstreamTaskInfo()}
          contractDetail={task.upstreamContractDetail}
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
          onPanelShow={() => usePropertyPanelStore.getState().showContent('upstream-contract', task.id)}
        />
      )}
      {/* 下游依赖箭头 - 蓝色，单击切换属性面板下游契约内容 */}
      {hasDownstream && (
        <DependencyBadge
          type="downstream"
          count={downstreamDeps.length}
          taskInfo={getDownstreamTaskInfo()}
          contractDetail={task.downstreamContractDetail}
          fontSize={displaySettings.fontSize}
          requireAltForTooltip={displaySettings.requireAltForTooltip}
          tooltipScale={displaySettings.tooltipScale}
          onPanelShow={() => usePropertyPanelStore.getState().showContent('downstream-contract', task.id)}
        />
      )}
      {displaySettings.showStatus && (
        <Tooltip
          title={
            <div>
              {task.issueDetails && (
                <div style={{ marginBottom: 8 }}>
                  <div style={{ fontWeight: 600, color: '#ff6b6b' }}>⚠️ 问题详情</div>
                  <div>{task.issueDetails}</div>
                </div>
              )}
              {task.bugLog && (
                <div>
                  <div style={{ fontWeight: 600, color: '#ffa500' }}>🐛 Bug 日志</div>
                  <div style={{ whiteSpace: 'pre-wrap' }}>{task.bugLog}</div>
                </div>
              )}
              {!task.issueDetails && !task.bugLog && '无问题记录'}
            </div>
          }
        >
          <span style={{
            fontSize: `${displaySettings.fontSize - 3}px`,
            padding: '1px 6px',
            background: 'rgba(255,255,255,0.2)',
            borderRadius: '3px',
            flexShrink: 0,
          }}>
            {task.status}
          </span>
        </Tooltip>
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

// 信息标识组件 - 单击切换属性面板内容，悬停显示详细信息（使用Ant Design Tooltip）
const InfoBadge: React.FC<{
  icon: string;
  title: string;
  content: string;
  type?: 'default' | 'error' | 'warning' | 'success' | 'tests';
  subItems?: { label: string; value: string }[];
  fontSize?: number;
  requireAltForTooltip?: boolean;
  tooltipScale?: number;
  onPanelShow?: () => void;   // 单击：切换属性面板显示对应内容
  onEditClick?: () => void;   // shift+点击：打开编辑窗口
}> = ({ icon, title, content, type = 'default', subItems, fontSize = 12, requireAltForTooltip = true, tooltipScale = 1, onPanelShow, onEditClick }) => {
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
      case 'tests': return '#a855f7';
      default: return '#00d9ff';
    }
  };

  const color = getColor();
  const scaledFontSize = Math.round(fontSize * tooltipScale);
  const tooltipWidth = Math.round(350 * tooltipScale);

  // 解析测试用例内容 - MCP格式 {target, api}
  const renderContent = () => {
    if (type === 'tests') {
      try {
        const tests = JSON.parse(content);
        if (Array.isArray(tests)) {
          return (
            <div style={{
              background: `${color}15`,
              padding: 10,
              borderRadius: 4,
              fontSize: scaledFontSize,
              color: '#ccc',
              lineHeight: 1.5,
              maxHeight: 300,
              overflowY: 'auto',
            }}>
              {tests.map((test, idx) => (
                <div key={idx} style={{
                  background: '#2d2d44',
                  padding: 8,
                  borderRadius: 4,
                  marginBottom: idx < tests.length - 1 ? 8 : 0,
                }}>
                  <div style={{ color: '#fff', fontWeight: 600, marginBottom: 4 }}>
                    🧪 {test.target || test.name || '未命名测试'}
                  </div>
                  <div style={{
                    color: '#a855f7',
                    fontSize: scaledFontSize - 1,
                    fontFamily: 'monospace',
                    background: '#1e1e2e',
                    padding: '4px 8px',
                    borderRadius: 4,
                  }}>
                    {test.api || test.expected || '未定义API'}
                  </div>
                </div>
              ))}
            </div>
          );
        }
      } catch {
        // 如果解析失败，显示原始内容
      }
    }
    return (
      <div style={{
        background: `${color}15`,
        padding: 10,
        borderRadius: 4,
        fontSize: scaledFontSize,
        color: '#ccc',
        whiteSpace: 'pre-wrap',
        lineHeight: 1.5,
      }}>
        {content}
      </div>
    );
  };

  const tooltipContent = (
    <div style={{ width: tooltipWidth }}>
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
      {renderContent()}
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

  // 处理图标点击：单击→切换属性面板，shift+click→打开编辑窗口
  const handleIconClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (e.shiftKey) {
      onEditClick?.();
    } else {
      onPanelShow?.();
    }
  };

  // 如果需要Alt键但Alt未按下，不显示Tooltip
  if (requireAltForTooltip && !isAltPressed) {
    return (
      <span
        style={{
          fontSize: `${fontSize - 2}px`,
          cursor: 'pointer',
          opacity: 0.7,
        }}
        onClick={handleIconClick}
        title={`单击查看${title}，Shift+单击打开编辑`}
      >
        {icon}
      </span>
    );
  }

  return (
    <Tooltip
      title={tooltipContent}
      color="#1a1a2e"
      overlayInnerStyle={{ padding: 16 * tooltipScale, width: tooltipWidth + 32 * tooltipScale }}
      mouseEnterDelay={0}
      mouseLeaveDelay={0.1}
      open={showTooltip}
      onOpenChange={(visible) => {
        if (!requireAltForTooltip || isAltPressed) {
          setShowTooltip(visible);
        }
      }}
    >
      <span
        style={{
          fontSize: `${fontSize - 2}px`,
          cursor: 'pointer',
        }}
        onClick={handleIconClick}
        title={`单击查看${title}，Shift+单击打开编辑`}
      >
        {icon}
      </span>
    </Tooltip>
  );
};

// 解析契约详情 JSON
interface ContractDetailItem {
  label: string;
  contract_api: string;
  from?: string;
}

interface ContractDetail {
  title: string;
  list: ContractDetailItem[];
}

const parseContractDetailJSON = (jsonStr: string | undefined): ContractDetail | null => {
  if (!jsonStr) return null;
  try {
    return JSON.parse(jsonStr);
  } catch {
    return null;
  }
};

// 依赖箭头组件 - 单击切换属性面板内容，带Tooltip显示上下游接口信息
const DependencyBadge: React.FC<{
  type: 'upstream' | 'downstream';
  count: number;
  taskInfo: { task?: Task; contract: string }[];
  contractDetail?: string;
  fontSize?: number;
  requireAltForTooltip?: boolean;
  tooltipScale?: number;
  onPanelShow?: () => void;  // 单击：切换属性面板显示契约内容
}> = ({ type, count, taskInfo: _taskInfo, contractDetail, fontSize = 12, requireAltForTooltip = true, tooltipScale = 1, onPanelShow }) => {
  const [isAltPressed, setIsAltPressed] = React.useState(false);
  void React.useState(false); // tooltip state reserved for future use

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
  const title = isUpstream ? '上游契约接口' : '下游契约接口';
  const arrow = isUpstream ? '↑' : '↓';

  const scaledFontSize = Math.round(fontSize * tooltipScale);
  const tooltipWidth = Math.round(350 * tooltipScale);

  // 解析契约详情
  const contractData = parseContractDetailJSON(contractDetail);

  const tooltipContent = (
    <div style={{ width: tooltipWidth }}>
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
      
      {/* 契约描述 - 显示title */}
      {contractData?.title && (
        <div style={{
          color: '#aaa',
          fontSize: scaledFontSize,
          marginBottom: 12,
          fontStyle: 'italic',
        }}>
          {contractData.title}
        </div>
      )}
      
      {/* 接口列表 */}
      {contractData?.list && contractData.list.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <div style={{ fontSize: scaledFontSize, color: '#888', marginBottom: 8, fontWeight: 500 }}>
            接口列表：
          </div>
          {contractData.list.map((item, idx) => (
            <div key={idx} style={{
              background: `${color}15`,
              padding: 10,
              borderRadius: 4,
              marginBottom: 6,
              borderLeft: `3px solid ${color}`,
            }}>
              <div style={{ color: '#e0e0e0', fontSize: scaledFontSize, marginBottom: 6 }}>
                {item.label}
              </div>
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                flexWrap: 'wrap',
              }}>
                <span style={{
                  color: color,
                  fontSize: scaledFontSize - 1,
                  fontFamily: 'monospace',
                  background: '#1a1a2e',
                  padding: '4px 8px',
                  borderRadius: 3,
                }}>
                  {item.contract_api}
                </span>
                {item.from && (
                  <>
                    <span style={{ color: '#666', fontSize: scaledFontSize - 2 }}>|</span>
                    <span style={{
                      color: '#a78bfa',
                      fontSize: scaledFontSize - 2,
                      background: '#2d2d44',
                      padding: '2px 6px',
                      borderRadius: 3,
                    }}>
                      {item.from}
                    </span>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
      
      {/* 如果没有解析成功，显示原始内容 */}
      {!contractData && contractDetail && (
        <div style={{
          background: `${color}15`,
          padding: 10,
          borderRadius: 4,
          fontSize: scaledFontSize,
          color: '#ccc',
          whiteSpace: 'pre-wrap',
          lineHeight: 1.5,
          marginBottom: 8,
        }}>
          {contractDetail}
        </div>
      )}
    </div>
  );

  // 处理点击：切换属性面板显示契约内容
  const handleBadgeClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onPanelShow?.();
  };

  // 如果需要Alt键但Alt未按下，显示提示
  if (requireAltForTooltip && !isAltPressed) {
    return (
      <Tooltip
        title={<span style={{ fontSize: scaledFontSize }}>按住 Alt 键查看详情，单击查看属性面板</span>}
        color="#1a1a2e"
        overlayInnerStyle={{ padding: 8 * tooltipScale }}
        mouseEnterDelay={0}
        mouseLeaveDelay={0.1}
      >
        <span
          style={{
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
          }}
          onClick={handleBadgeClick}
        >
          {arrow}[{count}]
        </span>
      </Tooltip>
    );
  }

  return (
    <Tooltip
      title={tooltipContent}
      color="#1a1a2e"
      overlayInnerStyle={{ padding: 16 * tooltipScale, width: tooltipWidth + 32 * tooltipScale }}
      mouseEnterDelay={0}
      mouseLeaveDelay={0.1}
    >
      <span
        style={{
          fontSize: `${fontSize - 3}px`,
          padding: '1px 5px',
          background: bgColor,
          color: textColor,
          borderRadius: '3px',
          flexShrink: 0,
          cursor: 'pointer',
          fontWeight: 600,
          fontFamily: 'monospace',
        }}
        onClick={handleBadgeClick}
      >
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

// 包装组件，提供 ReactFlowProvider
const ModuleGraphViewWrapper: React.FC<ModuleGraphViewProps> = (props) => {
  return (
    <ReactFlowProvider>
      <ModuleGraphView {...props} />
    </ReactFlowProvider>
  );
};

export default ModuleGraphViewWrapper;
