package resp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/internal/infra/logger"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// BizError 标准业务错误
type BizError struct {
	HttpCode int    // HTTP 状态码 (如 200, 400, 401)
	Code     int    // 业务自定义 Code (如 1001, 2001)
	Msg      string // 默认提示信息
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return e.Msg
}

// WithMsg 动态替换提示信息
func (e *BizError) WithMsg(msg string) *BizError {
	return &BizError{
		HttpCode: e.HttpCode,
		Code:     e.Code,
		Msg:      msg,
	}
}

// NewError 创建一个业务错误
func NewError(httpCode, code int, msg string) *BizError {
	return &BizError{
		HttpCode: httpCode,
		Code:     code,
		Msg:      msg,
	}
}

// HandleError Handler 层的统一错误处理器
func HandleError(c *gin.Context, err error) {
	var bizErr *BizError
	// 如果是预定义的业务错误
	if errors.As(err, &bizErr) {
		logResponse(c, getLogLevel(bizErr.HttpCode), bizErr.Code, bizErr.Msg)
		c.JSON(bizErr.HttpCode, Response{
			Code: bizErr.Code,
			Msg:  bizErr.Msg,
			Data: nil,
		})
		return
	}

	// 未知错误，统一按 500 处理，避免真实报错暴露给前端
	logResponse(c, "error", 500, err.Error())
	c.JSON(http.StatusInternalServerError, Response{
		Code: 500,
		Msg:  "服务器内部错误",
		Data: nil,
	})
}

func getLogLevel(httpCode int) string {
	if httpCode >= 500 {
		return "error"
	}
	if httpCode >= 400 {
		return "warn"
	}
	return "info"
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
