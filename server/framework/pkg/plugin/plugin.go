package plugin

import "github.com/gin-gonic/gin"

// Module is the contract every business module implements.
//
// Lifecycle:
//
//   1. Register(m)        — called at process start, before Run().
//                            The module is appended to the global list.
//                            No side effects beyond self-registration.
//
//   2. Init(ctx)          — called by boot.Init for every module whose
//                            name is in cfg.Module. The module MUST
//                            publish whatever it owns into ctx via
//                            the Writer methods. The module MAY read
//                            other modules' contributions, but a nil
//                            repository is a documented signal that
//                            the producer module is not enabled.
//
//   3. Register(ctx, public, protected) — called by framework after
//                            every module has finished Init. By this
//                            point ctx.Reader exposes a stable view
//                            of all contributions. The module wires
//                            HTTP routes into public/protected.
//
//   4. Shutdown(ctx)      — graceful teardown. ctx is the same one
//                            passed during Init, so cleanup can use
//                            the same dependencies (e.g. flush
//                            caches, close connections it opened
//                            itself).
//
// Reader/Writer split: a module only sees Writer for slots it owns.
// Compile-time enforcement via SetX methods on Writer. See appcontext.go.
type Module interface {
	Name() string
	Init(ctx Reader, w Writer) error
	Register(ctx Reader, public *gin.RouterGroup, protected *gin.RouterGroup)
	Shutdown(ctx Reader) error
}

// ModuleFunc is the register-callback shape used by simple modules
// that have no Init() or Shutdown() work.
type ModuleFunc func(ctx Reader, public *gin.RouterGroup, protected *gin.RouterGroup)

// BaseModule is a concrete Module implementation that delegates
// every method to optional function pointers. It is the recommended
// type for new modules because:
//
//   - It is a struct, so the type system can be extended with new
//     methods (HealthCheck, DependsOn, Migrate) without breaking
//     callers.
//   - It does not require modules to define their own method-set.
//
// Existing modules that today use plugin.NewModule("name", fn) can
// stay on the function form: NewModule is preserved as a thin
// wrapper that builds a BaseModule under the hood.
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

// NewModule builds a Module from a pre-AppContext register callback.
// The callback ignores the Reader argument.
//
// Kept on the old signature deliberately so all 14 existing module.go
// files compile without modification during Phase 2 verification.
// Phase 3-5 gradually migrates each module to NewModuleCtx (which
// takes the new ModuleFunc that receives the Reader).
func NewModule(name string, register LegacyRegisterFunc) Module {
	return &legacyModule{name: name, regFn: register}
}

// NewModuleCtx builds a BaseModule for a register callback that
// needs the AppContext.Reader. This is the signature NEW modules
// must use.
func NewModuleCtx(name string, register ModuleFunc) Module {
	return &BaseModule{
		NameStr: name,
		RegFn:   register,
	}
}

// ------------------------------------------------------------------
// Backwards-compat layer for the pre-AppContext API.
//
// These wrappers exist so existing modules (weixin, etc.) that still
// use the legacy signature `func(public, protected)` and `func() error`
// keep compiling. They are removed when each module migrates in Phase
// 3-5. New modules must NOT use these.
// ------------------------------------------------------------------

// LegacyRegisterFunc is the pre-AppContext register callback shape.
type LegacyRegisterFunc func(public *gin.RouterGroup, protected *gin.RouterGroup)

// LegacyInitFunc is the pre-AppContext init callback shape.
type LegacyInitFunc func() error

// ModuleOption configures a legacy module.
type ModuleOption func(*legacyModule)

// WithInit attaches an Init hook (legacy signature).
func WithInit(fn LegacyInitFunc) ModuleOption {
	return func(m *legacyModule) { m.initFn = fn }
}

// NewModuleWithOpts builds a Module using legacy signatures. Phase 3-5
// rewrites each caller to BaseModule; once that work finishes this
// helper is removed.
func NewModuleWithOpts(name string, register LegacyRegisterFunc, opts ...ModuleOption) Module {
	lm := &legacyModule{name: name, regFn: register}
	for _, opt := range opts {
		opt(lm)
	}
	return lm
}

// legacyModule adapts the old (no-ctx) API to the new (ctx-based)
// Module interface. Register/Init/Shutdown swallow the new Reader
// argument; modules that need a real Reader must migrate to BaseModule.
type legacyModule struct {
	name   string
	regFn  LegacyRegisterFunc
	initFn LegacyInitFunc
}

func (m *legacyModule) Name() string { return m.name }
func (m *legacyModule) Init(_ Reader, _ Writer) error {
	if m.initFn == nil {
		return nil
	}
	return m.initFn()
}
func (m *legacyModule) Register(_ Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
	if m.regFn != nil {
		m.regFn(public, protected)
	}
}
func (m *legacyModule) Shutdown(_ Reader) error { return nil }

// NewModuleLegacy builds a Module from a pre-AppContext register
// callback. The callback ignores the Reader argument; modules that
// need a real Reader must migrate to BaseModule directly.
//
// Use this only as a transitional helper during Phase 3-5. New
// modules must use NewModule (which takes the new ModuleFunc shape)
// or BaseModule directly.
func NewModuleLegacy(name string, register LegacyRegisterFunc) Module {
	return &legacyModule{name: name, regFn: register}
}

// apps 存储通过 plugin.Register() 注册的外部应用模块
//
// 历史: 框架内置模块通过 framework.go 中的 builtinMap 注册,外部
// app 通过 init() 中的 plugin.Register() 注册。这种双轨制让 main.go
// 不得不硬编码所有内置模块名。
//
// 重构后所有模块(内置 + 外部)一律通过 plugin.Register() 注册。
// builtinMap 已删除;原先的"内置"模块在各自包内通过 init() 调用
// plugin.Register(self) 完成注册,与外部 app 走完全相同的路径。
var apps []Module

// Register 将模块注册到全局列表。
// 重复注册同名模块将被忽略(保护措施,避免 main.go 多次 import 同一模块)。
func Register(m Module) {
	for _, existing := range apps {
		if existing.Name() == m.Name() {
			return
		}
	}
	apps = append(apps, m)
}

// Apps 返回所有已注册的模块(顺序为注册顺序)。
func Apps() []Module {
	return apps
}
