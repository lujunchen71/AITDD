import React, { useState, useEffect } from 'react';
import { Layout, Button, Badge, Tooltip, Dropdown, Avatar, Space, Checkbox } from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BellOutlined,
  SettingOutlined,
  QuestionCircleOutlined,
  FolderOutlined,
  DownOutlined,
  CheckOutlined,
} from '@ant-design/icons';
import { useUIStore } from '../../../stores/useUIStore';
import { useProjectStore, useProjects, useSelectedProjectIds } from '../../../stores/useProjectStore';
import type { MenuProps } from 'antd';
import './index.css';

const { Header: AntHeader } = Layout;

const Header: React.FC = () => {
  const { sidebarCollapsed, toggleSidebar, setNotificationDrawerVisible } = useUIStore();
  const { fetchProjects, toggleProjectSelection, selectAllProjects, deselectAllProjects } = useProjectStore();
  const projects = useProjects();
  const selectedProjectIds = useSelectedProjectIds();
  const [dropdownVisible, setDropdownVisible] = useState(false);

  // [DEBUG] Header 渲染调试
  useEffect(() => {
    console.log('[Header] Rendered, projects count:', projects.length, 'selectedProjectIds:', selectedProjectIds);
  });

  // 加载项目列表
  useEffect(() => {
    console.log('[Header] useEffect: fetchProjects called');
    fetchProjects().then(() => {
      console.log('[Header] fetchProjects completed');
      // 注意：这里 projects.length 可能还是旧值，因为状态更新是异步的
    });
  }, [fetchProjects]);

  // 监控 projects 变化
  useEffect(() => {
    console.log('[Header] projects 变化:', projects.length, 'items');
  }, [projects]);

  // 监控 selectedProjectIds 变化
  useEffect(() => {
    console.log('[Header] selectedProjectIds 变化:', selectedProjectIds);
  }, [selectedProjectIds]);

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

  // 项目选择下拉内容
  const projectDropdownContent = (
    <div className="project-dropdown">
      <div className="project-dropdown-header">
        <span className="project-dropdown-title">选择项目</span>
        <div className="project-dropdown-actions">
          <Button 
            type="link" 
            size="small" 
            onClick={(e) => {
              e.stopPropagation();
              selectAllProjects();
            }}
          >
            全选
          </Button>
          <Button 
            type="link" 
            size="small" 
            onClick={(e) => {
              e.stopPropagation();
              deselectAllProjects();
            }}
          >
            清空
          </Button>
        </div>
      </div>
      <div className="project-dropdown-list">
        {projects.map((project) => {
          const isSelected = selectedProjectIds.includes(project.id);
          return (
            <div
              key={project.id}
              className={`project-dropdown-item ${isSelected ? 'selected' : ''}`}
              onClick={(e) => {
                e.stopPropagation();
                toggleProjectSelection(project.id);
              }}
            >
              <Checkbox 
                checked={isSelected}
                onClick={(e) => e.stopPropagation()}
                onChange={() => toggleProjectSelection(project.id)}
              />
              <FolderOutlined className="project-icon" />
              <span className="project-name">{project.name}</span>
              {isSelected && <CheckOutlined className="check-icon" />}
            </div>
          );
        })}
        {projects.length === 0 && (
          <div className="project-dropdown-empty">
            暂无项目
          </div>
        )}
      </div>
      <div className="project-dropdown-footer">
        <span>已选择 {selectedProjectIds.length} / {projects.length} 个项目</span>
      </div>
    </div>
  );

  // 获取显示的选中项目文本
  const getSelectedProjectsText = () => {
    if (selectedProjectIds.length === 0) {
      return '未选择项目';
    }
    if (selectedProjectIds.length === projects.length && projects.length > 0) {
      return '全部项目';
    }
    if (selectedProjectIds.length === 1) {
      const selectedProject = projects.find(p => p.id === selectedProjectIds[0]);
      return selectedProject?.name || '已选择1个项目';
    }
    return `已选择 ${selectedProjectIds.length} 个项目`;
  };

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
        
        {/* 项目选择器 - 移到左侧 */}
        <Dropdown
          dropdownRender={() => projectDropdownContent}
          trigger={['click']}
          open={dropdownVisible}
          onOpenChange={setDropdownVisible}
          placement="bottomLeft"
        >
          <Button className="project-selector-btn">
            <FolderOutlined />
            <span className="project-selector-text">{getSelectedProjectsText()}</span>
            <DownOutlined style={{ fontSize: 10, marginLeft: 4 }} />
          </Button>
        </Dropdown>
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
