import React, { useState } from 'react';
import { Descriptions, Tag, Progress, Empty, Spin, Button, Tabs, Table, message, Dropdown, Modal } from 'antd';
import { EditOutlined, DeleteOutlined, PlusOutlined, FileTextOutlined, FolderOutlined, MoreOutlined, CheckCircleOutlined, ReloadOutlined, CopyOutlined, LockOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';
import type { ColumnsType } from 'antd/es/table';
import type { MenuProps } from 'antd';

const { confirm } = Modal;

interface ModuleDetailProps {
  moduleId: string | null;
  onEdit?: (moduleId: string) => void;
  onDelete?: (moduleId: string) => void;
  onAddTask?: (moduleId: string) => void;
  onEditTask?: (taskId: string, moduleId: string) => void;
}

const statusColors: Record<string, string> = {
  designing: '#3b82f6',
  developing: '#f59e0b',
  testing: '#06b6d4',
  done: '#10b981',
};

const statusLabels: Record<string, string> = {
  designing: '设计中',
  developing: '开发中',
  testing: '测试中',
  done: '已完成',
};

const taskStatusColors: Record<string, string> = {
  ready: '#808080',
  in_progress: '#3b82f6',
  done: '#10b981',
  blocked: '#ef4444',
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
  const [activeTab, setActiveTab] = useState('tasks');

  const { data: moduleData, isLoading: moduleLoading, error: moduleError } = useQuery({
    queryKey: ['module', moduleId],
    queryFn: async () => {
      if (!moduleId) return null;
      try {
        const response = await apiClient.get(`/modules/${moduleId}`);
        // API 响应格式：{success: true, data: {module: {...}}, timestamp: ...}
        // 添加空值检查，防止 undefined
        return response.data?.data?.module || null;
      } catch (error) {
        console.error('获取模块详情失败:', error);
        throw error;
      }
    },
    enabled: !!moduleId,
    retry: 1,
  });

  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['moduleTasks', moduleId],
    queryFn: async () => {
      if (!moduleId) return { tasks: [], total: 0 };
      try {
        const response = await apiClient.get(`/modules/${moduleId}/tasks`);
        console.log('ModuleDetail tasks response:', response);
        // API 响应格式：{success: true, data: {tasks: [], total: 0}} 或 {success: true, data: {data: {tasks: []}}}
        // 添加空值检查，防止 undefined
        const responseData = response.data?.data;
        if (!responseData) {
          return { tasks: [], total: 0 };
        }
        return responseData?.tasks ? responseData : { tasks: responseData?.data?.tasks || [], total: responseData?.total || 0 };
      } catch (error) {
        console.error('获取任务列表失败:', error);
        return { tasks: [], total: 0 };
      }
    },
    enabled: !!moduleId && activeTab === 'tasks',
    retry: 1,
  });

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

  // 检查任务
  const handleCheckTask = async (taskId: string) => {
    try {
      const response = await apiClient.post(`/tasks/${taskId}/check`);
      if (response.data.success) {
        message.success('任务检查通过');
      } else {
        message.warning(response.data.message || '任务检查发现问题');
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '检查失败';
      message.error(errorMsg);
    }
  };

  // 重构任务
  const handleRefactorTask = async (taskId: string) => {
    confirm({
      title: '确认重构',
      icon: <ExclamationCircleOutlined />,
      content: '确定要重构这个任务吗？这将重新分析任务的结构和依赖。',
      okText: '重构',
      cancelText: '取消',
      onOk: async () => {
        try {
          await apiClient.post(`/tasks/${taskId}/refactor`);
          message.success('任务重构成功');
          queryClient.invalidateQueries({ queryKey: ['moduleTasks', moduleId] });
        } catch (error: any) {
          const errorMsg = error?.response?.data?.error?.message || '重构失败';
          message.error(errorMsg);
        }
      },
    });
  };

  // 复制任务
  const handleDuplicateTask = async (taskId: string) => {
    try {
      await apiClient.post(`/tasks/${taskId}/duplicate`);
      message.success('任务复制成功');
      queryClient.invalidateQueries({ queryKey: ['moduleTasks', moduleId] });
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '复制失败';
      message.error(errorMsg);
    }
  };

  // 锁定/解锁任务
  const handleLockTask = async (taskId: string) => {
    try {
      const response = await apiClient.post(`/tasks/${taskId}/toggle-lock`);
      message.success(response.data.locked ? '任务已锁定' : '任务已解锁');
      queryClient.invalidateQueries({ queryKey: ['moduleTasks', moduleId] });
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error?.message || '操作失败';
      message.error(errorMsg);
    }
  };

  const taskColumns: ColumnsType<any> = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text, record) => (
        <a onClick={() => onEditTask?.(record.id, moduleId!)} style={{ color: '#00d9ff' }}>{text}</a>
      ),
      fixed: 'left',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status) => (
        <Tag style={{ background: taskStatusColors[status] || '#404040', border: 'none', color: '#0d3d80' }}>
          {taskStatusLabels[status] || status}
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
      width: 60,
      fixed: 'right',
      render: (_, record) => {
        const menuItems: MenuProps['items'] = [
          {
            key: 'edit',
            label: '编辑',
            icon: <EditOutlined />,
            onClick: () => onEditTask?.(record.id, moduleId!),
          },
          {
            key: 'check',
            label: '检查',
            icon: <CheckCircleOutlined />,
            onClick: () => handleCheckTask(record.id),
          },
          {
            key: 'refactor',
            label: '重构',
            icon: <ReloadOutlined />,
            onClick: () => handleRefactorTask(record.id),
          },
          {
            key: 'duplicate',
            label: '复制',
            icon: <CopyOutlined />,
            onClick: () => handleDuplicateTask(record.id),
          },
          {
            key: 'lock',
            label: record.locked ? '解锁' : '锁定',
            icon: <LockOutlined />,
            onClick: () => handleLockTask(record.id),
          },
          {
            type: 'divider',
          },
          {
            key: 'delete',
            label: '删除',
            icon: <DeleteOutlined />,
            danger: true,
            onClick: () => {
              confirm({
                title: '确认删除',
                icon: <ExclamationCircleOutlined />,
                content: '确定要删除这个任务吗？此操作不可恢复。',
                okText: '删除',
                okType: 'danger',
                cancelText: '取消',
                onOk: () => handleDeleteTask(record.id),
              });
            },
          },
        ];

        return (
          <Dropdown menu={{ items: menuItems }} trigger={['click']}>
            <Button type="text" size="small" icon={<MoreOutlined style={{ color: '#a0a0a0' }} />} />
          </Dropdown>
        );
      },
    },
  ];

  if (!moduleId) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Empty description="请选择一个模块查看详情" />
      </div>
    );
  }

  if (moduleLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}>
        <Spin tip="加载中..." />
      </div>
    );
  }

  if (moduleError) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Empty
          description={
            <div style={{ color: '#ff4d4f', fontSize: '13px' }}>
              加载模块失败，请稍后重试
              <div style={{ marginTop: '8px', fontSize: '12px', color: '#6b7280' }}>
                {moduleError.message}
              </div>
            </div>
          }
        />
      </div>
    );
  }

  if (!moduleData?.module) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Empty description="未找到模块信息" />
      </div>
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
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="状态">
            <Tag style={{ background: statusColors[module.status] || '#404040', border: 'none', color: '#fff' }}>
              {statusLabels[module.status] || module.status}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="版本">
            <span style={{ color: '#e0e0e0' }}>v{module.version}</span>
          </Descriptions.Item>
          <Descriptions.Item label="测试覆盖率">
            {module.testCoverage !== undefined ? (
              <Progress 
                percent={module.testCoverage} 
                size="small" 
                strokeColor={{ '0%': '#06b6d4', '100%': '#10b981' }}
                trailColor="#2d2d44"
              />
            ) : (
              <span style={{ color: '#666' }}>未设置</span>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="描述">
            <span style={{ color: '#e0e0e0' }}>{module.description || '无描述'}</span>
          </Descriptions.Item>
          <Descriptions.Item label="Prompt">
            <pre style={{ 
              margin: 0, 
              whiteSpace: 'pre-wrap', 
              fontSize: '12px', 
              background: '#1a1a2e', 
              padding: '8px', 
              borderRadius: '4px',
              color: '#a0a0a0',
              border: '1px solid #2d2d44',
            }}>
              {module.prompt || '无 Prompt'}
            </pre>
          </Descriptions.Item>
          <Descriptions.Item label="上游契约摘要">
            <span style={{ color: '#e0e0e0' }}>{module.upstreamContractSummary || '无'}</span>
          </Descriptions.Item>
          <Descriptions.Item label="下游契约摘要">
            <span style={{ color: '#e0e0e0' }}>{module.downstreamContractSummary || '无'}</span>
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            <span style={{ color: '#a0a0a0' }}>{new Date(module.createdAt).toLocaleString()}</span>
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            <span style={{ color: '#a0a0a0' }}>{new Date(module.updatedAt).toLocaleString()}</span>
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
            <span style={{ marginLeft: '8px', color: '#666' }}>({tasksData.total})</span>
          )}
        </span>
      ),
      children: (
        <div>
          <div style={{ marginBottom: '12px', display: 'flex', justifyContent: 'flex-end' }}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => onAddTask?.(moduleId)}
              style={{
                background: 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)',
                border: 'none',
              }}
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
              pageSize: 5,
              showSizeChanger: false,
              showTotal: (total) => `共 ${total} 条`,
            }}
            scroll={{ x: 600 }}
            locale={{
              emptyText: <Empty description="暂无任务，点击上方按钮创建" styles={{ image: { opacity: 0.5 } }} />,
            }}
            style={{ 
              background: '#16213e',
              borderRadius: '8px',
              overflow: 'hidden',
            }}
          />
        </div>
      ),
    },
  ];

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ 
        padding: '12px 16px', 
        borderBottom: '1px solid #2d2d44',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        background: '#1a1a2e',
      }}>
        <span style={{ fontSize: '14px', fontWeight: 500, color: '#ffffff' }}>{module.name}</span>
        <div style={{ display: 'flex', gap: '8px' }}>
          {onEdit && (
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={() => onEdit(moduleId)}
              style={{ color: '#00d9ff' }}
            >
              编辑
            </Button>
          )}
          {onDelete && (
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => onDelete(moduleId)}
            >
              删除
            </Button>
          )}
        </div>
      </div>
      <div style={{ flex: 1, overflow: 'auto', padding: '16px' }}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          type="card"
          size="small"
          style={{ 
            background: 'transparent',
          }}
          tabBarStyle={{ 
            borderBottom: '1px solid #2d2d44',
            marginBottom: '16px',
          }}
        />
      </div>
    </div>
  );
};

export default ModuleDetail;
