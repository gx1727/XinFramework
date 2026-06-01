package flag

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *Handler) {
	// Frames
	public.GET("/flag/frames", h.ListFrames)
	public.GET("/flag/frames/:id", h.GetFrame)
	protected.POST("/flag/frames", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateFrame)
	protected.PUT("/flag/frames/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateFrame)
	protected.DELETE("/flag/frames/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteFrame)

	// Categories
	public.GET("/flag/frames-categories", h.ListFrameCategories)
	protected.POST("/flag/frames-categories", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateFrameCategory)
	protected.PUT("/flag/frames-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateFrameCategory)
	protected.DELETE("/flag/frames-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteFrameCategory)

	// Spaces
	public.GET("/flag/spaces/:code", h.GetSpaceByCode)
	protected.POST("/flag/spaces", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateSpace)
	protected.PUT("/flag/spaces/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateSpace)
	protected.DELETE("/flag/spaces/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteSpace)
	protected.GET("/flag/spaces", middleware.Require(permission.P(permission.ResFlag, permission.ActList)), h.ListSpaces)

	// Avatar Categories
	public.GET("/flag/avatar-categories", h.ListAvatarCategories)
	protected.POST("/flag/avatar-categories", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateAvatarCategory)
	protected.PUT("/flag/avatar-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateAvatarCategory)
	protected.DELETE("/flag/avatar-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteAvatarCategory)

	// Avatars
	public.GET("/flag/avatars", h.ListAvatars)
	public.GET("/flag/avatars/:id", h.GetAvatar)
	protected.POST("/flag/avatars", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateAvatar)
	protected.PUT("/flag/avatars/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateAvatar)
	protected.DELETE("/flag/avatars/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteAvatar)

	// Generate
	protected.POST("/flag/generate", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.GenerateAvatar)
	protected.GET("/flag/my-avatars", middleware.Require(permission.P(permission.ResFlag, permission.ActList)), h.ListMyAvatars)
}
