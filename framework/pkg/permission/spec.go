package permission


// Spec describes an authorization requirement for a route or action.
type Spec struct {
	Resource      string
	Action        string
	Authenticated bool
}

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
