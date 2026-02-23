import React from 'react';
import { Card, Descriptions, Tag, Typography } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../../services/api';

const { Title } = Typography;

const ProjectInfo: React.FC = () => {
  const { data, isLoading } = useQuery({
    queryKey: ['project'],
    queryFn: () => api.get<{ project: unknown }>('/project'),
  });

  const project = data?.data?.project;

  if (isLoading) {
    return <Card loading={true} />;
  }

  if (!project) {
    return (
      <Card>
        <Title level={5}>暂无项目</Title>
        <p>请先运行 aitdd init 初始化项目</p>
      </Card>
    );
  }

  const projectData = project as {
    id: string;
    name: string;
    constitution: string;
    createdAt: number;
    updatedAt: number;
    version: number;
    syncStatus: string;
  };

  return (
    <Card title="项目信息">
      <Descriptions column={2}>
        <Descriptions.Item label="项目名称">{projectData.name}</Descriptions.Item>
        <Descriptions.Item label="版本">{projectData.version}</Descriptions.Item>
        <Descriptions.Item label="同步状态">
          <Tag color={projectData.syncStatus === 'SYNCED' ? 'green' : 'orange'}>
            {projectData.syncStatus}
          </Tag>
        </Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {new Date(projectData.createdAt).toLocaleString()}
        </Descriptions.Item>
      </Descriptions>
    </Card>
  );
};

export default ProjectInfo;
