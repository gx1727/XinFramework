package server

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin-framework/configs"
)

type XinServer struct {
	Engine *gin.Engine
}

func New(cfg *configs.Config) *XinServer {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	return &XinServer{Engine: engine}
}

func (s *XinServer) Start(addr string) error {
	return s.Engine.Run(addr)
}
