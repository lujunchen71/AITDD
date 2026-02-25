import React, { useState } from 'react';
import { Tree, Button, Empty, Spin, Dropdown } from 'antd';
import { PlusOutlined, FolderOutlined, FolderOpenOutlined, MoreOutlined, FileTextOutlined, FileOutlined } from '@ant-design/icons';
import type { TreeDataNode, TreeProps, MenuProps } from 'antd';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';

interface ModuleTreeProps {
  projectId: string;
  onSelectModule: (moduleId: string) => void;
  selectedModuleId?: string | null;
  onAddModule?: (parentId?: string) => void;
  onAddTask?: (moduleId: string) => void;
  onDeleteModule?: (moduleId: string) => void;
}

const ModuleTree: React.FC<ModuleTreeProps> = ({
  projectId,
  onSelectModule,
  selectedModuleId,
  onAddModule,
  onAddTask,
  onDeleteModule,
}) => {
  const queryClient = useQueryClient();
  const [contextMenuOpen, setContextMenuOpen] = useState(false);
  const [selectedModuleForMenu, setSelectedModuleForMenu] = useState<{ id: string; name: string } | null>(null);

  const { data: modules, isLoading, error } = useQuery({
    queryKey: ['modules', projectId],
    queryFn: async () => {
      const response = await apiClient.get('/modules', {
        params: { projectId },
      });
      console.log('ModuleTree API response:', response);
      // API 响应格式：{success: true, data: {modules: [], total: 0}}
      const modulesList = response.data?.data?.modules || response.data?.modules || [];
      
      // 为每个模块获取其任务
      const modulesWithTasks = await Promise.all(
        modulesList.map(async (module: any) => {
          try {
            const tasksResponse = await apiClient.get(`/modules/${module.id}/tasks`);
            const tasks = tasksResponse.data?.data?.tasks || tasksResponse.data?.tasks || [];
            return { ...module, tasks };
          } catch (err) {
            console.error(`Failed to fetch tasks for module ${module.id}:`, err);
            return { ...module, tasks: [] };
          }
        })
      );
      
      return modulesWithTasks;
    },
  });

  const convertToTreeData = (modulesData: any[]): TreeDataNode[] => {
    const moduleMap = new Map<string, TreeDataNode>();
    const rootNodes: TreeDataNode[] = [];

    // 先排序，确保父模块在子模块之前
    const sortedModules = [...modulesData].sort((a, b) => {
      if (a.parentId && !b.parentId) return 1;
      if (!a.parentId && b.parentId) return -1;
      return 0;
    });

    sortedModules.forEach((module) => {
      const menuItems: MenuProps['items'] = [
        {
          key: 'add-task',
          label: '创建任务',
          icon: <FileTextOutlined style={{ color: '#00d9ff' }} />,
          onClick: (e) => {
            e?.domEvent?.stopPropagation();
            onAddTask?.(module.id);
          },
        },
        {
          key: 'add',
          label: '添加子模块',
          icon: <PlusOutlined style={{ color: '#00d9ff' }} />,
          onClick: (e) => {
            e?.domEvent?.stopPropagation();
            onAddModule?.(module.id);
          },
        },
        {
          type: 'divider',
        },
        {
          key: 'delete',
          label: '删除模块',
          icon: <span style={{ color: '#ff4d4f' }}>🗑️</span>,
          danger: true,
          onClick: (e) => {
            e?.domEvent?.stopPropagation();
            if (window.confirm(`确定要删除模块 "${module.name}" 吗？`)) {
              onDeleteModule?.(module.id);
            }
          },
        },
      ];

      moduleMap.set(module.id, {
        key: module.id,
        title: (
          <div style={{
            display: 'flex',
            flexDirection: 'row',
            justifyContent: 'space-between',
            alignItems: 'center',
            width: '100%',
            paddingRight: '4px',
            height: '28px',
            gap: '8px',
          }}>
            <span style={{
              flex: 1,
              fontSize: '13px',
              color: '#400a18',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}>{module.name}</span>
            <Dropdown
              menu={{ items: menuItems }}
              trigger={['click']}
              open={contextMenuOpen && selectedModuleForMenu?.id === module.id}
              onOpenChange={(open) => {
                setContextMenuOpen(open);
                if (open) {
                  setSelectedModuleForMenu({ id: module.id, name: module.name });
                } else {
                  setSelectedModuleForMenu(null);
                }
              }}
              getPopupContainer={(trigger) => trigger.parentNode as HTMLElement}
            >
              <Button
                type="text"
                size="small"
                icon={<MoreOutlined />}
                onClick={(e) => e.stopPropagation()}
                style={{
                  transition: 'all 0.2s',
                  color: '#fff',
                  padding: '0 6px',
                  height: '24px',
                  minWidth: '28px',
                  background: 'rgba(233, 69, 96, 0.8)',
                  border: '1px solid rgba(233, 69, 96, 0.6)',
                  boxShadow: '0 2px 6px rgba(233, 69, 96, 0.3)',
                  borderRadius: '4px',
                  fontWeight: 'bold',
                }}
                className="module-tree-action"
                onMouseEnter={(e) => {
                  e.currentTarget.style.color = '#fff';
                  e.currentTarget.style.background = '#e94560';
                  e.currentTarget.style.borderColor = '#e94560';
                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(233, 69, 96, 0.5)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.color = '#fff';
                  e.currentTarget.style.background = 'rgba(233, 69, 96, 0.8)';
                  e.currentTarget.style.borderColor = 'rgba(233, 69, 96, 0.6)';
                  e.currentTarget.style.boxShadow = '0 2px 6px rgba(233, 69, 96, 0.3)';
                }}
              />
            </Dropdown>
          </div>
        ),
        icon: ({ expanded }) => expanded ?
          <FolderOpenOutlined style={{ color: '#00d9ff', fontSize: '14px' }} /> :
          <FolderOutlined style={{ color: '#00d9ff', fontSize: '14px' }} />,
        children: [],
        isLeaf: false, // 始终允许展开，因为可能有子模块或任务
        switcherIcon: ({ expanded }) => (
          <span style={{
            color: '#6b7280',
            fontSize: '10px',
            transition: 'transform 0.2s',
            display: 'inline-block',
            transform: expanded ? 'rotate(90deg)' : 'rotate(0deg)',
          }}>
            ▶
          </span>
        ),
      });
    });

    // 构建父子关系
    modulesData.forEach((module) => {
      const node = moduleMap.get(module.id)!;
      if (module.parentId && moduleMap.has(module.parentId)) {
        const parent = moduleMap.get(module.parentId)!;
        (parent.children as TreeDataNode[]).push(node);
      } else {
        rootNodes.push(node);
      }
    });

    // 为每个模块添加任务子节点
    modulesData.forEach((module) => {
      const node = moduleMap.get(module.id);
      if (node && module.tasks && Array.isArray(module.tasks) && module.tasks.length > 0) {
        const taskChildren: TreeDataNode[] = module.tasks.map((task: any) => ({
          key: `task-${task.id}`,
          title: (
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              fontSize: '12px',
              color: '#9ca3af',
            }}>
              <FileOutlined style={{ fontSize: '12px', color: '#6b7280' }} />
              <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{task.name}</span>
            </div>
          ),
          icon: <FileOutlined style={{ color: '#6b7280', fontSize: '12px' }} />,
          isLeaf: true,
        }));
        node.children = [...(node.children || []), ...taskChildren];
        // 更新 isLeaf：有任务的模块不是叶子节点
        node.isLeaf = false;
      }
    });

    return rootNodes;
  };

  const treeData = modules && Array.isArray(modules) ? convertToTreeData(modules) : [];

  const handleSelect: TreeProps['onSelect'] = (selectedKeys) => {
    if (selectedKeys.length > 0) {
      onSelectModule(selectedKeys[0] as string);
    }
  };

  // 处理模块添加后的刷新
  React.useEffect(() => {
    if (onAddModule || onAddTask) {
      // 订阅查询客户端，在添加操作后刷新数据
      const unsubscribe = queryClient.getQueryCache().subscribe((event) => {
        if (event?.query.queryKey[0] === 'modules') {
          // 数据已更新，无需操作
        }
      });
      return () => unsubscribe();
    }
  }, [queryClient, onAddModule, onAddTask]);

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}>
        <Spin size="small" />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '16px', color: '#ff4d4f', textAlign: 'center', fontSize: '13px' }}>
        加载模块失败
      </div>
    );
  }

  return (
    <div style={{ 
      height: '100%', 
      display: 'flex', 
      flexDirection: 'column',
      background: '#16213e',
    }}>
      <style>{`
        .module-tree-with-actions .ant-tree-treenode:hover .module-tree-action,
        .module-tree-with-actions .module-tree-action,
        .module-tree-with-actions .ant-tree-treenode .module-tree-action {
          opacity: 1 !important;
          visibility: visible !important;
          display: inline-flex !important;
        }
        .module-tree-with-actions .ant-dropdown-menu {
          background: #1f1f1f !important;
          border: 1px solid #2d2d44 !important;
        }
        .module-tree-with-actions .ant-dropdown-menu-item {
          color: #d1d5db !important;
        }
        .module-tree-with-actions .ant-dropdown-menu-item:hover {
          background: rgba(255, 255, 255, 0.08) !important;
        }
        .module-tree-with-actions .ant-tree-node-content-wrapper {
          padding: 4px 8px !important;
          border-radius: 4px;
          margin: 1px 4px;
          min-height: 28px !important;
          height: 28px !important;
          display: flex !important;
          align-items: center !important;
        }
        .module-tree-with-actions .ant-tree-node-content-wrapper:hover {
          background-color: rgba(255, 255, 255, 0.08);
        }
        .module-tree-with-actions .ant-tree-node-content-wrapper.ant-tree-node-selected {
          background-color: rgba(0, 217, 255, 0.2);
        }
        .module-tree-with-actions .ant-tree-indent-unit {
          width: 16px;
        }
        .module-tree-with-actions .ant-tree-switcher {
          width: 16px;
        }
        .module-tree-with-actions .ant-tree {
          background: transparent !important;
          color: #d1d5db;
        }
        .module-tree-with-actions .ant-tree-treenode {
          color: #d1d5db;
        }
        .module-tree-with-actions .ant-tree-title {
          color: #d1d5db;
        }
        .module-tree-with-actions .ant-tree-node-content-wrapper .ant-tree-icon {
          color: #00d9ff;
        }
        .module-tree-with-actions .ant-tree-title {
          display: flex !important;
          align-items: center !important;
          height: 100% !important;
          width: 100% !important;
          flex: 1 !important;
        }
        .module-tree-with-actions .ant-tree-treenode {
          width: 100% !important;
        }
        .module-tree-with-actions .ant-tree-node-content-wrapper {
          flex: 1 !important;
          width: 100% !important;
        }
        .module-tree-with-actions .ant-tree-iconEle {
          background: transparent !important;
        }
        .module-tree-with-actions .ant-tree-icon__customize {
          background: transparent !important;
        }
        .module-tree-with-actions .ant-tree-switcher {
          background: transparent !important;
        }
      `}</style>
      <div style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center', 
        marginBottom: '12px',
        paddingBottom: '10px',
        borderBottom: '1px solid #2d2d44',
        paddingLeft: '8px',
        paddingRight: '8px',
      }}>
        <span style={{ 
          fontSize: '12px', 
          fontWeight: 600, 
          color: '#a0a0a0',
          textTransform: 'uppercase',
          letterSpacing: '0.5px',
        }}>模块列表</span>
        {onAddModule && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            size="small"
            onClick={() => {
              onAddModule();
              // 添加后刷新数据
              queryClient.invalidateQueries({ queryKey: ['modules', projectId] });
            }}
            style={{
              background: 'linear-gradient(135deg, #e94560 0%, #ff6b6b 100%)',
              border: 'none',
              height: 26,
              fontSize: 12,
              fontWeight: 500,
              boxShadow: '0 2px 8px rgba(233, 69, 96, 0.3)',
            }}
          >
            新建
          </Button>
        )}
      </div>

      {treeData.length === 0 ? (
        <Empty 
          description={<span style={{ color: '#6b7280', fontSize: '13px' }}>暂无模块</span>} 
          styles={{ 
            image: { opacity: 0.4, filter: 'grayscale(100%)' },
            description: { color: '#6b7280' }
          }}
          style={{ padding: '30px 0' }}
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
            padding: '4px',
          }}
          blockNode
          className="module-tree-with-actions"
          expandAction="click"
        />
      )}
    </div>
  );
};

export default ModuleTree;
