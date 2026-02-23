import React from 'react';
import { Card, Row, Col, Typography } from 'antd';
import {
  CodeOutlined,
  ThunderboltOutlined,
  RobotOutlined,
} from '@ant-design/icons';

const { Title, Paragraph } = Typography;

interface PluginSelectorProps {
  onSelect: (plugin: string) => void;
}

interface PluginOption {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  color: string;
}

const plugins: PluginOption[] = [
  {
    id: 'kilocode',
    name: 'KiloCode',
    description: '强大的AI编程助手，支持多种编程语言和框架',
    icon: <CodeOutlined style={{ fontSize: 48 }} />,
    color: '#1890ff',
  },
  {
    id: 'opencode',
    name: 'OpenCode',
    description: '开源的AI代码生成工具，灵活可扩展',
    icon: <ThunderboltOutlined style={{ fontSize: 48 }} />,
    color: '#52c41a',
  },
  {
    id: 'claudecode',
    name: 'ClaudeCode',
    description: 'Anthropic Claude驱动的智能编程助手',
    icon: <RobotOutlined style={{ fontSize: 48 }} />,
    color: '#722ed1',
  },
];

const PluginSelector: React.FC<PluginSelectorProps> = ({ onSelect }) => {
  return (
    <div className="plugin-selector">
      <div className="plugin-selector-header">
        <Title level={4}>选择您使用的AI编程工具</Title>
        <Paragraph type="secondary">
          选择后，AITDD将为您生成对应的工作流模板和配置文件
        </Paragraph>
      </div>

      <Row gutter={[24, 24]}>
        {plugins.map((plugin) => (
          <Col xs={24} sm={8} key={plugin.id}>
            <Card
              hoverable
              className="plugin-card"
              onClick={() => onSelect(plugin.id)}
              style={{
                borderColor: plugin.color,
                borderWidth: 2,
              }}
            >
              <div className="plugin-card-content">
                <div
                  className="plugin-icon"
                  style={{ color: plugin.color }}
                >
                  {plugin.icon}
                </div>
                <Title level={4}>{plugin.name}</Title>
                <Paragraph type="secondary">{plugin.description}</Paragraph>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
};

export default PluginSelector;
