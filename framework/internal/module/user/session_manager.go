package user

import "time"

type SessionManager interface {
	Create(sessionID string, userID, tenantID uint, role string, ttl time.Duration) error
	Revoke(sessionID string) error
}
