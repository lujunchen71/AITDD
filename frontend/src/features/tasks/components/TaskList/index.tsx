import React from 'react';
import { Table, Tag, Button, Space, Empty, Spin, Input, Select } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined, SearchOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import type { ColumnsType } from 'antd/es/table';
import { apiClient } from '../../../../services/api';

interface TaskListProps {
  moduleId?: string;
  onEdit?: (taskId: string) => void;
  onDelete?: (taskId: string) => void;
  onAdd?: () => void;
}

const statusColors: Record<string, string> = {
  ready: 'default',
  in_progress: 'processing',
  done: 'success',
  blocked: 'error',
};

const statusLabels: Record<string, string> = {
  ready: '就绪',
  in_progress: '进行中',
  done: '已完成',
  blocked: '阻塞',
};

const TaskList: React.FC<TaskListProps> = ({
  moduleId,
  onEdit,
  onDelete,
  onAdd,
}) => {
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(10);
  const [statusFilter, setStatusFilter] = React.useState<string>('');
  const [searchText, setSearchText] = React.useState('');

  // 获取任务列表
  const { data, isLoading, error } = useQuery({
    queryKey: ['tasks', moduleId, page, pageSize, statusFilter],
    queryFn: async () => {
      const params: any = { page, pageSize };
      if (moduleId) params.moduleId = moduleId;
      if (statusFilter) params.status = statusFilter;

      const response = await apiClient.get('/api/v1/tasks', { params });
      return response.data;
    },
  });

  const columns: ColumnsType<any> = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <a onClick={() => onEdit?.(record.id)}>{text}</a>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => (
        <Tag color={statusColors[status] || 'default'}>
          {statusLabels[status] || status}
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
            onClick={() => onEdit?.(record.id)}
          />
          <Button
            type="text"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => onDelete?.(record.id)}
          />
        </Space>
      ),
    },
  ];

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 text-red-500">
        加载任务失败
      </div>
    );
  }

  const tasks = data?.tasks || [];
  const total = data?.total || 0;

  // 过滤搜索文本
  const filteredTasks = searchText
    ? tasks.filter((t: any) => t.name.toLowerCase().includes(searchText.toLowerCase()))
    : tasks;

  return (
    <div className="task-list">
      <div className="task-list-header flex justify-between items-center mb-4">
        <Space>
          <Input
            placeholder="搜索任务"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 200 }}
            allowClear
          />
          <Select
            placeholder="状态筛选"
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            allowClear
            options={[
              { value: 'ready', label: '就绪' },
              { value: 'in_progress', label: '进行中' },
              { value: 'done', label: '已完成' },
              { value: 'blocked', label: '阻塞' },
            ]}
          />
        </Space>
        {onAdd && (
          <Button type="primary" icon={<PlusOutlined />} onClick={onAdd}>
            新建任务
          </Button>
        )}
      </div>

      {filteredTasks.length === 0 ? (
        <Empty description="暂无任务" />
      ) : (
        <Table
          columns={columns}
          dataSource={filteredTasks}
          rowKey="id"
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (p, ps) => {
              setPage(p);
              setPageSize(ps);
            },
          }}
        />
      )}
    </div>
  );
};

export default TaskList;
