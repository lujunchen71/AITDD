/**
 * AITDD 测试数据导入脚本
 * 通过后端 API 导入数据，确保数据结构一致性
 * 
 * 使用方法: node scripts/import-aitdd-data.js [json文件路径]
 * 默认使用 debug/aitdd_example.json
 * 
 * 前提条件: 后端服务必须已经启动 (http://localhost:34567)
 */

const fs = require('fs');
const path = require('path');

const API_BASE = 'http://localhost:34567/api/v1';
const DEFAULT_JSON = path.join(__dirname, '..', 'debug', 'aitdd_example.json');

// 获取 JSON 文件路径
const jsonPath = process.argv[2] || DEFAULT_JSON;

console.log('📁 JSON文件:', jsonPath);
console.log('🌐 API地址:', API_BASE);

// 检查后端是否运行
async function checkBackend() {
  try {
    const response = await fetch(`${API_BASE.replace('/api/v1', '')}/health`);
    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }
    console.log('✅ 后端服务正常运行');
    return true;
  } catch (err) {
    console.error('❌ 后端服务未运行，请先启动后端: start-backend.bat');
    console.error('   错误:', err.message);
    return false;
  }
}

// 读取 JSON 文件
let jsonData;
try {
  const content = fs.readFileSync(jsonPath, 'utf-8');
  jsonData = JSON.parse(content);
  console.log('✅ JSON 文件读取成功');
} catch (err) {
  console.error('❌ 读取 JSON 文件失败:', err.message);
  process.exit(1);
}

// API 请求封装
async function apiRequest(method, endpoint, data = null) {
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
  };
  if (data) {
    options.body = JSON.stringify(data);
  }
  const response = await fetch(`${API_BASE}${endpoint}`, options);
  const responseText = await response.text();
  if (!response.ok) {
    throw new Error(`API Error ${response.status}: ${responseText}`);
  }
  return JSON.parse(responseText);
}

// 导入数据
async function importData() {
  // 检查后端
  if (!await checkBackend()) {
    process.exit(1);
  }

  console.log('\n🚀 开始导入数据...\n');

  try {
    // 1. 创建或获取项目（upsert 模式）
    console.log('📝 初始化项目...');
    let projectId;
    
    const projectName = jsonData.project?.name || 'New Project';
    
    // 先获取所有现有项目，检查是否存在同名项目
    let existingProject = null;
    try {
      const projectsResp = await apiRequest('GET', '/projects');
      const existingProjects = projectsResp?.data?.projects || projectsResp?.projects || [];
      existingProject = existingProjects.find(p => p.name === projectName);
      if (existingProject) {
        console.log(`   📋 发现已存在项目: ${projectName} (${existingProject.id})`);
      }
    } catch (err) {
      console.log(`   ⚠️  获取现有项目失败: ${err.message}`);
    }
    
    if (existingProject) {
      // 项目已存在，使用现有项目，并删除该项目的所有现有数据（实现完全覆盖）
      projectId = existingProject.id;
      console.log(`   🔄 使用现有项目: ${projectName} (${projectId})`);
      
      // 删除该项目的所有现有任务依赖
      console.log(`   🗑️  清理现有任务依赖...`);
      try {
        // 先获取所有任务
        const tasksResp = await apiRequest('GET', `/tasks?projectId=${projectId}`);
        const existingTasks = tasksResp?.data?.tasks || tasksResp?.tasks || [];
        
        // 删除每个任务的依赖
        for (const task of existingTasks) {
          try {
            await apiRequest('DELETE', `/tasks/${task.id}/dependencies`);
          } catch (err) {
            // 静默失败，依赖可能不存在
          }
        }
        console.log(`   ✅ 已清理任务依赖`);
      } catch (err) {
        console.log(`   ⚠️  清理任务依赖失败: ${err.message}`);
      }
      
      // 删除该项目的所有现有任务
      console.log(`   🗑️  清理现有任务...`);
      try {
        const tasksResp = await apiRequest('GET', `/tasks?projectId=${projectId}`);
        const existingTasks = tasksResp?.data?.tasks || tasksResp?.tasks || [];
        
        for (const task of existingTasks) {
          try {
            await apiRequest('DELETE', `/tasks/${task.id}`);
          } catch (err) {
            // 静默失败
          }
        }
        console.log(`   ✅ 已清理 ${existingTasks.length} 个任务`);
      } catch (err) {
        console.log(`   ⚠️  清理任务失败: ${err.message}`);
      }
      
      // 删除该项目的所有现有模块
      console.log(`   🗑️  清理现有模块...`);
      try {
        const modulesResp = await apiRequest('GET', `/modules?projectId=${projectId}`);
        const existingModules = modulesResp?.data?.modules || modulesResp?.modules || [];
        
        for (const mod of existingModules) {
          try {
            await apiRequest('DELETE', `/modules/${mod.id}`);
          } catch (err) {
            // 静默失败
          }
        }
        console.log(`   ✅ 已清理 ${existingModules.length} 个模块`);
      } catch (err) {
        console.log(`   ⚠️  清理模块失败: ${err.message}`);
      }
    } else {
      // 项目不存在，创建新项目
      const createResp = await apiRequest('POST', '/projects', {
        name: projectName,
        description: jsonData.project?.description || ''
      });
      const newProject = createResp?.data?.project || createResp?.project;
      if (newProject) {
        projectId = newProject.id;
        console.log(`   ✅ 创建新项目: ${projectName} (${projectId})`);
      } else {
        throw new Error('创建项目失败: 响应无效');
      }
    }

    // 2. 创建或更新模块（upsert 模式）
    console.log('\n📝 导入模块...');
    const moduleIdMap = {};
    
    // 先获取该项目下所有现有模块，用于判断是否需要更新
    const existingModulesMap = new Map(); // key: `${projectId}:${moduleName}`, value: module
    try {
      const modulesResp = await apiRequest('GET', `/modules?projectId=${projectId}`);
      const existingModules = modulesResp?.data?.modules || modulesResp?.modules || [];
      for (const m of existingModules) {
        const key = `${m.projectId}:${m.name}`;
        existingModulesMap.set(key, m);
      }
      console.log(`   📋 已有模块数量: ${existingModules.length}`);
    } catch (err) {
      console.log(`   ⚠️  获取现有模块失败: ${err.message}`);
    }
    
    for (const mod of jsonData.modules || []) {
      const moduleData = {
        projectId: projectId,  // 必填字段
        name: mod.name,
        description: mod.description || '',
        prompt: mod.prompt || '',
      };
      
      // 如果有父模块
      if (mod.parentId && moduleIdMap[mod.parentId]) {
        moduleData.parentId = mod.parentId;
      }
      
      // 检查是否已存在同名模块（同一项目下）
      const moduleKey = `${projectId}:${mod.name}`;
      const existingModule = existingModulesMap.get(moduleKey);
      
      try {
        let result;
        if (existingModule) {
          // 模块已存在，使用 PUT 更新（需要包含 version字段）
          moduleData.version = existingModule.version;
          result = await apiRequest('PUT', `/modules/${existingModule.id}`, moduleData);
          moduleIdMap[mod.id] = existingModule.id;
          console.log(`   🔄 模块更新: ${mod.name} (${existingModule.id})`);
        } else {
          // 模块不存在，使用 POST 创建
          result = await apiRequest('POST', '/modules', moduleData);
          const createdModule = result?.data?.module || result?.module;
          moduleIdMap[mod.id] = createdModule.id;
          console.log(`   ✅ 模块创建: ${mod.name} (${createdModule.id})`);
        }
      } catch (err) {
        if (err.message.includes('already exists') || err.message.includes('duplicate')) {
          console.log(`   ⏭️  模块已存在: ${mod.name}`);
          moduleIdMap[mod.id] = mod.id;
        } else {
          console.log(`   ❌ 模块操作失败: ${mod.name} - ${err.message}`);
        }
      }
    }

    // 3. 创建或更新任务（upsert 模式）
    console.log('\n📝 导入任务...');
    const taskIdMap = {};
    
    // 先获取该项目下所有现有任务，用于判断是否需要更新
    const existingTasksMap = new Map(); // key: `${moduleId}:${taskName}`, value: task
    try {
      const tasksResp = await apiRequest('GET', `/tasks?projectId=${projectId}`);
      const existingTasks = tasksResp?.data?.tasks || tasksResp?.tasks || [];
      for (const t of existingTasks) {
        const key = `${t.moduleId}:${t.name}`;
        existingTasksMap.set(key, t);
      }
      console.log(`   📋 已有任务数量: ${existingTasks.length}`);
    } catch (err) {
      console.log(`   ⚠️  获取现有任务失败: ${err.message}`);
    }
    
    for (const task of jsonData.tasks || []) {
      // 获取实际的模块ID
      const actualModuleId = moduleIdMap[task.moduleId];
      if (!actualModuleId) {
        console.log(`   ⚠️  跳过任务 ${task.name}: 模块 ${task.moduleId} 不存在`);
        console.log(`   🔍 moduleIdMap:`, JSON.stringify(moduleIdMap));
        continue;
      }
      
      // 只发送 API 支持的字段，将对象转换为 JSON 字符串
      const taskData = {
        moduleId: actualModuleId,
        name: task.name,
        description: task.description || '',
        prompt: task.prompt || '',
        // API 期望字符串，如果是对象则转换为 JSON 字符串
        upstreamContractDetail: typeof task.upstreamContractDetail === 'object'
          ? JSON.stringify(task.upstreamContractDetail)
          : (task.upstreamContractDetail || ''),
        downstreamContractDetail: typeof task.downstreamContractDetail === 'object'
          ? JSON.stringify(task.downstreamContractDetail)
          : (task.downstreamContractDetail || ''),
      };
      
      // 检查是否已存在同名任务（同一模块下）
      const taskKey = `${actualModuleId}:${task.name}`;
      const existingTask = existingTasksMap.get(taskKey);
      
      try {
        let result;
        if (existingTask) {
          // 任务已存在，使用 PUT 更新（需要包含 version字段）
          taskData.version = existingTask.version;
          result = await apiRequest('PUT', `/tasks/${existingTask.id}`, taskData);
          taskIdMap[task.id] = existingTask.id;
          console.log(`   🔄 任务更新: ${task.name} (${existingTask.id}) - ${task.status}`);
        } else {
          // 任务不存在，使用 POST 创建
          result = await apiRequest('POST', '/tasks', taskData);
          const createdTask = result?.data?.task || result?.task;
          taskIdMap[task.id] = createdTask.id;
          console.log(`   ✅ 任务创建: ${task.name} (${createdTask.id}) - ${task.status}`);
        }
      } catch (err) {
        console.log(`   ❌ 任务操作失败: ${task.name} - ${err.message}`);
      }
    }

    // 4. 创建任务依赖
    console.log('\n📝 导入任务依赖...');
    let depCount = 0;
    
    // 支持 taskDependencies 或 dependencies 两种字段名
    const dependencies = jsonData.taskDependencies || jsonData.dependencies || [];
    for (const dep of dependencies) {
      const upstreamId = taskIdMap[dep.upstreamTaskId];
      const downstreamId = taskIdMap[dep.downstreamTaskId];
      
      if (!upstreamId || !downstreamId) {
        console.log(`   ⚠️  跳过依赖: 任务ID不存在`);
        continue;
      }
      
      const depData = {
        upstreamTaskId: upstreamId,
        downstreamTaskId: downstreamId,
        interfaceContract: dep.interfaceContract || '',
      };
      
      try {
        await apiRequest('POST', '/dependencies', depData);
        console.log(`   ✅ 依赖: ${dep.upstreamTaskId} -> ${dep.downstreamTaskId}`);
        depCount++;
      } catch (err) {
        if (err.message.includes('already exists') || err.message.includes('duplicate')) {
          console.log(`   ⏭️  依赖已存在: ${dep.upstreamTaskId} -> ${dep.downstreamTaskId}`);
        } else {
          console.log(`   ❌ 依赖创建失败: ${err.message}`);
        }
      }
    }

    // 5. 创建模块依赖
    console.log('\n📝 计算模块依赖...');
    let modDepCount = 0;
    
    const moduleDeps = new Set();
    for (const dep of dependencies) {
      const task = jsonData.tasks.find(t => t.id === dep.upstreamTaskId);
      const downstreamTask = jsonData.tasks.find(t => t.id === dep.downstreamTaskId);
      
      if (task && downstreamTask && task.moduleId !== downstreamTask.moduleId) {
        const key = `${task.moduleId}->${downstreamTask.moduleId}`;
        if (!moduleDeps.has(key)) {
          moduleDeps.add(key);
          const upstreamModId = moduleIdMap[task.moduleId];
          const downstreamModId = moduleIdMap[downstreamTask.moduleId];
          
          if (upstreamModId && downstreamModId) {
            try {
              await apiRequest('POST', `/modules/${upstreamModId}/dependencies`, {
                upstreamModuleId: upstreamModId,
                downstreamModuleId: downstreamModId,
                contractSummary: dep.interfaceContract || '',
              });
              console.log(`   ✅ 模块依赖: ${task.moduleId} -> ${downstreamTask.moduleId}`);
              modDepCount++;
            } catch (err) {
              console.log(`   ⏭️  模块依赖已存在或失败`);
            }
          }
        }
      }
    }

    console.log('\n========================================');
    console.log('✅ 数据导入完成!');
    console.log(`   - 模块: ${Object.keys(moduleIdMap).length}`);
    console.log(`   - 任务: ${Object.keys(taskIdMap).length}`);
    console.log(`   - 任务依赖: ${depCount}`);
    console.log(`   - 模块依赖: ${modDepCount}`);
    console.log('========================================\n');

  } catch (err) {
    console.error('\n❌ 导入失败:', err.message);
    process.exit(1);
  }
}

// 执行导入
importData().catch(err => {
  console.error('❌ 未预期的错误:', err);
  process.exit(1);
});
