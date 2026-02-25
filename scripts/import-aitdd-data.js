/**
 * AITDD 测试数据导入脚本
 * 将 aitdd_example.json 中的数据导入到 SQLite 数据库
 * 
 * 使用方法: node scripts/import-aitdd-data.js [json文件路径]
 * 默认使用 ../aitdd_example.json
 */

const fs = require('fs');
const path = require('path');
const os = require('os');
const Database = require('better-sqlite3');

// 数据库路径 - 使用与后端相同的路径 (~/.aitdd/aitdd.db)
const DB_PATH = path.join(os.homedir(), '.aitdd', 'aitdd.db');
const DEFAULT_JSON = path.join(__dirname, '..', 'aitdd_example.json');

// 获取 JSON 文件路径
const jsonPath = process.argv[2] || DEFAULT_JSON;

console.log('📁 JSON文件:', jsonPath);
console.log('🗄️  数据库路径:', DB_PATH);

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

// 打开数据库
const db = new Database(DB_PATH);
db.pragma('foreign_keys = OFF'); // 暂时禁用外键约束

// 获取或创建项目ID
let projectId;
const existingProject = db.prepare('SELECT id FROM projects LIMIT 1').get();
if (existingProject) {
  projectId = existingProject.id;
  console.log(`📌 使用现有项目 ID: ${projectId}`);
} else {
  projectId = jsonData.project?.id || 'proj-001';
  console.log(`📌 创建新项目 ID: ${projectId}`);
}

// 初始化表结构
console.log('\n🔧 初始化表结构...');
db.exec(`
  CREATE TABLE IF NOT EXISTS projects (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      constitution TEXT,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      version INTEGER NOT NULL DEFAULT 1,
      sync_status TEXT NOT NULL DEFAULT 'SYNCED'
  );

  CREATE TABLE IF NOT EXISTS modules (
      id TEXT PRIMARY KEY,
      parent_id TEXT,
      project_id TEXT NOT NULL,
      name TEXT NOT NULL,
      description TEXT,
      prompt TEXT,
      status TEXT NOT NULL DEFAULT 'designing',
      test_coverage REAL DEFAULT 0,
      upstream_contract_summary TEXT,
      downstream_contract_summary TEXT,
      file_path TEXT,
      locked INTEGER NOT NULL DEFAULT 0,
      locked_by TEXT,
      locked_at INTEGER,
      lock_expires_at INTEGER,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      version INTEGER NOT NULL DEFAULT 1,
      sync_status TEXT NOT NULL DEFAULT 'SYNCED'
  );

  CREATE TABLE IF NOT EXISTS tasks (
      id TEXT PRIMARY KEY,
      module_id TEXT NOT NULL,
      name TEXT NOT NULL,
      description TEXT,
      status TEXT NOT NULL DEFAULT 'ready',
      assignee TEXT,
      upstream_contract_detail TEXT,
      downstream_contract_detail TEXT,
      prompt TEXT,
      tests TEXT,
      logs TEXT,
      code_paths TEXT,
      human_assistance TEXT,
      locked INTEGER NOT NULL DEFAULT 0,
      locked_by TEXT,
      locked_at INTEGER,
      lock_expires_at INTEGER,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL,
      version INTEGER NOT NULL DEFAULT 1,
      sync_status TEXT NOT NULL DEFAULT 'SYNCED'
  );

  CREATE TABLE IF NOT EXISTS dependencies (
      id TEXT PRIMARY KEY,
      upstream_task_id TEXT NOT NULL,
      downstream_task_id TEXT NOT NULL,
      interface_contract TEXT,
      created_at INTEGER NOT NULL,
      version INTEGER NOT NULL DEFAULT 1,
      sync_status TEXT NOT NULL DEFAULT 'SYNCED'
  );

  CREATE TABLE IF NOT EXISTS module_dependencies (
      id TEXT PRIMARY KEY,
      upstream_module_id TEXT NOT NULL,
      downstream_module_id TEXT NOT NULL,
      contract_summary TEXT,
      created_at INTEGER NOT NULL,
      version INTEGER NOT NULL DEFAULT 1,
      sync_status TEXT NOT NULL DEFAULT 'SYNCED'
  );
`);
console.log('   表结构已就绪');

// 开始事务
const importData = db.transaction(() => {
  const now = Date.now();
  
  // 1. 清理旧数据（不删除项目，只删除模块和任务）
  console.log('\n🧹 清理旧数据...');
  db.exec(`
    DELETE FROM dependencies;
    DELETE FROM module_dependencies;
    DELETE FROM tasks;
    DELETE FROM modules;
  `);
  
  // 2. 插入/更新项目
  console.log('\n📝 插入项目...');
  const project = db.prepare('SELECT 1 FROM projects WHERE id = ?').get(projectId);
  if (!project) {
    db.prepare(`
      INSERT INTO projects (id, name, constitution, created_at, updated_at, version, sync_status)
      VALUES (?, ?, ?, ?, ?, 1, 'SYNCED')
    `).run(projectId, jsonData.project?.name || '示例项目', '', now, now);
    console.log(`   创建项目: ${projectId}`);
  } else {
    db.prepare(`
      UPDATE projects SET name = ?, updated_at = ? WHERE id = ?
    `).run(jsonData.project?.name || '示例项目', now, projectId);
    console.log(`   更新项目: ${projectId}`);
  }

  // 3. 插入模块
  console.log('\n📝 插入模块...');
  const insertModule = db.prepare(`
    INSERT INTO modules (id, project_id, name, description, status, created_at, updated_at, version, sync_status)
    VALUES (?, ?, ?, ?, ?, ?, ?, 1, 'SYNCED')
  `);
  
  for (const mod of jsonData.modules || []) {
    insertModule.run(
      mod.id,
      projectId,
      mod.name,
      mod.description || '',
      mod.status || 'designing',
      now,
      now
    );
    console.log(`   模块: ${mod.name} (${mod.id})`);
  }

  // 4. 插入任务
  console.log('\n📝 插入任务...');
  const insertTask = db.prepare(`
    INSERT INTO tasks (
      id, module_id, name, description, status,
      upstream_contract_detail, downstream_contract_detail,
      prompt, tests, logs, code_paths, human_assistance,
      locked, created_at, updated_at, version, sync_status
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, 1, 'SYNCED')
  `);
  
  for (const task of jsonData.tasks || []) {
    insertTask.run(
      task.id,
      task.moduleId,
      task.name,
      task.description || '',
      task.status || 'ready',
      task.upstreamContractDetail || '',
      task.downstreamContractDetail || '',
      task.prompt || '',
      typeof task.tests === 'string' ? task.tests : JSON.stringify(task.tests || []),
      task.logs || '',
      typeof task.codePaths === 'string' ? task.codePaths : JSON.stringify(task.codePaths || []),
      typeof task.humanAssistance === 'string' ? task.humanAssistance : JSON.stringify(task.humanAssistance || {}),
      now,
      now
    );
    console.log(`   任务: ${task.name} (${task.id}) - ${task.status}`);
  }

  // 5. 插入任务依赖
  console.log('\n📝 插入任务依赖...');
  const insertDep = db.prepare(`
    INSERT INTO dependencies (id, upstream_task_id, downstream_task_id, contract_summary, created_at, version, sync_status)
    VALUES (?, ?, ?, ?, ?, 1, 'SYNCED')
  `);
  
  for (const dep of jsonData.taskDependencies || []) {
    insertDep.run(
      dep.id,
      dep.upstreamTaskId,
      dep.downstreamTaskId,
      dep.interfaceContract || '',
      now
    );
    console.log(`   依赖: ${dep.upstreamTaskId} -> ${dep.downstreamTaskId}`);
    if (dep.interfaceContract) {
      console.log(`      接口: ${dep.interfaceContract}`);
    }
  }

  // 6. 计算模块依赖
  console.log('\n📝 计算模块依赖...');
  const moduleDeps = new Map();
  
  for (const dep of jsonData.taskDependencies || []) {
    const upstreamTask = jsonData.tasks.find(t => t.id === dep.upstreamTaskId);
    const downstreamTask = jsonData.tasks.find(t => t.id === dep.downstreamTaskId);
    
    if (upstreamTask && downstreamTask && upstreamTask.moduleId !== downstreamTask.moduleId) {
      const key = `${upstreamTask.moduleId}->${downstreamTask.moduleId}`;
      if (!moduleDeps.has(key)) {
        moduleDeps.set(key, {
          upstream: upstreamTask.moduleId,
          downstream: downstreamTask.moduleId,
          interfaces: []
        });
      }
      if (dep.interfaceContract) {
        moduleDeps.get(key).interfaces.push(dep.interfaceContract);
      }
    }
  }
  
  const insertModuleDep = db.prepare(`
    INSERT INTO module_dependencies (id, module_id, depends_on_module_id, contract_summary, created_at, updated_at, version, sync_status)
    VALUES (?, ?, ?, ?, ?, ?, 1, 'SYNCED')
  `);
  
  let depIndex = 1;
  for (const [key, dep] of moduleDeps) {
    const depId = `moddep-${String(depIndex).padStart(3, '0')}`;
    insertModuleDep.run(
      depId,
      dep.downstream,  // module_id (当前模块)
      dep.upstream,    // depends_on_module_id (依赖的模块)
      dep.interfaces.join('\n'),
      now,
      now
    );
    const upMod = jsonData.modules.find(m => m.id === dep.upstream);
    const downMod = jsonData.modules.find(m => m.id === dep.downstream);
    console.log(`   模块依赖: ${upMod?.name || dep.upstream} -> ${downMod?.name || dep.downstream}`);
    depIndex++;
  }

  console.log('\n✅ 数据导入完成!');
  console.log(`   - 项目: 1`);
  console.log(`   - 模块: ${jsonData.modules?.length || 0}`);
  console.log(`   - 任务: ${jsonData.tasks?.length || 0}`);
  console.log(`   - 任务依赖: ${jsonData.taskDependencies?.length || 0}`);
  console.log(`   - 模块依赖: ${moduleDeps.size}`);
});

// 执行导入
try {
  importData();
} catch (err) {
  console.error('❌ 导入失败:', err.message);
  console.error(err);
  process.exit(1);
}

// 关闭数据库
db.close();
console.log('\n👋 数据库已关闭');
