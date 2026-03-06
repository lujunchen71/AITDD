import React, { useState, useMemo } from 'react';
import { Button, Typography, Tooltip } from 'antd';
import { CloseOutlined, BugOutlined, DownOutlined } from '@ant-design/icons';
import { Module, Task, Project } from '../../../../types';
import './index.css';

// ===== BugLog 解析工具 =====
export interface BugLogEntry {
  level: 'error' | 'warn' | 'info';
  message: string;
}

/** 新格式的 BugLog section */
export interface BugLogSection {
  label: string;   // "静态分析" / "动态运行" / "执行"
  key: 'static' | 'dynamic' | 'exe';
  errors: string[];
  warnings: string[];
}

/** 新格式接口 */
interface NewBugLogFormat {
  static?: { error?: string[]; warning?: string[] };
  dynamic?: { error?: string[]; warning?: string[] };
  exe?: { error?: string[]; warning?: string[] };
}

/**
 * 解析 bugLog 字段，支持新格式（对象）和旧格式（数组）
 * 新格式: { static: { error: [...], warning: [...] }, dynamic: {...}, exe: {...} }
 * 旧格式: [{ level: 'error'|'warn'|'info', message: string }]
 */
export function parseBugLog(bugLog: string | object | null | undefined): BugLogEntry[] {
  if (!bugLog) return [];
  try {
    const parsed = typeof bugLog === 'string' ? JSON.parse(bugLog) : bugLog;

    // 新格式：对象含 static/dynamic/exe 键
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      const newFmt = parsed as NewBugLogFormat;
      const entries: BugLogEntry[] = [];
      const sections: Array<{ key: keyof NewBugLogFormat }> = [
        { key: 'static' }, { key: 'dynamic' }, { key: 'exe' }
      ];
      for (const { key } of sections) {
        const section = newFmt[key];
        if (!section) continue;
        (section.error || []).forEach(msg => {
          if (msg && typeof msg === 'string') {
            entries.push({ level: 'error', message: msg });
          }
        });
        (section.warning || []).forEach(msg => {
          if (msg && typeof msg === 'string') {
            entries.push({ level: 'warn', message: msg });
          }
        });
      }
      return entries;
    }

    // 旧格式：数组
    if (Array.isArray(parsed)) {
      return parsed.filter(
        (item): item is BugLogEntry =>
          item && typeof item === 'object' &&
          ['error', 'warn', 'info'].includes(item.level) &&
          typeof item.message === 'string'
      );
    }

    // 纯字符串
    if (typeof parsed === 'string' && parsed.trim()) {
      return [{ level: 'info', message: parsed }];
    }
  } catch {
    if (typeof bugLog === 'string' && bugLog.trim()) {
      return [{ level: 'info', message: bugLog }];
    }
  }
  return [];
}

/**
 * 解析 bugLog 为分 section 的详细结构（仅适用于新格式）
 * 旧格式不分 section，返回空数组
 */
export function parseBugLogSections(bugLog: string | object | null | undefined): BugLogSection[] {
  if (!bugLog) return [];
  try {
    const parsed = typeof bugLog === 'string' ? JSON.parse(bugLog) : bugLog;
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return [];

    const newFmt = parsed as NewBugLogFormat;
    const sectionDefs: Array<{ key: keyof NewBugLogFormat; label: string }> = [
      { key: 'static', label: '静态分析' },
      { key: 'dynamic', label: '动态运行' },
      { key: 'exe', label: '执行结果' },
    ];
    const result: BugLogSection[] = [];
    for (const { key, label } of sectionDefs) {
      const section = newFmt[key];
      if (!section) continue;
      const errors = (section.error || []).filter((m): m is string => typeof m === 'string' && m.trim().length > 0);
      const warnings = (section.warning || []).filter((m): m is string => typeof m === 'string' && m.trim().length > 0);
      if (errors.length > 0 || warnings.length > 0) {
        result.push({ label, key, errors, warnings });
      }
    }
    return result;
  } catch {
    return [];
  }
}

// ===== 统计接口 =====
interface TaskStats {
  task: Task;
  entries: BugLogEntry[];
  sections: BugLogSection[];
  errorCount: number;
  warnCount: number;
  infoCount: number;
  hasIssueDetails: boolean;
  issueDetailsText: string;
}

interface ModuleStats {
  module: Module;
  tasks: TaskStats[];
  totalError: number;
  totalWarn: number;
  totalInfo: number;
  // 模块自身的 bugLog
  moduleBugLogEntries: BugLogEntry[];
  moduleBugLogSections: BugLogSection[];
  moduleOwnError: number;
  moduleOwnWarn: number;
}

// ===== Badge 组件 =====
const StatBadge: React.FC<{ count: number; type: 'error' | 'warn' | 'info' }> = ({ count, type }) => {
  if (count === 0) return null;
  return (
    <span className={`buglog-badge ${type}`}>{count}</span>
  );
};

// ===== 任务行组件（支持展开/折叠） =====
const TaskRow: React.FC<{ taskStats: TaskStats }> = ({ taskStats }) => {
  const [expanded, setExpanded] = useState(false);
  const { task, errorCount, warnCount, infoCount, hasIssueDetails, issueDetailsText, sections, entries } = taskStats;
  const hasAny = errorCount + warnCount + infoCount > 0 || hasIssueDetails;
  const canExpand = hasAny && (sections.length > 0 || entries.length > 0 || hasIssueDetails);

  if (!hasAny) {
    return (
      <div className="buglog-task-item">
        <span className="buglog-task-name" style={{ color: '#666' }}>{task.name}</span>
        <span style={{ color: '#444', fontSize: 11 }}>—</span>
      </div>
    );
  }

  return (
    <div className="buglog-task-item-wrapper">
      <div
        className={`buglog-task-item${canExpand ? ' clickable' : ''}`}
        onClick={canExpand ? () => setExpanded(v => !v) : undefined}
      >
        <span className="buglog-task-name">{task.name}</span>
        <div className="buglog-task-badges">
          <StatBadge count={errorCount} type="error" />
          <StatBadge count={warnCount} type="warn" />
          <StatBadge count={infoCount} type="info" />
          {hasIssueDetails && (
            <Tooltip title="有待处理的 issueDetails">
              <span className="buglog-issue-mark">!</span>
            </Tooltip>
          )}
          {canExpand && (
            <DownOutlined
              className={`buglog-expand-icon${expanded ? ' expanded' : ''}`}
              style={{ fontSize: 10, marginLeft: 2 }}
            />
          )}
        </div>
      </div>

      {/* 展开区域 */}
      {expanded && (
        <div className="buglog-task-detail">
          {/* 新格式：按 section 显示 */}
          {sections.length > 0 ? (
            sections.map(sec => (
              <div key={sec.key} className="buglog-section">
                <div className="buglog-section-title">{sec.label}</div>
                {sec.errors.map((msg, i) => (
                  <div key={`e-${i}`} className="buglog-message error">
                    <span className="buglog-msg-level">✕</span>
                    <span className="buglog-msg-text">{msg}</span>
                  </div>
                ))}
                {sec.warnings.map((msg, i) => (
                  <div key={`w-${i}`} className="buglog-message warn">
                    <span className="buglog-msg-level">⚠</span>
                    <span className="buglog-msg-text">{msg}</span>
                  </div>
                ))}
              </div>
            ))
          ) : (
            /* 旧格式：直接列出条目 */
            entries.length > 0 && (
              <div className="buglog-section">
                {entries.map((entry, i) => (
                  <div key={i} className={`buglog-message ${entry.level}`}>
                    <span className="buglog-msg-level">
                      {entry.level === 'error' ? '✕' : entry.level === 'warn' ? '⚠' : 'ℹ'}
                    </span>
                    <span className="buglog-msg-text">{entry.message}</span>
                  </div>
                ))}
              </div>
            )
          )}

          {/* issueDetails 显示 */}
          {hasIssueDetails && (
            <div className="buglog-issue-details">
              <div className="buglog-section-title" style={{ color: '#ff7a00' }}>问题备注</div>
              <div className="buglog-message" style={{ color: '#ffb347', background: 'rgba(255,122,0,0.08)' }}>
                <span style={{ marginRight: 4 }}>📝</span>
                <span className="buglog-msg-text">{issueDetailsText}</span>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

// ===== 模块行组件 =====
const ModuleRow: React.FC<{ moduleStats: ModuleStats }> = ({ moduleStats }) => {
  const [expanded, setExpanded] = useState(false);
  const { module, tasks, totalError, totalWarn, moduleBugLogEntries, moduleBugLogSections, moduleOwnError, moduleOwnWarn } = moduleStats;

  const hasAny = totalError + totalWarn > 0;

  return (
    <div className="buglog-module-item">
      <div
        className="buglog-module-header"
        onClick={() => setExpanded(v => !v)}
      >
        <span className="buglog-module-name">{module.name}</span>
        <div className="buglog-module-badges">
          {totalError > 0 && <span className="buglog-badge error">{totalError}</span>}
          {totalWarn > 0 && <span className="buglog-badge warn">{totalWarn}</span>}
          {!hasAny && <span style={{ color: '#444', fontSize: 11 }}>—</span>}
          <DownOutlined className={`buglog-expand-icon${expanded ? ' expanded' : ''}`} />
        </div>
      </div>
      {expanded && (
        <div className="buglog-task-list">
          {/* 模块自身的 bugLog */}
          {(moduleOwnError > 0 || moduleOwnWarn > 0) && (
            <div className="buglog-module-own-buglog">
              <div className="buglog-section-title" style={{ color: '#aaa', fontSize: 11, padding: '4px 8px' }}>
                📦 模块级错误
              </div>
              {moduleBugLogSections.length > 0 ? (
                moduleBugLogSections.map(sec => (
                  <div key={sec.key} className="buglog-section" style={{ marginLeft: 8 }}>
                    <div className="buglog-section-title">{sec.label}</div>
                    {sec.errors.map((msg, i) => (
                      <div key={`e-${i}`} className="buglog-message error">
                        <span className="buglog-msg-level">✕</span>
                        <span className="buglog-msg-text">{msg}</span>
                      </div>
                    ))}
                    {sec.warnings.map((msg, i) => (
                      <div key={`w-${i}`} className="buglog-message warn">
                        <span className="buglog-msg-level">⚠</span>
                        <span className="buglog-msg-text">{msg}</span>
                      </div>
                    ))}
                  </div>
                ))
              ) : (
                moduleBugLogEntries.length > 0 && (
                  <div className="buglog-section" style={{ marginLeft: 8 }}>
                    {moduleBugLogEntries.map((entry, i) => (
                      <div key={i} className={`buglog-message ${entry.level}`}>
                        <span className="buglog-msg-level">
                          {entry.level === 'error' ? '✕' : entry.level === 'warn' ? '⚠' : 'ℹ'}
                        </span>
                        <span className="buglog-msg-text">{entry.message}</span>
                      </div>
                    ))}
                  </div>
                )
              )}
            </div>
          )}
          {/* 任务列表 */}
          {tasks.length === 0 ? (
            <div className="buglog-task-item">
              <span style={{ color: '#555', fontSize: 12 }}>无任务</span>
            </div>
          ) : (
            tasks.map(ts => <TaskRow key={ts.task.id} taskStats={ts} />)
          )}
        </div>
      )}
    </div>
  );
};

// ===== 项目级 BugLog 展示组件 =====
const ProjectBugLogSection: React.FC<{ entries: BugLogEntry[]; sections: BugLogSection[] }> = ({ entries, sections }) => {
  const hasAny = entries.length > 0;
  return (
    <div className="buglog-project-own-buglog">
      <div className="buglog-section-title" style={{ color: '#ff9f3f', fontSize: 11, padding: '4px 0 2px 0', fontWeight: 600 }}>
        🏗️ 项目级错误
      </div>
      {!hasAny ? (
        <div style={{ color: '#555', fontSize: 11, padding: '2px 0 4px 0' }}>✅ 无错误 / 无警告</div>
      ) : sections.length > 0 ? (
        sections.map(sec => (
          <div key={sec.key} className="buglog-section">
            <div className="buglog-section-title">{sec.label}</div>
            {sec.errors.map((msg, i) => (
              <div key={`e-${i}`} className="buglog-message error">
                <span className="buglog-msg-level">✕</span>
                <span className="buglog-msg-text">{msg}</span>
              </div>
            ))}
            {sec.warnings.map((msg, i) => (
              <div key={`w-${i}`} className="buglog-message warn">
                <span className="buglog-msg-level">⚠</span>
                <span className="buglog-msg-text">{msg}</span>
              </div>
            ))}
          </div>
        ))
      ) : (
        entries.map((entry, i) => (
          <div key={i} className={`buglog-message ${entry.level}`}>
            <span className="buglog-msg-level">
              {entry.level === 'error' ? '✕' : entry.level === 'warn' ? '⚠' : 'ℹ'}
            </span>
            <span className="buglog-msg-text">{entry.message}</span>
          </div>
        ))
      )}
    </div>
  );
};

// ===== 主面板组件 =====
interface BugLogPanelProps {
  project?: Project;
  modules: Module[];
  tasks: Task[];
  anchorRight?: number;
  anchorTop?: number;
  onClose?: () => void;
}

const BugLogPanel: React.FC<BugLogPanelProps> = ({
  project,
  modules,
  tasks,
  anchorRight = 16,
  anchorTop = 160,
  onClose,
}) => {
  // 解析项目自身的 bugLog
  const projectBugLogEntries = useMemo(() => parseBugLog(project?.bugLog), [project?.bugLog]);
  const projectBugLogSections = useMemo(() => parseBugLogSections(project?.bugLog), [project?.bugLog]);
  const projectOwnError = useMemo(() => projectBugLogEntries.filter(e => e.level === 'error').length, [projectBugLogEntries]);
  const projectOwnWarn = useMemo(() => projectBugLogEntries.filter(e => e.level === 'warn').length, [projectBugLogEntries]);

  // 计算统计数据
  const stats = useMemo<ModuleStats[]>(() => {
    return modules.map(module => {
      const moduleTasks = tasks.filter(t => t.moduleId === module.id);
      const taskStats: TaskStats[] = moduleTasks.map(task => {
        const rawBugLog = task.bugLog;
        const entries = parseBugLog(rawBugLog);
        const sections = parseBugLogSections(rawBugLog);
        const issueDetailsText = task.issueDetails ? String(task.issueDetails).trim() : '';
        const hasIssueDetails = issueDetailsText.length > 0;

        // issueDetails 非空计为 1 个 error
        const errorCount = entries.filter(e => e.level === 'error').length + (hasIssueDetails ? 1 : 0);
        const warnCount = entries.filter(e => e.level === 'warn').length;
        const infoCount = entries.filter(e => e.level === 'info').length;
        return { task, entries, sections, errorCount, warnCount, infoCount, hasIssueDetails, issueDetailsText };
      });

      const totalError = taskStats.reduce((s, t) => s + t.errorCount, 0);
      const totalWarn = taskStats.reduce((s, t) => s + t.warnCount, 0);
      const totalInfo = taskStats.reduce((s, t) => s + t.infoCount, 0);

      // 模块自身的 bugLog
      const moduleBugLogEntries = parseBugLog(module.bugLog);
      const moduleBugLogSections = parseBugLogSections(module.bugLog);
      const moduleOwnError = moduleBugLogEntries.filter(e => e.level === 'error').length;
      const moduleOwnWarn = moduleBugLogEntries.filter(e => e.level === 'warn').length;

      return {
        module, tasks: taskStats,
        totalError: totalError + moduleOwnError,
        totalWarn: totalWarn + moduleOwnWarn,
        totalInfo,
        moduleBugLogEntries,
        moduleBugLogSections,
        moduleOwnError,
        moduleOwnWarn,
      };
    });
  }, [modules, tasks]);

  // 项目级统计（汇总：项目自身 + 所有模块）
  const projectError = useMemo(() => projectOwnError + stats.reduce((s, m) => s + m.totalError, 0), [projectOwnError, stats]);
  const projectWarn = useMemo(() => projectOwnWarn + stats.reduce((s, m) => s + m.totalWarn, 0), [projectOwnWarn, stats]);
  const projectInfo = useMemo(() => stats.reduce((s, m) => s + m.totalInfo, 0), [stats]);
  const modulesWithError = useMemo(() => stats.filter(m => m.totalError > 0).length, [stats]);
  const tasksWithError = useMemo(
    () => stats.reduce((s, m) => s + m.tasks.filter(t => t.errorCount > 0).length, 0),
    [stats]
  );

  return (
    <div
      className="buglog-panel"
      style={{
        right: anchorRight,
        top: anchorTop,
        bottom: 40,
      }}
    >
      {/* 面板标题栏 */}
      <div className="buglog-panel-header">
        <div className="buglog-panel-title">
          <BugOutlined style={{ color: '#ff4d4f' }} />
          <Typography.Text style={{ color: '#e0e0e0', fontSize: 13, fontWeight: 600 }}>
            BugLog 统计
          </Typography.Text>
        </div>
        <Button
          type="text"
          icon={<CloseOutlined />}
          size="small"
          onClick={onClose}
          style={{ color: '#888', padding: '0 4px' }}
        />
      </div>

      {/* 可滚动内容区 */}
      <div className="buglog-panel-content">
        {/* 项目级统计数字 */}
        <div className="buglog-project-stats">
          <div className="buglog-stat-item">
            <span className="buglog-stat-number error">{projectError}</span>
            <span className="buglog-stat-label">错误</span>
          </div>
          <div className="buglog-stat-item">
            <span className="buglog-stat-number warn">{projectWarn}</span>
            <span className="buglog-stat-label">警告</span>
          </div>
          <div className="buglog-stat-item">
            <span className="buglog-stat-number info">{projectInfo}</span>
            <span className="buglog-stat-label">信息</span>
          </div>
        </div>

        {/* 项目自身 bugLog 错误（在最上面，始终显示） */}
        <ProjectBugLogSection entries={projectBugLogEntries} sections={projectBugLogSections} />

        {/* 趋势 */}
        {(projectError > 0 || projectWarn > 0) && (
          <div className="buglog-trend">
            {projectError > 0
              ? `⚠️ ${modulesWithError} 个模块 / ${tasksWithError} 个任务有错误`
              : `📊 ${projectWarn} 个警告待处理`}
          </div>
        )}
        {projectError === 0 && projectWarn === 0 && projectInfo === 0 && (
          <div className="buglog-empty">
            <div className="buglog-empty-icon">✅</div>
            <div>所有任务无 BugLog</div>
          </div>
        )}

        {/* 模块级列表 */}
        {(projectError > 0 || projectWarn > 0 || projectInfo > 0) && (
          <div className="buglog-module-list">
            {stats.map(ms => (
              <ModuleRow key={ms.module.id} moduleStats={ms} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default BugLogPanel;
