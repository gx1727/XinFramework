package permission

// Spec describes an authorization requirement for a route or action.
type Spec struct {
	Resource      string
	Action        string
	Authenticated bool
}

// MatchMode 决定 Require / RequireAny / RequireAll 的判定方式。
// 用具名类型代替原来的 "one"/"any"/"all" 字符串,让拼写错误编译期失败。
type MatchMode int

const (
	// MatchAll 全部 spec 都必须通过（Require / RequireAuthenticated）。
	MatchAll MatchMode = iota
	// MatchAny 至少一个 spec 通过即可（RequireAny）。
	MatchAny
)

// P creates a permission-based authorization spec.
func P(resource, action string) Spec {
	return Spec{
		Resource:      resource,
		Action:        action,
		Authenticated: true,
	}
}

// AuthOnly creates a "login required, no RBAC check" spec.
func AuthOnly() Spec {
	return Spec{Authenticated: true}
}

func (s Spec) IsPermission() bool {
	return s.Resource != "" || s.Action != ""
}

func (s Spec) IsAuthOnly() bool {
	return s.Authenticated && !s.IsPermission()
}

func (s Spec) IsValid() bool {
	if !s.Authenticated {
		return false
	}
	if s.IsAuthOnly() {
		return true
	}
	return s.Resource != "" && s.Action != ""
}

func (s Spec) String() string {
	if s.IsAuthOnly() {
		return "auth"
	}
	return s.Resource + ":" + s.Action
}
