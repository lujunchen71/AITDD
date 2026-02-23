import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const SettingsPage: React.FC = () => {
  return (
    <div>
      <Title level={2}>系统设置</Title>
      <Card>
        <Paragraph>系统设置页面 - 开发中</Paragraph>
      </Card>
    </div>
  );
};

export default SettingsPage;
