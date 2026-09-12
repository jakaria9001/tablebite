package session

import (
	"context"
	"time"
)

type AdminUser struct {
	ID      int64
	Email   string
	Role    string
	Active  bool
	Expires time.Time
}

// Store persists admin sessions outside process memory so they survive
// across serverless invocations and multiple backend instances.
type Store interface {
	Register(ctx context.Context, token string, user AdminUser) error
	Lookup(ctx context.Context, token string) (AdminUser, bool, error)
	Clear(ctx context.Context, token string) error
}
