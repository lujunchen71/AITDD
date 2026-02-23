import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Tag } from 'antd';

interface TaskNodeData {
  label: string;
  status: string;
  onClick?: () => void;
}

const statusColors: Record<string, string> = {
  ready: '#1890ff',
  in_progress: '#fa8c16',
  done: '#52c41a',
  blocked: '#f5222d',
};

const statusLabels: Record<string, string> = {
  ready: '就绪',
  in_progress: '进行中',
  done: '已完成',
  blocked: '阻塞',
};

const TaskNode: React.FC<NodeProps<TaskNodeData>> = ({ data }) => {
  return (
    <div
      className="task-node"
      style={{
        padding: '10px 15px',
        borderRadius: '8px',
        background: '#16213e',
        border: `2px solid ${statusColors[data.status] || '#1890ff'}`,
        minWidth: '120px',
        cursor: 'pointer',
      }}
      onClick={data.onClick}
    >
      <Handle type="target" position={Position.Top} />
      
      <div className="task-node-content">
        <div className="task-node-label" style={{ color: '#fff', marginBottom: '5px' }}>
          {data.label}
        </div>
        <Tag color={statusColors[data.status] || 'default'}>
          {statusLabels[data.status] || data.status}
        </Tag>
      </div>
      
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
};

export default memo(TaskNode);
