package user

import (
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/core/boot"
	"gx1727.com/xin/framework/internal/module/asset"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

// Module 返回 user 模块的完整定义
func Module(app *boot.App) plugin.Module {
	return plugin.NewModule("user", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
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
				log.Fatalf("failed to init cos storage for user: %v", err)
			}
			s = cosStorage
		} else {
			s = storage_local.NewLocalStorage(
				app.Config.Storage.LocalDir,
				app.Config.Storage.LocalBaseURL,
			)
		}

		assetSvc := asset.NewFileService(s, app.Repository.Attachment())
		h := NewHandler(NewService(app.Repository.User(), app.Repository.Role(), assetSvc))
		Register(protected, h)
	})
}
