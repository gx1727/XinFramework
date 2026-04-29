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

var modules []Module

func Register(m Module) {
	modules = append(modules, m)
}

func All() []Module {
	return modules
}
