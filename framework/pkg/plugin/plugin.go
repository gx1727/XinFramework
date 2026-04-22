package plugin

import "github.com/gin-gonic/gin"

// Module 模块接口，定义插件模块的基本行为
type Module interface {
	Name() string                                                 // 获取模块名称
	Init() error                                                  // 初始化模块
	Migrate() error                                               // 执行数据库迁移
	Register(public *gin.RouterGroup, protected *gin.RouterGroup) // 注册API路由（公开和受保护）
}

// ModuleFunc 模块路由注册函数类型
type ModuleFunc func(public *gin.RouterGroup, protected *gin.RouterGroup)

// SimpleModule 简单模块实现，提供基础的模块功能
type SimpleModule struct {
	name      string       // 模块名称
	register  ModuleFunc   // 路由注册函数
	initFn    func() error // 初始化函数（可选）
	migrateFn func() error // 迁移函数（可选）
}

// NewModule 创建一个新的简单模块实例
func NewModule(name string, register ModuleFunc) *SimpleModule {
	return &SimpleModule{name: name, register: register}
}

// Name 返回模块名称
func (m *SimpleModule) Name() string {
	return m.name
}

// Register 注册API路由到指定的路由组
func (m *SimpleModule) Register(public, protected *gin.RouterGroup) {
	m.register(public, protected)
}

// Init 执行模块初始化，如果定义了初始化函数则调用
func (m *SimpleModule) Init() error {
	if m.initFn != nil {
		return m.initFn()
	}
	return nil
}

// Migrate 执行数据库迁移，如果定义了迁移函数则调用
func (m *SimpleModule) Migrate() error {
	if m.migrateFn != nil {
		return m.migrateFn()
	}
	return nil
}

// ModuleOption 模块配置选项函数类型，用于自定义模块行为
type ModuleOption func(*SimpleModule)

// WithInit 设置模块的初始化函数
func WithInit(fn func() error) ModuleOption {
	return func(m *SimpleModule) { m.initFn = fn }
}

// WithMigrate 设置模块的数据库迁移函数
func WithMigrate(fn func() error) ModuleOption {
	return func(m *SimpleModule) { m.migrateFn = fn }
}

// NewModuleWithOpts 使用配置选项创建模块实例
func NewModuleWithOpts(name string, register ModuleFunc, opts ...ModuleOption) *SimpleModule {
	m := &SimpleModule{name: name, register: register}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// modules 全局模块注册表
var modules []Module

// Register 注册一个模块到全局模块列表
func Register(m Module) {
	modules = append(modules, m)
}

// All 返回所有已注册的模块
func All() []Module {
	return modules
}
