package asset

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/middleware"
	"gx1727.com/xin/framework/pkg/permission"
)

func Register(public *gin.RouterGroup, protected *gin.RouterGroup, h *FileHandler) {
	// Asset routes group
	assetGroup := protected.Group("/asset")

	{
		// Upload endpoint
		assetGroup.POST("/upload", middleware.Require(permission.P(permission.ResAsset, permission.ActCreate)), h.Upload)

		// Delete endpoint
		assetGroup.DELETE("/:id", middleware.Require(permission.P(permission.ResAsset, permission.ActDelete)), h.Delete)
	}
}
