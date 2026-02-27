/**
 * AITDD pathName 数据修复脚本 (直接操作数据库版本)
 * 修复现有模块和任务的 pathName，使其符合正确的格式：
 * - 模块: 项目pathName/模块名称
 * - 任务: 模块pathName/任务名称
 * 
 * 使用方法: node scripts/fix-pathname-direct.js
 */

const sqlite3 = require('sqlite3').verbose();
const path = require('path');
const os = require('os');

// 数据库路径：使用用户主目录下的 .aitdd/aitdd.db
const DB_PATH = path.join(os.homedir(), '.aitdd', 'aitdd.db');

console.log('📂 数据库路径:', DB_PATH);

const db = new sqlite3.Database(DB_PATH, (err) => {
    if (err) {
        console.error('❌ 无法连接数据库:', err.message);
        process.exit(1);
    }
    console.log('✅ 已连接数据库\n');
});

async function run(sql, params = []) {
    return new Promise((resolve, reject) => {
        db.run(sql, params, function(err) {
            if (err) reject(err);
            else resolve(this);
        });
    });
}

async function all(sql, params = []) {
    return new Promise((resolve, reject) => {
        db.all(sql, params, (err, rows) => {
            if (err) reject(err);
            else resolve(rows);
        });
    });
}

async function fixPathNames() {
    console.log('🔧 开始修复 pathName 数据...\n');

    try {
        // 1. 获取所有项目
        const projects = await all('SELECT id, name, path_name FROM projects');
        console.log(`📋 找到 ${projects.length} 个项目`);

        // 构建项目 pathName 映射
        const projectPathNameMap = new Map();
        for (const project of projects) {
            projectPathNameMap.set(project.id, project.path_name);
            console.log(`   - ${project.name}: ${project.path_name}`);
        }

        // 2. 获取所有模块
        const modules = await all('SELECT id, project_id, name, path_name FROM modules');
        console.log(`\n📋 找到 ${modules.length} 个模块`);

        // 3. 修复模块 pathName
        console.log('\n📝 修复模块 pathName...');
        let fixedModules = 0;
        const modulePathNameMap = new Map();

        for (const module of modules) {
            const projectPathName = projectPathNameMap.get(module.project_id);
            if (!projectPathName) {
                console.log(`   ⚠️ 模块 ${module.name} 的项目不存在，跳过`);
                continue;
            }

            const correctPathName = `${projectPathName}/${module.name}`;
            modulePathNameMap.set(module.id, correctPathName);

            if (module.path_name !== correctPathName) {
                console.log(`   🔧 修复模块: ${module.name}`);
                console.log(`      旧 pathName: ${module.path_name}`);
                console.log(`      新 pathName: ${correctPathName}`);

                await run('UPDATE modules SET path_name = ?, updated_at = ? WHERE id = ?', [
                    correctPathName,
                    Date.now(),
                    module.id
                ]);
                fixedModules++;
            }
        }
        console.log(`   ✅ 修复了 ${fixedModules} 个模块`);

        // 4. 获取所有任务
        const tasks = await all('SELECT id, module_id, name, path_name FROM tasks');
        console.log(`\n📋 找到 ${tasks.length} 个任务`);

        // 5. 修复任务 pathName
        console.log('\n📝 修复任务 pathName...');
        let fixedTasks = 0;

        for (const task of tasks) {
            let modulePathName = modulePathNameMap.get(task.module_id);
            if (!modulePathName) {
                // 重新查询模块的 pathName
                const mod = await all('SELECT path_name FROM modules WHERE id = ?', [task.module_id]);
                if (mod.length > 0) {
                    modulePathName = mod[0].path_name;
                    modulePathNameMap.set(task.module_id, modulePathName);
                }
            }

            if (!modulePathName) {
                console.log(`   ⚠️ 任务 ${task.name} 的模块不存在，跳过`);
                continue;
            }

            const correctPathName = `${modulePathName}/${task.name}`;

            if (task.path_name !== correctPathName) {
                console.log(`   🔧 修复任务: ${task.name}`);
                console.log(`      旧 pathName: ${task.path_name}`);
                console.log(`      新 pathName: ${correctPathName}`);

                await run('UPDATE tasks SET path_name = ?, updated_at = ? WHERE id = ?', [
                    correctPathName,
                    Date.now(),
                    task.id
                ]);
                fixedTasks++;
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
    } finally {
        db.close();
    }
}

// 执行修复
fixPathNames().catch(err => {
    console.error('❌ 未预期的错误:', err);
    process.exit(1);
});
