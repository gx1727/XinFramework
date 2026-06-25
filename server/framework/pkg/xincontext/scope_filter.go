package xincontext

import (
	"context"

	"gx1727.com/xin/framework/pkg/permission"
)

// ScopeFilterFrom 从 ctx 中自动取出 UserContext 的 DataScope，并生成
// 可直接拼接到 WHERE 子句的 SQL 过滤片段。
//
// 这是 DataScope 生成的 ctx-aware 一行式入口。Repository 层不再需要手动
// 取出 UserContext、判空、再调用 BuildDataScopeFilter。
//
// 取值规则：
//   - ctx 中没有 UserContext（中间件链漏挂 Auth、或后台异步任务调用）：
//     返回空 ScopeFilter（等同于 DataScopeAll，不施加任何过滤）。
//   - UserContext.DataScope 为空（零值）时同样返回空过滤。
//
// 配合 WHERE 拼接的标准用法：
//
//	filter, err := xincontext.ScopeFilterFrom(ctx, permission.ScopeColumns{
//	    SelfColumn: "u.id",
//	    OrgID:      "u.org_id",
//	})
//	if err != nil {
//	    return err
//	}
//	if !filter.IsEmpty() {
//	    where = append(where, filter.SQL)
//	    args  = append(args,  filter.Args...)
//	}
//
// 如果你需要更复杂的列名映射，请改用 permission.BuildDataScopeFilter 直接调用。
//
// 本函数定义在 xincontext 包而非 permission 包，避免与 permission → db
// 之间的依赖形成循环（xincontext 已依赖 permission）。
func ScopeFilterFrom(ctx context.Context, columns permission.ScopeColumns) (permission.ScopeFilter, error) {
	// ctx=nil 是容错边界：CLI / 后台异步任务可能传入 nil context。
	// 这里不能 panic，应视为"无 UserContext"，返回空过滤。
	if ctx == nil {
		return permission.ScopeFilter{}, nil
	}
	uc, ok := UserContextFrom(ctx)
	if !ok || uc == nil {
		return permission.ScopeFilter{}, nil
	}
	return permission.BuildDataScopeFilter(uc.DataScope, uc.UserID, uc.OrgID, columns)
}