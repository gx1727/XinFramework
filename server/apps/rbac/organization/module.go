package organization

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/bootx"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the organization module as a BaseModule.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "organization",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := bootx.Pool()
			w.SetOrgRepo(NewOrganizationRepository(pool))
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := bootx.Pool()
			h := NewHandler(NewService(pool, NewOrganizationRepository(pool)))
			Register(protected, h)
		},
	}
}
