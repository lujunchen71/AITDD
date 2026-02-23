import React, { useState } from 'react';
import { Card, Typography, Button, Steps, Result } from 'antd';
import { CheckCircleOutlined } from '@ant-design/icons';
import PluginSelector from './components/PluginSelector';

const { Title, Paragraph } = Typography;
const { Step } = Steps;

const SetupPage: React.FC = () => {
  const [currentStep, setCurrentStep] = useState(0);
  const [selectedPlugin, setSelectedPlugin] = useState<string | null>(null);
  const [isComplete, setIsComplete] = useState(false);

  const handlePluginSelect = (plugin: string) => {
    setSelectedPlugin(plugin);
    setCurrentStep(1);
  };

  const handleConfirm = () => {
    // 这里调用API初始化项目
    setCurrentStep(2);
    setTimeout(() => {
      setIsComplete(true);
    }, 1500);
  };

  const handleBack = () => {
    setCurrentStep(0);
    setSelectedPlugin(null);
  };

  if (isComplete) {
    return (
      <div className="setup-page">
        <Card className="setup-card">
          <Result
            icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            status="success"
            title="项目初始化完成"
            subTitle="您已成功初始化AITDD项目，现在可以开始使用了"
            extra={[
              <Button type="primary" key="dashboard" href="/dashboard">
                进入仪表盘
              </Button>,
            ]}
          />
        </Card>
      </div>
    );
  }

  return (
    <div className="setup-page">
      <Card className="setup-card">
        <div className="setup-header">
          <Title level={2}>欢迎使用 AITDD</Title>
          <Paragraph type="secondary">
            AI辅助可视化任务治理系统 - 让AI更好地理解和管理您的项目
          </Paragraph>
        </div>

        <Steps current={currentStep} className="setup-steps">
          <Step title="选择插件" description="选择您使用的AI编程工具" />
          <Step title="确认配置" description="确认初始化配置" />
          <Step title="完成" description="开始使用" />
        </Steps>

        <div className="setup-content">
          {currentStep === 0 && (
            <PluginSelector onSelect={handlePluginSelect} />
          )}

          {currentStep === 1 && (
            <div className="setup-confirm">
              <Title level={4}>确认配置</Title>
              <Paragraph>
                您选择的AI插件: <strong>{selectedPlugin}</strong>
              </Paragraph>
              <Paragraph type="secondary">
                初始化将创建以下内容:
              </Paragraph>
              <ul>
                <li>.aitdd/ - 配置目录</li>
                <li>.kilocode/ - 工作流模板</li>
                <li>specs/ - 规格文档目录</li>
                <li>AITDD_GUIDE.md - 使用指南</li>
              </ul>
              <div className="setup-actions">
                <Button onClick={handleBack}>返回</Button>
                <Button type="primary" onClick={handleConfirm}>
                  确认初始化
                </Button>
              </div>
            </div>
          )}

          {currentStep === 2 && (
            <div className="setup-loading">
              <Paragraph>正在初始化项目...</Paragraph>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
};

export default SetupPage;
