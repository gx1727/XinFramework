package tenants

import "encoding/json"

// TenantConfig 租户纀JSONB 配置契约。前竀DynamicForm 挀Schema 渲染，避免黑盒　
// 新增字段请同步更新前竀schema，保持向后兼容（旧字段允许缺失）　
type TenantConfig struct {
	// LogoURL 租户 logo（CDN/cos 路径：
	LogoURL string `json:"logo_url,omitempty"`
	// Theme light | dark | auto
	Theme string `json:"theme,omitempty"`
	// Locale zh-CN | en-US
	Locale string `json:"locale,omitempty"`
	// ModuleFlags 模块启用位图（true=启用）。前端按位渲染菜单与导航　
	ModuleFlags map[string]bool `json:"module_flags,omitempty"`
	// SubscriptionOverrides 单租户覆盖默认套餐限额（妀max_users=1000：
	SubscriptionOverrides map[string]int `json:"subscription_overrides,omitempty"`
}

// Dashboard 租户仪表盘布局：左中右三栏皀widget key 列表　
// key 对应 UI/src/components/widgets/<key>.tsx　
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
	// AdminAccountID 首装 admin 用户绑定的账叀ID（accounts.id）　
	// 必填：必须先圀accounts 表创建账号，再创建租户并绑定　
	// 留空时跳迀admin user 创建，只建空租户（用于纯数据迁移场景）　
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
	Status *int16 `json:"status" binding:"required,oneof=0 1"`
}

type ListTenantReq struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Size    int    `form:"size" binding:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword"`
	Status  *int16 `form:"status"`
}

// ImpersonateResp 模拟登录响应。前端收到后应保存原 platform refresh_token（用于"退出模拟"），
// 并将 token/refresh_token 替换为模拟 token，跳转到租户域首页。
type ImpersonateResp struct {
	Scope         LoginScopeString `json:"scope"`
	Token         string           `json:"token"`
	RefreshToken  string           `json:"refresh_token"`
	ExpiresIn     int              `json:"expires_in"`
	TenantID      uint             `json:"tenant_id"`
	TenantName    string           `json:"tenant_name"`
	ImpersonatedUserID uint        `json:"impersonated_user_id"`
	// ImpersonatedBy 原 super_admin 的 account_id（用于审计展示）
	ImpersonatedBy uint   `json:"impersonated_by"`
	// ImpersonationSID 原 platform 会话 ID；前端"退出模拟"时调 /auth/refresh 即可恢复
	ImpersonationSID string `json:"impersonation_sid"`
}

// LoginScopeString 复用 auth.LoginScope 的字面值，避免反向依赖 auth 包
type LoginScopeString = string

const (
	ImpersonateScopeTenant LoginScopeString = "tenant"
)

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
	// Config / Dashboard 结构化返回（JSONB ↀ强类型）。旧数据空字符串时为 nil　
	Config    *TenantConfig `json:"config,omitempty"`
	Dashboard *Dashboard    `json:"dashboard,omitempty"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// parseJSONConfig 技JSONB 字符串解刀TenantConfig；空字符串返囀nil　
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

// parseJSONDashboard 技JSONB 字符串解刀Dashboard　
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
