package resp

import (
	"gx1727.com/xin/internal/infra/logger"
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
	logResponse(c, "error", code, msg)
	c.JSON(200, Response{Code: code, Msg: msg, Data: nil})
}

// Unauthorized 未认证（401）
func Unauthorized(c *gin.Context, msg string) {
	logResponse(c, "warn", 401, msg)
	c.JSON(401, Response{Code: 401, Msg: msg, Data: nil})
}

// Forbidden 无权限（403）
func Forbidden(c *gin.Context, msg string) {
	logResponse(c, "warn", 403, msg)
	c.JSON(403, Response{Code: 403, Msg: msg, Data: nil})
}

// BadRequest 参数校验失败（400）
func BadRequest(c *gin.Context, msg string) {
	logResponse(c, "warn", 400, msg)
	c.JSON(400, Response{Code: 400, Msg: msg, Data: nil})
}

// NotFound 资源不存在（404）
func NotFound(c *gin.Context, msg string) {
	logResponse(c, "warn", 404, msg)
	c.JSON(404, Response{Code: 404, Msg: msg, Data: nil})
}

// ServerError 系统内部错误（500）
func ServerError(c *gin.Context, msg string) {
	logResponse(c, "error", 500, msg)
	c.JSON(500, Response{Code: 500, Msg: msg, Data: nil})
}

func logResponse(c *gin.Context, level string, code int, msg string) {
	requestID, _ := c.Get("request_id")
	reqID, _ := requestID.(string)
	if reqID == "" {
		reqID = "-"
	}
	switch level {
	case "error":
		logger.Errorf("[%s] %s %s | %d | %s", reqID, c.Request.Method, c.Request.URL.Path, code, msg)
	case "warn":
		logger.Warnf("[%s] %s %s | %d | %s", reqID, c.Request.Method, c.Request.URL.Path, code, msg)
	default:
		logger.Infof("[%s] %s %s | %d | %s", reqID, c.Request.Method, c.Request.URL.Path, code, msg)
	}
}

// Paginate 列表分页返回
func Paginate(c *gin.Context, total int64, data interface{}) {
	c.JSON(200, Response{Code: 0, Msg: "ok", Data: gin.H{
		"total": total,
		"list":  data,
	}})
}
