import React from 'react';
import { Card, Button, Empty } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../../services/api';

const ConstitutionView: React.FC = () => {
  const { data, isLoading } = useQuery({
    queryKey: ['constitution'],
    queryFn: () => api.get<{ constitution: string }>('/project/constitution'),
  });

  const constitution = data?.data?.constitution;

  if (isLoading) {
    return <Card loading={true} />;
  }

  return (
    <Card
      title="项目宪法"
      extra={
        <Button type="text" icon={<EditOutlined />}>
          编辑
        </Button>
      }
    >
      {constitution ? (
        <div className="constitution-content">
          <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
            {constitution}
          </pre>
        </div>
      ) : (
        <Empty
          description="暂无项目宪法"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        >
          <Button type="primary">添加宪法</Button>
        </Empty>
      )}
    </Card>
  );
};

export default ConstitutionView;
