package resource

import "errors"

var (
	ErrResourceNotFound     = errors.New("resource not found")
	ErrResourceCodeExists   = errors.New("resource code already exists")
	ErrCannotDeleteResource = errors.New("cannot delete system resource")
	ErrBackendUnavailable   = errors.New("backend unavailable")
)
