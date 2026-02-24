import React, { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, Spin } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import MainLayout from './components/layout/MainLayout';
import { useProjectStore } from './stores/useProjectStore';
import './App.css';

// 占位页面组件
const DashboardPage = React.lazy(() => import('./features/dashboard'));
const ModulesPage = React.lazy(() => import('./features/modules'));
const TasksPage = React.lazy(() => import('./features/tasks'));
const NotificationsPage = React.lazy(() => import('./features/notifications'));
const SettingsPage = React.lazy(() => import('./features/settings'));

// 项目初始化组件
const ProjectInitializer: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { ensureProject, isInitialized, isLoading, projectId, error } = useProjectStore();

  useEffect(() => {
    // 应用启动时确保项目存在
    if (!isInitialized) {
      ensureProject();
    }
  }, [ensureProject, isInitialized]);

  // 显示加载状态
  if (isLoading && !isInitialized) {
    return (
      <div style={{ 
        display: 'flex', 
        flexDirection: 'column',
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        gap: 16
      }}>
        <Spin size="large" />
        <span>正在初始化项目...</span>
      </div>
    );
  }

  // 显示错误状态
  if (error && !projectId) {
    return (
      <div style={{ 
        display: 'flex', 
        flexDirection: 'column',
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh',
        gap: 16,
        color: '#ff4d4f'
      }}>
        <span>项目初始化失败: {error}</span>
        <button 
          onClick={() => ensureProject()}
          style={{
            padding: '8px 16px',
            cursor: 'pointer'
          }}
        >
          重试
        </button>
      </div>
    );
  }

  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN}>
      <ProjectInitializer>
        <MainLayout>
          <React.Suspense fallback={<div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}><Spin /></div>}>
            <Routes>
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/modules" element={<ModulesPage />} />
              <Route path="/tasks" element={<TasksPage />} />
              <Route path="/notifications" element={<NotificationsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </React.Suspense>
        </MainLayout>
      </ProjectInitializer>
    </ConfigProvider>
  );
};

export default App;
