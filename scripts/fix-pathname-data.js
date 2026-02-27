/**
 * AITDD pathName 数据修复脚本
 * 修复现有模块和任务的 pathName，使其符合正确的格式：
 * - 模块: 项目pathName/模块名称
 * - 任务: 模块pathName/任务名称
 * 
 * 使用方法: node scripts/fix-pathname-data.js
 * 
 * 前提条件: 后端服务必须已经启动 (http://localhost:34567)
 */

const API_BASE = 'http://localhost:34567/api/v1';

async function fixPathNames() {
    console.log('🔧 开始修复 pathName 数据...\n');

    try {
        // 1. 获取所有项目
        const projectsResp = await fetch(`${API_BASE}/projects`);
        const projectsData = await projectsResp.json();
        const projects = projectsData?.data?.projects || projectsData?.projects || [];
        console.log(`📋 找到 ${projects.length} 个项目`);

        
        // 2. 获取所有模块
        const modulesResp = await fetch(`${API_BASE}/modules`);
        const modulesData = await modulesResp.json();
        const modules = modulesData?.data?.modules || modulesData?.modules || [];
        console.log(`📋 找到 ${modules.length} 个模块`);

        // 3. 获取所有任务
        const tasksResp = await fetch(`${API_BASE}/tasks`);
        const tasksData = await tasksResp.json();
        const tasks = tasksData?.data?.tasks || tasksData?.tasks || [];
        console.log(`📋 找到 ${tasks.length} 个任务`);

        // 4. 构建项目 pathName 映射
        const projectPathNameMap = new Map();
        for (const project of projects) {
            projectPathNameMap.set(project.id, project.pathName);
        }

        // 5. 构建模块 pathName 映射（新的正确格式）
        const modulePathNameMap = new Map();
        
        // 6. 修复模块 pathName
        console.log('\n📝 修复模块 pathName...');
        let fixedModules = 0;
        for (const module of modules) {
            const projectPathName = projectPathNameMap.get(module.projectId);
            if (!projectPathName) {
                console.log(`   ⚠️ 模块 ${module.name} 的项目不存在，跳过`);
                continue;
            }

            const correctPathName = `${projectPathName}/${module.name}`;
            
            // 检查是否需要修复
            if (module.pathName !== correctPathName) {
                console.log(`   🔧 修复模块: ${module.name}`);
                console.log(`      旧 pathName: ${module.pathName}`);
                console.log(`      新 pathName: ${correctPathName}`);
                
                // 更新模块
                const updateResp = await fetch(`${API_BASE}/modules/${module.id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: module.name,
                        description: module.description,
                        prompt: module.prompt,
                        status: module.status,
                        version: module.version
                    })
                });
                
                if (updateResp.ok) {
                    fixedModules++;
                    modulePathNameMap.set(module.id, correctPathName);
                } else {
                    const errorText = await updateResp.text();
                    console.log(`      ❌ 更新失败: ${errorText}`);
                }
            } else {
                // pathName 已经正确，记录下来用于后续任务修复
                modulePathNameMap.set(module.id, module.pathName);
            }
        }
        console.log(`   ✅ 修复了 ${fixedModules} 个模块`);

        
        // 7. 修复任务 pathName
        console.log('\n📝 修复任务 pathName...');
        let fixedTasks = 0;
        for (const task of tasks) {
                // 查找任务所属模块的正确 pathName
                let modulePathName = modulePathNameMap.get(task.moduleId);
                
                // 如果没有找到，可能是模块刚刚更新，需要重新获取
                if (!modulePathName) {
                    const modResp = await fetch(`${API_BASE}/modules/${task.moduleId}`);
                    const modData = await modResp.json();
                    const mod = modData?.data?.module || modData?.module;
                    if (mod) {
                        modulePathName = mod.pathName;
                        modulePathNameMap.set(task.moduleId, mod.pathName);
                    }
                }
                
                if (!modulePathName) {
                    console.log(`   ⚠️ 任务 ${task.name} 的模块不存在，跳过`);
                    continue;
                }

                const correctPathName = `${modulePathName}/${task.name}`;
                
                // 检查是否需要修复
                if (task.pathName !== correctPathName) {
                    console.log(`   🔧 修复任务: ${task.name}`);
                    console.log(`      旧 pathName: ${task.pathName}`);
                    console.log(`      新 pathName: ${correctPathName}`);
                    
                    // 更新任务
                    const updateResp = await fetch(`${API_BASE}/tasks/${task.id}`, {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            name: task.name,
                            description: task.description,
                            prompt: task.prompt,
                            status: task.status,
                            version: task.version,
                            upstreamContractDetail: task.upstreamContractDetail,
                            downstreamContractDetail: task.downstreamContractDetail
                        })
                    });
                    
                    if (updateResp.ok) {
                        fixedTasks++;
                    } else {
                        const errorText = await updateResp.text();
                        console.log(`      ❌ 更新失败: ${errorText}`);
                    }
                }
        }
        console.log(`   ✅ 修复了 ${fixedTasks} 个任务`);

        console.log('\n========================================');
        console.log('✅ pathName 数据修复完成!');
        console.log(`   - 修复模块: ${fixedModules}`);
        console.log(`   - 修复任务: ${fixedTasks}`);
        console.log('========================================\n');

    } catch (err) {
        console.error('❌ 修复失败:', err.message);
        process.exit(1);
    }
}

// 执行修复
fixPathNames().catch(err => {
    console.error('❌ 未预期的错误:', err);
    process.exit(1);
});
