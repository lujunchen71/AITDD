import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Button, Space, Spin, Alert, Empty, message, Tooltip } from 'antd';
import { CopyOutlined, ReloadOutlined, ZoomInOutlined, ZoomOutOutlined, ExpandOutlined } from '@ant-design/icons';
import mermaid from 'mermaid';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../../../services/api';
import { Task, TaskDependency, TaskStatus } from '../../../../types';
import { generateMermaidGraph, STATUS_LABELS } from '../../utils/mermaidGenerator';

interface TaskMermaidViewProps {
  moduleId?: string;
  projectIds?: string[];
  onTaskSelect?: (taskId: string) => void;
  refreshKey?: number;
}

/** 状态图例颜色 */
const LEGEND_COLORS: Record<TaskStatus, string> = {
  ready: '#1890ff',
  claimed: '#722ed1',
  in_progress: '#fa8c16',
  pending_review: '#13c2c2',
  completed: '#52c41a',
  failed: '#f5222d',
  blocked: '#8c8c8c',
};

let mermaidInitialized = false;

const TaskMermaidView: React.FC<TaskMermaidViewProps> = ({
  moduleId,
  projectIds,
  onTaskSelect,
  refreshKey,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewportRef = useRef<HTMLDivElement>(null);
  const [svgContent, setSvgContent] = useState<string>('');
  const [renderError, setRenderError] = useState<string>('');
  const [isRendering, setIsRendering] = useState(false);
  const renderCountRef = useRef(0);

  // 交互状态：缩放和平移（使用 ref 存储当前值，避免事件处理器闭包问题）
  const [scale, setScale] = useState(1);
  const [translateX, setTranslateX] = useState(0);
  const [translateY, setTranslateY] = useState(0);
  const scaleRef = useRef(1);
  const translateXRef = useRef(0);
  const translateYRef = useRef(0);
  const isPanningRef = useRef(false);
  const lastMousePosRef = useRef({ x: 0, y: 0 });
  const isSelectingRef = useRef(false);
  const selectionStartRef = useRef({ x: 0, y: 0 });
  const [selectionBox, setSelectionBox] = useState<{ x: number; y: number; w: number; h: number } | null>(null);

  // 保持 ref 与 state 同步
  useEffect(() => { scaleRef.current = scale; }, [scale]);
  useEffect(() => { translateXRef.current = translateX; }, [translateX]);
  useEffect(() => { translateYRef.current = translateY; }, [translateY]);

  // 构造查询键
  const queryKey = moduleId ? moduleId : (projectIds?.join(',') ?? 'all');

  // 获取任务列表
  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['tasks', queryKey, refreshKey],
    queryFn: async () => {
      const params: Record<string, any> = { pageSize: 500 };
      if (moduleId) {
        params.moduleId = moduleId;
      } else if (projectIds && projectIds.length > 0) {
        params.projectIds = projectIds.join(',');
      }
      const response = await apiClient.get('/tasks', { params });
      return response.data;
    },
  });

  // 获取依赖关系
  const { data: depsData, isLoading: depsLoading } = useQuery({
    queryKey: ['dependencies', queryKey, refreshKey],
    queryFn: async () => {
      const params: Record<string, any> = {};
      if (moduleId) {
        params.moduleId = moduleId;
      } else if (projectIds && projectIds.length > 0) {
        params.projectIds = projectIds.join(',');
      }
      const response = await apiClient.get('/dependencies', { params });
      return response.data;
    },
  });

  const tasks: Task[] = tasksData?.tasks || [];
  const dependencies: TaskDependency[] = depsData?.dependencies || [];

  // 初始化 mermaid
  useEffect(() => {
    if (!mermaidInitialized) {
      mermaid.initialize({
        startOnLoad: false,
        theme: 'dark',
        flowchart: {
          useMaxWidth: true,
          htmlLabels: true,
          curve: 'basis',
        },
        securityLevel: 'loose',
      });
      mermaidInitialized = true;
    }
  }, []);

  // 生成 mermaid 图定义
  const graphDefinition = generateMermaidGraph(tasks, dependencies);

  // 渲染 mermaid 图
  const renderMermaid = useCallback(async () => {
    if (tasksLoading || depsLoading) return;

    setIsRendering(true);
    setRenderError('');

    try {
      renderCountRef.current += 1;
      const uniqueId = `mermaid-graph-${Date.now()}-${renderCountRef.current}`;
      const { svg } = await mermaid.render(uniqueId, graphDefinition);
      setSvgContent(svg);
    } catch (err: any) {
      console.error('Mermaid render error:', err);
      setRenderError(err?.message || '图表渲染失败');
    } finally {
      setIsRendering(false);
    }
  }, [graphDefinition, tasksLoading, depsLoading]);

  useEffect(() => {
    renderMermaid();
  }, [renderMermaid]);

  // 注入 SVG 到容器
  useEffect(() => {
    if (containerRef.current && svgContent) {
      containerRef.current.innerHTML = svgContent;

      // 为节点添加点击事件
      if (onTaskSelect) {
        const svgEl = containerRef.current.querySelector('svg');
        if (svgEl) {
          // 为每个任务节点绑定点击
          tasks.forEach((task) => {
            const safeId = `task_${task.id.replace(/-/g, '_')}`;
            const nodeEl = svgEl.querySelector(`#${safeId}, [id="${safeId}"]`);
            if (nodeEl) {
              (nodeEl as HTMLElement).style.cursor = 'pointer';
              nodeEl.addEventListener('click', () => {
                onTaskSelect(task.id);
              });
            }
          });
        }
      }
    }
  }, [svgContent, tasks, onTaskSelect]);

  // 复制 Mermaid 代码
  const handleCopyCode = () => {
    navigator.clipboard
      .writeText(graphDefinition)
      .then(() => {
        message.success('Mermaid 代码已复制到剪贴板');
      })
      .catch(() => {
        message.error('复制失败，请手动复制');
      });
  };

  const isLoading = tasksLoading || depsLoading || isRendering;

  return (
    <div style={{ position: 'relative' }}>
      {/* 工具栏 */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
        }}
      >
        {/* 状态图例 */}
        <Space wrap size={8}>
          {(Object.entries(STATUS_LABELS) as [TaskStatus, string][]).map(([status, label]) => (
            <span
              key={status}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
                fontSize: 12,
              }}
            >
              <span
                style={{
                  display: 'inline-block',
                  width: 10,
                  height: 10,
                  borderRadius: 2,
                  backgroundColor: LEGEND_COLORS[status],
                }}
              />
              {label}
            </span>
          ))}
        </Space>

        {/* 操作按钮 */}
        <Space>
          <Tooltip title="重新渲染">
            <Button
              icon={<ReloadOutlined />}
              size="small"
              onClick={renderMermaid}
              loading={isRendering}
            >
              刷新
            </Button>
          </Tooltip>
          <Tooltip title="复制 Mermaid 代码">
            <Button
              icon={<CopyOutlined />}
              size="small"
              onClick={handleCopyCode}
            >
              复制代码
            </Button>
          </Tooltip>
        </Space>
      </div>

      {/* 图表区域 */}
      {isLoading && !svgContent ? (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: 400,
          }}
        >
          <Spin tip="正在渲染流程图..." size="large" />
        </div>
      ) : renderError ? (
        <Alert
          type="error"
          message="图表渲染失败"
          description={
            <div>
              <p>{renderError}</p>
              <details style={{ marginTop: 8 }}>
                <summary style={{ cursor: 'pointer', color: '#1890ff' }}>
                  查看 Mermaid 代码
                </summary>
                <pre
                  style={{
                    marginTop: 8,
                    padding: 8,
                    background: '#1a1a2e',
                    borderRadius: 4,
                    overflow: 'auto',
                    fontSize: 12,
                    color: '#e0e0e0',
                  }}
                >
                  {graphDefinition}
                </pre>
              </details>
            </div>
          }
          action={
            <Button size="small" onClick={renderMermaid}>
              重试
            </Button>
          }
        />
      ) : tasks.length === 0 ? (
        <Empty description="暂无任务数据，请先创建任务" style={{ padding: '60px 0' }} />
      ) : (
        <div
          ref={viewportRef}
          style={{
            position: 'relative',
            height: 'calc(100vh - 300px)',
            minHeight: 400,
            background: '#1a1a2e',
            borderRadius: 8,
            overflow: 'hidden',
            cursor: isPanningRef.current ? 'grabbing' : 'default',
            userSelect: 'none',
          }}
          onWheel={(e) => {
            e.preventDefault();
            const delta = e.deltaY > 0 ? 0.9 : 1.1;
            const oldScale = scaleRef.current;
            const newScale = Math.max(0.1, Math.min(10, oldScale * delta));
            if (!viewportRef.current) return;
            const rect = viewportRef.current.getBoundingClientRect();
            const mouseX = e.clientX - rect.left;
            const mouseY = e.clientY - rect.top;
            // 缩放中心为鼠标位置：保持鼠标下方的点不动
            const newTX = mouseX - (mouseX - translateXRef.current) * (newScale / oldScale);
            const newTY = mouseY - (mouseY - translateYRef.current) * (newScale / oldScale);
            setScale(newScale);
            setTranslateX(newTX);
            setTranslateY(newTY);
          }}
          onMouseDown={(e) => {
            if (e.button === 1) {
              // 中键：平移
              e.preventDefault();
              isPanningRef.current = true;
              lastMousePosRef.current = { x: e.clientX, y: e.clientY };
            } else if (e.button === 0) {
              // 左键：框选
              if (!viewportRef.current) return;
              const rect = viewportRef.current.getBoundingClientRect();
              const x = e.clientX - rect.left;
              const y = e.clientY - rect.top;
              isSelectingRef.current = true;
              selectionStartRef.current = { x, y };
              setSelectionBox({ x, y, w: 0, h: 0 });
            }
          }}
          onMouseMove={(e) => {
            if (isPanningRef.current) {
              const dx = e.clientX - lastMousePosRef.current.x;
              const dy = e.clientY - lastMousePosRef.current.y;
              lastMousePosRef.current = { x: e.clientX, y: e.clientY };
              setTranslateX(tx => tx + dx);
              setTranslateY(ty => ty + dy);
            } else if (isSelectingRef.current && viewportRef.current) {
              const rect = viewportRef.current.getBoundingClientRect();
              const cx = e.clientX - rect.left;
              const cy = e.clientY - rect.top;
              const sx = selectionStartRef.current.x;
              const sy = selectionStartRef.current.y;
              setSelectionBox({
                x: Math.min(sx, cx),
                y: Math.min(sy, cy),
                w: Math.abs(cx - sx),
                h: Math.abs(cy - sy),
              });
            }
          }}
          onMouseUp={(_e) => {
            isPanningRef.current = false;
            isSelectingRef.current = false;
            setSelectionBox(null);
          }}
          onMouseLeave={() => {
            isPanningRef.current = false;
            isSelectingRef.current = false;
            setSelectionBox(null);
          }}
          onContextMenu={(e) => e.preventDefault()}
        >
          {isRendering && (
            <div
              style={{
                position: 'absolute',
                top: '50%',
                left: '50%',
                transform: 'translate(-50%, -50%)',
                zIndex: 10,
              }}
            >
              <Spin />
            </div>
          )}
          {/* 缩放/平移控制按钮 */}
          <div style={{
            position: 'absolute',
            bottom: 16,
            right: 16,
            display: 'flex',
            gap: 4,
            zIndex: 20,
          }}>
            <Tooltip title="放大">
              <Button size="small" icon={<ZoomInOutlined />} onClick={() => setScale(s => Math.min(10, s * 1.2))} />
            </Tooltip>
            <Tooltip title="缩小">
              <Button size="small" icon={<ZoomOutOutlined />} onClick={() => setScale(s => Math.max(0.1, s / 1.2))} />
            </Tooltip>
            <Tooltip title="重置视图">
              <Button size="small" icon={<ExpandOutlined />} onClick={() => { setScale(1); setTranslateX(0); setTranslateY(0); }} />
            </Tooltip>
            <span style={{ color: '#666', fontSize: 12, lineHeight: '24px', padding: '0 4px' }}>
              {Math.round(scale * 100)}%
            </span>
          </div>
          {/* 框选遮罩 */}
          {selectionBox && selectionBox.w > 2 && selectionBox.h > 2 && (
            <div style={{
              position: 'absolute',
              left: selectionBox.x,
              top: selectionBox.y,
              width: selectionBox.w,
              height: selectionBox.h,
              border: '1px dashed #1890ff',
              background: 'rgba(24, 144, 255, 0.1)',
              pointerEvents: 'none',
              zIndex: 15,
            }} />
          )}
          {/* 图表内容：支持缩放和平移 */}
          <div
            ref={containerRef}
            style={{
              transformOrigin: '0 0',
              transform: `translate(${translateX}px, ${translateY}px) scale(${scale})`,
              opacity: isRendering ? 0.5 : 1,
              transition: isRendering ? 'opacity 0.2s' : 'none',
              padding: 16,
              display: 'inline-block',
              minWidth: '100%',
            }}
          />
        </div>
      )}
    </div>
  );
};

export default TaskMermaidView;
