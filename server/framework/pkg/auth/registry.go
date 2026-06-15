package auth

// AccountRepositoryFactory produces an AccountRepository backed by
// the apps/boot/auth implementation. Registered at process start by
// the auth module via Register; consumed by user (and any other
// framework-internal module that needs account access).
//
// Phase 2 rationale: framework/internal cannot import apps/ directly,
// so user module cannot construct apps/boot/auth.NewAccountRepository.
// Instead, apps/boot/auth's init() pushes its factory here, and user
// fetches it through Get.
var globalAccountFactory func() AccountRepository

// Register wires an AccountRepository factory. Typically called from
// apps/boot/auth's package init().
func Register(f func() AccountRepository) {
	globalAccountFactory = f
}

// Get returns the registered factory, or nil if auth is not loaded.
// user module's localAccountAdapter treats nil as a configuration
// error and returns clean error responses.
func Get() func() AccountRepository {
	return globalAccountFactory
}