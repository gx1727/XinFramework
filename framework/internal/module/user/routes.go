package user

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/users", middleware.RequirePermission(permission.ResUser, permission.ActList), h.List)
	protected.GET("/users/:id", middleware.RequirePermission(permission.ResUser, permission.ActList), h.Get)
	protected.PUT("/users/:id/status", middleware.RequirePermission(permission.ResUser, permission.ActUpdate), h.UpdateStatus)
	protected.GET("/user/profile", h.Profile)
	protected.POST("/user/avatar", h.UploadAvatar)
	protected.PUT("/user/profile", h.UpdateProfile)
}
