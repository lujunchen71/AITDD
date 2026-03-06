import React from 'react';
import { Button, Typography } from 'antd';
import { CloseOutlined, AppstoreOutlined, BlockOutlined, CheckSquareOutlined, FileTextOutlined, LinkOutlined, BugOutlined, ExperimentOutlined, ApartmentOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { Module, Task, TaskDependency, Project } from '../../../../types';
import { usePropertyPanelStore, PanelContentType } from '../../../../stores/usePropertyPanelStore';
import './index.css';

// 子视图
import ProjectView from './views/ProjectView';
import ModuleView from './views/ModuleView';
import TaskView from './views/TaskView';
import PromptView from './views/PromptView';
import ContractView from './views/ContractView';
import ErrorView from './views/ErrorView';
import TestsView from './views/TestsView';
import DependenciesView from './views/DependenciesView';

interface PropertyPanelProps {
  project: Project | null;
  modules: Module[];
  tasks: Task[];
  taskDependencies: TaskDependency[];
  // 面板定位：右侧偏移（按钮位置）
  anchorRight?: number;
  anchorTop?: number;
}

// 标题图标和文字映射
const contentTypeInfo: Record<PanelContentType, { icon: React.ReactNode; label: string }> = {
  project: { icon: <AppstoreOutlined />, label: '项目信息' },
  module: { icon: <BlockOutlined />, label: '模块信息' },
  task: { icon: <CheckSquareOutlined />, label: '任务信息' },
  prompt: { icon: <FileTextOutlined />, label: '提示词' },
  'upstream-contract': { icon: <LinkOutlined />, label: '上游契约' },
  'downstream-contract': { icon: <LinkOutlined />, label: '下游契约' },
  error: { icon: <BugOutlined />, label: '错误信息' },
  tests: { icon: <ExperimentOutlined />, label: '测试用例' },
  dependencies: { icon: <ApartmentOutlined />, label: '依赖关系' },
};

const PropertyPanel: React.FC<PropertyPanelProps> = ({
  project,
  modules,
  tasks,
  taskDependencies,
  anchorRight = 16,
  anchorTop = 96,
}) => {
  const { visible, contentType, selectedModuleId, selectedTaskId, close, canGoBack, goBack } = usePropertyPanelStore();

  if (!visible) return null;

  const { icon, label } = contentTypeInfo[contentType] || { icon: null, label: '属性面板' };

  // 根据 contentType 渲染对应子视图
  const renderContent = () => {
    switch (contentType) {
      case 'project':
        return (
          <ProjectView
            project={project}
            modules={modules}
            tasks={tasks}
            taskDependencies={taskDependencies}
          />
        );
      case 'module':
        return (
          <ModuleView
            moduleId={selectedModuleId}
            modules={modules}
            tasks={tasks}
          />
        );
      case 'task':
        return (
          <TaskView
            taskId={selectedTaskId}
            modules={modules}
            tasks={tasks}
          />
        );
      case 'prompt':
        return (
          <PromptView
            taskId={selectedTaskId}
            tasks={tasks}
          />
        );
      case 'upstream-contract':
        return (
          <ContractView
            taskId={selectedTaskId}
            tasks={tasks}
            type="upstream"
          />
        );
      case 'downstream-contract':
        return (
          <ContractView
            taskId={selectedTaskId}
            tasks={tasks}
            type="downstream"
          />
        );
      case 'error':
        return (
          <ErrorView
            taskId={selectedTaskId}
            moduleId={selectedModuleId}
            tasks={tasks}
            modules={modules}
            project={project}
          />
        );
      case 'tests':
        return (
          <TestsView
            taskId={selectedTaskId}
            tasks={tasks}
          />
        );
      case 'dependencies':
        return (
          <DependenciesView
            taskId={selectedTaskId}
            tasks={tasks}
            modules={modules}
            taskDependencies={taskDependencies}
          />
        );
      default:
        return <div style={{ color: '#666', fontSize: 12 }}>未知内容类型</div>;
    }
  };

  return (
    <div
      className="property-panel"
      style={{
        right: anchorRight,
        top: anchorTop,
        bottom: 40,
      }}
    >
      {/* 面板标题栏 */}
      <div className="property-panel-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          {/* 返回按钮 */}
          {canGoBack && (
            <Button
              type="text"
              icon={<ArrowLeftOutlined />}
              size="small"
              onClick={goBack}
              style={{ color: '#00d9ff', padding: '0 4px' }}
              title="返回"
            />
          )}
          <div className="property-panel-title">
            {icon && <span style={{ color: '#00d9ff' }}>{icon}</span>}
            <Typography.Text style={{ color: '#e0e0e0', fontSize: 13, fontWeight: 600 }}>
              {label}
            </Typography.Text>
          </div>
        </div>
        <Button
          type="text"
          icon={<CloseOutlined />}
          size="small"
          onClick={close}
          style={{ color: '#888', padding: '0 4px' }}
        />
      </div>

      {/* 可滚动内容区 */}
      <div className="property-panel-content">
        {renderContent()}
      </div>
    </div>
  );
};

export default PropertyPanel;
