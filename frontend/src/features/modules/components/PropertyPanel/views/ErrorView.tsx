import React, { useState } from 'react';
import { List, Tag, Typography, Empty, Collapse, Divider } from 'antd';
import { DownOutlined } from '@ant-design/icons';
import { Task, Module, Project } from '../../../../../types';
import { parseBugLog, parseBugLogSections } from '../../BugLogPanel';

interface ErrorViewProps {
  /** 任务模式：展示单个任务的错误 */
  taskId?: string | null;
  /** 模块模式：展示该模块下所有任务的错误聚合 */
  moduleId?: string | null;
  /** 传入所有任务（用于模块/项目模式聚合） */
  tasks: Task[];
  /** 传入所有模块（用于项目模式） */
  modules?: Module[];
  /** 项目信息 */
  project?: Project | null;
}

const levelColors: Record<string, string> = {
  error: 'error',
  warn: 'warning',
  info: 'processing',
  debug: 'default',
};

// ===== 任务错误统计 =====
interface TaskErrorStats {
  task: Task;
  errorCount: number;
  warnCount: number;
  hasIssueDetails: boolean;
  sections: ReturnType<typeof parseBugLogSections>;
  entries: ReturnType<typeof parseBugLog>;
}

function getTaskErrorStats(task: Task): TaskErrorStats {
  const entries = parseBugLog(task.bugLog as string);
  const sections = parseBugLogSections(task.bugLog as string);
  const errorCount = entries.filter(e => e.level === 'error').length
    + sections.reduce((sum, s) => sum + s.errors.length, 0);
  const warnCount = entries.filter(e => e.level === 'warn').length
    + sections.reduce((sum, s) => sum + s.warnings.length, 0);
  return {
    task,
    errorCount,
    warnCount,
    hasIssueDetails: !!task.issueDetails,
    sections,
    entries,
  };
}

// ===== 单任务详情（可展开） =====
const TaskErrorRow: React.FC<{ stats: TaskErrorStats }> = ({ stats }) => {
  const [expanded, setExpanded] = useState(false);
  const { task, errorCount, warnCount, hasIssueDetails, sections, entries } = stats;
  const hasAny = errorCount + warnCount > 0 || hasIssueDetails;
  const canExpand = hasAny;

  if (!hasAny) {
    return (
      <div style={{ padding: '6px 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #2a2a3c' }}>
        <Typography.Text style={{ color: '#666', fontSize: 12 }}>{task.name}</Typography.Text>
        <span style={{ color: '#444', fontSize: 11 }}>—</span>
      </div>
    );
  }

  return (
    <div style={{ borderBottom: '1px solid #2a2a3c' }}>
      <div
        style={{
          padding: '6px 0',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          cursor: canExpand ? 'pointer' : 'default',
        }}
        onClick={canExpand ? () => setExpanded(v => !v) : undefined}
      >
        <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, flex: 1, marginRight: 8 }} ellipsis={{ tooltip: task.name }}>
          {task.name}
        </Typography.Text>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          {errorCount > 0 && <Tag color="error" style={{ fontSize: 11, margin: 0 }}>{errorCount} 错误</Tag>}
          {warnCount > 0 && <Tag color="warning" style={{ fontSize: 11, margin: 0 }}>{warnCount} 警告</Tag>}
          {hasIssueDetails && <Tag color="orange" style={{ fontSize: 11, margin: 0 }}>!</Tag>}
          {canExpand && (
            <DownOutlined
              style={{
                fontSize: 10,
                color: '#888',
                transform: expanded ? 'rotate(180deg)' : 'rotate(0)',
                transition: 'transform 0.2s',
              }}
            />
          )}
        </div>
      </div>

      {expanded && (
        <div style={{ paddingLeft: 12, paddingBottom: 8 }}>
          {/* 新格式：按 section 展示 */}
          {sections.length > 0 ? (
            sections.map(sec => (
              <div key={sec.key} style={{ marginBottom: 6 }}>
                <div style={{ color: '#888', fontSize: 11, marginBottom: 3 }}>{sec.label}</div>
                {sec.errors.map((msg, i) => (
                  <div key={`e-${i}`} style={{ display: 'flex', gap: 6, marginBottom: 2 }}>
                    <span style={{ color: '#ff4d4f', fontSize: 11 }}>✕</span>
                    <Typography.Text style={{ color: '#ffb3b3', fontSize: 11, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{msg}</Typography.Text>
                  </div>
                ))}
                {sec.warnings.map((msg, i) => (
                  <div key={`w-${i}`} style={{ display: 'flex', gap: 6, marginBottom: 2 }}>
                    <span style={{ color: '#faad14', fontSize: 11 }}>⚠</span>
                    <Typography.Text style={{ color: '#ffe58f', fontSize: 11, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{msg}</Typography.Text>
                  </div>
                ))}
              </div>
            ))
          ) : (
            entries.length > 0 && (
              <div>
                {entries.map((entry, i) => (
                  <div key={i} style={{ display: 'flex', gap: 6, marginBottom: 2 }}>
                    <span style={{ color: entry.level === 'error' ? '#ff4d4f' : entry.level === 'warn' ? '#faad14' : '#1890ff', fontSize: 11 }}>
                      {entry.level === 'error' ? '✕' : entry.level === 'warn' ? '⚠' : 'ℹ'}
                    </span>
                    <Typography.Text style={{ color: '#c0c0c0', fontSize: 11, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{entry.message}</Typography.Text>
                  </div>
                ))}
              </div>
            )
          )}

          {hasIssueDetails && (
            <div style={{ marginTop: 4, padding: '4px 6px', background: 'rgba(255,122,0,0.08)', borderRadius: 4, borderLeft: '2px solid #ff7a00' }}>
              <div style={{ color: '#ff7a00', fontSize: 11, marginBottom: 2 }}>问题备注</div>
              <Typography.Text style={{ color: '#ffb347', fontSize: 11, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {task.issueDetails}
              </Typography.Text>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

// ===== 单任务错误视图（原有功能） =====
const TaskErrorView: React.FC<{ task: Task }> = ({ task }) => {
  const bugLogRaw = task.bugLog;
  const bugLogParsed = bugLogRaw ? parseBugLog(bugLogRaw as string) : [];
  const bugLog: any[] = Array.isArray(bugLogParsed) ? bugLogParsed as any[] : [];

  const formatTimestamp = (ts: string | number | undefined) => {
    if (!ts) return '';
    const d = new Date(typeof ts === 'number' ? ts * 1000 : ts);
    if (isNaN(d.getTime())) return String(ts);
    return d.toLocaleString('zh-CN');
  };

  return (
    <div>
      {/* 任务基本信息 */}
      <div style={{ marginBottom: 8, paddingBottom: 8, borderBottom: '1px solid #2a2a3c' }}>
        <Typography.Text style={{ color: '#888', fontSize: 11 }}>任务：</Typography.Text>
        <Typography.Text style={{ color: '#e0e0e0', fontSize: 12 }}>{task.name}</Typography.Text>
      </div>

      {bugLog.length === 0 && !task.issueDetails ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>无错误信息</span>}
        />
      ) : (
        <>
          {bugLog.length > 0 && (
            <List
              size="small"
              dataSource={bugLog}
              renderItem={(entry: any, index: number) => (
                <List.Item
                  key={index}
                  style={{ borderColor: '#3d3d5c', padding: '8px 0', flexDirection: 'column', alignItems: 'flex-start' }}
                >
                  <div
                    style={{
                      width: '100%',
                      borderLeft: `3px solid ${entry.level === 'error' ? '#ff4d4f' : entry.level === 'warn' ? '#faad14' : '#1890ff'}`,
                      paddingLeft: 8,
                      marginBottom: 4,
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                      <Tag color={levelColors[entry.level] || 'default'} style={{ fontSize: 11 }}>
                        {(entry.level || 'info').toUpperCase()}
                      </Tag>
                      {entry.timestamp && (
                        <Typography.Text style={{ color: '#555', fontSize: 11 }}>
                          {formatTimestamp(entry.timestamp)}
                        </Typography.Text>
                      )}
                    </div>
                    <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-word', display: 'block' }}>
                      {entry.message || String(entry)}
                    </Typography.Text>
                  </div>
                </List.Item>
              )}
            />
          )}

          {/* issueDetails */}
          {task.issueDetails && (
            <Collapse
              ghost
              style={{ marginTop: 8 }}
            >
              <Collapse.Panel
                header={<span style={{ color: '#a0a0a0', fontSize: 12 }}>AI 发现的问题详情</span>}
                key="issueDetails"
              >
                <pre style={{
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  color: '#c0c0c0',
                  fontSize: 12,
                  lineHeight: 1.6,
                  margin: 0,
                }}>
                  {task.issueDetails}
                </pre>
              </Collapse.Panel>
            </Collapse>
          )}
        </>
      )}
    </div>
  );
};

// ===== 模块错误聚合视图 =====
const ModuleErrorView: React.FC<{ module: Module; tasks: Task[] }> = ({ module, tasks }) => {
  const moduleTasks = tasks.filter(t => t.moduleId === module.id);
  const taskStats = moduleTasks.map(getTaskErrorStats);
  const totalErrors = taskStats.reduce((sum, s) => sum + s.errorCount, 0);
  const totalWarns = taskStats.reduce((sum, s) => sum + s.warnCount, 0);
  const tasksWithIssues = taskStats.filter(s => s.errorCount + s.warnCount > 0 || s.hasIssueDetails);

  return (
    <div>
      {/* 模块基本信息 */}
      <div style={{ marginBottom: 8, paddingBottom: 8, borderBottom: '1px solid #2a2a3c' }}>
        <Typography.Text style={{ color: '#888', fontSize: 11 }}>模块：</Typography.Text>
        <Typography.Text style={{ color: '#e0e0e0', fontSize: 12 }}>{module.name}</Typography.Text>
      </div>

      {/* 汇总统计 */}
      {(totalErrors + totalWarns) > 0 && (
        <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
          {totalErrors > 0 && <Tag color="error">{totalErrors} 个错误</Tag>}
          {totalWarns > 0 && <Tag color="warning">{totalWarns} 个警告</Tag>}
          <Tag color="default">共 {moduleTasks.length} 个任务</Tag>
        </div>
      )}

      {tasksWithIssues.length === 0 ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>该模块下无错误信息</span>}
        />
      ) : (
        <div>
          <Typography.Text style={{ color: '#888', fontSize: 11, display: 'block', marginBottom: 6 }}>
            有问题的任务（{tasksWithIssues.length}/{moduleTasks.length}）
          </Typography.Text>
          {taskStats.map(stats => (
            <TaskErrorRow key={stats.task.id} stats={stats} />
          ))}
        </div>
      )}
    </div>
  );
};

// ===== 项目错误全局视图 =====
const ProjectErrorView: React.FC<{ project: Project; modules: Module[]; tasks: Task[] }> = ({ project, modules, tasks }) => {
  const [expandedModules, setExpandedModules] = useState<Set<string>>(new Set());

  const toggleModule = (moduleId: string) => {
    setExpandedModules(prev => {
      const next = new Set(prev);
      if (next.has(moduleId)) next.delete(moduleId);
      else next.add(moduleId);
      return next;
    });
  };

  // 按模块统计
  const moduleStats = modules.map(mod => {
    const moduleTasks = tasks.filter(t => t.moduleId === mod.id);
    const taskStats = moduleTasks.map(getTaskErrorStats);
    const totalErrors = taskStats.reduce((sum, s) => sum + s.errorCount, 0);
    const totalWarns = taskStats.reduce((sum, s) => sum + s.warnCount, 0);
    const hasIssueDetails = taskStats.some(s => s.hasIssueDetails);
    return { module: mod, taskStats, totalErrors, totalWarns, hasIssueDetails };
  }).filter(ms => ms.totalErrors + ms.totalWarns > 0 || ms.hasIssueDetails);

  const grandTotalErrors = moduleStats.reduce((sum, ms) => sum + ms.totalErrors, 0);
  const grandTotalWarns = moduleStats.reduce((sum, ms) => sum + ms.totalWarns, 0);

  return (
    <div>
      {/* 项目基本信息 */}
      <div style={{ marginBottom: 8, paddingBottom: 8, borderBottom: '1px solid #2a2a3c' }}>
        <Typography.Text style={{ color: '#888', fontSize: 11 }}>项目：</Typography.Text>
        <Typography.Text style={{ color: '#e0e0e0', fontSize: 12 }}>{project.name}</Typography.Text>
      </div>

      {/* 汇总统计 */}
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 12 }}>
        {grandTotalErrors > 0 && <Tag color="error">{grandTotalErrors} 个错误</Tag>}
        {grandTotalWarns > 0 && <Tag color="warning">{grandTotalWarns} 个警告</Tag>}
        <Tag color="default">{modules.length} 个模块 / {tasks.length} 个任务</Tag>
      </div>

      {moduleStats.length === 0 ? (
        <Empty
          description={<span style={{ color: '#555', fontSize: 12 }}>项目中无错误信息</span>}
        />
      ) : (
        <div>
          <Typography.Text style={{ color: '#888', fontSize: 11, display: 'block', marginBottom: 6 }}>
            有问题的模块（{moduleStats.length}/{modules.length}）
          </Typography.Text>
          {moduleStats.map(ms => {
            const isExpanded = expandedModules.has(ms.module.id);
            return (
              <div key={ms.module.id} style={{ marginBottom: 6, border: '1px solid #2a2a3c', borderRadius: 6, overflow: 'hidden' }}>
                {/* 模块标题行 */}
                <div
                  style={{
                    padding: '8px 10px',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    cursor: 'pointer',
                    background: isExpanded ? '#1a1a32' : '#141428',
                  }}
                  onClick={() => toggleModule(ms.module.id)}
                >
                  <Typography.Text style={{ color: '#c0c0c0', fontSize: 12, flex: 1 }} ellipsis={{ tooltip: ms.module.name }}>
                    {ms.module.name}
                  </Typography.Text>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    {ms.totalErrors > 0 && <Tag color="error" style={{ fontSize: 11, margin: 0 }}>{ms.totalErrors}</Tag>}
                    {ms.totalWarns > 0 && <Tag color="warning" style={{ fontSize: 11, margin: 0 }}>{ms.totalWarns}</Tag>}
                    <DownOutlined
                      style={{
                        fontSize: 10,
                        color: '#888',
                        transform: isExpanded ? 'rotate(180deg)' : 'rotate(0)',
                        transition: 'transform 0.2s',
                      }}
                    />
                  </div>
                </div>

                {/* 展开后：显示任务列表 */}
                {isExpanded && (
                  <div style={{ padding: '6px 12px', background: '#0f0f23' }}>
                    {ms.taskStats.map(stats => (
                      <TaskErrorRow key={stats.task.id} stats={stats} />
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

// ===== 主组件 =====
const ErrorView: React.FC<ErrorViewProps> = ({ taskId, moduleId, tasks, modules = [], project }) => {
  // 优先级：taskId > moduleId > project
  if (taskId) {
    const task = tasks.find(t => t.id === taskId);
    if (!task) {
      return (
        <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
          未找到任务信息
        </div>
      );
    }
    return <TaskErrorView task={task} />;
  }

  if (moduleId) {
    const module = modules.find(m => m.id === moduleId);
    if (!module) {
      return (
        <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
          未找到模块信息
        </div>
      );
    }
    return <ModuleErrorView module={module} tasks={tasks} />;
  }

  if (project) {
    return <ProjectErrorView project={project} modules={modules} tasks={tasks} />;
  }

  return (
    <div style={{ color: '#666', textAlign: 'center', padding: 20 }}>
      请选择一个节点查看错误信息
    </div>
  );
};

export default ErrorView;
