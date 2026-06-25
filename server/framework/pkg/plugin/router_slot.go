package plugin

import "github.com/gin-gonic/gin"

// Slot 名称常量。
//
// 这些是 framework.Register 阶段向业务模块"广播"的路由槽位，
// 模块按名字订阅自己关心的路由组。framework 启动时构造 RouterSlots，
// 业务模块通过 MustGet/SlotByName 取用。
const (
	// SlotPublic 公开域：/api/v1/* （OptionalAuth）。
	// 适用于登录、注册、健康检查、公开字典/配置读取等不需要登录的接口。
	SlotPublic = "public"

	// SlotTenant 租户域：/api/v1/* （Auth + RequireTenantContext）。
	// 适用于租户内业务（user / role / menu / dict / asset / config 等）。
	SlotTenant = "tenant"

	// SlotProtected 平台域：/api/v1/platform/* （Auth）。
	// 适用于平台管理（tenants / sys_users / sys_roles / sys_menus 等）。
	// 平台模块内部需自行追加 RequirePlatformRole("super_admin")。
	SlotProtected = "protected"
)

// RouterSlot 是单个路由槽位。Name 用于业务模块按名字查找；
// Group 是已经装好中间件的 gin.RouterGroup，业务模块直接挂路由即可。
type RouterSlot struct {
	Name        string
	Group       *gin.RouterGroup
	Description string // 可选：描述该槽位的用途（仅 godoc）
}

// RouterSlots 是 framework → Module.Register 阶段传递的路由槽位 map。
//
// 设计动机：原先 Register(ctx, public, tenant, protected) 把三个固定的
// *gin.RouterGroup 写死在签名里，新增第 4 类路由（如 /api/v2、内部灰度）时
// 需要修改 Module 接口本身。改为 slots map 后，framework 决定要注册哪几类，
// 业务模块按名字订阅，扩展性更好。
//
// 取值约定：
//   - Get(name)         安全取，未注册返回 nil
//   - MustGet(name)     严格取，未注册时 panic（开发期应暴露的路由配置错误）
type RouterSlots map[string]*RouterSlot

// Get 安全取 slot，未注册时返回 nil。
//
// 适用于"我知道这个 slot 可能不存在"的场景。
func (s RouterSlots) Get(name string) *RouterSlot {
	if s == nil {
		return nil
	}
	return s[name]
}

// MustGet 严格取 slot，未注册时 panic。
//
// 适用于"这个 slot 必须存在，否则就是框架装配错了"的场景，
// 业务模块在 Register 阶段应优先使用 MustGet——一旦 slot 缺失，
// panic 能在启动期立即暴露，比上线后 404 更友好。
func (s RouterSlots) MustGet(name string) *RouterSlot {
	if s == nil {
		panic("plugin: RouterSlots is nil — framework did not construct slots before calling Register")
	}
	slot, ok := s[name]
	if !ok {
		panic("plugin: router slot \"" + name + "\" is not registered in framework")
	}
	return slot
}

// Names 返回所有已注册 slot 的名字列表，方便日志/调试。
func (s RouterSlots) Names() []string {
	if s == nil {
		return nil
	}
	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, name)
	}
	return names
}