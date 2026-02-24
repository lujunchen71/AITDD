import React from 'react';
import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  AppstoreOutlined,
  NotificationOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useUIStore } from '../../../stores/useUIStore';
import './index.css';

const { Sider } = Layout;

const Sidebar: React.FC = () => {
  const { sidebarCollapsed } = useUIStore();
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/modules',
      icon: <AppstoreOutlined />,
      label: '模块与任务',
    },
    {
      key: '/notifications',
      icon: <NotificationOutlined />,
      label: '通知中心',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={sidebarCollapsed}
      className="sidebar"
      width={240}
      collapsedWidth={80}
    >
      <div className="sidebar-logo">
        {sidebarCollapsed ? 'AT' : 'AITDD'}
      </div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
        onClick={handleMenuClick}
      />
    </Sider>
  );
};

export default Sidebar;
