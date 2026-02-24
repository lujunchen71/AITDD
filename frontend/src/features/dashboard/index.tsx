import React from 'react';
import { Row, Col, Card, Statistic, Typography, Progress, List, Tag, Space } from 'antd';
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  BlockOutlined,
  AppstoreOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../services/api';
import { useProjectId } from '../../stores/useProjectStore';

const { Title, Text } = Typography;

const DashboardPage: React.FC = () => {
  const projectId = useProjectId();
  
  const { data: modulesData } = useQuery({
    queryKey: ['modules', projectId],
    queryFn: async () => {
      const response = await api.get<{ modules: any[] }>('/modules', {
        params: { projectId: projectId || undefined },
      });
      return response;
    },
    enabled: !!projectId,
  });

  const { data: tasksData } = useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      const response = await api.get<{ tasks: any[] }>('/tasks', {
        params: { pageSize: 100 },
      });
      return response;
    },
  });

  const modules = modulesData?.data?.modules || [];
  const tasks = tasksData?.data?.tasks || [];

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

  return (
    <div style={{ padding: 24 }}>
      <Title level={4} style={{ marginBottom: 24 }}>
        仪表盘
      </Title>

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
  );
};

export default DashboardPage;
