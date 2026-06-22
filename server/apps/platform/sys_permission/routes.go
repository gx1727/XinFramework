package syspermission

import (
	"github.com/gin-gonic/gin"

	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/sys-permissions",
		pkgmiddleware.RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin),
	)
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
