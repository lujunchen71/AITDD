import React from 'react';
import { Row, Col, Card, Statistic, Typography, Progress, List, Tag, Space, Empty } from 'antd';
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  BlockOutlined,
  AppstoreOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../services/api';
import { useSelectedProjectIds } from '../../stores/useProjectStore';

const { Title, Text } = Typography;

const DashboardPage: React.FC = () => {
  const selectedProjectIds = useSelectedProjectIds();
  
  // [DEBUG] Dashboard 渲染调试
  console.log('[Dashboard] Rendered, selectedProjectIds:', selectedProjectIds);
  console.log('[Dashboard] selectedProjectIds type:', typeof selectedProjectIds, Array.isArray(selectedProjectIds));
  
  // 当没有选中项目时，显示空状态
  const hasSelectedProjects = selectedProjectIds.length > 0;
  console.log('[Dashboard] hasSelectedProjects:', hasSelectedProjects);
  
  const { data: modulesData, isLoading: modulesLoading, error: modulesError } = useQuery({
    queryKey: ['dashboard-modules', selectedProjectIds],
    queryFn: async () => {
      console.log('[Dashboard] Fetching modules for projectIds:', selectedProjectIds);
      // 使用 selectedProjectIds 获取多个项目的模块
      const projectIdsParam = selectedProjectIds.join(',');
      const response = await api.get<{ modules: any[] }>('/modules', {
        params: { projectIds: projectIdsParam },
      });
      console.log('[Dashboard] Modules fetched:', response.data?.modules?.length);
      return response;
    },
    enabled: hasSelectedProjects,
  });

  const { data: tasksData, isLoading: tasksLoading, error: tasksError } = useQuery({
    queryKey: ['dashboard-tasks', selectedProjectIds],
    queryFn: async () => {
      console.log('[Dashboard] Fetching tasks for projectIds:', selectedProjectIds);
      // 使用 selectedProjectIds 获取多个项目的任务
      const projectIdsParam = selectedProjectIds.join(',');
      const response = await api.get<{ tasks: any[] }>('/tasks', {
        params: { projectIds: projectIdsParam, pageSize: 100 },
      });
      console.log('[Dashboard] Tasks fetched:', response.data?.tasks?.length);
      return response;
    },
    enabled: hasSelectedProjects,
  });

  const modules = modulesData?.data?.modules || [];
  const tasks = tasksData?.data?.tasks || [];
  
  // [DEBUG] 数据状态日志
  console.log('[Dashboard] Data state:', {
    modulesLoading,
    tasksLoading,
    modulesError,
    tasksError,
    modulesCount: modules.length,
    tasksCount: tasks.length,
  });

  // 统计任务状态
  const taskStats = {
    total: tasks.length,
    pending: tasks.filter((t: any) => t.status === 'pending').length,
    inProgress: tasks.filter((t: any) => t.status === 'in_progress').length,
    completed: tasks.filter((t: any) => t.status === 'completed').length,
    blocked: tasks.filter((t: any) => t.status === 'blocked').length,
  };

  const completionRate = taskStats.total > 0
    ? Math.round((taskStats.completed / taskStats.total) * 100)
    : 0;

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return '#52c41a';
      case 'in_progress':
        return '#1890ff';
      case 'pending':
        return '#faad14';
      case 'blocked':
        return '#ff4d4f';
      default:
        return '#888';
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'completed':
        return '已完成';
      case 'in_progress':
        return '进行中';
      case 'pending':
        return '待处理';
      case 'blocked':
        return '已阻塞';
      default:
        return status;
    }
  };

  // 如果没有选中项目，显示空状态
  if (!hasSelectedProjects) {
    return (
      <div style={{ padding: 24 }}>
        <Title level={4} style={{ marginBottom: 24 }}>
          仪表盘
        </Title>
        <Empty
          description="请选择至少一个项目查看统计数据"
          style={{ padding: '40px 0' }}
        />
      </div>
    );
  }

  return (
    <div style={{
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      overflow: 'hidden',
    }}>
      {/* 固定标题 */}
      <div style={{
        padding: '16px 24px',
        flexShrink: 0,
        borderBottom: '1px solid #2d2d44',
        background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',
      }}>
        <Title level={4} style={{ margin: 0 }}>
          仪表盘
        </Title>
      </div>
      {/* 可滚动内容 */}
      <div style={{ flex: 1, overflow: 'auto', padding: 24 }}>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总模块数"
              value={modules.length}
              prefix={<AppstoreOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总任务数"
              value={taskStats.total}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="进行中"
              value={taskStats.inProgress}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="已阻塞"
              value={taskStats.blocked}
              prefix={<BlockOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 完成进度 */}
      <Card title="任务完成进度" style={{ marginTop: 16 }}>
        <Progress
          percent={completionRate}
          status={completionRate === 100 ? 'success' : 'active'}
          strokeColor={{
            '0%': '#e94560',
            '100%': '#52c41a',
          }}
        />
        <Row gutter={16} style={{ marginTop: 16 }}>
          <Col span={6}>
            <Text type="secondary">待处理: {taskStats.pending}</Text>
          </Col>
          <Col span={6}>
            <Text type="secondary">进行中: {taskStats.inProgress}</Text>
          </Col>
          <Col span={6}>
            <Text type="secondary">已完成: {taskStats.completed}</Text>
          </Col>
          <Col span={6}>
            <Text type="secondary">已阻塞: {taskStats.blocked}</Text>
          </Col>
        </Row>
      </Card>

      {/* 最近任务 */}
      <Card title="最近任务" style={{ marginTop: 16 }}>
        <List
          dataSource={tasks.slice(0, 5)}
          renderItem={(task: any) => (
            <List.Item>
              <Space>
                <Tag color={getStatusColor(task.status)}>
                  {getStatusLabel(task.status)}
                </Tag>
                <Text>{task.title}</Text>
              </Space>
            </List.Item>
          )}
          locale={{ emptyText: '暂无任务' }}
        />
      </Card>
      </div>
    </div>
  );
};

export default DashboardPage;
