// Package plugin 模块契约与全局注册表。
//
// Module 接口（精简后）：
//
//	type Module interface {
//	    Name() string
//	    Init(ctx Reader, w Writer) error
//	    Register(ctx Reader, public, protected *gin.RouterGroup)
//	    Shutdown(ctx Reader) error
//	}
//
// 历史背景：旧版本有 NewModule / NewModuleLegacy / NewModuleWithOpts
// 三种构造器外加 ModuleOption / WithInit 等兼容 API，全部在 Phase 2
// 删除。新模块直接用 BaseModule 或实现 Module 接口。
package plugin

import "github.com/gin-gonic/gin"

// Module 业务模块对外的契约。
type Module interface {
	Name() string
	Init(ctx Reader, w Writer) error
	Register(ctx Reader, public *gin.RouterGroup, protected *gin.RouterGroup)
	Shutdown(ctx Reader) error
}

// ModuleFunc 是简单模块的 Register 回调形状（无需 Init/Shutdown）。
type ModuleFunc func(ctx Reader, public *gin.RouterGroup, protected *gin.RouterGroup)

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

func (m *BaseModule) Register(ctx Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
	if m.RegFn != nil {
		m.RegFn(ctx, public, protected)
	}
}

func (m *BaseModule) Shutdown(ctx Reader) error {
	if m.StopFn == nil {
		return nil
	}
	return m.StopFn(ctx)
}

// apps 存储通过 Register 注册的全部模块。
var apps []Module

// Register 将模块加入全局列表。同名模块重复注册被忽略。
func Register(m Module) {
	for _, existing := range apps {
		if existing.Name() == m.Name() {
			return
		}
	}
	apps = append(apps, m)
}

// Apps 返回所有已注册模块（注册顺序）。
func Apps() []Module { return apps }
