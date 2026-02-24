package models

import (
	"time"

	"gorm.io/gorm"
)

// ModuleDependency 模块依赖关系
type ModuleDependency struct {
	ID               string `json:"id" gorm:"primaryKey;type:text"`
	ModuleID         string `json:"moduleId" gorm:"not null;type:text;index"`         // 依赖方模块ID
	DependsOnModuleID string `json:"dependsOnModuleId" gorm:"not null;type:text;index"`        // 被依赖模块ID
	DependencyType   string `json:"dependencyType" gorm:"type:text;default:'required'"`     // 依赖类型: required, optional, conditional
	ContractSummary  string `json:"contractSummary" gorm:"type:text"`                        // 合约摘要
	CreatedAt        int64  `json:"createdAt" gorm:"not null"`
	UpdatedAt        int64  `json:"updatedAt" gorm:"not null"`
	Version          int    `json:"version" gorm:"not null;default:1"`
	SyncStatus       string `json:"syncStatus" gorm:"not null;default:'SYNCED';type:text"`
}

// ModuleDependencyType 依赖类型枚举
const (
	ModuleDependencyRequired    = "required"    // 必需依赖
	ModuleDependencyOptional    = "optional"    // 可选依赖
	ModuleDependencyConditional = "conditional" // 条件依赖
)

// BeforeCreate 创建前钩子
func (d *ModuleDependency) BeforeCreate(_ *gorm.DB) error {
	if d.CreatedAt == 0 {
		d.CreatedAt = time.Now().UnixMilli()
	}
	if d.UpdatedAt == 0 {
		d.UpdatedAt = time.Now().UnixMilli()
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (d *ModuleDependency) BeforeUpdate(_ *gorm.DB) error {
	d.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// TableName 指定表名
func (ModuleDependency) TableName() string {
	return "module_dependencies"
}

// ModuleDependencyDetail 模块依赖详情（包含被依赖模块的信息）
type ModuleDependencyDetail struct {
	ModuleDependency
	DependsOnModule *Module `json:"dependsOnModule"` // 被依赖的模块信息
}

// ModuleDependentDetail 模块被依赖详情（包含依赖方模块的信息）
type ModuleDependentDetail struct {
	ModuleDependency
	Module *Module `json:"module"` // 依赖方的模块信息
}
