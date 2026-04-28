package asset

import (
	"github.com/gin-gonic/gin"
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
