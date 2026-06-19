package resource

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the resource module as a BaseModule.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "resource",
		InitFn: func(_ plugin.Reader, _ plugin.Writer) error {
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			authzSvc := ctx.Authz()
			h := NewHandler(NewService(NewResourceRepository(pool), authzSvc))
			Register(protected, h)
		},
	}
}
