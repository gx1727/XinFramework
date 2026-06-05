package session

import "errors"

var (
	ErrEmptySessionID     = errors.New("empty session id")
	ErrInvalidSessionTTL  = errors.New("invalid session ttl")
	ErrBackendUnavailable = errors.New("session backend unavailable: db not initialized")
)
