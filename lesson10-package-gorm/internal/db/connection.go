package db

import (
	"context"
	"fmt"
	"hoc-gin/internal/config"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	connStr := config.NewConfig().DNS()

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN: connStr,
	}), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	// Get generic database object sql.DB to use its functions
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm DB: %v", err)
	}
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(2 * time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to ping database: %v", err)
	}
	log.Println("connected")
	return nil
}
