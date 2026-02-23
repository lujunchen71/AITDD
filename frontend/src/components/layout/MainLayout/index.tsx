import React from 'react';
import { Layout } from 'antd';
import { useUIStore } from '../../../stores/useUIStore';
import Header from '../Header';
import Sidebar from '../Sidebar';
import './index.css';

const { Content } = Layout;

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const { sidebarCollapsed } = useUIStore();

  return (
    <Layout className="main-layout">
      <Sidebar />
      <Layout
        className="main-layout-content-wrapper"
        style={{
          marginLeft: sidebarCollapsed ? 80 : 240,
          transition: 'margin-left 0.2s ease',
        }}
      >
        <Header />
        <Content className="main-layout-content">
          {children}
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
