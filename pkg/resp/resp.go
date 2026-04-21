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
// 所有业务错误统一返回 HTTP 200，HTTP 语义全由业务 Code 表达
type BizError struct {
	Code int    // 业务自定义 Code (如 1001, 2001)
	Msg  string // 默认提示信息
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return e.Msg
}

// WithMsg 动态替换提示信息
func (e *BizError) WithMsg(msg string) *BizError {
	return &BizError{
		Code: e.Code,
		Msg:  msg,
	}
}

// NewError 创建一个业务错误
func NewError(code int, msg string) *BizError {
	return &BizError{
		Code: code,
		Msg:  msg,
	}
}

// HandleError Handler 层的统一错误处理器
// 所有业务错误统一返回 HTTP 200，前端只需判断业务 Code
func HandleError(c *gin.Context, err error) {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		logResponse(c, getLogLevelByCode(bizErr.Code), bizErr.Code, bizErr.Msg)
		c.JSON(http.StatusOK, Response{
			Code: bizErr.Code,
			Msg:  bizErr.Msg,
			Data: nil,
		})
		return
	}

	// 未知错误
	logResponse(c, "error", 500, err.Error())
	c.JSON(http.StatusInternalServerError, Response{
		Code: 500,
		Msg:  "服务器内部错误",
		Data: nil,
	})
}

// getLogLevelByCode 根据业务 Code 判定日志级别
func getLogLevelByCode(code int) string {
	if code >= 5000 {
		return "error"
	}
	if code >= 4000 {
		return "warn"
	}
	return "warn"
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: data})
}

func Error(c *gin.Context, code int, msg string) {
	logResponse(c, "error", code, msg)
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: nil})
}

// Unauthorized 未认证（业务 Code 由调用方指定）
func Unauthorized(c *gin.Context, msg string) {
	logResponse(c, "warn", 401, msg)
	c.JSON(http.StatusOK, Response{Code: 401, Msg: msg, Data: nil})
}

// Forbidden 无权限（业务 Code 由调用方指定）
func Forbidden(c *gin.Context, msg string) {
	logResponse(c, "warn", 403, msg)
	c.JSON(http.StatusOK, Response{Code: 403, Msg: msg, Data: nil})
}

// BadRequest 参数校验失败（业务 Code 由调用方指定）
func BadRequest(c *gin.Context, msg string) {
	logResponse(c, "warn", 400, msg)
	c.JSON(http.StatusOK, Response{Code: 400, Msg: msg, Data: nil})
}

// NotFound 资源不存在（业务 Code 由调用方指定）
func NotFound(c *gin.Context, msg string) {
	logResponse(c, "warn", 404, msg)
	c.JSON(http.StatusOK, Response{Code: 404, Msg: msg, Data: nil})
}

// ServerError 系统内部错误
func ServerError(c *gin.Context, msg string) {
	logResponse(c, "error", 500, msg)
	c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: msg, Data: nil})
}

func logResponse(c *gin.Context, level string, code int, msg string) {
	requestID, _ := c.Get("request_id")
	reqID, _ := requestID.(string)
	if reqID == "" {
		reqID = "-"
	}
	method := "-"
	path := "-"
	if c.Request != nil {
		method = c.Request.Method
		if c.Request.URL != nil {
			path = c.Request.URL.Path
		}
	}
	switch level {
	case "error":
		logger.Errorf("[%s] %s %s | %d | %s", reqID, method, path, code, msg)
	case "warn":
		logger.Warnf("[%s] %s %s | %d | %s", reqID, method, path, code, msg)
	default:
		logger.Infof("[%s] %s %s | %d | %s", reqID, method, path, code, msg)
	}
}

// Paginate 列表分页返回
func Paginate(c *gin.Context, total int64, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: gin.H{
		"total": total,
		"list":  data,
	}})
}
