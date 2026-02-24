import React, { useState } from 'react';
import { Card, Descriptions, Tag, Progress, Empty, Spin, Button, Tabs, Table, Space, Popconfirm, message } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined, FileTextOutlined, FolderOutlined } from '@ant-design/icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';
import type { ColumnsType } from 'antd/es/table';

interface ModuleDetailProps {
  moduleId: string | null;
  onEdit?: (moduleId: string) => void;
  onDelete?: (moduleId: string) => void;
  onAddTask?: (moduleId: string) => void;
  onEditTask?: (taskId: string, moduleId: string) => void;
}

const statusColors: Record<string, string> = {
  designing: 'blue',
  developing: 'orange',
  testing: 'cyan',
  done: 'green',
};

const statusLabels: Record<string, string> = {
  designing: '设计中',
  developing: '开发中',
  testing: '测试中',
  done: '已完成',
};

const taskStatusColors: Record<string, string> = {
  ready: 'default',
  in_progress: 'processing',
  done: 'success',
  blocked: 'error',
};

const taskStatusLabels: Record<string, string> = {
  ready: '就绪',
  in_progress: '进行中',
  done: '已完成',
  blocked: '阻塞',
};

const ModuleDetail: React.FC<ModuleDetailProps> = ({
  moduleId,
  onEdit,
  onDelete,
  onAddTask,
  onEditTask,
}) => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('info');

  // 获取模块详情
  const { data: moduleData, isLoading: moduleLoading, error: moduleError } = useQuery({
    queryKey: ['module', moduleId],
    queryFn: async () => {
      if (!moduleId) return null;
      const response = await apiClient.get(`/modules/${moduleId}`);
      return response.data?.module;
    },
    enabled: !!moduleId,
  });

  // 获取模块内的任务列表
  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['moduleTasks', moduleId],
    queryFn: async () => {
      if (!moduleId) return { tasks: [], total: 0 };
      const response = await apiClient.get(`/modules/${moduleId}/tasks`);
      return response.data || { tasks: [], total: 0 };
    },
    enabled: !!moduleId && activeTab === 'tasks',
  });

  // 删除任务
  const handleDeleteTask = async (taskId: string) => {
    try {
      await apiClient.delete(`/tasks/${taskId}`);
      message.success('任务删除成功');
      queryClient.invalidateQueries({ queryKey: ['moduleTasks', moduleId] });
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '删除任务失败';
      message.error(errorMsg);
    }
  };

  const taskColumns: ColumnsType<any> = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <a onClick={() => onEditTask?.(record.id, moduleId!)}>{text}</a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => (
        <Tag color={taskStatusColors[status] || 'default'}>
          {taskStatusLabels[status] || status}
        </Tag>
      ),
    },
    {
      title: '分配给',
      dataIndex: 'assignee',
      key: 'assignee',
      width: 120,
      render: (assignee) => assignee || <span className="text-gray-400">未分配</span>,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 80,
      render: (v) => `v${v}`,
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
      width: 180,
      render: (time) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => onEditTask?.(record.id, moduleId!)}
          />
          <Popconfirm
            title="确定删除此任务？"
            onConfirm={() => handleDeleteTask(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  if (!moduleId) {
    return (
      <Card>
        <Empty description="请选择一个模块查看详情" />
      </Card>
    );
  }

  if (moduleLoading) {
    return (
      <Card>
        <div className="flex justify-center items-center h-64">
          <Spin />
        </div>
      </Card>
    );
  }

  if (moduleError || !moduleData?.module) {
    return (
      <Card>
        <Empty description="加载模块失败" />
      </Card>
    );
  }

  const module = moduleData.module;

  const tabItems = [
    {
      key: 'info',
      label: (
        <span>
          <FolderOutlined />
          模块信息
        </span>
      ),
      children: (
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="状态">
            <Tag color={statusColors[module.status] || 'default'}>
              {statusLabels[module.status] || module.status}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="版本">
            v{module.version}
          </Descriptions.Item>
          <Descriptions.Item label="测试覆盖率" span={2}>
            {module.testCoverage !== undefined ? (
              <Progress percent={module.testCoverage} size="small" />
            ) : (
              <span className="text-gray-400">未设置</span>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>
            {module.description || <span className="text-gray-400">无描述</span>}
          </Descriptions.Item>
          <Descriptions.Item label="Prompt" span={2}>
            <pre className="whitespace-pre-wrap text-sm bg-gray-50 p-2 rounded">
              {module.prompt || <span className="text-gray-400">无Prompt</span>}
            </pre>
          </Descriptions.Item>
          <Descriptions.Item label="上游契约摘要" span={2}>
            {module.upstreamContractSummary || <span className="text-gray-400">无</span>}
          </Descriptions.Item>
          <Descriptions.Item label="下游契约摘要" span={2}>
            {module.downstreamContractSummary || <span className="text-gray-400">无</span>}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {new Date(module.createdAt).toLocaleString()}
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            {new Date(module.updatedAt).toLocaleString()}
          </Descriptions.Item>
        </Descriptions>
      ),
    },
    {
      key: 'tasks',
      label: (
        <span>
          <FileTextOutlined />
          任务列表
          {tasksData?.total !== undefined && (
            <span className="ml-1 text-gray-400">({tasksData.total})</span>
          )}
        </span>
      ),
      children: (
        <div>
          <div className="mb-4 flex justify-end">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => onAddTask?.(moduleId)}
            >
              新建任务
            </Button>
          </div>
          <Table
            columns={taskColumns}
            dataSource={tasksData?.tasks || []}
            rowKey="id"
            loading={tasksLoading}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            locale={{
              emptyText: <Empty description="暂无任务，点击上方按钮创建" />,
            }}
          />
        </div>
      ),
    },
  ];

  return (
    <Card
      title={module.name}
      extra={
        <div className="flex gap-2">
          {onEdit && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => onEdit(moduleId)}
            >
              编辑
            </Button>
          )}
          {onDelete && (
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => onDelete(moduleId)}
            >
              删除
            </Button>
          )}
        </div>
      }
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
      />
    </Card>
  );
};

export default ModuleDetail;
