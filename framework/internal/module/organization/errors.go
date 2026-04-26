package organization

import "errors"

var (
	ErrOrgNotFound        = errors.New("organization not found")
	ErrOrgCodeExists      = errors.New("organization code already exists")
	ErrCannotDeleteRoot   = errors.New("cannot delete root organization")
	ErrBackendUnavailable = errors.New("backend unavailable")
)
