package role

import "errors"

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleCodeExists     = errors.New("role code already exists")
	ErrCannotDeleteAdmin  = errors.New("cannot delete admin role")
	ErrBackendUnavailable = errors.New("backend unavailable")
)
