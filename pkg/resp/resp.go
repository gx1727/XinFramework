package resp

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: data})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(200, Response{Code: code, Msg: msg, Data: nil})
}

func Paginate(c *gin.Context, total int64, data interface{}) {
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: gin.H{
		"total": total,
		"list":  data,
	}})
}
