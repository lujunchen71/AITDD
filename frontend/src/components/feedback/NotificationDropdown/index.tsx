import React, { useState } from 'react';
import { Badge, Dropdown, List, Button, Empty, Spin, Typography } from 'antd';
import { BellOutlined, CheckOutlined, DeleteOutlined } from '@ant-design/icons';
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

const NotificationDropdown: React.FC = () => {
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['notifications', 'unread'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/notifications', {
        params: { unread: true, pageSize: 10 },
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
  const unreadCount = notifications.length;

  const formatTime = (timestamp: number) => {
    const now = Date.now();
    const diff = now - timestamp;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    return `${days}天前`;
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'task_completed':
        return '✅';
      case 'task_blocked':
        return '🚫';
      case 'mention':
        return '👤';
      case 'system':
        return '⚙️';
      default:
        return '📢';
    }
  };

  const handleMarkRead = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    markReadMutation.mutate(id);
  };

  const items = {
    items: [
      {
        key: 'notifications',
        label: (
          <div style={{ width: 320, maxHeight: 400, overflow: 'auto' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '8px 0',
                borderBottom: '1px solid #333',
                marginBottom: 8,
              }}
            >
              <Typography.Text strong>通知</Typography.Text>
              {unreadCount > 0 && (
                <Button type="link" size="small">
                  全部已读
                </Button>
              )}
            </div>

            {isLoading ? (
              <div style={{ textAlign: 'center', padding: 20 }}>
                <Spin />
              </div>
            ) : notifications.length === 0 ? (
              <Empty description="暂无新通知" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <List
                dataSource={notifications}
                renderItem={(item: Notification) => (
                  <List.Item
                    style={{
                      padding: '8px 0',
                      borderBottom: '1px solid #222',
                      cursor: 'pointer',
                    }}
                    onClick={() => {
                      if (item.link) {
                        window.location.href = item.link;
                      }
                      setOpen(false);
                    }}
                  >
                    <List.Item.Meta
                      avatar={<span style={{ fontSize: 18 }}>{getTypeIcon(item.type)}</span>}
                      title={
                        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                          <span style={{ fontSize: 13 }}>{item.title}</span>
                          <Button
                            type="text"
                            size="small"
                            icon={<CheckOutlined />}
                            onClick={(e) => handleMarkRead(item.id, e)}
                          />
                        </div>
                      }
                      description={
                        <div>
                          <div style={{ fontSize: 12, color: '#888' }}>
                            {item.content}
                          </div>
                          <div style={{ fontSize: 11, color: '#666', marginTop: 4 }}>
                            {formatTime(item.createdAt)}
                          </div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            )}
          </div>
        ),
      },
    ],
  };

  return (
    <Dropdown
      menu={items}
      open={open}
      onOpenChange={setOpen}
      trigger={['click']}
      placement="bottomRight"
    >
      <Badge count={unreadCount} size="small" offset={[0, 2]}>
        <Button
          type="text"
          icon={<BellOutlined style={{ fontSize: 18, color: '#eaeaea' }} />}
        />
      </Badge>
    </Dropdown>
  );
};

export default NotificationDropdown;
