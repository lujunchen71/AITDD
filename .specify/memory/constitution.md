<!--
=============================================================================
SYNC IMPACT REPORT
=============================================================================
Version change: N/A → 1.0.0 (Initial creation)

Modified principles: None (initial version)

Added sections:
  - I. Code Quality (代码质量)
  - II. Testing Standards (测试标准) [NON-NEGOTIABLE]
  - III. User Experience Consistency (用户体验一致性)
  - IV. Performance Requirements (性能要求)
  - Development Workflow
  - Quality Gates
  - Governance

Removed sections: None (initial version)

Templates requiring updates:
  - .specify/templates/plan-template.md ✅ (no changes needed - generic)
  - .specify/templates/spec-template.md ✅ (no changes needed - generic)
  - .specify/templates/tasks-template.md ✅ (no changes needed - generic)

Follow-up TODOs: None
=============================================================================
-->

# AITDD Project Constitution

## Core Principles

### I. Code Quality (代码质量)

代码质量是项目成功的基石。所有代码必须遵循以下非协商标准：

**Mandatory Standards:**
- 所有代码 MUST 遵循项目定义的编码规范（linting 配置、格式化规则）
- 所有代码 MUST 通过静态分析检查，无严重或高危警告
- 所有公共 API MUST 包含文档注释（描述、参数、返回值、异常）
- 代码复杂度 MUST 保持在可维护范围内（单函数不超过 50 行，圈复杂度不超过 10）
- 所有依赖项 MUST 明确声明版本，禁止使用浮动版本号

**Rationale:**
高质量的代码减少技术债务，提高可维护性，并使团队协作更加高效。一致的代码风格降低认知负担，使开发者能够专注于业务逻辑而非代码格式。

### II. Testing Standards (测试标准) [NON-NEGOTIABLE]

测试是质量保证的核心，此原则不可协商。

**Mandatory Testing Practices:**
- TDD 强制执行：测试先行 → 用户确认 → 测试失败 → 然后实现
- Red-Green-Refactor 循环严格执行
- 单元测试覆盖率 MUST 达到 80% 以上（核心业务逻辑 90%+）
- 所有测试 MUST 独立运行，无相互依赖
- 测试 MUST 幂等：相同输入产生相同结果，可重复执行

**Required Test Types:**
- **Unit Tests**: 验证单个函数/方法的行为
- **Integration Tests**: 验证模块间交互（API 契约、数据库操作）
- **Contract Tests**: 验证服务间接口契约
- **End-to-End Tests**: 验证关键用户旅程

**Test Naming Convention:**
```
test_[unit]_[scenario]_[expected_result]
Example: test_user_login_invalid_password_returns_401
```

**Rationale:**
测试先行确保代码可测试性，减少缺陷流入生产环境。高覆盖率提供回归保护，使重构更加安全。

### III. User Experience Consistency (用户体验一致性)

用户体验一致性确保产品专业性和用户满意度。

**Mandatory UX Standards:**
- 所有用户界面 MUST 遵循设计系统规范（颜色、字体、间距、组件）
- 所有交互 MUST 提供即时反馈（加载状态、成功/失败提示）
- 响应式设计 MUST 支持目标设备范围（桌面、平板、移动端）
- 错误信息 MUST 清晰、可操作，禁止显示技术堆栈给终端用户

**Accessibility Requirements:**
- 所有界面 MUST 符合 WCAG 2.1 AA 级标准
- 键盘导航 MUST 完全可用
- 屏幕阅读器兼容性 MUST 验证

**Internationalization:**
- 所有用户可见文本 MUST 支持国际化（i18n）
- 日期、时间、数字格式 MUST 根据区域设置本地化
- RTL（从右到左）布局 MUST 支持（如适用）

**Rationale:**
一致的用户体验建立用户信任，减少学习成本。可访问性确保产品对所有用户可用，同时满足法律合规要求。

### IV. Performance Requirements (性能要求)

性能是功能需求的一部分，非事后考虑。

**Mandatory Performance Standards:**
- API 响应时间 MUST 满足以下目标：
  - P95 延迟 < 200ms（读操作）
  - P95 延迟 < 500ms（写操作）
  - P99 延迟 < 1000ms（所有操作）
- 前端首屏加载时间 MUST < 3 秒（3G 网络）
- 内存使用 MUST 不超过定义的限制（需在规格中明确）
- 数据库查询 MUST 使用索引，禁止全表扫描（大数据表）

**Performance Testing Requirements:**
- 负载测试 MUST 在发布前执行
- 性能回归测试 MUST 集成到 CI/CD 流水线
- 性能基准 MUST 记录并监控

**Optimization Principles:**
- 过早优化是万恶之源，但性能设计必须提前考虑
- 使用性能分析工具定位瓶颈，而非猜测
- 缓存策略 MUST 明确文档化（缓存键、失效策略、TTL）

**Rationale:**
性能直接影响用户体验和业务指标。性能问题在生产环境中修复成本高昂，必须在开发阶段纳入考量。

## Development Workflow (开发工作流)

**Code Review Requirements:**
- 所有代码变更 MUST 经过至少一名其他开发者审查
- 审查者 MUST 验证：
  - 代码符合本宪法原则
  - 测试充分且通过
  - 文档已更新
- 审查反馈 MUST 在 24 小时内响应

**Branch Strategy:**
- `main` 分支 MUST 始终处于可部署状态
- 功能开发 MUST 在特性分支进行
- 合并 MUST 通过 Pull Request/Merge Request
- 合并前 MUST 通过所有自动化检查

**Commit Standards:**
- 提交信息 MUST 遵循 Conventional Commits 规范
- 每个提交 MUST 是原子性的（单一逻辑变更）
- 提交 MUST 引用相关 Issue/Ticket

## Quality Gates (质量门禁)

**Pre-Commit:**
- 代码格式化检查
- Lint 检查
- 单元测试执行

**Pre-Merge:**
- 所有 Pre-Commit 检查
- 集成测试执行
- 代码审查批准
- 文档完整性检查

**Pre-Release:**
- 所有 Pre-Merge 检查
- 端到端测试
- 性能测试
- 安全扫描
- 部署到 Staging 环境验证

## Governance

**Constitution Authority:**
- 本宪法优先于所有其他开发实践文档
- 当其他文档与本宪法冲突时，以本宪法为准

**Amendment Procedure:**
1. 提出修订建议（Issue 或文档提案）
2. 团队讨论并达成共识
3. 更新宪法文档，增加版本号
4. 通知所有相关方
5. 更新受影响的模板和文档

**Versioning Policy:**
- **MAJOR**: 向后不兼容的原则移除或重新定义
- **MINOR**: 新增原则/章节或实质性扩展指导
- **PATCH**: 澄清、措辞、错别字修复、非语义性改进

**Compliance Review:**
- 每个 Sprint MUST 包含宪法合规性审查
- 违规 MUST 记录并跟踪至解决
- 重复违规 MUST 触发流程改进讨论

**Complexity Justification:**
- 任何违反简化原则的设计 MUST 在实现计划中明确记录理由
- 复杂性 MUST 有明确的业务或技术驱动因素
- 定期审查复杂设计，评估简化可能性

**Version**: 1.0.0 | **Ratified**: 2026-02-23 | **Last Amended**: 2026-02-23
