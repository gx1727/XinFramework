package extapi

import "time"

// User DTO for external apps (kept for cms/handler.go JSON response).
// 历史背景：cms module 通过 ctx.UserRepo().GetByID() 拿到 pkgrbac.User，
// 然后转成这个 DTO 返回前端——DTO 还在这里，但 Provider/Facade 已删。
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

// Tenant DTO for external apps (kept for cms/handler.go JSON response).
// pkg/tenant.TenantRecord 没有 json tag，这里提供带 tag 的版本。
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