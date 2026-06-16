package organization

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	orgs := protected.Group("/organizations")
	{
		orgs.GET("/tree", middleware.Require(permission.P(permission.ResOrganization, permission.ActList)), h.Tree)
		orgs.GET("", middleware.Require(permission.P(permission.ResOrganization, permission.ActList)), h.List)
		orgs.GET("/:id", middleware.Require(permission.P(permission.ResOrganization, permission.ActList)), h.Get)
		orgs.POST("", middleware.Require(permission.P(permission.ResOrganization, permission.ActCreate)), h.Create)
		orgs.PUT("/:id", middleware.Require(permission.P(permission.ResOrganization, permission.ActUpdate)), h.Update)
		orgs.DELETE("/:id", middleware.Require(permission.P(permission.ResOrganization, permission.ActDelete)), h.Delete)
	}
}
