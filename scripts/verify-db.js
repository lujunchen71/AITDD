/**
 * 验证数据库内容
 */
const Database = require('better-sqlite3');
const path = require('path');
const os = require('os');

const DB_PATH = path.join(os.homedir(), '.aitdd', 'aitdd.db');
console.log('数据库路径:', DB_PATH);

const db = new Database(DB_PATH);

// 列出所有表
const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all();
console.log('\n=== 数据库表 ===');
tables.forEach(t => console.log('  - ' + t.name));

// 统计各表数据
console.log('\n=== 数据统计 ===');
try {
  const modules = db.prepare('SELECT id, name FROM modules').all();
  console.log('模块数:', modules.length);
  modules.forEach(m => console.log('  - ' + m.id + ': ' + m.name));
} catch (e) {
  console.log('模块表不存在');
}

try {
  const tasks = db.prepare('SELECT id, name, status FROM tasks').all();
  console.log('\n任务数:', tasks.length);
  tasks.forEach(t => console.log('  - ' + t.id + ': ' + t.name + ' (' + t.status + ')'));
} catch (e) {
  console.log('任务表不存在');
}

try {
  const deps = db.prepare('SELECT COUNT(*) as count FROM dependencies').get();
  console.log('\n依赖数:', deps.count);
} catch (e) {
  console.log('依赖表不存在');
}

db.close();
console.log('\n验证完成');
