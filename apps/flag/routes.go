package flag

import (
	"github.com/gin-gonic/gin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	// Frames
	public.GET("/flag/frames", h.ListFrames)
	public.GET("/flag/frames/:id", h.GetFrame)
	protected.POST("/flag/frames", h.CreateFrame)
	protected.PUT("/flag/frames/:id", h.UpdateFrame)
	protected.DELETE("/flag/frames/:id", h.DeleteFrame)

	// Categories
	public.GET("/flag/categories", h.ListCategories)
	protected.POST("/flag/categories", h.CreateFrameCategory)
	protected.PUT("/flag/categories/:id", h.UpdateFrameCategory)
	protected.DELETE("/flag/categories/:id", h.DeleteFrameCategory)

	// Spaces
	public.GET("/flag/spaces/:code", h.GetSpaceByCode)
	protected.POST("/flag/spaces", h.CreateSpace)
	protected.PUT("/flag/spaces/:id", h.UpdateSpace)
	protected.DELETE("/flag/spaces/:id", h.DeleteSpace)
	protected.GET("/flag/spaces", h.ListSpaces)

	// Generate
	protected.POST("/flag/generate", h.GenerateAvatar)
	protected.GET("/flag/my-avatars", h.ListMyAvatars)

	// Avatar Categories
	public.GET("/flag/avatar-categories", h.ListAvatarCategories)
	protected.POST("/flag/avatar-categories", h.CreateAvatarCategory)
	protected.PUT("/flag/avatar-categories/:id", h.UpdateAvatarCategory)
	protected.DELETE("/flag/avatar-categories/:id", h.DeleteAvatarCategory)

	// Avatars
	public.GET("/flag/avatars", h.ListAvatars)
	public.GET("/flag/avatars/:id", h.GetAvatar)
	protected.POST("/flag/avatars", h.CreateAvatar)
	protected.PUT("/flag/avatars/:id", h.UpdateAvatar)
	protected.DELETE("/flag/avatars/:id", h.DeleteAvatar)
}
