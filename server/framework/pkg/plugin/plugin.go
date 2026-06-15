package plugin

import "github.com/gin-gonic/gin"

type Module interface {
	Name() string
	Init() error
	Register(public *gin.RouterGroup, protected *gin.RouterGroup)
	Shutdown() error
}

type ModuleFunc func(public *gin.RouterGroup, protected *gin.RouterGroup)

// module 是 Module 接口的标准实现
type module struct {
	name       string
	register   ModuleFunc
	initFn     func() error
	shutdownFn func() error
}

func (m *module) Name() string {
	return m.name
}

func (m *module) Register(public *gin.RouterGroup, protected *gin.RouterGroup) {
	m.register(public, protected)
}

func (m *module) Init() error {
	if m.initFn != nil {
		return m.initFn()
	}
	return nil
}

func (m *module) Shutdown() error {
	if m.shutdownFn != nil {
		return m.shutdownFn()
	}
	return nil
}

type ModuleOption func(*module)

func WithInit(fn func() error) ModuleOption {
	return func(m *module) { m.initFn = fn }
}

func WithShutdown(fn func() error) ModuleOption {
	return func(m *module) { m.shutdownFn = fn }
}

// NewModule 创建一个简单的插件模块
func NewModule(name string, register ModuleFunc) Module {
	return NewModuleWithOpts(name, register)
}

// NewModuleWithOpts 创建一个支持可选配置的插件模块
func NewModuleWithOpts(name string, register ModuleFunc, opts ...ModuleOption) Module {
	m := &module{name: name, register: register}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// apps 存储通过 plugin.Register() 注册的外部应用模块
//
// 历史上框架内置模块通过 framework.go 中的 builtinMap 注册，外部
// app 通过 init() 中的 plugin.Register() 注册。这种双轨制让 main.go
// 不得不硬编码所有内置模块名。
//
// 重构后所有模块（内置 + 外部）一律通过 plugin.Register() 注册。
// builtinMap 已删除；原先的"内置"模块在各自包内通过 init() 调用
// plugin.Register(self) 完成注册，与外部 app 走完全相同的路径。
var apps []Module

// Register 将模块注册到全局列表。
// 重复注册同名模块将被忽略（保护措施，避免 main.go 多次 import 同一模块）。
func Register(m Module) {
	for _, existing := range apps {
		if existing.Name() == m.Name() {
			return
		}
	}
	apps = append(apps, m)
}

// Apps 返回所有已注册的模块（顺序为注册顺序）。
func Apps() []Module {
	return apps
}
