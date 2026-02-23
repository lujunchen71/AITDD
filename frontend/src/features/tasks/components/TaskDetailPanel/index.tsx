import React from 'react';
import { Drawer, Tabs, Descriptions, Tag, Input, Spin } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface TaskDetailPanelProps {
  taskId: string | null;
  onClose: () => void;
}

const statusColors: Record<string, string> = {
  ready: 'blue',
  in_progress: 'orange',
  done: 'green',
  blocked: 'red',
};

const statusLabels: Record<string, string> = {
  ready: '就绪',
  in_progress: '进行中',
  done: '已完成',
  blocked: '阻塞',
};

const TaskDetailPanel: React.FC<TaskDetailPanelProps> = ({ taskId, onClose }) => {
  const { data, isLoading } = useQuery({
    queryKey: ['task', taskId],
    queryFn: async () => {
      if (!taskId) return null;
      const response = await apiClient.get(`/api/v1/tasks/${taskId}`);
      return response.data;
    },
    enabled: !!taskId,
  });

  const task = data?.task;

  return (
    <Drawer
      open={!!taskId}
      title={task?.name || '任务详情'}
      onClose={onClose}
      width={600}
      placement="right"
    >
      {isLoading ? (
        <div className="flex justify-center items-center h-64">
          <Spin />
        </div>
      ) : task ? (
        <Tabs
          items={[
            {
              key: 'basic',
              label: '基本信息',
              children: (
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="状态">
                    <Tag color={statusColors[task.status] || 'default'}>
                      {statusLabels[task.status] || task.status}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="分配给">
                    {task.assignee || <span className="text-gray-400">未分配</span>}
                  </Descriptions.Item>
                  <Descriptions.Item label="版本">v{task.version}</Descriptions.Item>
                  <Descriptions.Item label="描述">
                    {task.description || <span className="text-gray-400">无描述</span>}
                  </Descriptions.Item>
                  <Descriptions.Item label="创建时间">
                    {new Date(task.createdAt).toLocaleString()}
                  </Descriptions.Item>
                  <Descriptions.Item label="更新时间">
                    {new Date(task.updatedAt).toLocaleString()}
                  </Descriptions.Item>
                </Descriptions>
              ),
            },
            {
              key: 'prompt',
              label: 'Prompt',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.prompt || ''}
                  rows={10}
                  placeholder="无Prompt"
                />
              ),
            },
            {
              key: 'tests',
              label: '测试',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.tests || ''}
                  rows={10}
                  placeholder="无测试内容"
                />
              ),
            },
            {
              key: 'logs',
              label: '日志',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.logs || ''}
                  rows={10}
                  placeholder="无日志"
                  className="font-mono"
                />
              ),
            },
            {
              key: 'assistance',
              label: '人类协助',
              children: (
                <Input.TextArea
                  readOnly
                  value={task.humanAssistance || ''}
                  rows={10}
                  placeholder="无协助记录"
                />
              ),
            },
          ]}
        />
      ) : (
        <div className="text-gray-400">未找到任务</div>
      )}
    </Drawer>
  );
};

export default TaskDetailPanel;
