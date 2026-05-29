package resp

// 模块错误码分段定义
// 各模块错误码应在本范围内定义，禁止越界
const (
	CodeAuth         = 1000  // auth: 1001-1999
	CodeUser         = 2000  // user: 2001-2999
	CodeTenant       = 3000  // tenant: 3001-3999
	CodeRole         = 4000  // role: 4001-4999
	CodeMenu         = 5000  // menu: 5001-5999
	CodeOrganization = 6000  // organization: 6001-6999
	CodePermission   = 7000  // permission: 7001-7999
	CodeResource     = 8000  // resource: 8001-8999
	CodeAsset        = 9000  // asset: 9001-9999
	CodeDict         = 10000 // dict: 10001-10999
	CodeSystem       = 11000 // system: 11001-11999
	CodeWeixin       = 12000 // weixin: 12001-12999
)

// Err 模块错误构造函数，替代直接使用 NewError
// 用法: var ErrUserNotFound = resp.Err(2001, "用户不存在")
func Err(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}
