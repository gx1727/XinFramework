package user

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"index" json:"tenant_id"`
	Username  string    `gorm:"size:64;uniqueIndex" json:"username"`
	Password  string    `gorm:"size:255" json:"-"`
	Email     string    `gorm:"size:128" json:"email"`
	Status    int8      `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"index" json:"tenant_id"`
	Name      string    `gorm:"size:64" json:"name"`
	Code      string    `gorm:"size:64" json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	Code      string    `gorm:"size:128" json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type Tenant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128" json:"name"`
	Code      string    `gorm:"size:64;uniqueIndex" json:"code"`
	Plan      string    `gorm:"size:32;default:free" json:"plan"`
	Status    int8      `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
