package tenant

import "time"

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

func (t *Tenant) ToResp() TenantResp {
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
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
