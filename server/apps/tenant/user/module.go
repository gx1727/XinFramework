package user

import (
	"log"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/apps/tenant/organization"
	"gx1727.com/xin/apps/tenant/role"
	"gx1727.com/xin/apps/reference/asset"
	"gx1727.com/xin/framework/pkg/appx"
	"gx1727.com/xin/framework/pkg/plugin"
	"gx1727.com/xin/framework/pkg/storage"
	storage_cos "gx1727.com/xin/framework/pkg/storage/cos"
	storage_local "gx1727.com/xin/framework/pkg/storage/local"
)

// Module returns the user module as a BaseModule.
//
func Module(app *appx.App) plugin.Module {
	return &plugin.BaseModule{
		NameStr: "user",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := app.DB
			w.SetUserRepo(NewUserRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, tenant *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := app.DB
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}

			var s storage.Storage
			cfg := app.Config
			if cfg.Storage.Provider == "cos" {
				cosStorage, err := storage_cos.NewCosStorage(storage_cos.Config{
					URL:       cfg.Storage.CosURL,
					SecretID:  cfg.Storage.CosSecretID,
					SecretKey: cfg.Storage.CosSecretKey,
					BaseURL:   cfg.Storage.CosBaseURL,
				})
				if err != nil {
					log.Fatalf("failed to init cos storage for user: %v", err)
				}
				s = cosStorage
			} else {
				s = storage_local.NewLocalStorage(
					cfg.Storage.LocalDir,
					cfg.Storage.LocalBaseURL,
				)
			}

			assetSvc := asset.NewFileService(s, asset.NewAttachmentRepository(pool))

			accountRepo := ctx.AccountRepo()
			if accountRepo == nil {
				log.Printf("user: apps/boot/auth not loaded, skipping")
				return
			}

			h := NewHandler(NewService(
				pool,
				NewUserRepository(pool),
				role.NewRoleRepository(pool),
				organization.NewOrganizationRepository(pool),
				assetSvc,
				accountRepo,
			))
			Register(tenant, h)
		},
	}
}
