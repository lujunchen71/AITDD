# AITDD 开发文档

## 项目结构

```
AITDD/
├── backend/                    # Go 后端
│   ├── cmd/                    # CLI 命令
│   │   └── aitdd/             # 主入口
│   ├── internal/
│   │   ├── api/               # API 层
│   │   │   ├── handlers/      # 请求处理器
│   │   │   ├── middleware/    # 中间件
│   │   │   └── response.go    # 统一响应
│   │   ├── database/          # 数据库初始化
│   │   ├── models/            # 数据模型
│   │   ├── server/            # HTTP 服务器
│   │   ├── services/          # 业务逻辑
│   │   └── websocket/         # WebSocket
│   ├── migrations/            # 数据库迁移
│   └── go.mod                 # Go 依赖
│
├── frontend/                   # React 前端
│   ├── src/
│   │   ├── components/        # 通用组件
│   │   ├── features/          # 功能模块
│   │   ├── services/          # API 服务
│   │   ├── stores/            # 状态管理
│   │   ├── types/             # TypeScript 类型
│   │   └── hooks/             # 自定义 Hooks
│   └── package.json
│
└── specs/                      # 规格文档
```

## 技术栈

### 后端
- **Go 1.21+** - 编程语言
- **Gin** - HTTP 框架
- **GORM** - ORM
- **gorilla/websocket** - WebSocket
- **spf13/cobra** - CLI 框架
- **SQLite** - 本地数据库

### 前端
- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite 5** - 构建工具
- **Ant Design 5** - UI 组件库
- **React Flow 11** - 图形可视化
- **Zustand 4** - 状态管理
- **React Query 5** - 数据获取
- **TailwindCSS** - 样式

## 开发环境设置

### 后端

```bash
cd backend

# 安装依赖
go mod download

# 运行开发服务器
go run ./cmd/aitdd serve

# 运行测试
go test ./...

# 构建
go build -o aitdd ./cmd/aitdd
```

### 前端

```bash
cd frontend

# 安装依赖
npm install

# 运行开发服务器
npm run dev

# 构建
npm run build

# 预览生产版本
npm run preview
```

## API 开发

### 添加新的 API 端点

1. 在 `backend/internal/models/` 创建数据模型
2. 在 `backend/internal/api/handlers/` 创建处理器
3. 在 `backend/internal/api/routes.go` 注册路由

示例:

```go
// models/example.go
type Example struct {
    ID   string `gorm:"primaryKey"`
    Name string `json:"name"`
}

// handlers/example.go
func GetExamples(c *gin.Context) {
    var examples []models.Example
    database.DB.Find(&examples)
    api.Success(c, examples)
}

// routes.go
api.GET("/examples", handlers.GetExamples)
```

### 统一响应格式

```go
// 成功响应
api.Success(c, data)

// 错误响应
api.ValidationError(c, "参数错误", nil)
api.NotFound(c, "资源不存在")
api.InternalError(c, "服务器错误")
api.Locked(c, "资源被锁定")
```

## 前端开发

### 添加新页面

1. 在 `frontend/src/features/` 创建功能模块
2. 在 `App.tsx` 添加路由

示例:

```tsx
// features/example/index.tsx
const ExamplePage: React.FC = () => {
  return <div>Example Page</div>;
};

// App.tsx
<Route path="/example" element={<ExamplePage />} />
```

### 使用 API

```tsx
import { useQuery } from '@tanstack/react-query';
import apiClient from '../services/api';

const { data, isLoading } = useQuery({
  queryKey: ['examples'],
  queryFn: async () => {
    const response = await apiClient.get('/api/v1/examples');
    return response.data;
  },
});
```

### 状态管理

```tsx
import { useUIStore } from '../stores/useUIStore';

const { sidebarCollapsed, toggleSidebar } = useUIStore();
```

## 数据库

### 迁移

迁移文件位于 `backend/migrations/`

```sql
-- 001_init.sql
CREATE TABLE examples (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at INTEGER
);
```

### 模型定义

```go
type Example struct {
    ID        string `gorm:"primaryKey" json:"id"`
    Name      string `json:"name"`
    CreatedAt int64  `json:"createdAt"`
}
```

## WebSocket

### 后端

```go
// 创建 Hub
hub := websocket.NewHub()
go hub.Run()

// 广播消息
hub.BroadcastMessage("notification", payload)
```

### 前端

```tsx
const { connected, send } = useWebSocket({
  onMessage: (data) => {
    console.log('Received:', data);
  },
});
```

## 测试

### 后端测试

```go
func TestExample(t *testing.T) {
    // 设置测试数据库
    // 执行测试
    // 验证结果
}
```

### 前端测试

```tsx
import { render, screen } from '@testing-library/react';

test('renders example', () => {
  render(<ExampleComponent />);
  expect(screen.getByText('Example')).toBeInTheDocument();
});
```

## 构建和部署

### 构建

```bash
# 后端
cd backend && go build -o aitdd ./cmd/aitdd

# 前端
cd frontend && npm run build
```

### 部署

1. 将 `aitdd` 二进制文件复制到目标服务器
2. 将 `frontend/dist` 目录复制到服务器
3. 运行 `./aitdd serve`

## 代码规范

### Go
- 使用 gofmt 格式化代码
- 遵循 Go 命名约定
- 添加错误处理

### TypeScript
- 使用 ESLint 和 Prettier
- 使用函数组件和 Hooks
- 添加类型定义

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request
