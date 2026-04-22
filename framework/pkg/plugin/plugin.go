package plugin

import "github.com/gin-gonic/gin"

type Module interface {
	Name() string
	RegisterV1(public, protected *gin.RouterGroup)
}

type ModuleFunc func(public, protected *gin.RouterGroup)

type SimpleModule struct {
	name     string
	register ModuleFunc
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

var modules []Module

func Register(m Module) {
	modules = append(modules, m)
}

func All() []Module {
	return modules
}
