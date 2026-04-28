package user

import (
	"github.com/gin-gonic/gin"
)

func Register(protected *gin.RouterGroup, h *Handler) {
	protected.GET("/users", h.List)
	protected.GET("/users/:id", h.Get)
	protected.PUT("/users/:id/status", h.UpdateStatus)
	protected.GET("/user/profile", h.Profile)
	protected.POST("/user/avatar", h.UploadAvatar)
	protected.PUT("/user/profile", h.UpdateProfile)
}
