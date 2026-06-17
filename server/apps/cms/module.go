package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the cms module as a BaseModule.
//
// Migration note (Phase 5): cms will move from direct Repository
// construction to reading its DB pool off the AppContext.Reader.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "cms",
		RegFn: func(_ plugin.Reader, public *gin.RouterGroup, protected *gin.RouterGroup) {
			h := NewHandler()
			Register(h, public, protected)
		},
	}
}