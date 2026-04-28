package model

import (
	"context"
	"time"
)

// ============ User Repository ============

// User represents a user entity
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

// UserRepository defines data access operations for users
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByAccountID(ctx context.Context, accountID uint) (*User, error)
	GetByCode(ctx context.Context, code string) (*User, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]User, int64, error)
	Create(ctx context.Context, tenantID, accountID uint, code string) (*User, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	UpdatePhone(ctx context.Context, userID uint, phone string) error
	UpdateProfile(ctx context.Context, id uint, nickname, avatar string) error
	Delete(ctx context.Context, id uint) error
}

// ============ Role Repository ============

// Role represents a role entity
type Role struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	OrgID       uint      `json:"org_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DataScope   int8      `json:"data_scope"`
	Extend      string    `json:"extend"`
	IsDefault   bool      `json:"is_default"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleRepository defines data access operations for roles
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Role, error)
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]Role, int64, error)
	Create(ctx context.Context, tenantID uint, req CreateRoleRepoReq) (*Role, error)
	Update(ctx context.Context, id uint, req UpdateRoleRepoReq) (*Role, error)
	Delete(ctx context.Context, id uint) error
}

// CreateRoleRepoReq fields for role creation
type CreateRoleRepoReq struct {
	Code        string
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
}

// UpdateRoleRepoReq fields for role update
type UpdateRoleRepoReq struct {
	Name        string
	Description string
	DataScope   int8
	IsDefault   bool
	Sort        int
	Status      int8
}

// ============ Account Repository ============

// Account represents a global account (cross-tenant)
type Account struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	RealName  string    `json:"real_name"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountRepository defines data access operations for accounts
type AccountRepository interface {
	GetByID(ctx context.Context, id uint) (*Account, error)
	GetByUsername(ctx context.Context, username string) (*Account, error)
	GetByPhone(ctx context.Context, phone string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	Create(ctx context.Context, username, phone, email, realName, passwordHash string) (*Account, error)
	Exists(ctx context.Context, account string) (bool, error)
}

// ============ Account Auth Repository ============

// AccountAuth represents a third-party authentication binding
type AccountAuth struct {
	ID         uint      `json:"id"`
	TenantID   uint      `json:"tenant_id"`
	AccountID  uint      `json:"account_id"`
	Type       string    `json:"type"` // wechat, qq, weibo, wxxcx
	OpenID     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AccountAuthRepository defines data access operations for account auths
type AccountAuthRepository interface {
	GetByOpenID(ctx context.Context, tenantID uint, authType, openID string) (*AccountAuth, error)
	GetByAccountID(ctx context.Context, accountID uint) ([]AccountAuth, error)
	Create(ctx context.Context, tenantID, accountID uint, authType, openID, unionID, sessionKey string) (*AccountAuth, error)
	UpdateSessionKey(ctx context.Context, id uint, sessionKey string) error
	Delete(ctx context.Context, id uint) error
}

// ============ Tenant Repository ============

// Tenant represents a tenant entity
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
	CreatedBy uint      `json:"created_by"`
	UpdatedBy uint      `json:"updated_by"`
	IsDeleted bool      `json:"is_deleted"`
}

// TenantRepository defines data access operations for tenants
type TenantRepository interface {
	GetByID(ctx context.Context, id uint) (*Tenant, error)
	GetByCode(ctx context.Context, code string) (*Tenant, error)
	List(ctx context.Context, keyword string, status *int16, page, size int) ([]Tenant, int64, error)
	Create(ctx context.Context, code, name, contact, phone, email string) (*Tenant, error)
	Update(ctx context.Context, id uint, name, contact, phone, email, province, city, area, address string) (*Tenant, error)
	Delete(ctx context.Context, id uint) error
}

// ============ Organization Repository ============

// Organization represents an organization entity
type Organization struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AdminCode   string    `json:"admin_code"`
	ParentID    uint      `json:"parent_id"`
	Ancestors   string    `json:"ancestors"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrganizationRepository defines data access operations for organizations
type OrganizationRepository interface {
	GetByID(ctx context.Context, id uint) (*Organization, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Organization, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Organization, error)
	GetChildren(ctx context.Context, parentID uint) ([]Organization, error)
	GetTree(ctx context.Context, tenantID uint) ([]Organization, error)
	Create(ctx context.Context, tenantID uint, req CreateOrgRepoReq) (*Organization, error)
	Update(ctx context.Context, id uint, req UpdateOrgRepoReq) (*Organization, error)
	Delete(ctx context.Context, id uint) error
}

// CreateOrgRepoReq fields for organization creation
type CreateOrgRepoReq struct {
	Code        string
	Name        string
	Type        string
	Description string
	AdminCode   string
	ParentID    uint
	Ancestors   string
	Sort        int
	Status      int8
}

// UpdateOrgRepoReq fields for organization update
type UpdateOrgRepoReq struct {
	Name        string
	Type        string
	Description string
	AdminCode   string
	Sort        int
	Status      int8
}

// ============ Menu Repository ============

// Menu represents a menu entity
type Menu struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Subtitle  string    `json:"subtitle"`
	URL       string    `json:"url"`
	Path      string    `json:"path"`
	Icon      string    `json:"icon"`
	Sort      int       `json:"sort"`
	ParentID  uint      `json:"parent_id"`
	Ancestors string    `json:"ancestors"`
	Visible   bool      `json:"visible"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MenuRepository defines data access operations for menus
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*Menu, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Menu, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Menu, error)
	GetUserMenus(ctx context.Context, tenantID, userID uint) ([]Menu, error)
	Create(ctx context.Context, tenantID uint, req CreateMenuRepoReq) (*Menu, error)
	Update(ctx context.Context, id uint, req UpdateMenuRepoReq) (*Menu, error)
	Delete(ctx context.Context, id uint) error
}

// CreateMenuRepoReq fields for menu creation
type CreateMenuRepoReq struct {
	Code      string
	Name      string
	Subtitle  string
	URL       string
	Path      string
	Icon      string
	Sort      int
	ParentID  uint
	Ancestors string
	Visible   bool
	Enabled   bool
}

// UpdateMenuRepoReq fields for menu update
type UpdateMenuRepoReq struct {
	Code     string
	Name     string
	Subtitle string
	URL      string
	Path     string
	Icon     string
	Sort     int
	Visible  bool
	Enabled  bool
}

// ============ Resource Repository ============

// Resource represents a resource/permission entity
type Resource struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	MenuID      uint      `json:"menu_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ResourceRepository defines data access operations for resources
type ResourceRepository interface {
	GetByID(ctx context.Context, id uint) (*Resource, error)
	GetByCode(ctx context.Context, tenantID uint, code string) (*Resource, error)
	GetByTenant(ctx context.Context, tenantID uint) ([]Resource, error)
	GetByMenu(ctx context.Context, menuID uint) ([]Resource, error)
	GetUserResources(ctx context.Context, tenantID, userID uint) ([]Resource, error)
	Create(ctx context.Context, tenantID uint, req CreateResourceRepoReq) (*Resource, error)
	Update(ctx context.Context, id uint, req UpdateResourceRepoReq) (*Resource, error)
	Delete(ctx context.Context, id uint) error
}

// CreateResourceRepoReq fields for resource creation
type CreateResourceRepoReq struct {
	MenuID      uint
	Code        string
	Name        string
	Action      string
	Description string
	Sort        int
	Status      int8
}

// UpdateResourceRepoReq fields for resource update
type UpdateResourceRepoReq struct {
	Name        string
	Action      string
	Description string
	Sort        int
	Status      int8
}

// ============ CmsPost Repository ============

// CmsPost represents a CMS article entity
type CmsPost struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    int16     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// CmsPostRepository defines data access operations for CMS posts
type CmsPostRepository interface {
	GetByID(ctx context.Context, id uint) (*CmsPost, error)
	List(ctx context.Context, tenantID uint, keyword string, status *int16, page, size int) ([]CmsPost, int64, error)
	Create(ctx context.Context, tenantID uint, title, content string, status int16) (*CmsPost, error)
	Update(ctx context.Context, id uint, title, content string, status int16) error
	Delete(ctx context.Context, id uint) error
}

// ============ Attachment Repository ============

// Attachment represents a file uploaded by a user
type Attachment struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	UserID    uint      `json:"user_id"`
	FileName  string    `json:"file_name"`
	FileExt   string    `json:"file_ext"`
	MimeType  string    `json:"mime_type"`
	FileSize  int64     `json:"file_size"`
	Storage   string    `json:"storage"`
	ObjectKey string    `json:"object_key"`
	URL       string    `json:"url"`
	Hash      string    `json:"hash"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// AttachmentRepository defines data access operations for attachments
type AttachmentRepository interface {
	GetByID(ctx context.Context, id uint) (*Attachment, error)
	GetByHash(ctx context.Context, tenantID uint, hash string) (*Attachment, error)
	Create(ctx context.Context, attachment *Attachment) (*Attachment, error)
	UpdateStatus(ctx context.Context, id uint, status int8) error
	Delete(ctx context.Context, id uint) error
}
