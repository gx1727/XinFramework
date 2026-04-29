package cms

import (
	"time"

	"gx1727.com/xin/framework/pkg/model"
)

// CmsPost CMS 文章模型
type CmsPost = model.CmsPost

// User 用户模型（简化版）
type User struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Code      string    `json:"code"`
	RealName  string    `json:"real_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    int16     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tenant 租户模型（简化版）
type Tenant struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Province  string    `json:"province"`
	City      string    `json:"city"`
	Area      string    `json:"area"`
	Address   string    `json:"address"`
	Config    string    `json:"config"`
	Dashboard string    `json:"dashboard"`
	Status    int16     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uint      `json:"created_by"`
	UpdatedBy uint      `json:"updated_by"`
}
