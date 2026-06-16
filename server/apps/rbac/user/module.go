package user

import (
	"log"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/apps/reference/asset"
	"gx1727.com/xin/apps/rbac/organization"
	"gx1727.com/xin/apps/rbac/role"
	pkgrbac "gx1727.com/xin/framework/pkg/rbac"
	"gx1727.com/xin/framework/pkg/config"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

func init() {
	plugin.Register(Module())

	// Phase 3: register this module's UserRepository with the framework's
	// public pkg/rbac registry so that framework/internal consumers
	// (e.g. weixin) can resolve user data without importing apps/.
	pkgrbac.RegisterUserRepository(func() pkgrbac.UserRepository {
		return NewUserRepository(db.Get())
	})
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

		// Phase 3: user is now in apps/rbac/, so it can import
		// apps/boot/auth directly — the Phase 2 framework-side
		// adapter (newLocalAccountAdapter) is no longer needed.
		h := NewHandler(NewService(
			NewUserRepository(db.Get()),
			role.NewRoleRepository(db.Get()),
			organization.NewOrganizationRepository(db.Get()),
			assetSvc,
			newAccountAdapter(),
		))
		Register(protected, h)
	})
}