package utils

import "strings"

func NormalizeString(text string) string {
	return strings.ToLower(strings.TrimSpace(text)) // Chuyển đổi thành chữ thường và loại bỏ khoảng trắng ở đầu và cuối
}

func ConvertToInt32Pointer(value int32) *int32 {
	if value <= 0 {
		return nil
	}
	
	return &value
}