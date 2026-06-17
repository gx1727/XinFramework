package ext_impl

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	pkgtenant "gx1727.com/xin/framework/pkg/tenant"
	"gx1727.com/xin/framework/pkg/db"
	"gx1727.com/xin/framework/pkg/plugin"
)

// userRepoRef is the minimal user repository shape extapi needs.
// We define it locally to avoid pulling in framework/internal/module/user
// (which would create a cycle through framework/pkg/extapi).
type userRepoRef interface {
	GetByID(ctx context.Context, id uint) (*userRecord, error)
	List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]userRecord, int64, error)
}

// userRecord is the slim row representation extapi reads from users.
type userRecord struct {
	ID        uint
	TenantID  uint
	AccountID uint
	Code      string
	Nickname  string
	Status    int16
	RealName  string
	Avatar    string
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// userRepoAdapter reads users table directly. Phase 2 stop-gap:
// user.UserRepository still lives in framework/internal — importing
// it here would cycle. Phase 3 (user → apps/rbac/user/) removes
// this adapter entirely.
type userRepoAdapter struct{ db *pgxpool.Pool }

func newUserRepoAdapter(pool *pgxpool.Pool) userRepoRef {
	return &userRepoAdapter{db: pool}
}

func (a *userRepoAdapter) GetByID(ctx context.Context, id uint) (*userRecord, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, err
	}
	var u userRecord
	err = q.QueryRow(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status,
		       real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_deleted = FALSE
		LIMIT 1`, id).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
		&u.RealName, &u.Avatar, &u.Phone, &u.Email, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (a *userRepoAdapter) List(ctx context.Context, tenantID uint, keyword string, page, size int) ([]userRecord, int64, error) {
	q, err := db.GetQuerier(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	var total int64
	err = q.QueryRow(ctx, `
		SELECT COUNT(*) FROM users
		WHERE tenant_id = $1 AND is_deleted = FALSE
		  AND ($2 = '' OR code ILIKE '%' || $2 || '%' OR real_name ILIKE '%' || $2 || '%')`,
		tenantID, keyword).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.Query(ctx, `
		SELECT id, tenant_id, account_id, code, nickname, status,
		       real_name, avatar, phone, email, created_at, updated_at
		FROM users
		WHERE tenant_id = $1 AND is_deleted = FALSE
		  AND ($2 = '' OR code ILIKE '%' || $2 || '%' OR real_name ILIKE '%' || $2 || '%')
		ORDER BY id DESC
		LIMIT $3 OFFSET $4`,
		tenantID, keyword, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []userRecord
	for rows.Next() {
		var u userRecord
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.AccountID, &u.Code, &u.Nickname, &u.Status,
			&u.RealName, &u.Avatar, &u.Phone, &u.Email, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

// tenantRecord mirrors pkg/tenant.TenantRecord — duplicated locally
// because Go's structural typing only works on methods, not on data
// fields. apps/boot/tenant's factory returns pkg/tenant.TenantRecord;
// ext_impl converts to local tenantRecord for facade construction.
type tenantRecord struct {
	ID        uint
	Code      string
	Name      string
	Status    int16
	Contact   string
	Phone     string
	Email     string
	Province  string
	City      string
	Area      string
	Address   string
	Config    string
	Dashboard string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// pkgTenantGet returns a TenantRepoRef backed by the AppContext's
// TenantRepository. Returns nil if apps/boot/tenant is not loaded.
//
// Phase 3: replaced the previous pkgtenant.Get() global lookup with
// a closed-over ctx. The defaultProvider in provider.go populates
// pkgTenantGetCtx during boot.
var pkgTenantGetCtx plugin.Reader

func setTenantCtx(ctx plugin.Reader) { pkgTenantGetCtx = ctx }

func pkgTenantGet() TenantRepoRef {
	if pkgTenantGetCtx == nil {
		return nil
	}
	repo := pkgTenantGetCtx.TenantRepo()
	if repo == nil {
		return nil
	}
	return &tenantCtxAdapter{repo: repo}
}

// tenantCtxAdapter bridges pkg/tenant.TenantRepository to the local
// TenantRepoRef (same GetByID shape, different return type). Phase 6
// deletes this when ext_impl/registry.go is retired.
type tenantCtxAdapter struct {
	repo pkgtenant.TenantRepository
}

func (a *tenantCtxAdapter) GetByID(ctx context.Context, id uint) (*tenantRecord, error) {
	t, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	createdAt, _ := t.CreatedAt.(time.Time)
	updatedAt, _ := t.UpdatedAt.(time.Time)
	return &tenantRecord{
		ID: t.ID, Code: t.Code, Name: t.Name, Status: t.Status,
		Contact: t.Contact, Phone: t.Phone, Email: t.Email,
		Province: t.Province, City: t.City, Area: t.Area, Address: t.Address,
		Config: t.Config, Dashboard: t.Dashboard,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

// TenantRepoRef is the local minimal interface ext_impl uses.
type TenantRepoRef interface {
	GetByID(ctx context.Context, id uint) (*tenantRecord, error)
}

// tenantPkgAdapter wraps a pkg/tenant.TenantRepository so the local
// facade can consume it without depending on the apps/ side.
type tenantPkgAdapter struct {
	factory func() pkgtenant.TenantRepository
}

func (a *tenantPkgAdapter) GetByID(ctx context.Context, id uint) (*tenantRecord, error) {
	repo := a.factory()
	if repo == nil {
		return nil, errTenantNotLoaded
	}
	t, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Convert pkg/tenant.TenantRecord → local tenantRecord. The two
	// structs have identical field sets; we coerce timestamps via
	// the time.Time interface.
	createdAt, _ := t.CreatedAt.(time.Time)
	updatedAt, _ := t.UpdatedAt.(time.Time)
	return &tenantRecord{
		ID: t.ID, Code: t.Code, Name: t.Name, Status: t.Status,
		Contact: t.Contact, Phone: t.Phone, Email: t.Email,
		Province: t.Province, City: t.City, Area: t.Area, Address: t.Address,
		Config: t.Config, Dashboard: t.Dashboard,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

var errTenantNotLoaded = stringErrX("tenant module not loaded — register apps/boot/tenant in main.go")

type stringErrX string

func (e stringErrX) Error() string { return string(e) }