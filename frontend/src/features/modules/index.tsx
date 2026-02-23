import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const ModulesPage: React.FC = () => {
  return (
    <div>
      <Title level={2}>模块管理</Title>
      <Card>
        <Paragraph>模块管理页面 - 开发中</Paragraph>
      </Card>
    </div>
  );
};

export default ModulesPage;
