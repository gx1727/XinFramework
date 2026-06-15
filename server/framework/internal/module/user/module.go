package user

import (
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/internal/module/asset"
	"gx1727.com/xin/framework/internal/module/organization"
	"gx1727.com/xin/framework/internal/module/role"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

func init() {
	plugin.Register(Module())
}

// Module 返回 user 模块的完整定义
func Module() plugin.Module {
	return plugin.NewModule("user", func(public *gin.RouterGroup, protected *gin.RouterGroup) {
		var s storage.Storage
		if config.Get().Storage.Provider == "cos" {
			cosStorage, err := storage_cos.NewCosStorage(storage_cos.Config{
				URL:       config.Get().Storage.CosURL,
				SecretID:  config.Get().Storage.CosSecretID,
				SecretKey: config.Get().Storage.CosSecretKey,
				BaseURL:   config.Get().Storage.CosBaseURL,
			})
			if err != nil {
				log.Fatalf("failed to init cos storage for user: %v", err)
			}
			s = cosStorage
		} else {
			s = storage_local.NewLocalStorage(
				config.Get().Storage.LocalDir,
				config.Get().Storage.LocalBaseURL,
			)
		}

		assetSvc := asset.NewFileService(s, asset.NewAttachmentRepository(db.Get()))

		// Phase 2: AccountRepository moved to apps/boot/auth. user module
		// (which still lives in framework/internal) cannot import apps/,
		// so we instantiate the auth-side repository via a thin local
		// adapter. Once Phase 3 moves user to apps/rbac/user/, this
		// adapter goes away — user can then import apps/boot/auth directly.
		h := NewHandler(NewService(
			NewUserRepository(db.Get()),
			role.NewRoleRepository(db.Get()),
			organization.NewOrganizationRepository(db.Get()),
			assetSvc,
			newLocalAccountAdapter(db.Get()),
		))
		Register(protected, h)
	})
}