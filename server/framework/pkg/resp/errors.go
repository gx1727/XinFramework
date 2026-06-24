package resp

// 模块错误码分段常量。
//
// 分段与 HTTP 状态码的映射由 resp.CodeToHTTPStatus 集中实现，
// 新增模块请申请连续区间，并保证 code ∈ 区间 → 走对应 HTTP（见包级文档）。
const (
	CodeAuth         = 1000  // auth: 1001-1999        (1xxx 默认 200)
	CodeUser         = 2000  // user: 2001-2999        (2xxx → 400)
	CodeTenant       = 3000  // tenant: 3001-3999      (3xxx → 404)
	CodeRole         = 4000  // role: 4001-4999        (4xxx → 403)
	CodeMenu         = 5000  // menu: 5001-5999        (5xxx → 500)
	CodeOrganization = 6000  // organization: 6001-6999 (6xxx → 500；如需 4xx 复用 CodeRole 段)
	CodePermission   = 7000  // permission: 7001-7999  (7xxx → 500)
	CodeResource     = 8000  // resource: 8001-8999    (8xxx → 500)
	CodeAsset        = 9000  // asset: 9001-9999       (9xxx → 500)
	CodeDict         = 10000 // dict: 10001-10999      (10xxx → 500)
	CodeSystem       = 11000 // system: 11001-11999    (11xxx → 500)
	CodeWeixin       = 12000 // weixin: 12001-12999    (12xxx → 500)
	CodeFlag         = 13000 // flag: 13001-13999      (13xxx → 500)
	CodeCMS          = 14000 // cms: 14001-14999       (14xxx → 500；示例模块)
)

// Err 模块错误构造函数。推荐用法：
//
//	var ErrUserNotFound = resp.Err(2001, "用户不存在")
//
// Err 与 NewError 等价，仅函数名更短；所有业务错误统一用此构造。
func Err(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}
