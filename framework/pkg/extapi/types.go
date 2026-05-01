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

// CmsPost DTO for external apps
type CmsPost struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    int16     `json:"status"`
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

type CmsPostFacade interface {
	GetByID(ctx context.Context, id uint) (*CmsPost, error)
	List(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]CmsPost, int64, error)
	Create(ctx context.Context, tenantID uint, title, content string, status int16) (*CmsPost, error)
	Update(ctx context.Context, id uint, title, content string, status int16) error
	Delete(ctx context.Context, id uint) error
}

// Provider is the Facade for external apps to interact with internal modules.
type Provider interface {
	User() UserFacade
	Tenant() TenantFacade
	CmsPost() CmsPostFacade
}
