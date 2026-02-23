import React, { useState } from 'react';
import { Card, List, Button, Space, Tag, Modal, Typography, Empty, Spin, message } from 'antd';
import { WarningOutlined, CheckOutlined, CloseOutlined, MergeOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '../../../services/api';

const { Text, Paragraph } = Typography;

interface Conflict {
  id: string;
  entityType: string;
  entityId: string;
  localData: any;
  remoteData: any;
  conflictType: string;
  createdAt: number;
  resolution?: string;
}

const ConflictResolver: React.FC = () => {
  const [selectedConflict, setSelectedConflict] = useState<Conflict | null>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const queryClient = useQueryClient();

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['conflicts'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/sync/conflicts');
      return response.data;
    },
  });

  const resolveMutation = useMutation({
    mutationFn: async ({ conflictId, resolution, mergedData }: { conflictId: string; resolution: string; mergedData?: any }) => {
      await apiClient.post(`/api/v1/sync/conflicts/${conflictId}/resolve`, {
        resolution,
        mergedData,
      });
    },
    onSuccess: () => {
      message.success('冲突已解决');
      queryClient.invalidateQueries({ queryKey: ['conflicts'] });
      setModalVisible(false);
      setSelectedConflict(null);
    },
    onError: () => {
      message.error('解决冲突失败');
    },
  });

  const conflicts = data?.conflicts || [];

  const getEntityTypeLabel = (type: string) => {
    switch (type) {
      case 'module':
        return '模块';
      case 'task':
        return '任务';
      default:
        return type;
    }
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  const handleResolve = (resolution: string) => {
    if (selectedConflict) {
      resolveMutation.mutate({
        conflictId: selectedConflict.id,
        resolution,
      });
    }
  };

  const renderDataComparison = (conflict: Conflict) => {
    return (
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <Card title="本地数据" size="small" style={{ backgroundColor: 'rgba(233, 69, 96, 0.1)' }}>
          <pre style={{ fontSize: 11, overflow: 'auto', maxHeight: 200 }}>
            {JSON.stringify(conflict.localData, null, 2)}
          </pre>
        </Card>
        <Card title="远程数据" size="small" style={{ backgroundColor: 'rgba(52, 211, 153, 0.1)' }}>
          <pre style={{ fontSize: 11, overflow: 'auto', maxHeight: 200 }}>
            {JSON.stringify(conflict.remoteData, null, 2)}
          </pre>
        </Card>
      </div>
    );
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 40 }}>
        <Spin />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Text strong style={{ fontSize: 16 }}>
          <WarningOutlined style={{ color: '#faad14', marginRight: 8 }} />
          数据冲突 ({conflicts.length})
        </Text>
        <Button onClick={() => refetch()}>刷新</Button>
      </div>

      {conflicts.length === 0 ? (
        <Card>
          <Empty description="暂无数据冲突" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        </Card>
      ) : (
        <List
          dataSource={conflicts}
          renderItem={(item: Conflict) => (
            <List.Item>
              <Card
                style={{ width: '100%', cursor: 'pointer' }}
                onClick={() => {
                  setSelectedConflict(item);
                  setModalVisible(true);
                }}
              >
                <Space>
                  <Tag color="warning">{getEntityTypeLabel(item.entityType)}</Tag>
                  <Text>{item.entityId}</Text>
                  <Text type="secondary">{formatTime(item.createdAt)}</Text>
                </Space>
              </Card>
            </List.Item>
          )}
        />
      )}

      <Modal
        title="解决冲突"
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setSelectedConflict(null);
        }}
        footer={null}
        width={700}
      >
        {selectedConflict && (
          <div>
            <Paragraph>
              <Tag color="warning">{getEntityTypeLabel(selectedConflict.entityType)}</Tag>
              ID: {selectedConflict.entityId}
            </Paragraph>

            {renderDataComparison(selectedConflict)}

            <div style={{ marginTop: 24, textAlign: 'center' }}>
              <Space size="large">
                <Button
                  icon={<CheckOutlined />}
                  onClick={() => handleResolve('local_wins')}
                  loading={resolveMutation.isPending}
                >
                  使用本地版本
                </Button>
                <Button
                  icon={<CloseOutlined />}
                  onClick={() => handleResolve('remote_wins')}
                  loading={resolveMutation.isPending}
                >
                  使用远程版本
                </Button>
                <Button
                  type="primary"
                  icon={<MergeOutlined />}
                  onClick={() => handleResolve('merged')}
                  loading={resolveMutation.isPending}
                >
                  合并数据
                </Button>
              </Space>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default ConflictResolver;
