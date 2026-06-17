package user

import (
	"log"

	"github.com/gin-gonic/gin"

	"gx1727.com/xin/apps/rbac/organization"
	"gx1727.com/xin/apps/rbac/role"
	"gx1727.com/xin/apps/reference/asset"
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

// Module returns the user module as a BaseModule.
//
// Phase 4 changes:
//   - Init publishes UserRepository onto the AppContext.Writer for
//     framework-internal consumers (weixin).
//   - Register consumes AccountRepo from AppContext.Reader (cross-
//     framework-boundary dep). apps-internal deps (role, org) are
//     constructed directly because the framework-side interfaces are
//     narrower (subset of operations) than what user.Service requires.
func Module() plugin.Module {
	return &plugin.BaseModule{
		NameStr: "user",
		InitFn: func(_ plugin.Reader, w plugin.Writer) error {
			pool := db.Get()
			w.SetUserRepo(NewUserRepository(pool))
			return nil
		},
		RegFn: func(ctx plugin.Reader, _ *gin.RouterGroup, protected *gin.RouterGroup) {
			pool := db.Get()
			if ctx != nil {
				if p := ctx.DB(); p != nil {
					pool = p
				}
			}

			var s storage.Storage
			cfg := config.Get()
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

			// Cross-framework-boundary dep: account lives in
			// apps/boot/auth and is published via AppContext.
			accountRepo := ctx.AccountRepo()
			if accountRepo == nil {
				log.Printf("user: apps/boot/auth not loaded, skipping")
				return
			}

			h := NewHandler(NewService(
				NewUserRepository(pool),
				role.NewRoleRepository(pool),
				organization.NewOrganizationRepository(pool),
				assetSvc,
				accountRepo,
			))
			Register(protected, h)
		},
	}
}