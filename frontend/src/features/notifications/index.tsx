import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const NotificationsPage: React.FC = () => {
  return (
    <div>
      <Title level={2}>通知中心</Title>
      <Card>
        <Paragraph>通知中心页面 - 开发中</Paragraph>
      </Card>
    </div>
  );
};

export default NotificationsPage;
