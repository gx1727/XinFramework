package rbac

// Resource hook — reserved for future cross-module resource access.
//
// No framework-internal module currently consumes it cross-module.
// Apps that need resources import apps/rbac/resource directly.
//
// This file is exposed for symmetry with the other RBAC modules.
// apps/rbac/resource's init() does NOT need to register.

// RegisterResourceHook is a no-op placeholder for symmetry.
func RegisterResourceHook() {}