package db

import (
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"fullstack-cms/config"
)

func Connect(cfg config.Config, log *zap.Logger) *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = cfg.Connection.Postgresql.DSN
	}

	var gdb *gorm.DB
	var err error
	for i := 1; i <= 30; i++ {
		gdb, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: false, // pakai tabel jamak: users, roles
			},
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			sdb, e2 := gdb.DB()
			if e2 == nil {
				err = sdb.Ping()
			}
		}
		if err == nil {
			break
		}
		log.Warn("db.wait", zap.Int("attempt", i), zap.Error(err))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("db.connect.fail", zap.Error(err))
	}
	db := gdb

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("db.sql.fail", zap.Error(err))
	}

	sqlDB.SetMaxOpenConns(cfg.Connection.Postgresql.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.Connection.Postgresql.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Connection.Postgresql.MaxLifetimeConnections) * time.Second)

	return db
}
