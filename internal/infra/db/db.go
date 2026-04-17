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
