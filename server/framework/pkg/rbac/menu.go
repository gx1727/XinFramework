package rbac

// Menu hook — reserved for future cross-module menu access.
//
// Phase 3 note: menu's struct shape varies by app — apps/rbac/menu
// defines its own Menu struct. No framework-internal module currently
// consumes menu cross-module, so the hook below is dormant. It's
// exposed for symmetry with the other RBAC modules and for future
// use cases (e.g. the auth middleware eventually loading user menus).
//
// apps/rbac/menu's init() does NOT need to register — apps that need
// menus just import apps/rbac/menu directly (no type alias needed).

// RegisterMenuHook is a no-op placeholder for symmetry. Future code
// can register a menu factory here if needed.
func RegisterMenuHook() {}