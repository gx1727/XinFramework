package asset

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/plugin"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *FileHandler) {
	// Asset routes group
	assetGroup := protected.Group("/asset")

	{
		// Upload endpoint
		assetGroup.POST("/upload", h.Upload)

		// Delete endpoint
		assetGroup.DELETE("/:id", h.Delete)
	}
}

func Module(h *FileHandler) plugin.Module {
	return plugin.NewModule("asset", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		Register(public, protected, h)
	})
}
