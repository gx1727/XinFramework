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

// Unauthorized 未认证（401）
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(401, Response{Code: 401, Msg: msg, Data: nil})
}

// Forbidden 无权限（403）
func Forbidden(c *gin.Context, msg string) {
	c.JSON(403, Response{Code: 403, Msg: msg, Data: nil})
}

// BadRequest 参数校验失败（400）
func BadRequest(c *gin.Context, msg string) {
	c.JSON(400, Response{Code: 400, Msg: msg, Data: nil})
}

// NotFound 资源不存在（404）
func NotFound(c *gin.Context, msg string) {
	c.JSON(404, Response{Code: 404, Msg: msg, Data: nil})
}

// ServerError 系统内部错误（500）
func ServerError(c *gin.Context, msg string) {
	c.JSON(500, Response{Code: 500, Msg: msg, Data: nil})
}

// Paginate 列表分页返回
func Paginate(c *gin.Context, total int64, data interface{}) {
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: gin.H{
		"total": total,
		"list":  data,
	}})
}
