// Package authz 暴露 Authorization 接口，供中间件（Auth / Require / RequireAll）
// 与业务模块（需要手动清理权限缓存时）使用。
//
// 具体实现 *service.AuthorizationService 位于 framework/internal/service
// （Go 的 internal/ 规则阻止 apps/ 直接 import）。本包暴露一个与该服务
// 签名一致的小型接口，无需额外适配层。
//
// 装配路径：boot.Init 构造具体服务，编译期断言其满足 Authorization 接口，
// 然后通过 appCtx.SetAuthz(...) 写入 AppContext；业务模块在 Register 阶段
// 通过 ctx.Authz() 取出使用。
package authz

import (
	"context"

	"gx1727.com/xin/framework/pkg/permission"
)

// Authorization 是 apps 可消费的公开鉴权接口。
//
// 方法签名与 framework/internal/service.AuthorizationService 一致；
// boot 启动期将 *AuthorizationService 装配到本接口并通过 AppContext.SetAuthz 发布。
//
// 如果 AuthorizationService 新增了 apps 需要的方法，必须同步加到这里。
//
// 编译期一致性保证位于 framework/internal/core/boot/boot.go：
//
//	var _ Authorization = (*service.AuthorizationService)(nil)
type Authorization interface {
	// LoadPermissions 返回用户的有效权限码 map（resource_code → bool）。
	LoadPermissions(ctx context.Context, userID uint) (map[string]bool, error)

	// LoadRoles 返回用户被分配的角色编码列表。
	LoadRoles(ctx context.Context, userID uint) ([]string, error)

	// LoadDataScope 返回用户的数据范围限制。
	// 返回具体类型 *permission.DataScope 而非 any —— 调用方无需 type assert。
	LoadDataScope(ctx context.Context, userID uint) (*permission.DataScope, error)

	// LoadUserSecurityContext 一次性加载用户的完整鉴权上下文（权限码 / 角色 /
	// 数据范围 / 会话版本），供 Auth 中间件每个请求调用。
	// 签名与 framework/internal/core/middleware.SecurityContextLoader 共享，
	// 中间件可直接消费本接口而无需额外包装。
	LoadUserSecurityContext(ctx context.Context, userID uint) (map[string]bool, []string, *permission.DataScope, int64, error)

	// InvalidateUser 清理指定用户的权限 / 数据范围缓存。
	InvalidateUser(ctx context.Context, userID uint) error

	// InvalidateRole 清理所有持有该角色的用户的缓存。
	InvalidateRole(ctx context.Context, roleID uint) error

	// InvalidateResource 清理所有受该资源变更影响的用户的缓存。
	InvalidateResource(ctx context.Context, resourceID uint) error
}