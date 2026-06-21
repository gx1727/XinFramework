package platformtenant

import "encoding/json"

// TenantConfig 租户级 JSONB 配置契约。前端 DynamicForm 按 Schema 渲染，避免黑盒。
// 新增字段请同步更新前端 schema，保持向后兼容（旧字段允许缺失）。
type TenantConfig struct {
	// LogoURL 租户 logo（CDN/cos 路径）
	LogoURL string `json:"logo_url,omitempty"`
	// Theme light | dark | auto
	Theme string `json:"theme,omitempty"`
	// Locale zh-CN | en-US
	Locale string `json:"locale,omitempty"`
	// ModuleFlags 模块启用位图（true=启用）。前端按位渲染菜单与导航。
	ModuleFlags map[string]bool `json:"module_flags,omitempty"`
	// SubscriptionOverrides 单租户覆盖默认套餐限额（如 max_users=1000）
	SubscriptionOverrides map[string]int `json:"subscription_overrides,omitempty"`
}

// Dashboard 租户仪表盘布局：左中右三栏的 widget key 列表。
// key 对应 UI/src/components/widgets/<key>.tsx。
type Dashboard struct {
	Top    []string `json:"top,omitempty"`
	Left   []string `json:"left,omitempty"`
	Center []string `json:"center,omitempty"`
	Right  []string `json:"right,omitempty"`
}

type CreateTenantReq struct {
	Code    string `json:"code" binding:"required,min=1,max=50"`
	Name    string `json:"name" binding:"required,min=1,max=100"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Status  *int16 `json:"status"`
	// AdminAccountID 首装 admin 用户绑定的账号 ID（accounts.id）。
	// 必填：必须先在 accounts 表创建账号，再创建租户并绑定。
	// 留空时跳过 admin user 创建，只建空租户（用于纯数据迁移场景）。
	AdminAccountID *uint `json:"admin_account_id,omitempty"`
}

type UpdateTenantReq struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Contact  string `json:"contact"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   *int16 `json:"status"`
	Province string `json:"province"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Address  string `json:"address"`
}

type UpdateTenantStatusReq struct {
	Status int16 `json:"status" binding:"required,oneof=0 1"`
}

type ListTenantReq struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Size    int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword"`
	Status  *int16 `form:"status"`
}

type TenantResp struct {
	ID        uint        `json:"id"`
	Code      string      `json:"code"`
	Name      string      `json:"name"`
	Status    int16       `json:"status"`
	Contact   string      `json:"contact"`
	Phone     string      `json:"phone"`
	Email     string      `json:"email"`
	Province  string      `json:"province"`
	City      string      `json:"city"`
	Area      string      `json:"area"`
	Address   string      `json:"address"`
	// Config / Dashboard 结构化返回（JSONB → 强类型）。旧数据空字符串时为 nil。
	Config    *TenantConfig `json:"config,omitempty"`
	Dashboard *Dashboard    `json:"dashboard,omitempty"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// parseJSONConfig 把 JSONB 字符串解到 TenantConfig；空字符串返回 nil。
func parseJSONConfig(s string) *TenantConfig {
	if s == "" {
		return nil
	}
	var c TenantConfig
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		return nil
	}
	return &c
}

// parseJSONDashboard 把 JSONB 字符串解到 Dashboard。
func parseJSONDashboard(s string) *Dashboard {
	if s == "" {
		return nil
	}
	var d Dashboard
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return nil
	}
	return &d
}

func toResp(t *Tenant) TenantResp {
	return TenantResp{
		ID:        t.ID,
		Code:      t.Code,
		Name:      t.Name,
		Status:    t.Status,
		Contact:   t.Contact,
		Phone:     t.Phone,
		Email:     t.Email,
		Province:  t.Province,
		City:      t.City,
		Area:      t.Area,
		Address:   t.Address,
		Config:    parseJSONConfig(t.Config),
		Dashboard: parseJSONDashboard(t.Dashboard),
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
