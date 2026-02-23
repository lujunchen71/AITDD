import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const DashboardPage: React.FC = () => {
  return (
    <div>
      <Title level={2}>仪表盘</Title>
      <Card>
        <Paragraph>仪表盘页面 - 开发中</Paragraph>
      </Card>
    </div>
  );
};

export default DashboardPage;
