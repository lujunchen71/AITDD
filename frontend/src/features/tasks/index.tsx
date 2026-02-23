import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

const TasksPage: React.FC = () => {
  return (
    <div>
      <Title level={2}>任务管理</Title>
      <Card>
        <Paragraph>任务管理页面 - 开发中</Paragraph>
      </Card>
    </div>
  );
};

export default TasksPage;
