package mcp

// tools_status.go - 旧的状态相关函数已迁移到 tools_other.go
//
// 原有的函数已整合到新的统一接口中：
// - handleGetAllTaskStatusImpl -> 由 handleGetStatusImpl 替代
// - handleGetModuleTaskStatusImpl -> 由 handleGetStatusImpl 替代
// - handleGetProjectErrorsImpl -> 由 handleGetStatusImpl 替代
// - handleGetModuleErrorsImpl -> 由 handleGetStatusImpl 替代
//
// 新的 get_status 接口支持通过 type 参数区分不同节点类型
