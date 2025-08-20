package db

import (
	"context"
	"fmt"
	"time"
	"user-management-api/internal/config"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/utils"
	"user-management-api/pkg/logger"
	"user-management-api/pkg/pgx"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

var DB sqlc.Querier      // đang sử dụng interface nên bỏ con trỏ
var DBpool *pgxpool.Pool // Khởi tạo biến toàn cục cho pool kết nối
func InitDB() error {
	connStr := config.NewConfig().DNS()

	conf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse database connection string: %v", err)
	}
	sqlLogger := utils.NewLoggerWithPath("sql.log", "info")

	conf.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: &pgx.PgxZeroLogTracer{
			Logger:         *sqlLogger,
			SlowQueryLimit: 500 * time.Millisecond, // Thời gian chậm để ghi log
		},
		LogLevel: tracelog.LogLevelDebug, // Mức độ ghi log
	}

	conf.MaxConns = 50                       // Set maximum number of connections to the database
	conf.MinConns = 5                        // Set minimum number of connections to the database
	conf.MaxConnLifetime = 30 * time.Minute  // Set maximum lifetime of a connection
	conf.MaxConnIdleTime = 5 * time.Minute   // Set maximum idle time for a connection
	conf.HealthCheckPeriod = 1 * time.Minute // bao nhieu lâu thì kiểm tra kết nối một lần

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // set thoi gian timeout cho kết nối
	defer cancel()

	DBpool, err = pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to create database connection pool: %v", err)
	}
	DB = sqlc.New(DBpool) // Khởi tạo sqlc với DBpool
	if err := DBpool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	} // Kiểm tra kết nối đến cơ sở dữ liệu

	logger.Log.Info().Msg("🍻 Connected to database successfully")
	return nil

}
