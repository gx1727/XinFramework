package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg interface{}) error {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.(interface{ DSN() string }).DSN()), &gorm.Config{})
	return err
}

func Get() *gorm.DB {
	return DB
}

func SetTenantID(tenantID uint) {
	if DB != nil && tenantID > 0 {
		DB.Exec("SET app.tenant_id = ?", tenantID)
	}
}

func ClearTenantID() {
	if DB != nil {
		DB.Exec("RESET app.tenant_id")
	}
}
