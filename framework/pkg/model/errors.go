package model

import "errors"

var (
	// User errors
	ErrUserNotFound = errors.New("user not found")
	ErrUserDisabled = errors.New("user is disabled")

	// Role errors
	ErrRoleNotFound        = errors.New("role not found")
	ErrDefaultRoleNotFound = errors.New("default role not found")

	// Account errors
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")

	// Tenant errors
	ErrTenantNotFound   = errors.New("tenant not found")
	ErrTenantCodeExists = errors.New("tenant code already exists")

	// Menu errors
	ErrMenuNotFound = errors.New("menu not found")

	// Resource errors
	ErrResourceNotFound = errors.New("resource not found")

	// Backend errors
	ErrBackendUnavailable = errors.New("backend service unavailable")
)
