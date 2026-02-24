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
  const { data, isLoading, error } = useQuery({
    queryKey: ['modules', projectId],
    queryFn: async () => {
      const response = await apiClient.get('/modules', {
        params: { projectId },
      });
      // API 响应格式：{success: true, data: {modules: [], total: 0}}
      return response.data?.data?.modules || [];
    },
  });

  const convertToTreeData = (modules: any[]): TreeDataNode[] => {
    const moduleMap = new Map<string, TreeDataNode>();
    const rootNodes: TreeDataNode[] = [];

    modules.forEach((module) => {
      moduleMap.set(module.id, {
        key: module.id,
        title: module.name,
        icon: ({ expanded }) => expanded ? <FolderOpenOutlined style={{ color: '#00d9ff' }} /> : <FolderOutlined style={{ color: '#a0a0a0' }} />,
        children: [],
      });
    });

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

  const treeData = data && Array.isArray(data) ? convertToTreeData(data) : [];

  const handleSelect: TreeProps['onSelect'] = (selectedKeys) => {
    if (selectedKeys.length > 0) {
      onSelectModule(selectedKeys[0] as string);
    }
  };

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}>
        <Spin />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '16px', color: '#ff4d4f', textAlign: 'center' }}>
        加载模块失败
      </div>
    );
  }

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: '12px',
        paddingBottom: '8px',
        borderBottom: '1px solid #2d2d44',
      }}>
        <span style={{ fontSize: '13px', fontWeight: 500, color: '#a0a0a0' }}>模块列表</span>
        {onAddModule && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            size="small"
            onClick={() => onAddModule()}
            style={{
              background: 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)',
              border: 'none',
              height: 24,
              fontSize: 12,
            }}
          >
            新建
          </Button>
        )}
      </div>

      {treeData.length === 0 ? (
        <Empty 
          description="暂无模块" 
          imageStyle={{ opacity: 0.5 }}
          style={{ padding: '20px 0' }}
        />
      ) : (
        <Tree
          showIcon
          defaultExpandAll
          selectedKeys={selectedModuleId ? [selectedModuleId] : []}
          treeData={treeData}
          onSelect={handleSelect}
          style={{ 
            background: 'transparent',
            fontSize: '13px',
          }}
          blockNode
        />
      )}
    </div>
  );
};

export default ModuleTree;
