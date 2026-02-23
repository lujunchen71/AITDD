
import React from 'react';
import { Tree, Button, Empty, Spin } from 'antd';
import { PlusOutlined, FolderOutlined, FolderOpenOutlined } from '@ant-design/icons';
import type { TreeDataNode, TreeProps } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface ModuleTreeProps {
  projectId: string;
  onSelectModule: (moduleId: string) => void;
  selectedModuleId?: string | null;
  onAddModule?: (parentId?: string) => void;
}

const ModuleTree: React.FC<ModuleTreeProps> = ({
  projectId,
  onSelectModule,
  selectedModuleId,
  onAddModule,
}) => {
  // 获取模块列表
  const { data, isLoading, error } = useQuery({
    queryKey: ['modules', projectId],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/modules', {
        params: { projectId },
      });
      return response.data;
    },
  });

  // 转换为树形数据
  const convertToTreeData = (modules: any[]): TreeDataNode[] => {
    const moduleMap = new Map<string, TreeDataNode>();
    const rootNodes: TreeDataNode[] = [];

    // 先创建所有节点
    modules.forEach((module) => {
      moduleMap.set(module.id, {
        key: module.id,
        title: module.name,
        icon: ({ expanded }) => expanded ? <FolderOpenOutlined /> : <FolderOutlined />,
        children: [],
      });
    });

    // 构建树结构
    modules.forEach((module) => {
      const node = moduleMap.get(module.id)!;
      if (module.parentId && moduleMap.has(module.parentId)) {
        const parent = moduleMap.get(module.parentId)!;
        (parent.children as TreeDataNode[]).push(node);
      } else {
        rootNodes.push(node);
      }
    });

    return rootNodes;
  };

  const treeData = data?.modules ? convertToTreeData(data.modules) : [];

  const handleSelect: TreeProps['onSelect'] = (selectedKeys) => {
    if (selectedKeys.length > 0) {
      onSelectModule(selectedKeys[0] as string);
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 text-red-500">
        加载模块失败
      </div>
    );
  }

  return (
    <div className="module-tree">
      <div className="module-tree-header flex justify-between items-center mb-4">
        <span className="text-lg font-medium">模块列表</span>
        {onAddModule && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            size="small"
            onClick={() => onAddModule()}
          >
            新建
          </Button>
        )}
      </div>

      {treeData.length === 0 ? (
        <Empty description="暂无模块" />
      ) : (
        <Tree
          showIcon
          defaultExpandAll
          selectedKeys={selectedModuleId ? [selectedModuleId] : []}
          treeData={treeData}
          onSelect={handleSelect}
        />
      )}
    </div>
  );
};

export default ModuleTree;
