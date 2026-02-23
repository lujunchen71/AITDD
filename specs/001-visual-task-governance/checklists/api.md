# API契约质量检查清单: AITDD可视化任务治理系统

**Purpose**: 验证API需求定义的完整性、清晰度和一致性，确保API契约可供实现
**Created**: 2026-02-23
**Feature**: [spec.md](../spec.md) | [api-contracts.md](../contracts/api-contracts.md)
**Depth**: 标准PR评审
**Focus**: API契约完整性、错误处理、边界条件

**Note**: 本检查清单由 `/speckit.checklist` 命令生成，用于验证需求质量而非实现正确性。

---

## Requirement Completeness (需求完整性)

- [ ] CHK001 所有14个MCP端点（MCP-01至MCP-14）是否都有对应的详细API定义？ [Completeness, Spec §功能需求：给AI的MCP工具]
- [ ] CHK002 是否每个API端点都明确定义了请求体结构（字段、类型、必填性）？ [Completeness, Gap]
- [ ] CHK003 是否每个API端点都明确定义了成功响应的完整数据结构？ [Completeness, api-contracts.md]
- [ ] CHK004 是否所有查询参数都定义了默认值和行为？ [Completeness, Gap]
- [ ] CHK005 是否定义了API版本控制策略？ [Gap]
- [ ] CHK006 是否所有POST/PUT操作都定义了版本号（乐观锁）校验行为？ [Completeness, Spec §US3]

## Requirement Clarity (需求清晰度)

- [ ] CHK007 "默认端口34567"是否可配置？配置方式是否明确？ [Clarity, api-contracts.md §基础信息]
- [ ] CHK008 "可配置简单令牌"的认证机制是否详细说明？ [Clarity, Gap]
- [ ] CHK009 `fields` 查询参数的允许值是否明确列出？ [Clarity, api-contracts.md §2.6]
- [ ] CHK010 `flat` 参数为true时的扁平列表结构是否定义？ [Clarity, api-contracts.md §2.1]
- [ ] CHK011 分页参数（page, pageSize）的默认值和最大值是否定义？ [Clarity, api-contracts.md §3.1]
- [ ] CHK012 "force"删除参数的具体行为（级联删除规则）是否明确？ [Clarity, api-contracts.md §2.5]

## Error Handling Coverage (错误处理覆盖)

- [ ] CHK013 是否每个API端点都定义了所有可能的错误码？ [Coverage, Gap]
- [ ] CHK014 VERSION_CONFLICT错误的详细信息结构是否定义？ [Completeness, api-contracts.md §错误码]
- [ ] CHK015 LOCKED错误返回时是否包含锁定者信息和过期时间？ [Coverage, Gap]
- [ ] CHK016 VALIDATION_ERROR的具体字段错误格式是否定义？ [Clarity, Gap]
- [ ] CHK017 循环依赖检测失败时的错误详情是否定义？ [Completeness, Spec §US4]
- [ ] CHK018 资源不存在（NOT_FOUND）的错误响应是否包含资源类型信息？ [Clarity, Gap]

## Scenario Coverage (场景覆盖)

- [ ] CHK019 模块嵌套层级超过10层时的API行为是否定义？ [Edge Case, Spec §Edge Cases]
- [ ] CHK020 任务依赖链超过50个节点时的分页/过滤机制是否定义？ [Edge Case, Spec §Edge Cases]
- [ ] CHK021 并发锁定同一任务时的竞争处理规则是否明确？ [Exception Flow, Spec §Edge Cases]
- [ ] CHK022 空结果（无模块/无任务）的响应格式是否定义？ [Coverage, Gap]
- [ ] CHK023 批量操作（如批量创建任务）是否在需求中考虑？ [Gap]
- [ ] CHK024 部分成功场景（如批量操作中部分失败）的处理是否定义？ [Exception Flow, Gap]

## Consistency (一致性)

- [ ] CHK025 所有日期时间字段是否统一使用毫秒时间戳格式？ [Consistency, api-contracts.md]
- [ ] CHK026 所有ID字段是否统一使用UUID字符串格式？ [Consistency, data-model.md]
- [ ] CHK027 成功响应的`success: true`字段是否在所有端点保持一致？ [Consistency, api-contracts.md §通用响应格式]
- [ ] CHK028 错误响应结构是否在所有错误场景保持一致？ [Consistency, api-contracts.md §错误响应]
- [ ] CHK029 版本号字段命名（version vs Version）是否全局一致？ [Consistency, Gap]

## Measurability (可测量性)

- [ ] CHK030 "API响应时间P95 < 100ms"的测量方法是否定义？ [Measurability, Spec §SC-001]
- [ ] CHK031 "支持10个并发AI连接无阻塞"的阻塞定义是否明确？ [Measurability, Spec §SC-006]
- [ ] CHK032 "AI通过MCP完成一次CRUD操作 < 200ms"的起止点是否定义？ [Measurability, Spec §SC-005]

## Dependencies & Assumptions (依赖与假设)

- [ ] CHK033 "仅监听localhost"假设被违反时的行为是否定义？ [Assumption, Spec §Assumptions]
- [ ] CHK034 SQLite数据库文件损坏时的API错误响应是否定义？ [Exception Flow, Spec §Edge Cases]
- [ ] CHK035 远程同步中断时的API可用性保证是否说明？ [Dependency, Spec §Edge Cases]

## WebSocket & Real-time (实时通信)

- [ ] CHK036 WebSocket连接URL和协议是否定义？ [Gap]
- [ ] CHK037 WebSocket消息格式是否与REST API响应格式一致？ [Consistency, Gap]
- [ ] CHK038 WebSocket断线重连机制是否在需求中说明？ [Gap]
- [ ] CHK039 通知推送的实时性要求是否量化？ [Measurability, Gap]

## Traceability (可追溯性)

- [ ] CHK040 是否所有API端点都可追溯到用户故事？ [Traceability, Spec §User Scenarios]
- [ ] CHK041 是否所有API端点都有唯一标识符（如API-001）？ [Traceability, Gap]
- [ ] CHK042 API契约与数据模型字段的对应关系是否文档化？ [Traceability, data-model.md]

---

## Summary

| 类别 | 检查项数量 | 关键发现 |
|------|-----------|----------|
| 需求完整性 | 6 | 部分端点缺少请求/响应结构定义 |
| 需求清晰度 | 6 | 认证机制、分页参数需明确 |
| 错误处理覆盖 | 6 | 错误详情结构需补充 |
| 场景覆盖 | 6 | 边界条件和空结果处理需定义 |
| 一致性 | 5 | 字段命名和格式基本一致 |
| 可测量性 | 3 | 性能指标测量方法需明确 |
| 依赖与假设 | 3 | 异常场景行为需定义 |
| 实时通信 | 4 | WebSocket协议需详细定义 |
| 可追溯性 | 3 | API ID体系需建立 |

**总计**: 42 项检查

---

## Notes

- 本检查清单聚焦于API契约质量，不验证API实现正确性
- 检查项标记 `[Gap]` 表示需求文档中缺失的定义
- 检查项标记 `[Spec §X]` 表示可追溯到规格说明章节
- 完成检查后，将 `[ ]` 更改为 `[x]` 并添加备注
