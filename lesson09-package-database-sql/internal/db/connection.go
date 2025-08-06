package db

import (
	"context"
	"database/sql"
	"fmt"
	"hoc-gin/internal/config"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

func InitDB() error {
	connStr := config.NewConfig().DNS()
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal("unable to use data source name", err)
	}
	DB.SetMaxOpenConns(10) // Set maximum number of open connections to the database - số kết nối tối đa mở đến cơ sở dữ liệu
	DB.SetMaxIdleConns(5)  // Set maximum number of idle connections in the pool- số kết nối tối đa không sử dụng- nhàn dỗi
	DB.SetConnMaxLifetime(30 * time.Minute)
	/*
		Thời gian tối đa một connection được tồn tại trong pool, kể cả đang được sử dụng.
		Sau 30 phút, kết nối sẽ luôn bị đóng, dù nó đang hoạt động hay không.
	*/
	DB.SetConnMaxIdleTime(5 * time.Minute)
	/*
		Thời gian rảnh tối đa mà một connection không được sử dụng (idle).
		Nếu một connection không được sử dụng trong 5 phút, nó sẽ bị đóng.
		Nhưng nếu connection vẫn đang được dùng → không bị ảnh hưởng.
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	/*
		Tạo một context mới với giới hạn thời gian là 5 giây.
		Nếu sau 5 giây mà chưa hoàn thành, context sẽ tự động hủy (timeout).
		cancel() là hàm dùng để hủy thủ công context khi không cần nữa (được gọi qua defer cancel()).
		✅ Mục đích: tránh việc gọi đến DB bị treo mãi nếu DB không phản hồi.
	*/
	if err := DB.PingContext(ctx); err != nil {
		DB.Close()
		return fmt.Errorf("DB ping err: %w", err)
	}
	/*
			Gửi một lệnh PING đến database, dùng ctx để kiểm soát thời gian thực hiện.
		Nếu DB không phản hồi trong 5 giây → trả lỗi context deadline exceeded.
	*/
	log.Println("connected")
	return nil
}
