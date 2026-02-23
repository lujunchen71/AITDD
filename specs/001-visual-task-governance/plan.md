# Implementation Plan: AITDD可视化任务治理系统

**Branch**: `001-visual-task-governance` | **Date**: 2026-02-23 | **Spec**: [spec.md](./spec.md)

## Summary

构建一个可视化任务治理系统，通过MCP接口供第三方AI编程软件调用，提供本地SQLite数据库存储和React前端可视化界面。

## Technical Context

**Language/Version**: Go 1.21+ (后端), Node.js 20+ / TypeScript 5.x (前端)
**Primary Dependencies**: 
- 后端: Gin (HTTP), GORM (ORM), gorilla/websocket
- 前端: React 18, Vite 5, Ant Design 5, React Flow 11, Zustand 4, React Query 5
**Storage**: SQLite (本地), PostgreSQL (可选远程同步)
**Testing**: Go testing, Jest/Vitest (前端)
**Target Platform**: Windows/Linux/macOS
**Project Type**: CLI工具 + Web服务 + 前端应用
**Performance Goals**: API P95 < 100ms, 前端加载 < 2s, 图渲染500+节点60fps
**Constraints**: 本地优先, 无外部依赖

## Constitution Check

✅ I. 代码质量 - Go/TypeScript代码规范, 静态分析
✅ II. 测试标准 - 单元测试覆盖率80%+
✅ III. 用户体验一致性 - Ant Design设计系统
✅ IV. 性能要求 - API P95 < 100ms

## Project Structure

### 后端 (Go)

```text
backend/
├── cmd/
│   └── aitdd/
│       └── main.go           # CLI入口
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP处理器
│   │   ├── middleware/       # 中间件
│   │   └── routes.go         # 路由定义
│   ├── models/               # 数据模型
│   ├── services/             # 业务逻辑
│   ├── database/             # 数据库操作
│   └── websocket/            # WebSocket处理
├── pkg/
│   └── utils/                # 工具函数
├── migrations/               # 数据库迁移
└── go.mod
```

### 前端 (React/TypeScript)

```text
frontend/
├── src/
│   ├── components/           # 通用组件
│   ├── features/             # 功能模块
│   │   ├── dashboard/
│   │   ├── modules/
│   │   ├── tasks/
│   │   └── notifications/
│   ├── services/             # API服务
│   ├── stores/               # Zustand状态
│   ├── hooks/                # 自定义Hooks
│   └── types/                # 类型定义
├── package.json
└── vite.config.ts
```

### CLI工具

```text
cmd/
├── init.go                   # aitdd init
├── serve.go                  # aitdd serve
└── root.go                   # 根命令
```

## Complexity Tracking

无违规需要记录。
