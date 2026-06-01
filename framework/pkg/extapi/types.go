package extapi

import (
	"context"
	"time"
)

// User DTO for external apps
type User struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	AccountID uint      `json:"account_id"`
	Code      string    `json:"code"`
	Nickname  string    `json:"nickname"`
	Status    int8      `json:"status"`
	RealName  string    `json:"real_name"`
	Avatar    string    `json:"avatar"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tenant DTO for external apps
type Tenant struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    int16     `json:"status"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Area      string    `json:"area"`
	Address   string    `json:"address"`
	Config    string    `json:"config"`
	Dashboard string    `json:"dashboard"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserFacade interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error)
}

type TenantFacade interface {
	GetByID(ctx context.Context, id uint) (*Tenant, error)
}

// Provider is the Facade for external apps to interact with internal modules.
type Provider interface {
	User() UserFacade
	Tenant() TenantFacade
}
