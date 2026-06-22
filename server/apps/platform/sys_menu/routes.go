package sysmenu

import (
	"github.com/gin-gonic/gin"

	jwtpkg "gx1727.com/xin/framework/pkg/jwt"
	pkgmiddleware "gx1727.com/xin/framework/pkg/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	g := protected.Group("/sys-menus",
		pkgmiddleware.RequirePlatformRole(jwtpkg.PlatformRoleSuperAdmin),
	)
	{
		g.GET("", h.List)
		g.GET("/tree", h.Tree)
		g.GET("/:id", h.Get)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
