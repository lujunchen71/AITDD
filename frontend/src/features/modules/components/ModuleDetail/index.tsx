import React from 'react';
import { Card, Descriptions, Tag, Progress, Empty, Spin, Button } from 'antd';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface ModuleDetailProps {
  moduleId: string | null;
  onEdit?: (moduleId: string) => void;
  onDelete?: (moduleId: string) => void;
}

const statusColors: Record<string, string> = {
  designing: 'blue',
  developing: 'orange',
  testing: 'cyan',
  done: 'green',
};

const statusLabels: Record<string, string> = {
  designing: '设计中',
  developing: '开发中',
  testing: '测试中',
  done: '已完成',
};

const ModuleDetail: React.FC<ModuleDetailProps> = ({
  moduleId,
  onEdit,
  onDelete,
}) => {
  // 获取模块详情
  const { data, isLoading, error } = useQuery({
    queryKey: ['module', moduleId],
    queryFn: async () => {
      if (!moduleId) return null;
      const response = await apiClient.get(`/api/v1/modules/${moduleId}`);
      return response.data;
    },
    enabled: !!moduleId,
  });

  if (!moduleId) {
    return (
      <Card>
        <Empty description="请选择一个模块查看详情" />
      </Card>
    );
  }

  if (isLoading) {
    return (
      <Card>
        <div className="flex justify-center items-center h-64">
          <Spin />
        </div>
      </Card>
    );
  }

  if (error || !data?.module) {
    return (
      <Card>
        <Empty description="加载模块失败" />
      </Card>
    );
  }

  const module = data.module;

  return (
    <Card
      title={module.name}
      extra={
        <div className="flex gap-2">
          {onEdit && (
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => onEdit(moduleId)}
            >
              编辑
            </Button>
          )}
          {onDelete && (
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => onDelete(moduleId)}
            >
              删除
            </Button>
          )}
        </div>
      }
    >
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="状态">
          <Tag color={statusColors[module.status] || 'default'}>
            {statusLabels[module.status] || module.status}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="版本">
          v{module.version}
        </Descriptions.Item>
        <Descriptions.Item label="测试覆盖率" span={2}>
          {module.testCoverage !== undefined ? (
            <Progress percent={module.testCoverage} size="small" />
          ) : (
            <span className="text-gray-400">未设置</span>
          )}
        </Descriptions.Item>
        <Descriptions.Item label="描述" span={2}>
          {module.description || <span className="text-gray-400">无描述</span>}
        </Descriptions.Item>
        <Descriptions.Item label="Prompt" span={2}>
          <pre className="whitespace-pre-wrap text-sm bg-gray-50 p-2 rounded">
            {module.prompt || <span className="text-gray-400">无Prompt</span>}
          </pre>
        </Descriptions.Item>
        <Descriptions.Item label="上游契约摘要" span={2}>
          {module.upstreamContractSummary || <span className="text-gray-400">无</span>}
        </Descriptions.Item>
        <Descriptions.Item label="下游契约摘要" span={2}>
          {module.downstreamContractSummary || <span className="text-gray-400">无</span>}
        </Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {new Date(module.createdAt).toLocaleString()}
        </Descriptions.Item>
        <Descriptions.Item label="更新时间">
          {new Date(module.updatedAt).toLocaleString()}
        </Descriptions.Item>
      </Descriptions>
    </Card>
  );
};

export default ModuleDetail;
