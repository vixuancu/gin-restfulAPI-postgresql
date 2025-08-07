package utils

import "strings"

func NormalizeString(text string) string {
	return strings.ToLower(strings.TrimSpace(text)) // Chuyển đổi thành chữ thường và loại bỏ khoảng trắng ở đầu và cuối
}