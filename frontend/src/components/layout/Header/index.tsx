import React from 'react';
import { Layout, Button, Badge, Tooltip, Dropdown, Avatar, Space } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BellOutlined,
  SettingOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import { useUIStore } from '../../../stores/useUIStore';
import type { MenuProps } from 'antd';
import './index.css';

const { Header: AntHeader } = Layout;

const Header: React.FC = () => {
  const { sidebarCollapsed, toggleSidebar, setNotificationDrawerVisible } = useUIStore();

  // 用户菜单项
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'settings',
      label: '设置',
      icon: <SettingOutlined />,
    },
    {
      key: 'help',
      label: '帮助',
      icon: <QuestionCircleOutlined />,
    },
    {
      type: 'divider',
    },
    {
      key: 'about',
      label: '关于 AITDD',
    },
  ];

  return (
    <AntHeader className="header">
      <div className="header-left">
        <Button
          type="text"
          icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          onClick={toggleSidebar}
          className="header-trigger"
        />
        <span className="header-title">AITDD - 可视化任务治理系统</span>
      </div>

      <div className="header-right">
        <Space size="middle">
          {/* 通知按钮 */}
          <Tooltip title="通知">
            <Badge count={0} size="small">
              <Button
                type="text"
                icon={<BellOutlined />}
                onClick={() => setNotificationDrawerVisible(true)}
                className="header-icon-btn"
              />
            </Badge>
          </Tooltip>

          {/* 用户菜单 */}
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Avatar
              className="header-avatar"
              style={{ backgroundColor: '#1890ff', cursor: 'pointer' }}
            >
              AI
            </Avatar>
          </Dropdown>
        </Space>
      </div>
    </AntHeader>
  );
};

export default Header;
