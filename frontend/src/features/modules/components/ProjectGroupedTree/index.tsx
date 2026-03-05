import React, { useState, useEffect } from 'react';
import { Tree, Button, Empty, Spin, Dropdown } from 'antd';
import { 
  PlusOutlined, 
  FolderOutlined, 
  FolderOpenOutlined, 
  MoreOutlined, 
  FileTextOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  CheckCircleOutlined, 
  ReloadOutlined,
  ProjectOutlined,
  DownOutlined,
  RightOutlined,
} from '@ant-design/icons';
import type { TreeDataNode, TreeProps, MenuProps } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';
import { localStorageService } from '../../../../services/localStorageService';
import { Project } from '../../../../stores/useProjectStore';
import './index.css';

interface ProjectGroupedTreeProps {
  selectedProjectIds: string[];
  projects: Project[];
  onSelectModule: (moduleId: string, projectId: string) => void;
  selectedModuleId?: string | null;
  onAddModule?: (projectId: string, parentId?: string) => void;
  onAddTask?: (moduleId: string, projectId: string) => void;
  onDeleteModule?: (moduleId: string, projectId: string) => void;
  onEditTask?: (taskId: string, projectId: string) => void;
  onDeleteTask?: (taskId: string, projectId: string) => void;
  onCheckTask?: (taskId: string, projectId: string) => void;
  onRefactorTask?: (taskId: string, projectId: string) => void;
}

interface ModuleTreeByProjectProps {
  projectId: string;
  onSelectModule: (moduleId: string, projectId: string) => void;
  selectedModuleId?: string | null;
  onAddModule?: (projectId: string, parentId?: string) => void;
  onAddTask?: (moduleId: string, projectId: string) => void;
  onDeleteModule?: (moduleId: string, projectId: string) => void;
  onEditTask?: (taskId: string, projectId: string) => void;
  onDeleteTask?: (taskId: string, projectId: string) => void;
  onCheckTask?: (taskId: string, projectId: string) => void;
  onRefactorTask?: (taskId: string, projectId: string) => void;
}

// 单个项目的模块树组件
const ModuleTreeByProject: React.FC<ModuleTreeByProjectProps> = ({
  projectId,
  onSelectModule,
  selectedModuleId,
  onAddModule,
  onAddTask,
  onDeleteModule,
  onEditTask,
  onDeleteTask,
  onCheckTask,
  onRefactorTask,
}) => {
  const [contextMenuOpen, setContextMenuOpen] = useState(false);
  const [selectedModuleForMenu, setSelectedModuleForMenu] = useState<{ id: string; name: string } | null>(null);
  const [taskMenuOpen, setTaskMenuOpen] = useState(false);
  const [selectedTaskForMenu, setSelectedTaskForMenu] = useState<{ id: string; name: string } | null>(null);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  const { data: modules, isLoading, error } = useQuery({
    queryKey: ['modules', projectId],
    queryFn: async () => {
      const response = await apiClient.get('/modules', {
        params: { projectId },
      });
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
    enabled: !!projectId,
  });

  const convertToTreeData = (modulesData: any[]): TreeDataNode[] => {
    // 防御性检查：确保 modulesData 是有效的数组
    if (!modulesData || !Array.isArray(modulesData)) {
      return [];
    }

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
          icon: <FileTextOutlined style={{ color: '#e94560' }} />,
          onClick: (e) => {
            e?.domEvent?.stopPropagation();
            onAddTask?.(module.id, projectId);
          },
        },
        {
          key: 'add',
          label: '添加子模块',
          icon: <PlusOutlined style={{ color: '#e94560' }} />,
          onClick: (e) => {
            e?.domEvent?.stopPropagation();
            onAddModule?.(projectId, module.id);
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
              onDeleteModule?.(module.id, projectId);
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
            height: '24px',
            gap: '6px',
          }}>
            <span style={{
              flex: 1,
              fontSize: '14px',
              color: '#f87171',
              fontWeight: 700,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}>{module.name}</span>
            <Dropdown
              menu={{ items: menuItems }}
              trigger={['click']}
              placement="bottomRight"
              open={contextMenuOpen && selectedModuleForMenu?.id === module.id}
              onOpenChange={(open) => {
                setContextMenuOpen(open);
                if (open) {
                  setSelectedModuleForMenu({ id: module.id, name: module.name });
                } else {
                  setSelectedModuleForMenu(null);
                }
              }}
              destroyOnHidden
            >
              <Button
                type="text"
                size="small"
                icon={<MoreOutlined />}
                onClick={(e) => e.stopPropagation()}
                style={{
                  transition: 'all 0.2s',
                  color: '#a0a0a0',
                  padding: '0 6px',
                  height: '24px',
                  minWidth: '28px',
                  background: 'transparent',
                  border: '1px solid transparent',
                  borderRadius: '4px',
                }}
                className="module-tree-action"
                onMouseEnter={(e) => {
                  e.currentTarget.style.color = '#f87171';
                  e.currentTarget.style.background = 'rgba(248, 113, 113, 0.1)';
                  e.currentTarget.style.borderColor = 'rgba(248, 113, 113, 0.3)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.color = '#a0a0a0';
                  e.currentTarget.style.background = 'transparent';
                  e.currentTarget.style.borderColor = 'transparent';
                }}
              />
            </Dropdown>
          </div>
        ),
        icon: ({ expanded }) => expanded ?
          <FolderOpenOutlined style={{ color: '#f87171', fontSize: '14px' }} /> :
          <FolderOutlined style={{ color: '#f87171', fontSize: '14px' }} />,
        children: [],
        isLeaf: false,
        switcherIcon: ({ expanded }) => expanded ? (
          <DownOutlined style={{ fontSize: 12, color: '#f87171' }} />
        ) : (
          <RightOutlined style={{ fontSize: 12, color: '#f87171' }} />
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
        const taskChildren: TreeDataNode[] = module.tasks.map((task: any) => {
          const taskMenuItems: MenuProps['items'] = [
            {
              key: 'edit',
              label: '编辑',
              icon: <EditOutlined style={{ color: '#e94560' }} />,
              onClick: (e) => {
                e?.domEvent?.stopPropagation();
                onEditTask?.(task.id, projectId);
              },
            },
            {
              key: 'check',
              label: '检查',
              icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
              onClick: (e) => {
                e?.domEvent?.stopPropagation();
                onCheckTask?.(task.id, projectId);
              },
            },
            {
              key: 'refactor',
              label: '重构',
              icon: <ReloadOutlined style={{ color: '#1890ff' }} />,
              onClick: (e) => {
                e?.domEvent?.stopPropagation();
                onRefactorTask?.(task.id, projectId);
              },
            },
            {
              type: 'divider',
            },
            {
              key: 'delete',
              label: '删除',
              icon: <DeleteOutlined style={{ color: '#ff4d4f' }} />,
              danger: true,
              onClick: (e) => {
                e?.domEvent?.stopPropagation();
                if (window.confirm(`确定要删除任务 "${task.name}" 吗？`)) {
                  onDeleteTask?.(task.id, projectId);
                }
              },
            },
          ];

          return {
            key: `task-${task.id}`,
            title: (
              <div style={{
                display: 'flex',
                flexDirection: 'row',
                justifyContent: 'space-between',
                alignItems: 'center',
                width: '100%',
                paddingRight: '4px',
                height: '24px',
                gap: '6px',
              }}>
                <span style={{
                  flex: 1,
                  fontSize: '13px',
                  color: task.status === 'completed' ? '#10b981' : '#fff2f2',
                  fontWeight: 400,
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  textDecoration: task.status === 'completed' ? 'line-through' : 'none',
                }}>{task.name}</span>
                <Dropdown
                  menu={{ items: taskMenuItems }}
                  trigger={['click']}
                  placement="bottomRight"
                  open={taskMenuOpen && selectedTaskForMenu?.id === task.id}
                  onOpenChange={(open) => {
                    setTaskMenuOpen(open);
                    if (open) {
                      setSelectedTaskForMenu({ id: task.id, name: task.name });
                    } else {
                      setSelectedTaskForMenu(null);
                    }
                  }}
                  destroyOnHidden
                >
                  <Button
                    type="text"
                    size="small"
                    icon={<MoreOutlined />}
                    onClick={(e) => e.stopPropagation()}
                    style={{
                      transition: 'all 0.2s',
                      color: '#808080',
                      padding: '0 6px',
                      height: '24px',
                      minWidth: '28px',
                      background: 'transparent',
                      border: '1px solid transparent',
                      borderRadius: '4px',
                    }}
                    className="task-tree-action"
                    onMouseEnter={(e) => {
                      e.currentTarget.style.color = '#fff2f2';
                      e.currentTarget.style.background = 'rgba(255, 242, 242, 0.15)';
                      e.currentTarget.style.borderColor = 'rgba(255, 242, 242, 0.3)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.color = '#808080';
                      e.currentTarget.style.background = 'transparent';
                      e.currentTarget.style.borderColor = 'transparent';
                    }}
                  />
                </Dropdown>
              </div>
            ),
            icon: <FileTextOutlined style={{ color: task.status === 'completed' ? '#10b981' : '#fff2f2', fontSize: '14px' }} />,
            isLeaf: true,
          };
        });
        
        if (node.children) {
          node.children = [...taskChildren, ...node.children];
        } else {
          node.children = taskChildren;
        }
      }
    });

    return rootNodes;
  };

  const handleSelect: TreeProps['onSelect'] = (selectedKeys, _info) => {
    if (selectedKeys.length > 0) {
      const key = selectedKeys[0] as string;
      if (!key.startsWith('task-')) {
        onSelectModule(key, projectId);
      }
    }
  };

  if (isLoading) {
    return (
      <div style={{ padding: '20px', textAlign: 'center' }}>
        <Spin size="small" />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '20px', textAlign: 'center', color: '#ff4d4f' }}>
        加载失败
      </div>
    );
  }

  const treeData = modules ? convertToTreeData(modules as any[]) : [];

  if (treeData.length === 0) {
    return (
      <div style={{ padding: '20px', textAlign: 'center', color: '#808080' }}>
        暂无模块
      </div>
    );
  }

  const handleExpand: TreeProps['onExpand'] = (expandedKeys, info) => {
    // 只有通过 switcher（三角图标）触发的展开才允许
    // 通过点击节点标题触发的展开会被忽略
    if (info.nativeEvent) {
      const target = info.nativeEvent.target as HTMLElement;
      // 检查点击是否来自 switcher
      if (target.closest('.ant-tree-switcher')) {
        setExpandedKeys(expandedKeys);
      }
    }
  };

  return (
    <Tree
      showIcon
      blockNode
      expandedKeys={expandedKeys}
      onExpand={handleExpand}
      selectedKeys={selectedModuleId ? [selectedModuleId] : []}
      treeData={treeData}
      onSelect={handleSelect}
      style={{ background: 'transparent' }}
    />
  );
};

// 主组件：按Project分组显示模块树
const ProjectGroupedTree: React.FC<ProjectGroupedTreeProps> = ({
  selectedProjectIds,
  projects,
  onSelectModule,
  selectedModuleId,
  onAddModule,
  onAddTask,
  onDeleteModule,
  onEditTask,
  onDeleteTask,
  onCheckTask,
  onRefactorTask,
}) => {
  // 从本地存储加载展开状态，如果没有则使用选中的项目ID
  const [expandedProjects, setExpandedProjects] = useState<string[]>(() => {
    const savedExpandedKeys = localStorageService.getExpandedKeys();
    // 过滤出仍然有效的项目ID
    const validProjectIds = selectedProjectIds.filter(id => 
      projects.some(p => p.id === id)
    );
    if (savedExpandedKeys.length > 0) {
      // 使用保存的展开状态，但只保留当前选中的项目
      return savedExpandedKeys.filter(id => validProjectIds.includes(id));
    }
    return validProjectIds;
  });

  // 当选中的项目变化时，更新展开状态
  useEffect(() => {
    const validProjectIds = selectedProjectIds.filter(id => 
      projects.some(p => p.id === id)
    );
    setExpandedProjects(prev => {
      // 合并之前保存的展开状态和新选中的项目
      const newExpanded = [...new Set([...prev, ...validProjectIds])];
      return newExpanded.filter(id => validProjectIds.includes(id));
    });
  }, [selectedProjectIds, projects]);

  // 保存展开状态到本地存储
  useEffect(() => {
    console.log('[ProjectGroupedTree] 保存展开状态:', expandedProjects);
    localStorageService.setExpandedKeys(expandedProjects);
  }, [expandedProjects]);

  const toggleProjectExpand = (projectId: string) => {
    setExpandedProjects(prev => 
      prev.includes(projectId)
        ? prev.filter(id => id !== projectId)
        : [...prev, projectId]
    );
  };

  // 过滤出选中的项目
  const selectedProjects = projects.filter(p => selectedProjectIds.includes(p.id));

  if (selectedProjects.length === 0) {
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="请选择至少一个项目"
        style={{ padding: '40px 0', color: '#808080' }}
      />
    );
  }

  return (
    <div className="project-grouped-tree">
      {selectedProjects.map((project) => {
        const isExpanded = expandedProjects.includes(project.id);
        
        return (
          <div key={project.id} className="project-tree-group">
            <div className="project-tree-header">
              <div
                className="project-tree-header-left"
                onClick={() => toggleProjectExpand(project.id)}
                style={{ cursor: 'pointer' }}
              >
                {isExpanded ? (
                  <DownOutlined style={{ fontSize: 12, color: '#ec4899' }} />
                ) : (
                  <RightOutlined style={{ fontSize: 12, color: '#ec4899' }} />
                )}
                <ProjectOutlined style={{ fontSize: 16, color: '#ec4899' }} />
                <span className="project-tree-name">{project.name}</span>
              </div>
              <Button
                type="text"
                size="small"
                icon={<PlusOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  onAddModule?.(project.id);
                }}
                style={{
                  color: '#ec4899',
                  opacity: 0.7,
                }}
                title="添加模块"
              />
            </div>
            {isExpanded && (
              <div className="project-tree-content">
                <ModuleTreeByProject
                  projectId={project.id}
                  onSelectModule={onSelectModule}
                  selectedModuleId={selectedModuleId}
                  onAddModule={onAddModule}
                  onAddTask={onAddTask}
                  onDeleteModule={onDeleteModule}
                  onEditTask={onEditTask}
                  onDeleteTask={onDeleteTask}
                  onCheckTask={onCheckTask}
                  onRefactorTask={onRefactorTask}
                />
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
};

export default ProjectGroupedTree;
