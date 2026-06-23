package user

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(tenant *gin.RouterGroup, h *Handler) {
	tenant.GET("/users", middleware.Require(permission.P(permission.ResUser, permission.ActList)), h.List)
	tenant.POST("/users", middleware.Require(permission.P(permission.ResUser, permission.ActCreate)), h.Create)
	tenant.GET("/users/:id", middleware.Require(permission.P(permission.ResUser, permission.ActList)), h.Get)
	tenant.PUT("/users/:id", middleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.Update)
	tenant.PATCH("/users/:id", middleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.Patch)
	tenant.PUT("/users/:id/status", middleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.UpdateStatus)
	tenant.PUT("/users/:id/org", middleware.Require(permission.P(permission.ResUser, permission.ActUpdate)), h.UpdateOrg)
	tenant.GET("/user/profile", h.Profile)
	tenant.POST("/user/avatar", h.UploadAvatar)
	tenant.PUT("/user/profile", h.UpdateProfile)
}
