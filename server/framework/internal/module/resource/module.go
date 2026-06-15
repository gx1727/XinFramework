package resource

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

func Module() plugin.Module {
	return plugin.NewModule("resource", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		h := NewHandler(NewService(NewResourceRepository(db.Get())))
		Register(protected, h)
	})
}
