import React from 'react';
import { Card, Form, Input, Select, Button, Switch, Space, message, Divider, Typography } from 'antd';
import { SyncOutlined, CloudServerOutlined, HistoryOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '../../../services/api';

const { Title, Text } = Typography;

const SyncSettings: React.FC = () => {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  const { data: syncConfig, isLoading } = useQuery({
    queryKey: ['sync', 'config'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/sync/config');
      return response.data;
    },
  });

  const { data: syncStatus } = useQuery({
    queryKey: ['sync', 'status'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/sync/status');
      return response.data;
    },
    refetchInterval: 30000, // 每30秒刷新
  });

  const saveConfigMutation = useMutation({
    mutationFn: async (values: any) => {
      await apiClient.post('/api/v1/sync/config', values);
    },
    onSuccess: () => {
      message.success('同步配置已保存');
      queryClient.invalidateQueries({ queryKey: ['sync'] });
    },
    onError: () => {
      message.error('保存配置失败');
    },
  });

  const syncNowMutation = useMutation({
    mutationFn: async () => {
      const response = await apiClient.post('/api/v1/sync/execute');
      return response.data;
    },
    onSuccess: (data) => {
      message.success(`同步完成: ${data.syncedItems} 项已同步`);
      queryClient.invalidateQueries({ queryKey: ['sync'] });
    },
    onError: () => {
      message.error('同步失败');
    },
  });

  const formatTime = (timestamp: number) => {
    if (!timestamp) return '从未同步';
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'synced':
        return '#52c41a';
      case 'pending':
        return '#faad14';
      case 'error':
        return '#ff4d4f';
      default:
        return '#888';
    }
  };

  return (
    <div style={{ padding: 24, maxWidth: 800 }}>
      <Title level={4}>
        <CloudServerOutlined style={{ marginRight: 8 }} />
        数据同步设置
      </Title>

      {/* 同步状态卡片 */}
      <Card style={{ marginBottom: 16 }}>
        <Space size="large">
          <div>
            <Text type="secondary">同步状态</Text>
            <div>
              <span
                style={{
                  display: 'inline-block',
                  width: 8,
                  height: 8,
                  borderRadius: '50%',
                  backgroundColor: getStatusColor(syncStatus?.status || 'pending'),
                  marginRight: 8,
                }}
              />
              {syncStatus?.status === 'synced' ? '已同步' : syncStatus?.status === 'pending' ? '待同步' : '同步错误'}
            </div>
          </div>
          <div>
            <Text type="secondary">上次同步</Text>
            <div>
              <HistoryOutlined style={{ marginRight: 4 }} />
              {formatTime(syncStatus?.lastSyncTime)}
            </div>
          </div>
          <Button
            type="primary"
            icon={<SyncOutlined spin={syncNowMutation.isPending} />}
            loading={syncNowMutation.isPending}
            onClick={() => syncNowMutation.mutate()}
          >
            立即同步
          </Button>
        </Space>
      </Card>

      {/* 同步配置表单 */}
      <Card title="同步配置">
        <Form
          form={form}
          layout="vertical"
          initialValues={syncConfig || {
            syncEnabled: false,
            syncInterval: '30',
            remoteUrl: '',
          }}
          onFinish={(values) => saveConfigMutation.mutate(values)}
        >
          <Form.Item name="syncEnabled" label="启用自动同步" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item name="remoteUrl" label="远程数据库地址" rules={[{ required: false }]}>
            <Input placeholder="postgresql://user:password@host:port/database" />
          </Form.Item>

          <Form.Item name="syncInterval" label="同步间隔">
            <Select>
              <Select.Option value="5">每 5 分钟</Select.Option>
              <Select.Option value="15">每 15 分钟</Select.Option>
              <Select.Option value="30">每 30 分钟</Select.Option>
              <Select.Option value="60">每 1 小时</Select.Option>
              <Select.Option value="360">每 6 小时</Select.Option>
            </Select>
          </Form.Item>

          <Divider />

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saveConfigMutation.isPending}>
                保存配置
              </Button>
              <Button onClick={() => form.resetFields()}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default SyncSettings;
