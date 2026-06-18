# Phase 0 - Global Variable Inventory

> Auto-generated. Re-run: `powershell scripts/phase0_scan.ps1`

- Repo root: `d:/work/xin/XinFramework/server`
- Tracked globals: **16**
- Total usages: **442**

## 1. Cross-module globals (Phase B must remove)

| Variable | Package | Defined at | Writes | Reads |
|---|---|---|---:|---:|
| `globalAccountFactory` | `framework/pkg/auth` | (not found):0 | 0 | 0 |
| `globalAccountAuthFactory` | `framework/pkg/auth` | (not found):0 | 0 | 0 |
| `globalFactory` | `framework/pkg/tenant` | (not found):0 | 0 | 0 |
| `globalUserFactory` | `framework/pkg/rbac` | (not found):0 | 0 | 0 |
| `globalRoleFactory` | `framework/pkg/rbac` | (not found):0 | 0 | 0 |
| `globalOrganizationFactory` | `framework/pkg/rbac` | (not found):0 | 0 | 0 |
| `globalPermissionFactory` | `framework/pkg/rbac` | (not found):0 | 0 | 0 |
| `globalProvider` | `framework/pkg/extapi` | framework/pkg/extapi/provider.go:3 | 0 | 4 |
| `global` | `framework/pkg/authz` | (not found):0 | 0 | 5 |
| `globalAuthorizationService` | `framework/internal/service` | (not found):0 | 0 | 0 |
| `globalApp` | `framework/internal/core/boot` | (not found):0 | 0 | 0 |
| `globalCache` | `framework/pkg/dict` | (not found):0 | 2 | 30 |

## 2. Infrastructure globals (keep, surface through AppContext reader)

| Variable | Package | Defined at | Reads |
|---|---|---|---:|
| `Pool` | `framework/pkg/db` | framework/pkg/db/db.go:15 | 166 |
| `Client` | `framework/pkg/cache` | framework/pkg/cache/cache.go:12 | 28 |
| `cfg` | `framework/pkg/config` | framework/pkg/config/config.go:108 | 164 |
| `defaultManager` | `framework/pkg/session` | framework/pkg/session/session.go:29 | 12 |

## 3. Detailed call sites

### `globalAccountFactory`

- Package: `framework/pkg/auth`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalAccountAuthFactory`

- Package: `framework/pkg/auth`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalFactory`

- Package: `framework/pkg/tenant`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalUserFactory`

- Package: `framework/pkg/rbac`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalRoleFactory`

- Package: `framework/pkg/rbac`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalOrganizationFactory`

- Package: `framework/pkg/rbac`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalPermissionFactory`

- Package: `framework/pkg/rbac`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalProvider`

- Package: `framework/pkg/extapi`
- Definition: `framework/pkg/extapi/provider.go:3`
- Usages: 4 total (write 0 / read 4)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\framework\pkg\extapi\provider.go` | 3 | `var globalProvider Provider` |
| `D:\work\xin\XinFramework\server\framework\pkg\extapi\provider.go` | 6 | `if globalProvider == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\extapi\provider.go` | 9 | `return globalProvider` |
| `D:\work\xin\XinFramework\server\framework\pkg\extapi\provider.go` | 13 | `globalProvider = p` |

### `global`

- Package: `framework/pkg/authz`
- Definition: `(not found):0`
- Usages: 5 total (write 0 / read 5)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\apps\rbac\menu\service.go` | 181 | `// Phase 2: break all cycles using DFS with a global visited set` |
| `D:\work\xin\XinFramework\server\framework\pkg\auth\types.go` | 20 | `// Account is the global (cross-tenant) account record. Same struct` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\types.go` | 51 | `// Check global wildcard (super admin)` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\plugin.go` | 10 | `// The module is appended to the global list.` |
| `D:\work\xin\XinFramework\server\framework\pkg\tenant\tenant.go` | 6 | `// Phase 3 cleanup: the historical Register/Get global variables` |

### `globalAuthorizationService`

- Package: `framework/internal/service`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalApp`

- Package: `framework/internal/core/boot`
- Definition: `(not found):0`
- Usages: 0 total (write 0 / read 0)

| File | Line | Snippet |
|---|---:|---|

### `globalCache`

- Package: `framework/pkg/dict`
- Definition: `(not found):0`
- Usages: 32 total (write 2 / read 30)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 37 | `globalCache = &Cache{` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 54 | `globalCache.mu.Lock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 55 | `defer globalCache.mu.Unlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 57 | `if globalCache.data[tenantID] == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 58 | `globalCache.data[tenantID] = make(map[string]*Dict)` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 60 | `globalCache.data[tenantID] = make(map[string]*Dict)` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 96 | `merged := globalCache.data[tenantID][d.Code]` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 108 | `globalCache.data[tenantID][d.Code] = merged` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 137 | `for _, merged := range globalCache.data[tenantID] {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 163 | `globalCache.mu.RLock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 164 | `defer globalCache.mu.RUnlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 166 | `if globalCache.data[tenantID] == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 169 | `d, ok := globalCache.data[tenantID][dictCode]` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 195 | `globalCache.mu.Lock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 196 | `delete(globalCache.data, tenantID)` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 197 | `globalCache.mu.Unlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 205 | `globalCache.mu.Lock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 206 | `if globalCache.data[tenantID] != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 207 | `delete(globalCache.data[tenantID], dictCode)` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 209 | `globalCache.mu.Unlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 283 | `globalCache.mu.Lock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 284 | `if globalCache.data[tenantID] == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 285 | `globalCache.data[tenantID] = make(map[string]*Dict)` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 287 | `globalCache.data[tenantID][dictCode] = merged` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 288 | `globalCache.mu.Unlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 295 | `globalCache.mu.Lock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 296 | `defer globalCache.mu.Unlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 298 | `if globalCache.data[tenantID] != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 299 | `delete(globalCache.data[tenantID], dictCode)` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 304 | `globalCache.mu.Lock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 305 | `defer globalCache.mu.Unlock()` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 307 | `delete(globalCache.data, tenantID)` |

### `Pool`

- Package: `framework/pkg/db`
- Definition: `framework/pkg/db/db.go:15`
- Usages: 175 total (write 9 / read 166)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\apps\boot\auth\account_auth_repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\account_auth_repository.go` | 18 | `func NewAccountAuthRepository(db *pgxpool.Pool) AccountAuthRepository {` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\account_repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\account_repository.go` | 18 | `func NewAccountRepository(db *pgxpool.Pool) AccountRepository {` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\deps.go` | 24 | `DB *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\deps.go` | 42 | `func DefaultDependencies(cfg *config.Config, db *pgxpool.Pool, repos Repositories) Dependencies {` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 30 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 31 | `w.SetAccountRepo(NewAccountRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 32 | `w.SetAccountAuthRepo(NewAccountAuthRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 36 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 39 | `pool = p` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 43 | `tenantRepo := tenant.NewTenantRepository(pool)` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 56 | `Account: NewAccountRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 58 | `Platform: permission.NewPlatformRoleRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\module.go` | 60 | `deps := DefaultDependencies(config.Get(), pool, repos)` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\service.go` | 31 | `func ResolveLoginIdentity(ctx context.Context, d *pgxpool.Pool, account string, tenantID uint) (*LoginIdentity, error) {` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\service.go` | 131 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\usercode.go` | 68 | `func SetUserCodeFormat(ctx context.Context, db *pgxpool.Pool, tenantID uint, format UserCodeFormat) error {` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\usercode.go` | 79 | `func GetUserCodeFormat(ctx context.Context, db *pgxpool.Pool, tenantID uint) (UserCodeFormat, error) {` |
| `D:\work\xin\XinFramework\server\apps\boot\tenant\module.go` | 27 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\boot\tenant\module.go` | 28 | `w.SetTenantRepo(&tenantPkgAdapter{repo: NewTenantRepository(pool)})` |
| `D:\work\xin\XinFramework\server\apps\boot\tenant\module.go` | 32 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\boot\tenant\module.go` | 33 | `h := NewHandler(NewService(NewTenantRepository(pool)))` |
| `D:\work\xin\XinFramework\server\apps\boot\tenant\repository.go` | 16 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\boot\tenant\repository.go` | 19 | `func NewTenantRepository(db *pgxpool.Pool) TenantRepository {` |
| `D:\work\xin\XinFramework\server\apps\cms\module.go` | 15 | `// construction to reading its DB pool off the AppContext.Reader.` |
| `D:\work\xin\XinFramework\server\apps\cms\repository.go` | 13 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\cms\repository.go` | 16 | `func NewCmsPostRepository(db *pgxpool.Pool) CmsPostRepository {` |
| `D:\work\xin\XinFramework\server\apps\flag\avatar_category_repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\flag\avatar_category_repository.go` | 18 | `func NewAvatarCategoryRepository(pool *pgxpool.Pool) *AvatarCategoryRepository {` |
| `D:\work\xin\XinFramework\server\apps\flag\avatar_category_repository.go` | 19 | `return &AvatarCategoryRepository{db: pool}` |
| `D:\work\xin\XinFramework\server\apps\flag\avatar_repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\flag\avatar_repository.go` | 18 | `func NewAvatarRepository(pool *pgxpool.Pool) *AvatarRepository {` |
| `D:\work\xin\XinFramework\server\apps\flag\avatar_repository.go` | 19 | `return &AvatarRepository{db: pool}` |
| `D:\work\xin\XinFramework\server\apps\flag\frame_category_repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\flag\frame_category_repository.go` | 18 | `func NewFrameCategoryRepository(pool *pgxpool.Pool) *FrameCategoryRepository {` |
| `D:\work\xin\XinFramework\server\apps\flag\frame_category_repository.go` | 19 | `return &FrameCategoryRepository{db: pool}` |
| `D:\work\xin\XinFramework\server\apps\flag\frame_repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\flag\frame_repository.go` | 18 | `func NewFrameRepository(pool *pgxpool.Pool) *FrameRepository {` |
| `D:\work\xin\XinFramework\server\apps\flag\frame_repository.go` | 19 | `return &FrameRepository{db: pool}` |
| `D:\work\xin\XinFramework\server\apps\flag\init.go` | 16 | `func InitRepositories(pool *pgxpool.Pool) {` |
| `D:\work\xin\XinFramework\server\apps\flag\init.go` | 17 | `frameRepo = NewFrameRepository(pool)` |
| `D:\work\xin\XinFramework\server\apps\flag\init.go` | 18 | `avatarRepo = NewAvatarRepository(pool)` |
| `D:\work\xin\XinFramework\server\apps\flag\init.go` | 19 | `frameCatRepo = NewFrameCategoryRepository(pool)` |
| `D:\work\xin\XinFramework\server\apps\flag\init.go` | 20 | `avatarCatRepo = NewAvatarCategoryRepository(pool)` |
| `D:\work\xin\XinFramework\server\apps\flag\module.go` | 21 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\flag\module.go` | 24 | `pool = p` |
| `D:\work\xin\XinFramework\server\apps\flag\module.go` | 27 | `InitRepositories(pool)` |
| `D:\work\xin\XinFramework\server\apps\rbac\menu\repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\menu\repository.go` | 18 | `func NewMenuRepository(db *pgxpool.Pool) MenuRepository {` |
| `D:\work\xin\XinFramework\server\apps\rbac\organization\module.go` | 21 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\organization\module.go` | 22 | `w.SetOrgRepo(NewOrganizationRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\rbac\organization\module.go` | 26 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\organization\module.go` | 27 | `h := NewHandler(NewService(NewOrganizationRepository(pool)))` |
| `D:\work\xin\XinFramework\server\apps\rbac\organization\repository.go` | 17 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\organization\repository.go` | 20 | `func NewOrganizationRepository(db *pgxpool.Pool) OrganizationRepository {` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\module.go` | 26 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\module.go` | 27 | `w.SetPermRepo(NewRoleResourceRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\module.go` | 31 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\module.go` | 32 | `roleResourceRepo := NewRoleResourceRepository(pool)` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\module.go` | 34 | `h := NewHandler(NewService(pool, roleResourceRepo, authzSvc))` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\role_resource_repository.go` | 20 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\role_resource_repository.go` | 23 | `func NewRoleResourceRepository(db *pgxpool.Pool) RoleResourceRepository {` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\service.go` | 16 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\permission\service.go` | 21 | `func NewService(db *pgxpool.Pool, roleResourceRepo RoleResourceRepository, authzSvc authz.Authorization) *Service {` |
| `D:\work\xin\XinFramework\server\apps\rbac\resource\module.go` | 25 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\resource\module.go` | 27 | `h := NewHandler(NewService(NewResourceRepository(pool), authzSvc))` |
| `D:\work\xin\XinFramework\server\apps\rbac\resource\repository.go` | 15 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\resource\repository.go` | 18 | `func NewResourceRepository(db *pgxpool.Pool) ResourceRepository {` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\module.go` | 22 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\module.go` | 23 | `w.SetRoleRepo(NewRoleRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\module.go` | 27 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\module.go` | 30 | `NewRoleRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\module.go` | 31 | `permission.NewDataScopeRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\module.go` | 32 | `NewRoleMenuRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\repository.go` | 17 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\repository.go` | 20 | `func NewRoleRepository(db *pgxpool.Pool) RoleRepository {` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\role_menu_repository.go` | 20 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\role\role_menu_repository.go` | 23 | `func NewRoleMenuRepository(db *pgxpool.Pool) RoleMenuRepository {` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 36 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 37 | `w.SetUserRepo(NewUserRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 41 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 44 | `pool = p` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 68 | `assetSvc := asset.NewFileService(s, asset.NewAttachmentRepository(pool))` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 79 | `NewUserRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 80 | `role.NewRoleRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 81 | `organization.NewOrganizationRepository(pool),` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\repository.go` | 18 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\repository.go` | 21 | `func NewUserRepository(db *pgxpool.Pool) UserRepository {` |
| `D:\work\xin\XinFramework\server\apps\reference\asset\repository.go` | 14 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\reference\asset\repository.go` | 17 | `func NewAttachmentRepository(db *pgxpool.Pool) *PostgresAttachmentRepository {` |
| `D:\work\xin\XinFramework\server\apps\reference\dict\repository.go` | 17 | `pool *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\reference\dict\repository.go` | 20 | `func NewPostgresDictRepository(pool *pgxpool.Pool) *PostgresDictRepository {` |
| `D:\work\xin\XinFramework\server\apps\reference\dict\repository.go` | 21 | `return &PostgresDictRepository{pool: pool}` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 30 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 46 | `db *pgxpool.Pool,` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 15 | `pool, err := pgxpool.New(context.Background(), dsn)` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 17 | `fmt.Println("pool:", err)` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 20 | `defer pool.Close()` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 23 | `pool.QueryRow(ctx, "SELECT current_user").Scan(&u)` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 26 | `repo := dict.NewPostgresDictRepository(pool)` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 33 | `err := db.RunInTenantTx(ctx, pool, tid, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 67 | `err = db.RunInTenantTx(ctx, pool, 1, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 92 | `err = db.RunInTenantTx(ctx, pool, 2, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\cmd\dict_e2e\main.go` | 95 | `err := db.RunInTenantTx(ctx, pool, 0, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 125 | `// Phase 3 will wire this to a real AppContext with the DB pool, cache,` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 25 | `DB *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 50 | `func RunBootstrap(ctx context.Context, pool *pgxpool.Pool, cfg BootstrapConfig) error {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 61 | `accountID, created, err := upsertBootstrapAccount(ctx, pool, cfg)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 74 | `if _, err := pool.Exec(ctx, `` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 84 | `if err := upsertBootstrapUser(ctx, pool, accountID, cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 94 | `func upsertBootstrapUser(ctx context.Context, pool *pgxpool.Pool, accountID uint, cfg BootstrapConfig) error {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 98 | `err := pool.QueryRow(ctx, `` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 116 | `err = db.RunInTenantTx(ctx, pool, tenantID, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 187 | `func upsertBootstrapAccount(ctx context.Context, pool *pgxpool.Pool, cfg BootstrapConfig) (uint, bool, error) {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 193 | `tx, err := pool.Begin(ctx)` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 15 | `var Pool *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 39 | `Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 41 | `return fmt.Errorf("create pool: %w", err)` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 44 | `if err := Pool.Ping(ctx); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 51 | `func Get() *pgxpool.Pool {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 52 | `return Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 56 | `if Pool != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 57 | `Pool.Close()` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 75 | `func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 80 | `if pool == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 81 | `return fmt.Errorf("db pool is not initialized")` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 84 | `tx, err := pool.Begin(ctx)` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 98 | `func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uint, fn func(ctx context.Context) error) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 99 | `return RunInTx(ctx, pool, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 120 | `func RunInPlatformTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 121 | `return RunInTx(ctx, pool, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 138 | `if Pool == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 139 | `return nil, fmt.Errorf("db pool is not initialized")` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 142 | `return Pool, nil` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 40 | `dbPool *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 43 | `func Init(pool *pgxpool.Pool) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 44 | `dbPool = pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\dict\dict.go` | 48 | `func GetPool() *pgxpool.Pool {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 17 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 18 | `if pool == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 22 | `_, _ = pool.Exec(ctx, `` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 31 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 32 | `if pool == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 36 | `_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM _schema_migrations WHERE version = $1)", version).Scan(&exists)` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 42 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 43 | `if pool == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 46 | `_, err := pool.Exec(ctx, "INSERT INTO _schema_migrations (version) VALUES ($1)", version)` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 94 | `pool := db.Get()` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 95 | `if pool == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 126 | `if err := runMigration(ctx, pool, m.SQL); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 142 | `func runMigration(ctx context.Context, pool *pgxpool.Pool, sql string) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\migrate\migrate.go` | 143 | `return db.RunInTx(ctx, pool, func(ctx context.Context) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\permission_impl.go` | 13 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\permission_impl.go` | 16 | `func NewPermissionRepository(db *pgxpool.Pool) *PostgresPermissionRepository {` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\permission_impl.go` | 156 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\permission_impl.go` | 159 | `func NewDataScopeRepository(db *pgxpool.Pool) *PostgresDataScopeRepository {` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\platform_role.go` | 26 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\platform_role.go` | 29 | `func NewPlatformRoleRepository(pool *pgxpool.Pool) *PostgresPlatformRoleRepository {` |
| `D:\work\xin\XinFramework\server\framework\pkg\permission\platform_role.go` | 30 | `return &PostgresPlatformRoleRepository{db: pool}` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 42 | `// - Infrastructure: DB pool, Redis client, Config, Session manager.` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 58 | `DB() *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 107 | `db *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 126 | `// Both the db pool and the config must be non-nil. cache and session` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 130 | `db *pgxpool.Pool,` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 136 | `panic("NewAppContext: db pool must not be nil")` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 157 | `func (a *AppContext) DB() *pgxpool.Pool { return a.db }` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 106 | `pool *pgxpool.Pool` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 112 | `_, _ = m.pool.Exec(ctx, `` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 121 | `_, _ = m.pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions (expires_at)`)` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 136 | `_, err := m.pool.Exec(ctx, `` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 153 | `err := m.pool.QueryRow(ctx, `` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 171 | `_, err := m.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE session_id = $1`, sessionID)` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 181 | `func NewDBSessionManager(pool *pgxpool.Pool) SessionManager {` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 182 | `return &dbSessionManager{pool: pool}` |

### `Client`

- Package: `framework/pkg/cache`
- Definition: `framework/pkg/cache/cache.go:12`
- Usages: 43 total (write 15 / read 28)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 14 | `client := cache.Get()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 15 | `if client == nil {` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 23 | `infoStr, err := client.Info(ctx).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 44 | `dbSize, _ := client.DBSize(ctx).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 47 | `cmdStatsStr, _ := client.Info(ctx, "commandstats").Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 78 | `client := cache.Get()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 79 | `if client == nil {` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 86 | `keys, err := client.Keys(ctx, pattern).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 103 | `client := cache.Get()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 104 | `if client == nil {` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 112 | `keyType, err := client.Type(ctx, key).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 126 | `value, _ = client.Get(ctx, key).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 128 | `value, _ = client.HGetAll(ctx, key).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 130 | `value, _ = client.LRange(ctx, key, 0, -1).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 132 | `value, _ = client.SMembers(ctx, key).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 134 | `value, _ = client.ZRange(ctx, key, 0, -1).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 139 | `ttl, _ := client.TTL(ctx, key).Result()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 157 | `client := cache.Get()` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 158 | `if client == nil {` |
| `D:\work\xin\XinFramework\server\apps\system\cache_handler.go` | 164 | `err := client.Del(ctx, key).Err()` |
| `D:\work\xin\XinFramework\server\apps\system\handler.go` | 73 | `client := cache.Get()` |
| `D:\work\xin\XinFramework\server\apps\system\handler.go` | 74 | `if client == nil {` |
| `D:\work\xin\XinFramework\server\apps\system\handler.go` | 80 | `err := client.FlushDB(context.Background()).Err()` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 12 | `var Client *redis.Client` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 16 | `Client = nil` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 40 | `Client = redis.NewClient(opts)` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 44 | `if err := Client.Ping(ctx).Err(); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 48 | `Client = nil` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 53 | `func Get() *redis.Client {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 54 | `return Client` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 58 | `if Client == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 61 | `err := Client.Close()` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 62 | `Client = nil` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 42 | `// - Infrastructure: DB pool, Redis client, Config, Session manager.` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 59 | `Cache() *redis.Client // may return nil if Redis is disabled` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 108 | `cache *redis.Client` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 131 | `cache *redis.Client,` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 158 | `func (a *AppContext) Cache() *redis.Client { return a.cache }` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 17 | `client *cos.Client` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 35 | `client := cos.NewClient(b, &http.Client{` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 50 | `client: client,` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 56 | `_, err := s.client.Object.Put(ctx, key, file, nil)` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 65 | `_, err := s.client.Object.Delete(ctx, key)` |

### `cfg`

- Package: `framework/pkg/config`
- Definition: `framework/pkg/config/config.go:108`
- Usages: 171 total (write 7 / read 164)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\apps\boot\auth\deps.go` | 42 | `func DefaultDependencies(cfg *config.Config, db *pgxpool.Pool, repos Repositories) Dependencies {` |
| `D:\work\xin\XinFramework\server\apps\boot\auth\deps.go` | 45 | `Config: cfg,` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\config.go` | 19 | `func Cfg() *AuthConfig {` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 49 | `cfg := config.Get()` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 50 | `if cfg.Storage.Provider == "cos" {` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 52 | `URL: cfg.Storage.CosURL,` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 53 | `SecretID: cfg.Storage.CosSecretID,` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 54 | `SecretKey: cfg.Storage.CosSecretKey,` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 55 | `BaseURL: cfg.Storage.CosBaseURL,` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 63 | `cfg.Storage.LocalDir,` |
| `D:\work\xin\XinFramework\server\apps\rbac\user\module.go` | 64 | `cfg.Storage.LocalBaseURL,` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\config.go` | 16 | `func Cfg() *WxxcxConfig {` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 67 | `if Cfg().AppID == "" \|\| Cfg().AppSecret == "" {` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 73 | `Cfg().AppID,` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 74 | `Cfg().AppSecret,` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 122 | `if Cfg().AppID == "" \|\| Cfg().AppSecret == "" {` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 180 | `Cfg().AppID,` |
| `D:\work\xin\XinFramework\server\apps\reference\weixin\service.go` | 181 | `Cfg().AppSecret,` |
| `D:\work\xin\XinFramework\server\cmd\xin\main.go` | 33 | `cfg, err := config.Load("config/config.yaml")` |
| `D:\work\xin\XinFramework\server\cmd\xin\main.go` | 42 | `framework.Run(cfg)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 31 | `func Run(cfg *config.Config) {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 33 | `runServer(cfg)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 51 | `runServer(cfg)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 59 | `func runServer(cfg *config.Config) {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 60 | `app, err := boot.Init(cfg)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 66 | `if err := initModules(cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 77 | `addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 99 | `// 两者合二为一。cfg.Module 控制加载哪些模块；未在配置中启用的模块` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 101 | `func initModules(cfg *config.Config) error {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 102 | `enabled := make(map[string]bool, len(cfg.Module))` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 103 | `for _, name := range cfg.Module {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 115 | `ctx, w := buildAppContextForModule(m.Name(), cfg)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 128 | `func buildAppContextForModule(name string, cfg *config.Config) (plugin.Reader, plugin.Writer) {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 140 | `cfg := app.Config` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 145 | `srv.Engine.Use(middleware.CORS(&cfg.CORS)) // 3. CORS 预检请求处理` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 150 | `registerModules(srv.Engine, cfg, app)` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 154 | `func registerModules(r *gin.Engine, cfg *config.Config, app *boot.App) {` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 157 | `public.Use(middleware.OptionalAuth(&cfg.JWT, app.SessionMgr, app.Authz))` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 160 | `protected.Use(middleware.Auth(&cfg.JWT, app.SessionMgr, app.Authz))` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 162 | `enabled := make(map[string]bool, len(cfg.Module))` |
| `D:\work\xin\XinFramework\server\framework\framework.go` | 163 | `for _, name := range cfg.Module {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 32 | `func Init(cfg *config.Config) (*App, error) {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 33 | `logger.Init(cfg.Log.Dir, cfg.Log.Level)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 34 | `if err := db.Init(&cfg.Database); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 40 | `if err := cache.Init(&cfg.Redis); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 59 | `appCtx := plugin.NewAppContext(db.Get(), cache.Get(), cfg, sm)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 77 | `Config: cfg,` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\boot.go` | 80 | `Server: server.New(cfg),` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 29 | `cfg := BootstrapConfig{` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 37 | `if cfg.Role == "" {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 38 | `cfg.Role = "super_admin"` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 40 | `if cfg.TenantCode == "" {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 41 | `cfg.TenantCode = "default"` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 43 | `cfg.Enabled = cfg.Token != "" && cfg.Account != "" && cfg.Password != ""` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 44 | `return cfg` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 48 | `// 调用条件：cfg.Enabled == true` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 50 | `func RunBootstrap(ctx context.Context, pool *pgxpool.Pool, cfg BootstrapConfig) error {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 51 | `if !cfg.Enabled {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 56 | `if len(cfg.Token) < 16 {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 61 | `accountID, created, err := upsertBootstrapAccount(ctx, pool, cfg)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 66 | `log.Printf("[bootstrap] created bootstrap account %q (id=%d)", cfg.Account, accountID)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 68 | `log.Printf("[bootstrap] bootstrap account %q already exists (id=%d)", cfg.Account, accountID)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 78 | ``, accountID, cfg.Role); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 81 | `log.Printf("[bootstrap] granted platform role %q to account %d", cfg.Role, accountID)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 84 | `if err := upsertBootstrapUser(ctx, pool, accountID, cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 91 | `// upsertBootstrapUser 把 account 绑定到 cfg.TenantCode 租户。` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 94 | `func upsertBootstrapUser(ctx context.Context, pool *pgxpool.Pool, accountID uint, cfg BootstrapConfig) error {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 102 | ``, cfg.TenantCode).Scan(&tenantID, &tenantStatus)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 105 | `return errors.New("XIN_BOOTSTRAP_TENANT_CODE=" + cfg.TenantCode + " 的租户不存在")` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 110 | `return errors.New("租户 " + cfg.TenantCode + " 已禁用")` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 133 | ``, cfg.RealName, userID); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 145 | ``, tenantID, accountID, cfg.Account, cfg.RealName).Scan(&userID); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 179 | `log.Printf("[bootstrap] created user binding account %d -> tenant %q (user_id=%d)", accountID, cfg.TenantCode, userID)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 187 | `func upsertBootstrapAccount(ctx context.Context, pool *pgxpool.Pool, cfg BootstrapConfig) (uint, bool, error) {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 188 | `passwordHash, err := pkgauth.HashPassword(cfg.Password)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 206 | ``, cfg.Account).Scan(&accountID)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 212 | ``, passwordHash, cfg.RealName, accountID); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\boot\bootstrap.go` | 229 | ``, cfg.Account, cfg.RealName, passwordHash).Scan(&accountID)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 25 | `func processAuthToken(c *gin.Context, cfg *config.JWTConfig, sm session.SessionManager) (*jwtpkg.Claims, error) {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 34 | `return []byte(cfg.Secret), nil` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 109 | `func Auth(cfg *config.JWTConfig, sm session.SessionManager, permSvc SecurityContextLoader) gin.HandlerFunc {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 111 | `claims, err := processAuthToken(c, cfg, sm)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 137 | `func AuthLite(cfg *config.JWTConfig, sm session.SessionManager) gin.HandlerFunc {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 139 | `claims, err := processAuthToken(c, cfg, sm)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 180 | `func OptionalAuth(cfg *config.JWTConfig, sm session.SessionManager, permSvc SecurityContextLoader) gin.HandlerFunc {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\auth.go` | 182 | `claims, err := processAuthToken(c, cfg, sm)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 12 | `func CORS(cfg *config.CORSConfig) gin.HandlerFunc {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 13 | `if cfg == nil \|\| !cfg.Enabled \|\| len(cfg.AllowOrigins) == 0 {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 24 | `for _, o := range cfg.AllowOrigins {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 42 | `c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 43 | `c.Header("Access-Control-Allow-Headers", cfg.AllowHeaders)` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 44 | `c.Header("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))` |
| `D:\work\xin\XinFramework\server\framework\internal\core\middleware\cors.go` | 46 | `if cfg.AllowCredentials {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\server\server.go` | 19 | `func New(cfg *config.Config) *XinServer {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\server\server.go` | 20 | `if cfg.App.Env == "prod" {` |
| `D:\work\xin\XinFramework\server\framework\internal\core\server\server.go` | 26 | `engine.Static(cfg.Storage.LocalBaseURL, cfg.Storage.LocalDir)` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 14 | `func Init(cfg *config.RedisConfig) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 15 | `if !cfg.Enabled {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 21 | `Addr: cfg.Addr(),` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 22 | `Password: cfg.Password,` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 23 | `DB: cfg.DB,` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 25 | `if cfg.PoolSize > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 26 | `opts.PoolSize = cfg.PoolSize` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 28 | `if cfg.MinIdleConns > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 29 | `opts.MinIdleConns = cfg.MinIdleConns` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 31 | `if cfg.PoolTimeoutSec > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 32 | `opts.PoolTimeout = time.Duration(cfg.PoolTimeoutSec) * time.Second` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 34 | `if cfg.IdleTimeoutSec > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 35 | `opts.IdleTimeout = time.Duration(cfg.IdleTimeoutSec) * time.Second` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 37 | `if cfg.MaxConnAgeSec > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 38 | `opts.MaxConnAge = time.Duration(cfg.MaxConnAgeSec) * time.Second` |
| `D:\work\xin\XinFramework\server\framework\pkg\cache\cache.go` | 45 | `if cfg.Required {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 108 | `var cfg *Config` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 168 | `cfg = defaults()` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 176 | `if err := yaml.Unmarshal(data, cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 181 | `overrideWithEnv(cfg)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 182 | `if err := validateModules(cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 185 | `if err := validateJWTSecret(cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 189 | `return cfg, nil` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config.go` | 441 | `return cfg` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 11 | `// cfg.Module 列表与本次"显式白名单"预期一致。` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 27 | `cfg, err := Load(cfgPath)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 43 | `got := append([]string{}, cfg.Module...)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 49 | `t.Fatalf("cfg.Module mismatch\n got: %v\n want: %v", got, wantSorted)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 54 | `if cfg.Module[i] != m {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 55 | `t.Errorf("cfg.Module[%d]: got %q, want %q (full: %v)", i, cfg.Module[i], m, cfg.Module)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 89 | `cfg, err := Load(filepath.Join(tmpDir, "config.yaml"))` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 93 | `if cfg.App.Env != "dev" {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 94 | `t.Fatalf("expected env=dev, got %q", cfg.App.Env)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 98 | `cfg.App.Env = "prod"` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 99 | `if err := validateJWTSecret(cfg); err == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 109 | `cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 113 | `if cfg.App.Env != "prod" {` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 114 | `t.Fatalf("expected env=prod, got %q", cfg.App.Env)` |
| `D:\work\xin\XinFramework\server\framework\pkg\config\config_load_test.go` | 116 | `if err := validateJWTSecret(cfg); err != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 17 | `func Init(cfg *config.DatabaseConfig) error {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 21 | `poolConfig, err := pgxpool.ParseConfig(cfg.DSN())` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 26 | `if cfg.MaxOpenConns > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 27 | `poolConfig.MaxConns = int32(cfg.MaxOpenConns)` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 29 | `if cfg.MaxIdleConns > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 30 | `poolConfig.MinConns = int32(cfg.MaxIdleConns)` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 32 | `if cfg.ConnMaxLifetimeSec > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 33 | `poolConfig.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetimeSec) * time.Second` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 35 | `if cfg.ConnMaxIdleTimeSec > 0 {` |
| `D:\work\xin\XinFramework\server\framework\pkg\db\db.go` | 36 | `poolConfig.MaxConnIdleTime = time.Duration(cfg.ConnMaxIdleTimeSec) * time.Second` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 30 | `func Generate(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string) (string, error) {` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 31 | `return GenerateWithType(cfg, userID, tenantID, role, sessionID, TokenTypeAccess)` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 34 | `func GenerateWithType(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, tokenType string) (string, error) {` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 35 | `return GenerateWithPlatformRoles(cfg, userID, tenantID, role, sessionID, nil, tokenType)` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 40 | `func GenerateWithPlatformRoles(cfg *config.JWTConfig, userID, tenantID uint, role, sessionID string, platformRoles []string, tokenType string) (string, error) {` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 41 | `expire := cfg.Expire` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 43 | `expire = cfg.RefreshExpire` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 59 | `return token.SignedString([]byte(cfg.Secret))` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 75 | `func Validate(tokenString string, cfg *config.JWTConfig) (*Claims, error) {` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 77 | `return []byte(cfg.Secret), nil` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 91 | `func ValidateRefresh(tokenString string, cfg *config.JWTConfig) (*Claims, error) {` |
| `D:\work\xin\XinFramework\server\framework\pkg\jwt\jwt.go` | 92 | `claims, err := Validate(tokenString, cfg)` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 47 | `// corresponding module was not enabled in cfg.Module. Modules must` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 67 | `// when the producing module was not enabled in cfg.Module.` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 109 | `cfg *config.Config` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 132 | `cfg *config.Config,` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 138 | `if cfg == nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 144 | `cfg: cfg,` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\appcontext.go` | 159 | `func (a *AppContext) Config() *config.Config { return a.cfg }` |
| `D:\work\xin\XinFramework\server\framework\pkg\plugin\plugin.go` | 14 | `// name is in cfg.Module. The module MUST` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 28 | `func NewCosStorage(cfg Config) (storage.Storage, error) {` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 29 | `u, err := url.Parse(cfg.URL)` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 38 | `SecretID: cfg.SecretID,` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 39 | `SecretKey: cfg.SecretKey,` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 43 | `baseURL := cfg.BaseURL` |
| `D:\work\xin\XinFramework\server\framework\pkg\storage\cos\cos.go` | 45 | `baseURL = cfg.URL` |

### `defaultManager`

- Package: `framework/pkg/session`
- Definition: `framework/pkg/session/session.go:29`
- Usages: 12 total (write 0 / read 12)

| File | Line | Snippet |
|---|---:|---|
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 29 | `var defaultManager SessionManager` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 32 | `defaultManager = manager` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 36 | `return defaultManager` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 185 | `// Create delegates to defaultManager if available, otherwise returns error` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 187 | `if defaultManager != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 188 | `return defaultManager.Create(sessionID, userID, tenantID, role, ttl)` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 193 | `// Validate delegates to defaultManager if available, otherwise returns error` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 195 | `if defaultManager != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 196 | `return defaultManager.Validate(sessionID)` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 201 | `// Revoke delegates to defaultManager if available, otherwise returns error` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 203 | `if defaultManager != nil {` |
| `D:\work\xin\XinFramework\server\framework\pkg\session\session.go` | 204 | `return defaultManager.Revoke(sessionID)` |

