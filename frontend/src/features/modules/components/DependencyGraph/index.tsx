import React, { useCallback, useMemo } from 'react';
import ReactFlow, {
  Node,
  Edge,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  MarkerType,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Module, ModuleDependency } from '../../../../types';

interface DependencyGraphProps {
  modules?: Module[];
  moduleDependencies?: ModuleDependency[];
  onNodeClick?: (node: Node) => void;
}

interface GraphNodeData {
  label: string;
  status?: string;
  type: 'module';
  color: string;
}

const DependencyGraph: React.FC<DependencyGraphProps> = ({
  modules = [],
  moduleDependencies = [],
  onNodeClick,
}) => {
  const { nodes: initialNodes, edges: initialEdges } = useMemo(() => {
    // 模块模式
    const moduleNodes: Node<GraphNodeData>[] = modules.map((module) => ({
      id: module.id,
      type: 'default',
      position: { x: 0, y: 0 },
      data: {
        label: module.name,
        status: module.status,
        type: 'module',
        color: getModuleColor(module.status),
      },
      style: {
        background: getModuleColor(module.status),
        border: '1px solid #3d3d5c',
        borderRadius: '8px',
        padding: '12px 16px',
        minWidth: '150px',
        color: '#ffffff',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)',
      },
    }));

    const moduleEdges: Edge[] = moduleDependencies.map((dep) => ({
      id: dep.id,
      source: dep.dependsOnModuleId,
      target: dep.moduleId,
      type: 'bezier',
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: '#e94560',
      },
      style: {
        stroke: '#e94560',
        strokeWidth: 2,
      },
      label: dep.dependencyType,
      labelStyle: {
        fill: '#e94560',
        fontSize: 10,
      },
    }));

    return { nodes: moduleNodes, edges: moduleEdges };
  }, [modules, moduleDependencies]);

  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, , onEdgesChange] = useEdgesState(initialEdges);

  const onNodeClickHandler = useCallback(
    (_: React.MouseEvent, node: Node) => {
      if (onNodeClick) {
        onNodeClick(node);
      }
    },
    [onNodeClick]
  );

  return (
    <div style={{ width: '100%', height: '100%', background: '#0f0f23' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClickHandler}
        fitView
        attributionPosition="bottom-left"
        style={{ background: '#0f0f23' }}
        defaultEdgeOptions={{
          type: 'bezier',
          style: { stroke: '#e94560', strokeWidth: 2 },
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
      </ReactFlow>
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

export default DependencyGraph;
