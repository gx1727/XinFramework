package menu

import "errors"

var (
	ErrMenuNotFound       = errors.New("menu not found")
	ErrMenuCodeExists     = errors.New("menu code already exists")
	ErrBackendUnavailable = errors.New("backend service unavailable")
)
