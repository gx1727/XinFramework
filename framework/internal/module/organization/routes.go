package organization

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	orgs := protected.Group("/organizations")
	{
		orgs.GET("/tree", middleware.RequirePermission(permission.ResOrganization, permission.ActList), h.Tree)
		orgs.GET("", middleware.RequirePermission(permission.ResOrganization, permission.ActList), h.List)
		orgs.GET("/:id", middleware.RequirePermission(permission.ResOrganization, permission.ActList), h.Get)
		orgs.POST("", middleware.RequirePermission(permission.ResOrganization, permission.ActCreate), h.Create)
		orgs.PUT("/:id", middleware.RequirePermission(permission.ResOrganization, permission.ActUpdate), h.Update)
		orgs.DELETE("/:id", middleware.RequirePermission(permission.ResOrganization, permission.ActDelete), h.Delete)
	}
}
