package utils

import "strings"

func NormalizeString(text string) string {
	return strings.ToLower(strings.TrimSpace(text)) // Chuyển đổi thành chữ thường và loại bỏ khoảng trắng ở đầu và cuối
}

func ConvertToInt32Pointer(value int) *int32 {
	if value <= 0 {
		return nil
	}
	int32Value := int32(value)
	return &int32Value
}