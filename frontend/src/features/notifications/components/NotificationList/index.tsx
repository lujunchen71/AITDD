import React from 'react';
import { List, Card, Tag, Button, Empty, Spin, Typography, Space } from 'antd';
import { CheckOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '../../../services/api';

interface Notification {
  id: string;
  type: string;
  title: string;
  content: string;
  link?: string;
  readAt?: number | null;
  createdAt: number;
}

const NotificationList: React.FC = () => {
  const queryClient = useQueryClient();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['notifications', 'all'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/notifications', {
        params: { pageSize: 50 },
      });
      return response.data;
    },
  });

  const markReadMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.post(`/api/v1/notifications/${id}/read`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const notifications = data?.notifications || [];

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'task_completed':
        return 'success';
      case 'task_blocked':
        return 'error';
      case 'mention':
        return 'processing';
      case 'system':
        return 'warning';
      default:
        return 'default';
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'task_completed':
        return '任务完成';
      case 'task_blocked':
        return '任务阻塞';
      case 'mention':
        return '提及';
      case 'system':
        return '系统';
      default:
        return '通知';
    }
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 40 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          通知中心
        </Typography.Title>
        <Button onClick={() => refetch()}>刷新</Button>
      </div>

      {notifications.length === 0 ? (
        <Card>
          <Empty description="暂无通知" />
        </Card>
      ) : (
        <List
          grid={{ gutter: 16, xs: 1, sm: 1, md: 2, lg: 2, xl: 3 }}
          dataSource={notifications}
          renderItem={(item: Notification) => (
            <List.Item>
              <Card
                hoverable
                style={{
                  backgroundColor: item.readAt ? 'transparent' : 'rgba(233, 69, 96, 0.1)',
                  borderColor: item.readAt ? '#333' : '#e94560',
                }}
                onClick={() => {
                  if (item.link) {
                    window.location.href = item.link;
                  }
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <Tag color={getTypeColor(item.type)}>{getTypeLabel(item.type)}</Tag>
                  {!item.readAt && (
                    <Button
                      type="text"
                      size="small"
                      icon={<CheckOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        markReadMutation.mutate(item.id);
                      }}
                    >
                      标记已读
                    </Button>
                  )}
                </div>

                <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                  {item.title}
                </Typography.Text>

                <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                  {item.content}
                </Typography.Text>

                <Space style={{ fontSize: 12, color: '#666' }}>
                  <ClockCircleOutlined />
                  {formatTime(item.createdAt)}
                </Space>
              </Card>
            </List.Item>
          )}
        />
      )}
    </div>
  );
};

export default NotificationList;
