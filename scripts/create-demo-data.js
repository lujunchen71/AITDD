const axios = require('axios');

const API_BASE_URL = 'http://localhost:34567/api/v1';

// 创建 axios 实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 获取项目信息
async function getProject() {
  try {
    const response = await apiClient.get('/project');
    return response.data?.data?.project;
  } catch (error) {
    console.error('获取项目失败:', error.message);
    return null;
  }
}

// 创建项目（如果不存在）
async function createProject() {
  try {
    const response = await apiClient.put('/project/constitution', {
      constitution: '# AITDD 项目宪法\n\n这是自动创建的默认项目宪法。\n\n## 项目规则\n\n- 遵循 TDD 开发流程\n- 保持代码质量\n- 及时同步任务状态',
      version: 0,
    });
    return response.data?.data?.project;
  } catch (error) {
    console.error('创建项目失败:', error.message);
    return null;
  }
}

// 确保项目存在
async function ensureProject() {
  let project = await getProject();
  if (!project) {
    console.log('项目不存在，正在创建...');
    project = await createProject();
  }
  if (project) {
    const projectId = project.ID || project.id;
    console.log(`项目已就绪：${projectId}`);
    return projectId;
  }
  throw new Error('无法获取或创建项目');
}

// 创建模块
async function createModule(projectId, name, description, parentId = null) {
  try {
    const response = await apiClient.post('/modules', {
      projectId,
      parentId,
      name,
      description,
      prompt: `这是${name}模块的提示词`,
    });
    const module = response.data?.data?.module;
    console.log(`✓ 模块创建成功：${name} (${module?.id || 'unknown'})`);
    return module;
  } catch (error) {
    const errorMsg = error?.response?.data?.error?.message || error.message;
    console.error(`✗ 创建模块失败 [${name}]:`, errorMsg);
    return null;
  }
}

// 创建任务
async function createTask(moduleId, name, description, status = 'ready') {
  try {
    const response = await apiClient.post('/tasks', {
      moduleId,
      name,
      description,
      status,
      prompt: `这是${name}任务的提示词`,
    });
    const task = response.data?.data?.task;
    console.log(`  ✓ 任务创建成功：${name} (${task?.id || 'unknown'})`);
    return task;
  } catch (error) {
    const errorMsg = error?.response?.data?.error?.message || error.message;
    console.error(`  ✗ 创建任务失败 [${name}]:`, errorMsg);
    return null;
  }
}

// 主函数
async function main() {
  console.log('=== AITDD 演示数据创建脚本 ===\n');
  
  try {
    // 1. 确保项目存在
    const projectId = await ensureProject();
    
    // 2. 创建示例模块
    console.log('\n--- 创建模块 ---');
    
    // 根模块
    const frontendModule = await createModule(projectId, '前端模块', '负责用户界面和交互');
    const backendModule = await createModule(projectId, '后端模块', '负责业务逻辑和数据处理');
    const databaseModule = await createModule(projectId, '数据库模块', '负责数据存储和查询');
    
    if (!frontendModule || !backendModule || !databaseModule) {
      console.log('\n模块创建失败，退出');
      return;
    }
    
    // 子模块
    console.log('\n--- 创建子模块 ---');
    const componentsModule = await createModule(projectId, '组件库', '可复用的 UI 组件', frontendModule.id);
    const pagesModule = await createModule(projectId, '页面模块', '应用的主要页面', frontendModule.id);
    const apiModule = await createModule(projectId, 'API 模块', 'RESTful API 接口', backendModule.id);
    const authModule = await createModule(projectId, '认证模块', '用户认证和授权', backendModule.id);
    
    // 3. 为每个模块创建示例任务
    console.log('\n--- 创建任务 ---');
    
    // 前端模块任务
    console.log('\n[前端模块] 任务:');
    await createTask(frontendModule.id, '搭建项目框架', '配置 Vite、React、TypeScript 环境');
    await createTask(frontendModule.id, '实现路由系统', '使用 React Router 配置应用路由');
    
    // 组件库任务
    console.log('\n[组件库] 任务:');
    await createTask(componentsModule.id, '创建按钮组件', '实现基础按钮及其变体');
    await createTask(componentsModule.id, '创建表单组件', '实现输入框、下拉框等表单元素');
    
    // 页面模块任务
    console.log('\n[页面模块] 任务:');
    await createTask(pagesModule.id, '实现首页布局', '创建首页的整体布局结构');
    await createTask(pagesModule.id, '实现登录页面', '创建用户登录界面');
    
    // 后端模块任务
    console.log('\n[后端模块] 任务:');
    await createTask(backendModule.id, '设计数据模型', '定义核心业务数据模型');
    await createTask(backendModule.id, '实现中间件', '创建日志、错误处理等中间件');
    
    // API 模块任务
    console.log('\n[API 模块] 任务:');
    await createTask(apiModule.id, '实现用户 API', '创建用户相关的 CRUD 接口');
    await createTask(apiModule.id, '实现数据验证', '添加请求数据验证逻辑');
    
    // 认证模块任务
    console.log('\n[认证模块] 任务:');
    await createTask(authModule.id, '实现 JWT 认证', '集成 JWT 令牌认证机制');
    await createTask(authModule.id, '实现权限控制', '基于角色的访问控制');
    
    // 数据库模块任务
    console.log('\n[数据库模块] 任务:');
    await createTask(databaseModule.id, '设计数据库 schema', '创建数据库表结构');
    await createTask(databaseModule.id, '实现数据迁移', '配置数据库迁移脚本');
    
    console.log('\n=== 演示数据创建完成 ===');
    console.log('\n请访问 http://localhost:5173/modules 查看创建的模块和任务');
    
  } catch (error) {
    console.error('\n=== 脚本执行失败 ===');
    console.error('错误:', error.message);
  }
}

// 运行脚本
main();
