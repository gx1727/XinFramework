package plugin

import "github.com/gin-gonic/gin"

type Module interface {
	Name() string
	Init() error
	Migrate() error
	RegisterV1(public, protected *gin.RouterGroup)
}

type ModuleFunc func(public, protected *gin.RouterGroup)

type SimpleModule struct {
	name      string
	register  ModuleFunc
	initFn    func() error
	migrateFn func() error
}

func NewModule(name string, register ModuleFunc) *SimpleModule {
	return &SimpleModule{name: name, register: register}
}

func (m *SimpleModule) Name() string {
	return m.name
}

func (m *SimpleModule) RegisterV1(public, protected *gin.RouterGroup) {
	m.register(public, protected)
}

func (m *SimpleModule) Init() error {
	if m.initFn != nil {
		return m.initFn()
	}
	return nil
}

func (m *SimpleModule) Migrate() error {
	if m.migrateFn != nil {
		return m.migrateFn()
	}
	return nil
}

type ModuleOption func(*SimpleModule)

func WithInit(fn func() error) ModuleOption {
	return func(m *SimpleModule) { m.initFn = fn }
}

func WithMigrate(fn func() error) ModuleOption {
	return func(m *SimpleModule) { m.migrateFn = fn }
}

func NewModuleWithOpts(name string, register ModuleFunc, opts ...ModuleOption) *SimpleModule {
	m := &SimpleModule{name: name, register: register}
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
