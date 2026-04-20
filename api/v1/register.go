package v1

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/internal/infra/logger"
	"gx1727.com/xin/internal/module/auth"
	"gx1727.com/xin/internal/module/cms"
	"gx1727.com/xin/internal/module/system"
	"gx1727.com/xin/internal/module/weixin"
	"gx1727.com/xin/pkg/config"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	logger.Infof("register routes...........................")
	v1 := r.Group("/api/v1")
	if cfg.DomainEnabled("system") {
		system.RegisterV1(v1)
		auth.RegisterV1(v1)
	}
	if cfg.DomainEnabled("cms") {
		cms.RegisterV1(v1)
	}
	if cfg.DomainEnabled("weixin") {
		weixin.RegisterV1(v1)
	}
}
