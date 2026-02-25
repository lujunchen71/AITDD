import React from 'react';
import { Dropdown, Tag, Button } from 'antd';
import { MoreOutlined, EditOutlined, DeleteOutlined, PlusOutlined, FileTextOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';

interface ModuleTreeNodeProps {
  module: {
    id: string;
    name: string;
    status: string;
    testCoverage?: number;
  };
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAddChild?: (parentId: string) => void;
  onAddTask?: (moduleId: string) => void;
}

const statusColors: Record<string, string> = {
  designing: 'blue',
  developing: 'orange',
  testing: 'cyan',
  done: 'green',
};

const statusLabels: Record<string, string> = {
  designing: '设计中',
  developing: '开发中',
  testing: '测试中',
  done: '已完成',
};

const ModuleTreeNode: React.FC<ModuleTreeNodeProps> = ({
  module,
  onEdit,
  onDelete,
  onAddChild,
  onAddTask,
}) => {
  const menuItems: MenuProps['items'] = [
    {
      key: 'add-task',
      label: '创建任务',
      icon: <FileTextOutlined />,
      onClick: () => onAddTask?.(module.id),
    },
    {
      key: 'add',
      label: '添加子模块',
      icon: <PlusOutlined />,
      onClick: () => onAddChild?.(module.id),
    },
    {
      key: 'edit',
      label: '编辑',
      icon: <EditOutlined />,
      onClick: () => onEdit?.(module.id),
    },
    {
      type: 'divider',
    },
    {
      key: 'delete',
      label: '删除',
      icon: <DeleteOutlined />,
      danger: true,
      onClick: () => onDelete?.(module.id),
    },
  ];

  return (
    <div className="module-tree-node flex items-center justify-between w-full pr-2">
      <div className="flex items-center gap-2">
        <span className="module-name">{module.name}</span>
        <Tag color={statusColors[module.status] || 'default'}>
          {statusLabels[module.status] || module.status}
        </Tag>
        {module.testCoverage !== undefined && (
          <span className="text-xs text-gray-400">
            覆盖率: {module.testCoverage}%
          </span>
        )}
      </div>
      <Dropdown menu={{ items: menuItems }} trigger={['click']}>
        <Button type="text" size="small" icon={<MoreOutlined />} />
      </Dropdown>
    </div>
  );
};

export default ModuleTreeNode;
