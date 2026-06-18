package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/dispute-resolve/common/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB(cfg *config.DatabaseConfig) *gorm.DB {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

		var err error
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			PrepareStmt: true,
		})
		if err != nil {
			panic(fmt.Errorf("connect database failed: %v", err))
		}

		sqlDB, err := db.DB()
		if err != nil {
			panic(fmt.Errorf("get sql db failed: %v", err))
		}

		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
		sqlDB.SetMaxOpenConns(cfg.MaxOpen)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(time.Minute * 30)
	})
	return db
}

func GetDB() *gorm.DB {
	return db
}

func AutoMigrate(models ...interface{}) error {
	return db.AutoMigrate(models...)
}

func BeginTx() *gorm.DB {
	return db.Begin()
}
