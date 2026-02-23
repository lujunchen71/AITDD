import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import MainLayout from './components/layout/MainLayout';
import './App.css';

// 占位页面组件
const DashboardPage = React.lazy(() => import('./features/dashboard'));
const ModulesPage = React.lazy(() => import('./features/modules'));
const TasksPage = React.lazy(() => import('./features/tasks'));
const NotificationsPage = React.lazy(() => import('./features/notifications'));
const SettingsPage = React.lazy(() => import('./features/settings'));

const App: React.FC = () => {
  return (
    <ConfigProvider locale={zhCN}>
      <MainLayout>
        <React.Suspense fallback={<div>加载中...</div>}>
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
    </ConfigProvider>
  );
};

export default App;
