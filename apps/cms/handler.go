package cms

import (
	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/resp"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(c *gin.Context) {
	resp.Success(c, gin.H{
		"domain": "cms",
		"status": "enabled",
		"config": moduleCfg,
	})
}
