package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	Level      string
	Filename   string
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int  //
	Compress   bool // disabled by default
	IsDev      string
}

func NewLogger(config LoggerConfig) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	lvl, err := zerolog.ParseLevel(config.Level) // Chuyển đổi chuỗi mức độ log thành zerolog.Level
	if err != nil {
		lvl = zerolog.InfoLevel // Mặc định là info nếu không parse được
	}
	zerolog.SetGlobalLevel(lvl) // Thiết lập mức độ log toàn cục
	var writer io.Writer

	if config.IsDev == "development" {
		writer = &PrettyJSONWriter{
			Writer: os.Stdout, // Ghi log ra console trong môi trường phát triển
		}
	} else {
		writer = &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,   //
			Compress:   config.Compress, // disabled by default
		}
	}

	logger := zerolog.New(writer).With().Timestamp().Logger()
	return &logger
}

// ghi đè io.Writer để ghi log dạng JSON
type PrettyJSONWriter struct {
	Writer io.Writer
}

func (w *PrettyJSONWriter) Write(p []byte) (n int, err error) {
	var prettyJSON bytes.Buffer

	err = json.Indent(&prettyJSON, p, "", "  ")
	if err != nil {
		return w.Writer.Write(p) // Nếu không thể định dạng, ghi log gốc
	}
	return w.Writer.Write(prettyJSON.Bytes())
}
