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
  ready: '#808080',
  in_progress: '#3b82f6',
  done: '#10b981',
  blocked: '#ef4444',
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

  const { data, isLoading, error } = useQuery({
    queryKey: ['tasks', moduleId, page, pageSize, statusFilter],
    queryFn: async () => {
      const params: any = { page, pageSize };
      if (moduleId) params.moduleId = moduleId;
      if (statusFilter) params.status = statusFilter;

      const response = await apiClient.get('/tasks', { params });
      return response.data;
    },
  });

  const columns: ColumnsType<any> = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text, record) => (
        <a onClick={() => onEdit?.(record.id)} style={{ color: '#00d9ff' }}>{text}</a>
      ),
      fixed: 'left',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status) => (
        <Tag style={{ background: statusColors[status] || '#404040', border: 'none', color: '#fff' }}>
          {statusLabels[status] || status}
        </Tag>
      ),
    },
    {
      title: '分配给',
      dataIndex: 'assignee',
      key: 'assignee',
      width: 100,
      render: (assignee) => assignee || <span style={{ color: '#666' }}>未分配</span>,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 60,
      render: (v) => <span style={{ color: '#a0a0a0' }}>v{v}</span>,
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
      width: 150,
      render: (time) => <span style={{ color: '#a0a0a0' }}>{new Date(time).toLocaleString()}</span>,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<EditOutlined style={{ color: '#00d9ff' }} />}
            onClick={() => onEdit?.(record.id)}
          />
          <Button
            type="text"
            size="small"
            icon={<DeleteOutlined style={{ color: '#ff4d4f' }} />}
            onClick={() => onDelete?.(record.id)}
          />
        </Space>
      ),
    },
  ];

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}>
        <Spin />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '16px', color: '#ff4d4f', textAlign: 'center' }}>
        加载任务失败
      </div>
    );
  }

  const tasks = data?.tasks || [];
  const total = data?.total || 0;

  const filteredTasks = searchText
    ? tasks.filter((t: any) => t.name.toLowerCase().includes(searchText.toLowerCase()))
    : tasks;

  return (
    <div style={{ padding: '16px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <Space>
          <Input
            placeholder="搜索任务"
            prefix={<SearchOutlined style={{ color: '#666' }} />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ 
              width: 200, 
              background: '#1a1a2e',
              border: '1px solid #2d2d44',
              color: '#e0e0e0',
            }}
            allowClear
          />
          <Select
            placeholder="状态筛选"
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ 
              width: 120,
              background: '#1a1a2e',
              border: '1px solid #2d2d44',
            }}
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
          <Button 
            type="primary" 
            icon={<PlusOutlined />} 
            onClick={onAdd}
            style={{
              background: 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)',
              border: 'none',
            }}
          >
            新建任务
          </Button>
        )}
      </div>

      {filteredTasks.length === 0 ? (
        <Empty 
          description="暂无任务" 
          imageStyle={{ opacity: 0.5 }}
          style={{ padding: '40px 0' }}
        />
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
          scroll={{ x: 600 }}
          style={{ 
            background: '#16213e',
            borderRadius: '8px',
            overflow: 'hidden',
          }}
        />
      )}
    </div>
  );
};

export default TaskList;
