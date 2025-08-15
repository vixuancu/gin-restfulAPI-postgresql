package pgx

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	"user-management-api/pkg/logger"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
)

type PgxZeroLogTracer struct {
	Logger         zerolog.Logger
	SlowQueryLimit time.Duration
}
type QueryInfo struct {
	QueryName     string
	OperationType string
	CleanSQL      string
	OriginalSQL   string
}

var (
	sqlcNameRegex = regexp.MustCompile(`-- name:\s*(\w+)\s*:(\w+)`)
	spaceRegex    = regexp.MustCompile(`\s+`)
	commentRegex  = regexp.MustCompile(`-- [^\r\n]*`)
)

func parseSql(sql string) QueryInfo {
	info := QueryInfo{
		OriginalSQL: sql,
	}

	if mathes := sqlcNameRegex.FindStringSubmatch(sql); len(mathes) == 3 {
		info.QueryName = mathes[1]
		info.OperationType = strings.ToUpper(mathes[2])
	}

	cleanSQL := commentRegex.ReplaceAllString(sql, "")
	cleanSQL = strings.TrimSpace(cleanSQL)
	cleanSQL = spaceRegex.ReplaceAllString(cleanSQL, " ")
	info.CleanSQL = cleanSQL
	return info
}
func formatArg(arg any) string {
	val := reflect.ValueOf(arg) // Lấy giá trị của arg

	if arg == nil || (val.Kind() == reflect.Ptr && val.IsNil()) {
		return "NULL" // Trả về NULL nếu arg là nil hoặc con trỏ nil
	}
	if val.Kind() == reflect.Ptr {
		arg = val.Elem().Interface() // Lấy giá trị của con trỏ
	}
	switch v := arg.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case []byte:
		return "'" + strings.ReplaceAll(string(v), "'", "''") + "'"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return "'" + v.Format(time.RFC3339) + "'"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''"))
	}
}

// hafm này dùng để thay thế các placeholder trong câu lệnh SQL với các giá trị thực tế
func replacePlaceholders(sql string, args []any) string {
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		sql = strings.ReplaceAll(sql, placeholder, formatArg(arg))
	}
	return sql
}

func (t *PgxZeroLogTracer) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	// log.Printf("%+v",data)

	sql, _ := data["sql"].(string)
	args, _ := data["args"].([]any)
	duration, _ := data["time"].(time.Duration)

	queryInfo := parseSql(sql)

	var finalSQL string
	if len(args) > 0 {
		finalSQL = replacePlaceholders(queryInfo.CleanSQL, args)
	} else {
		finalSQL = queryInfo.CleanSQL
	}

	baseLogger := t.Logger.With().
		Str("trace_id", logger.GetTraceID(ctx)).
		Dur("duration", duration).
		Str("sql_orginal", queryInfo.OriginalSQL).
		Str("sql", finalSQL).
		Str("query_name", queryInfo.QueryName).
		Str("operation", queryInfo.OperationType).
		Interface("args", args)

	logger := baseLogger.Logger()

	if msg == "Query" && duration > t.SlowQueryLimit {
		logger.Warn().Str("event", "Slow Query").Msg("Slow sql query")
	}

	if msg == "Query" {
		logger.Info().Str("event", "Query").Msg("Executed Sql")
	}
}
