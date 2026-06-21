package flag

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(public *gin.RouterGroup, tenant *gin.RouterGroup, h *Handler) {
	// Frames
	public.GET("/flag/frames", h.ListFrames)
	public.GET("/flag/frames/:id", h.GetFrame)
	tenant.POST("/flag/frames", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateFrame)
	tenant.PUT("/flag/frames/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateFrame)
	tenant.DELETE("/flag/frames/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteFrame)

	// Categories
	public.GET("/flag/frames-categories", h.ListFrameCategories)
	tenant.POST("/flag/frames-categories", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateFrameCategory)
	tenant.PUT("/flag/frames-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateFrameCategory)
	tenant.DELETE("/flag/frames-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteFrameCategory)

	// Spaces
	public.GET("/flag/spaces/:code", h.GetSpaceByCode)
	tenant.POST("/flag/spaces", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateSpace)
	tenant.PUT("/flag/spaces/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateSpace)
	tenant.DELETE("/flag/spaces/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteSpace)
	tenant.GET("/flag/spaces", middleware.Require(permission.P(permission.ResFlag, permission.ActList)), h.ListSpaces)

	// Avatar Categories
	public.GET("/flag/avatar-categories", h.ListAvatarCategories)
	tenant.POST("/flag/avatar-categories", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateAvatarCategory)
	tenant.PUT("/flag/avatar-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateAvatarCategory)
	tenant.DELETE("/flag/avatar-categories/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteAvatarCategory)

	// Avatars
	public.GET("/flag/avatars", h.ListAvatars)
	public.GET("/flag/avatars/:id", h.GetAvatar)
	tenant.POST("/flag/avatars", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.CreateAvatar)
	tenant.PUT("/flag/avatars/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActUpdate)), h.UpdateAvatar)
	tenant.DELETE("/flag/avatars/:id", middleware.Require(permission.P(permission.ResFlag, permission.ActDelete)), h.DeleteAvatar)

	// Generate
	tenant.POST("/flag/generate", middleware.Require(permission.P(permission.ResFlag, permission.ActCreate)), h.GenerateAvatar)
	tenant.GET("/flag/my-avatars", middleware.Require(permission.P(permission.ResFlag, permission.ActList)), h.ListMyAvatars)
}
