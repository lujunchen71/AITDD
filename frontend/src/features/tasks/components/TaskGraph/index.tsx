import React, { useCallback } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';
import TaskNode from './TaskNode';

interface TaskGraphProps {
  moduleId: string;
  onTaskSelect?: (taskId: string) => void;
}

const nodeTypes = {
  task: TaskNode,
};

const TaskGraph: React.FC<TaskGraphProps> = ({ moduleId, onTaskSelect }) => {
  // 获取模块的任务
  const { data: tasksData } = useQuery({
    queryKey: ['tasks', moduleId],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/tasks', {
        params: { moduleId, pageSize: 100 },
      });
      return response.data;
    },
  });

  // 获取依赖关系
  const { data: depsData } = useQuery({
    queryKey: ['dependencies', moduleId],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/dependencies', {
        params: { moduleId },
      });
      return response.data;
    },
  });

  // 转换为节点和边
  const initialNodes: Node[] = (tasksData?.tasks || []).map((task: any, index: number) => ({
    id: task.id,
    type: 'task',
    position: { x: (index % 5) * 200, y: Math.floor(index / 5) * 150 },
    data: {
      label: task.name,
      status: task.status,
      onClick: () => onTaskSelect?.(task.id),
    },
  }));

  const initialEdges: Edge[] = (depsData?.dependencies || []).map((dep: any) => ({
    id: dep.id,
    source: dep.moduleId,
    target: dep.dependsOn,
    animated: true,
    style: { stroke: '#e94560' },
  }));

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  return (
    <div className="task-graph" style={{ width: '100%', height: '500px' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        fitView
      >
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  );
};

export default TaskGraph;
