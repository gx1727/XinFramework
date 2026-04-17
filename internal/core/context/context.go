package context

import (
	"github.com/gin-gonic/gin"
)

type XinContext struct {
	*gin.Context
	TenantID uint
	UserID   uint
}

func New(c *gin.Context) *XinContext {
	return &XinContext{Context: c}
}

func (x *XinContext) SetTenantID(id uint) {
	x.TenantID = id
}

func (x *XinContext) SetUserID(id uint) {
	x.UserID = id
}
