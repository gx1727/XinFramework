package asset

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/middleware"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *FileHandler) {
	// Asset routes group
	assetGroup := protected.Group("/asset")

	{
		// Upload endpoint
		assetGroup.POST("/upload", middleware.RequirePermission("asset", "create"), h.Upload)

		// Delete endpoint
		assetGroup.DELETE("/:id", middleware.RequirePermission("asset", "delete"), h.Delete)
	}
}
