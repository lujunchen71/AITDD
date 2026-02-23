import React from 'react';
import { Card, List, Button, Empty, Spin, Tag, Popconfirm } from 'antd';
import { DeleteOutlined, LinkOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface DependencyListProps {
  moduleId?: string;
  dependsOn?: string;
}

const DependencyList: React.FC<DependencyListProps> = ({ moduleId, dependsOn }) => {
  const queryClient = useQueryClient();

  // 获取依赖列表
  const { data, isLoading, error } = useQuery({
    queryKey: ['dependencies', moduleId, dependsOn],
    queryFn: async () => {
      const params: any = {};
      if (moduleId) params.moduleId = moduleId;
      if (dependsOn) params.dependsOn = dependsOn;

      const response = await apiClient.get('/api/v1/dependencies', { params });
      return response.data;
    },
  });

  // 删除依赖
  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/api/v1/dependencies/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dependencies'] });
    },
  });

  if (isLoading) {
    return (
      <Card size="small">
        <div className="flex justify-center items-center h-32">
          <Spin />
        </div>
      </Card>
    );
  }

  if (error) {
    return (
      <Card size="small">
        <div className="text-red-500">加载依赖失败</div>
      </Card>
    );
  }

  const dependencies = data?.dependencies || [];

  return (
    <Card 
      title={
        <span>
          <LinkOutlined className="mr-2" />
          依赖关系
        </span>
      }
      size="small"
    >
      {dependencies.length === 0 ? (
        <Empty description="暂无依赖关系" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        <List
          dataSource={dependencies}
          renderItem={(dep: any) => (
            <List.Item
              actions={[
                <Popconfirm
                  key="delete"
                  title="确定删除此依赖关系？"
                  onConfirm={() => deleteMutation.mutate(dep.id)}
                >
                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                  />
                </Popconfirm>,
              ]}
            >
              <List.Item.Meta
                title={
                  <span>
                    <Tag color="blue">{dep.moduleId?.slice(0, 8)}</Tag>
                    <span className="mx-2">→</span>
                    <Tag color="green">{dep.dependsOn?.slice(0, 8)}</Tag>
                  </span>
                }
                description={dep.dependency || '无描述'}
              />
            </List.Item>
          )}
        />
      )}
    </Card>
  );
};

export default DependencyList;
