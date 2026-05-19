package user

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/users", middleware.RequirePermission("user", "list"), h.List)
	protected.GET("/users/:id", middleware.RequirePermission("user", "list"), h.Get)
	protected.PUT("/users/:id/status", middleware.RequirePermission("user", "update"), h.UpdateStatus)
	protected.GET("/user/profile", h.Profile)
	protected.POST("/user/avatar", h.UploadAvatar)
	protected.PUT("/user/profile", h.UpdateProfile)
}
