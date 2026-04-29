package asset

import (
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

// Module 返回 asset 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("asset", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		// 创建 storage
		var s storage.Storage
		if app.Config.Storage.Provider == "cos" {
			cosStorage, err := storage_cos.NewCosStorage(storage_cos.Config{
				URL:       app.Config.Storage.CosURL,
				SecretID:  app.Config.Storage.CosSecretID,
				SecretKey: app.Config.Storage.CosSecretKey,
				BaseURL:   app.Config.Storage.CosBaseURL,
			})
			if err != nil {
				log.Fatalf("failed to init cos storage: %v", err)
			}
			s = cosStorage
		} else {
			s = storage_local.NewLocalStorage(
				app.Config.Storage.LocalDir,
				app.Config.Storage.LocalBaseURL,
			)
		}

		// 创建 service 和 handler
		svc := NewFileService(s, app.Repository.Attachment())
		h := NewFileHandler(svc)

		// 注册路由
		Register(public, protected, h)
	})
}
