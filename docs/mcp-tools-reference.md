# AITDD MCP 工具接口参考

本文档列出所有 MCP 工具的输入参数和返回字段，供重构参考。

从原有的 38 个工具精简为 27 个新接口。

---

## 1. 项目上下文 (2个)

### init_project
初始化项目配置。获取数据库中所有项目列表（包含id、名称、简介），返回给 AI 让用户选择。然后自动调用此工具设置当前项目。

- `get_context` - 获取当前上下文信息
- `query_node` - 通用查询节点信息
    `query_project_index_tree` - 查询项目计划索引树
    `query_file_code_path_tree` - 查询代码文件路径树
- `get_status` - 获取状态

    `get_config` - 获取配置
    `get_rule` - 获取规则
    `update_config` - 更新配置
    `update_rule` - 更新规则
    `acquire_lock` / `release_lock` - 获取/释放锁
    `compile_static` - 齐编译：检查项目结构完整性、契约对齐、依赖关系等，生成编译报告
    `compile_dynamic` - 动态编译：执行实际代码生成、测试运行等，生成执行报告

    `get_status` - 获取状态
    `get_rule` - 获取规则
    `update_rule` - 更新规则

