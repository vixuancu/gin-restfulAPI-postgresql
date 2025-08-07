package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"net/url"
	"strings"
	"time"
)

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *CustomResponseWriter) Write(data []byte) (n int, err error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func LoggerMiddleware() gin.HandlerFunc {
	// nơi chứa log request
	logPath := "../../internal/logs/http.log"

	// Tạo logger với zerolog
	logger := zerolog.New(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     5,    //
		Compress:   true, // disabled by default
		LocalTime:  true, //
	}).With().Timestamp().Logger()
	return func(c *gin.Context) {
		start := time.Now()
		// Đọc toàn bộ body của request để ghi log
		contentType := c.GetHeader("Content-Type")
		requestBody := make(map[string]any) //interface{} cũng thay thế any được
		var formFiles []map[string]any
		if strings.HasPrefix(contentType, "multipart/form-data") {
			//multipart/form-data
			if err := c.Request.ParseMultipartForm(32 << 20); err == nil && c.Request.MultipartForm != nil {
				// for value
				for key, value := range c.Request.MultipartForm.Value {
					if len(value) == 1 {
						requestBody[key] = value[0] // Nếu chỉ có một giá trị, lưu trực tiếp
					} else {
						requestBody[key] = value
					}
				}
				// for file
				for field, files := range c.Request.MultipartForm.File {
					for _, f := range files {
						formFiles = append(formFiles, map[string]any{
							"field":        field,                        // Tên trường
							"filename":     f.Filename,                   // Tên file
							"size":         formatFileSize(f.Size),       // Kích thước file
							"content_type": f.Header.Get("Content-Type"), // Loại nội dung của file
						})
					}
				}
				if len(formFiles) > 0 {
					requestBody["form_files"] = formFiles // Lưu thông tin file vào requestBody
				}
			}
		} else {
			//application/json  application/x-www-form-urlencoded
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to read request body") // Ghi log lỗi nếu không đọc được body
				c.AbortWithStatusJSON(500, gin.H{"error": "Internal Server Error"})
				return
			}
			// Đặt lại body để các handler khác có thể đọc được
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			//fmt.Println(string(bodyBytes))
			if strings.HasPrefix(contentType, "application/json") {
				_ = json.Unmarshal(bodyBytes, &requestBody) // Chuyển đổi JSON thành map
			} else {
				//application/x-www-form-urlencoded
				values, _ := url.ParseQuery(string(bodyBytes))
				for key, value := range values {
					if len(value) == 1 {
						requestBody[key] = value[0] // Nếu chỉ có một giá trị, lưu trực tiếp
					} else {
						requestBody[key] = value
					}

				}
			}
		}
		customeWriter := &CustomResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = customeWriter // Thay thế ResponseWriter của gin bằng CustomResponseWriter
		c.Next()
		// Tính toán thời gian xử lý request
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		responseContentType := c.Writer.Header().Get("Content-Type")
		responseBodyRaw := customeWriter.body.String()
		var responseBodyParsed interface{}
		if strings.HasPrefix(responseContentType, "image") {
			//
			responseBodyParsed = "[binary image data]"
		} else if strings.HasPrefix(responseContentType, "application/json") || strings.HasPrefix(strings.TrimSpace(responseBodyRaw), "{") || strings.HasPrefix(strings.TrimSpace(responseBodyRaw), "[") {
			if err := json.Unmarshal([]byte(responseBodyRaw), &responseBodyParsed); err != nil {
				responseBodyParsed = responseBodyRaw
			}
		} else {
			responseBodyParsed = responseBodyRaw
		}
		logEvent := logger.Info() // Mặc định ghi log ở mức Info
		if statusCode >= 500 {
			// Nếu mã trạng thái là 500 trở lên, ghi log ở mức Error
			logEvent = logger.WithLevel(zerolog.ErrorLevel)
		} else if statusCode >= 400 {
			// Nếu mã trạng thái là 400 trở lên, ghi log ở mức Warn
			logEvent = logger.WithLevel(zerolog.WarnLevel)
		} 

		logEvent.
			Str("method", c.Request.Method).                  // Ghi phương thức HTTP(GET, POST, PUT, DELETE, v.v.)
			Str("path", c.Request.URL.Path).                  // Ghi đường dẫn của request(ví dụ: /api/v1/users)
			Str("query", c.Request.URL.RawQuery).             // Ghi query string nếu có (ví dụ: ?page=1&limit=10)
			Str("client_ip", c.ClientIP()).                   // Ghi địa chỉ IP của client
			Str("user_agent", c.Request.UserAgent()).         // Ghi user agent của client (trình duyệt, ứng dụng, v.v.)
			Str("referer", c.Request.Referer()).              // Ghi referer nếu có (trang trước đó mà client đã truy cập)
			Str("protocol", c.Request.Proto).                 // Ghi giao thức HTTP (HTTP/1.1, HTTP/2, v.v.)
			Str("host", c.Request.Host).                      // Ghi host của request (ví dụ: example.com)
			Str("remote_address", c.Request.RemoteAddr).      // nếu địa chỉ IP của client không được cung cấp bởi c.ClientIP()
			Str("request_uri", c.Request.RequestURI).         // Ghi toàn bộ URI của request (bao gồm query string)
			Int64("content_length", c.Request.ContentLength). // Ghi độ dài của nội dung request  (nếu có)
			Interface("headers", c.Request.Header).           // Ghi tất cả các header của request
			Interface("request_body", requestBody).
			Interface("response_body", responseBodyParsed).
			Int("status_code", statusCode).                // Ghi mã trạng thái HTTP của response (ví dụ: 200, 404, 500, v.v.)
			Int64("duration_ms", duration.Milliseconds()). // Ghi thời gian xử lý request tính bằng mili giây

			Msg("HTTP Request Log")

	}
}
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}
