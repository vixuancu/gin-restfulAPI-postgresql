package utils

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var allowedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}
var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

const maxFileSize = 5 << 20 // 5MB
func ValidateAndSaveFile(fileHeader *multipart.FileHeader, uploadDir string) (string, error) {
	// check extension in filename
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedFileTypes[ext] {
		return "", errors.New("không hỗ trợ định dạng tệp này, chỉ hỗ trợ jpg, jpeg, png")
	}
	// check file size
	if fileHeader.Size > maxFileSize { // 5MB
		return "", errors.New("kích thước tệp không được lớn hơn 5MB")
	}
	// check content type xem no có phải là hình ảnh không
	file, err := fileHeader.Open()
	if err != nil {
		return "", errors.New("không thể mở tệp: ")
	}
	defer file.Close()
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", errors.New("không thể đọc tệp: ")
	}
	mimeType := http.DetectContentType(buffer)
	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("không hỗ trợ định dạng tệp này, chỉ hỗ trợ %v", allowedMimeTypes)
	}
	// change file name to uuid
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	err = os.MkdirAll("./upload", os.ModePerm)
	if err != nil {
		return "", errors.New("không thể tạo thư mục upload: ")
	}
	// uploadDir "./upload" + filename "abc123.jpg"
	savePath := filepath.Join(uploadDir, filename)
	if err := saveFile(fileHeader, savePath); err != nil {
		return "", err
	}
	return filename, nil
}
func saveFile(fileHeader *multipart.FileHeader, destination string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	if err != nil {
		return err
	}
	return nil
}
