package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrorCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrorCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrorCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrorCodeConflict       ErrorCode = "CONFLICT"
)

type AppError struct {
	Message string
	Code    ErrorCode
	Err     error
}

func (ae *AppError) Error() string {
	return ""
}

func NewError(message string, code ErrorCode) error {
	return &AppError{
		Message: message,
		Code:    code,
	}
}
func WrapError(err error, message string, code ErrorCode) error {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}
func ResponseError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		status :=  httpStatusFromCode(appErr.Code)
		response := gin.H{
			"error": appErr.Message,
			"code":  appErr.Code,
		}
		if appErr.Err != nil {
			response["details"] = appErr.Err.Error() // thêm thông tin chi tiết nếu có
		}
		c.JSON(status, response)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
		"code":  ErrorCodeInternalServer,
	})

}
func ResponSuccess(c *gin.Context,status int, data any) {
	c.JSON(status, gin.H{
		"status": "success",
		"data": data,
	})
}
func ResponseStatusCode(c *gin.Context,status int) {
	c.Status(status)
}
func ResponseValidator(c *gin.Context, data any) {
	c.JSON(http.StatusBadRequest, data)
}
func httpStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrorCodeBadRequest:
		return 400
	case ErrorCodeNotFound:
		return 404
	case ErrorCodeValidation:
		return 422
	case ErrorCodeInternalServer:
		return 500
	case ErrorCodeUnauthorized:
		return 401
	case ErrorCodeForbidden:
		return 403
	case ErrorCodeConflict:
		return 409
	default:
		return 500 // Mặc định trả về lỗi máy chủ nội bộ nếu không xác định được mã lỗi
	}
}