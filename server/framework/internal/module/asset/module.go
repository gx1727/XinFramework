package asset

import (
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

// Module 返回 asset 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("asset", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		// 创建 storage
		var s storage.Storage
		if config.Get().Storage.Provider == "cos" {
			cosStorage, err := storage_cos.NewCosStorage(storage_cos.Config{
				URL:       config.Get().Storage.CosURL,
				SecretID:  config.Get().Storage.CosSecretID,
				SecretKey: config.Get().Storage.CosSecretKey,
				BaseURL:   config.Get().Storage.CosBaseURL,
			})
			if err != nil {
				log.Fatalf("failed to init cos storage: %v", err)
			}
			s = cosStorage
		} else {
			s = storage_local.NewLocalStorage(
				config.Get().Storage.LocalDir,
				config.Get().Storage.LocalBaseURL,
			)
		}

		// 创建 service 和 handler
		svc := NewFileService(s, NewAttachmentRepository(db.Get()))
		h := NewFileHandler(svc)

		// 注册路由
		Register(public, protected, h)
	})
}
