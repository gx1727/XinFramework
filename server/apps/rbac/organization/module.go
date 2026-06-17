package organization

import (
	"github.com/gin-gonic/gin"

	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

func init() {
	plugin.Register(Module())
}

// Module returns the organization module as a BaseModule.
//
// Phase 4: publishes OrganizationRepository onto the AppContext.Writer.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "organization",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := db.Get()
			w.SetOrgRepo(NewOrganizationRepository(pool))
			return nil
		},
		RegFn: func(_ plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			h := NewHandler(NewService(NewOrganizationRepository(pool)))
			Register(protected, h)
		},
	}
}