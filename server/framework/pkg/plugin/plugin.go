// Package plugin 模块契约。
//
// Module 接口：
//
//	type Module interface {
//	    Name() string
//	    Init(ctx Reader, w Writer) error
//	    Register(ctx Reader, slots RouterSlots)
//	    Shutdown(ctx Reader) error
//	}
//
// slots 是 framework 在 Register 阶段构造的路由插槽 map，业务模块
// 通过 slots.MustGet(SlotPublic | SlotTenant | SlotProtected) 取路由组。
// 如需新增第 4 类路由（如 /api/v2、内部灰度），只需在 framework 侧
// 往 slots 里多注册一个名字，业务模块无需改接口。
//
// 三类内置 slot 语义：
//   - SlotPublic    → /api/v1/*             （OptionalAuth，公开；需隔离的子资源挂 /public/<x>）
//   - SlotTenant    → /api/v1/*             （Auth + RequireTenantContext，业务域；模块直接挂资源路径，无 /t 前缀）
//   - SlotProtected → /api/v1/platform/*    （Auth，平台域；模块内部追加 RequirePlatformRole）
//
// 历史背景：旧版本有 NewModule / NewModuleLegacy / NewModuleWithOpts
// 三种构造器外加 ModuleOption / WithInit 等兼容 API，全部在 Phase 2
// 删除；曾有 Register / Apps 全局注册表（Phase 0-4 兼容 API 残骸），
// 在 Phase 6 删除——main.go 现在显式构造 []Module 传给 framework.Serve。
//
// 新模块直接用 BaseModule 或实现 Module 接口。
package plugin

import "github.com/gin-gonic/gin"

// Module 业务模块对外的契约。
type Module interface {
	Name() string
	Init(ctx Reader, w Writer) error
	// Register 注册路由到 framework 提供的 slots。
	// 业务模块通过 slots.MustGet(SlotPublic|SlotTenant|SlotProtected)
	// 取对应的 *gin.RouterGroup，然后挂载自己的路由。
	Register(ctx Reader, slots RouterSlots)
	Shutdown(ctx Reader) error
}

// ModuleFunc 是简单模块的 Register 回调形状（无需 Init/Shutdown）。
type ModuleFunc func(ctx Reader, slots RouterSlots)

// BaseModule 是 Module 接口的默认实现。所有字段可选，nil 字段视为 noop。
//
// 推荐新模块使用 BaseModule：未来要扩展 HealthCheck / DependsOn / Migrate
// 等能力时只需在 struct 上加字段，不用改 interface。
type BaseModule struct {
	NameStr string
	InitFn  func(ctx Reader, w Writer) error
	RegFn   ModuleFunc
	StopFn  func(ctx Reader) error
}

func (m *BaseModule) Name() string { return m.NameStr }

func (m *BaseModule) Init(ctx Reader, w Writer) error {
	if m.InitFn == nil {
		return nil
	}
	return m.InitFn(ctx, w)
}

func (m *BaseModule) Register(ctx Reader, slots RouterSlots) {
	if m.RegFn != nil {
		m.RegFn(ctx, slots)
	}
}

func (m *BaseModule) Shutdown(ctx Reader) error {
	if m.StopFn == nil {
		return nil
	}
	return m.StopFn(ctx)
}

// 编译期断言：BaseModule 满足 Module 接口。
var _ Module = (*BaseModule)(nil)

// 抑制未使用的 gin 导入告警（gin.RouterGroup 仅在 RouterSlot.Group 字段间接使用）。
var _ = (*gin.RouterGroup)(nil)