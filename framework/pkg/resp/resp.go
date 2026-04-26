package resp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gx1727.com/xin/framework/pkg/logger"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// BizError 标准业务错误
type BizError struct {
	Code int    // 业务自定义 Code (如 1001, 2001)
	Msg  string // 默认提示信息
}

func (e *BizError) Error() string {
	return e.Msg
}

func (e *BizError) WithMsg(msg string) *BizError {
	return &BizError{Code: e.Code, Msg: msg}
}

func NewError(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// HandleError Handler 层的统一错误处理器
// 根据错误类型返回对应 HTTP 状态码
func HandleError(c *gin.Context, err error) {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		httpStatus := http.StatusOK
		level := "warn"
		if bizErr.Code >= 5000 {
			level = "error"
		}
		logResponse(c, level, bizErr.Code, bizErr.Msg)
		c.JSON(httpStatus, Response{Code: bizErr.Code, Msg: bizErr.Msg, Data: nil})
		return
	}

	// 未知错误
	logResponse(c, "error", 500, err.Error())
	c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: "服务器内部错误", Data: nil})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: data})
}

// Error 返回业务错误，HTTP 状态码由业务 code 决定
func Error(c *gin.Context, code int, msg string) {
	httpStatus := http.StatusOK
	if code >= 5000 {
		httpStatus = http.StatusInternalServerError
	} else if code >= 4000 {
		httpStatus = http.StatusForbidden
	} else if code >= 3000 {
		httpStatus = http.StatusNotFound
	} else if code >= 2000 {
		httpStatus = http.StatusBadRequest
	}
	logResponse(c, "error", code, msg)
	c.JSON(httpStatus, Response{Code: code, Msg: msg, Data: nil})
}

// Unauthorized 未认证 - HTTP 401
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "未登录"
	}
	logResponse(c, "warn", 401, msg)
	c.JSON(http.StatusUnauthorized, Response{Code: 401, Msg: msg, Data: nil})
}

// Forbidden 无权限 - HTTP 403
func Forbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "无权限访问"
	}
	logResponse(c, "warn", 403, msg)
	c.JSON(http.StatusForbidden, Response{Code: 403, Msg: msg, Data: nil})
}

// BadRequest 参数校验失败 - HTTP 400
func BadRequest(c *gin.Context, msg string) {
	if msg == "" {
		msg = "请求参数错误"
	}
	logResponse(c, "warn", 400, msg)
	c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: msg, Data: nil})
}

// NotFound 资源不存在 - HTTP 404
func NotFound(c *gin.Context, msg string) {
	if msg == "" {
		msg = "资源不存在"
	}
	logResponse(c, "warn", 404, msg)
	c.JSON(http.StatusNotFound, Response{Code: 404, Msg: msg, Data: nil})
}

// ServerError 系统内部错误 - HTTP 500
func ServerError(c *gin.Context, msg string) {
	if msg == "" {
		msg = "服务器内部错误"
	}
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
