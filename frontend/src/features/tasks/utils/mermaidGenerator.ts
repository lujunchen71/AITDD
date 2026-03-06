import { Task, TaskDependency, TaskStatus } from '../../../types';

/** 状态标签映射 */
const STATUS_LABELS: Record<TaskStatus, string> = {
  ready: '待开始',
  claimed: '已认领',
  in_progress: '进行中',
  pending_review: '待审核',
  completed: '已完成',
  failed: '失败',
  blocked: '被阻塞',
};

/** 状态 classDef 颜色配置 */
const STATUS_CLASS_DEFS: Record<TaskStatus, string> = {
  ready: 'fill:#1890ff,stroke:#096dd9,color:#fff',
  claimed: 'fill:#722ed1,stroke:#531dab,color:#fff',
  in_progress: 'fill:#fa8c16,stroke:#d46b08,color:#fff',
  pending_review: 'fill:#13c2c2,stroke:#08979c,color:#fff',
  completed: 'fill:#52c41a,stroke:#389e0d,color:#fff',
  failed: 'fill:#f5222d,stroke:#cf1322,color:#fff',
  blocked: 'fill:#8c8c8c,stroke:#595959,color:#fff',
};

/**
 * 将 UUID 转换为 Mermaid 安全的节点 ID（替换连字符为下划线）
 */
function toNodeId(taskId: string): string {
  return `task_${taskId.replace(/-/g, '_')}`;
}

/**
 * 截断 contractSummary 标签（超过 20 字符加省略号）
 */
function truncateLabel(label: string, maxLen = 20): string {
  if (label.length <= maxLen) return label;
  return label.slice(0, maxLen) + '...';
}

/**
 * 将 tasks 和 dependencies 转换为 Mermaid flowchart TD 字符串
 */
export function generateMermaidGraph(
  tasks: Task[],
  dependencies: TaskDependency[]
): string {
  if (tasks.length === 0) {
    return 'flowchart TD\n  empty["暂无任务数据"]';
  }

  const lines: string[] = ['flowchart TD'];

  // 收集用到的状态类型（用于生成 classDef）
  const usedStatuses = new Set<TaskStatus>();

  // 节点定义
  for (const task of tasks) {
    const nodeId = toNodeId(task.id);
    const statusLabel = STATUS_LABELS[task.status] || task.status;
    // 使用双引号包裹标签（支持中文），名称和状态换行显示
    const label = `${task.name}\\n[${statusLabel}]`;
    lines.push(`  ${nodeId}["${label}"]:::${task.status}`);
    usedStatuses.add(task.status);
  }

  lines.push('');

  // 边定义
  for (const dep of dependencies) {
    const fromId = toNodeId(dep.upstreamTaskId);
    const toId = toNodeId(dep.downstreamTaskId);

    if (dep.contractSummary && dep.contractSummary.trim()) {
      const edgeLabel = truncateLabel(dep.contractSummary.trim());
      lines.push(`  ${fromId} -->|"${edgeLabel}"| ${toId}`);
    } else {
      lines.push(`  ${fromId} --> ${toId}`);
    }
  }

  lines.push('');

  // classDef 定义（只包含用到的状态）
  for (const status of usedStatuses) {
    const colorDef = STATUS_CLASS_DEFS[status];
    lines.push(`  classDef ${status} ${colorDef}`);
  }

  return lines.join('\n');
}

export { STATUS_LABELS, STATUS_CLASS_DEFS };
