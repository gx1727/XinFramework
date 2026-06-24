// Package resp 提供统一的 HTTP 响应与错误处理约定。
//
// # 错误码分段 → HTTP 状态码映射
//
// 所有业务错误统一走 *BizError，code 分段与 HTTP 状态码有明确映射，
// 由 CodeToHTTPStatus 单点实现，禁止在调用处自行判断：
//
//	[1xxx]    鉴权 / 账号 / 通用业务  → 200  (body Code 表达真实结果)
//	[2xxx]    参数校验 / 业务规则     → 400 Bad Request
//	[3xxx]    资源不存在 / 租户级     → 404 Not Found
//	[4xxx]    权限不足 / 角色冲突     → 403 Forbidden
//	[5xxx+]   服务端故障 / 系统异常   → 500 Internal Server Error
//
// 注意：历史遗留——menu（5xxx）/ organization（6xxx）等模块的「资源不存在」
// 类错误目前共用 5xxx+ 段，按现有规则会被映射为 HTTP 500。这是已知 gap，
// 修复路径是给这些模块重新分配段位（如把 menu 调到 35xx 段），不是本包的事。
//
// 新模块在 errors.go 申请一段连续区间，并保证 code ∈ 区间 → 走对应 HTTP。
// 见 errors.go 的 Code* 常量表。
//
// # 错误构造
//
//   - 业务错误统一用 resp.Err(code, msg) 或 resp.NewError(code, msg) 构造。
//   - DB 层 sentinel 必须命名（Err*DB / err*DB），禁止在 repository 函数里
//     裸调 errors.New("xxx not found")——那样 service 层无法用 errors.Is
//     区分"未找到"与"DB 故障"，会导致所有错误被错误地归为 500。
//
// # Handler 调用约定
//
//   - 已知业务错误：service 返回 *BizError（或由 mapRepoError 翻译后的错误），
//     handler 调 resp.HandleError(c, err)。
//   - 显式分流：参数校验 → resp.BadRequest；权限 → resp.Forbidden 等。
//   - 未识别 error：resp.HandleError 已兜底返回 500 + 通用文案，
//     业务模块不要自己拼 JSON。
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
	Data any `json:"data"`
}

// BizError 标准业务错误。
//
// 实现 errors.Is 接口：两个 BizError 只要 Code 相同即视为同一错误
// （Msg 不一致时调用 errors.Is 仍返回 true），因此可以安全地用作哨兵错误。
type BizError struct {
	Code int    // 业务自定义 Code (如 1001, 2001)
	Msg  string // 默认提示信息
}

func (e *BizError) Error() string {
	return e.Msg
}

// Unwrap 实现隐式错误链语义（io/fs.PathError 风格的兼容接口）。
func (e *BizError) Unwrap() error { return nil }

// Is 使 errors.Is 按 Code 匹配，而非指针相等。
// 例如 errors.Is(fmt.Errorf("%w: extra", resp.Err(2001)), resp.Err(2001)) 返回 true。
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	return ok && e.Code == t.Code
}

func (e *BizError) WithMsg(msg string) *BizError {
	return &BizError{Code: e.Code, Msg: msg}
}

func NewError(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// CodeToHTTPStatus 根据 BizError.Code 决定 HTTP 状态码。
//
// 分段规则见包级文档。该函数是 code → HTTP 映射的唯一入口，
// HandleError / Error 都必须调用它，禁止重复实现 if-else 链。
func CodeToHTTPStatus(code int) int {
	switch {
	case code >= 5000:
		return http.StatusInternalServerError
	case code >= 4000:
		return http.StatusForbidden
	case code >= 3000:
		return http.StatusNotFound
	case code >= 2000:
		return http.StatusBadRequest
	default:
		return http.StatusOK
	}
}

// HandleError Handler 层的统一错误处理器
// 业务错误按 code 分段返回对应 HTTP 状态码（与 Error 函数一致），
// 未识别 error 兜底返回 500 + 通用文案。
func HandleError(c *gin.Context, err error) {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		httpStatus := CodeToHTTPStatus(bizErr.Code)
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

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: data})
}

// Error 返回业务错误，HTTP 状态码由业务 code 决定（CodeToHTTPStatus）。
func Error(c *gin.Context, code int, msg string) {
	logResponse(c, "error", code, msg)
	c.JSON(CodeToHTTPStatus(code), Response{Code: code, Msg: msg, Data: nil})
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
func Paginate(c *gin.Context, total int64, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: gin.H{
		"total": total,
		"list":  data,
	}})
}
