package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

// mã hóa AES
func EncryptAES(plainText,key []byte) (string, error) {
	block,err :=aes.NewCipher(key) // Tạo một khối mã hóa mới với khóa AES
	if err != nil {
		return "", err
	}
	// seal (mã hóa) + Open (giải mã)
	aesGCM,err:=cipher.NewGCM(block) // Tạo một chế độ mã hóa GCM mới
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize()) // Tạo nonce với kích thước nonce của GCM 
	if _,err:=rand.Read(nonce); err !=nil {
		return "", err 
	} // Sinh nonce ngẫu nhiên
	// ciphertext = nonce +ciphertext + tag
	cipherText:=aesGCM.Seal(nonce, nonce, plainText, nil) // Mã hóa dữ liệu

	return base64.URLEncoding.EncodeToString(cipherText), nil // Trả về chuỗi base64 của dữ liệu đã mã hóa
}

func DecryptAES(cipherBase64 string,key []byte) ([]byte, error) {
	cipherText,err := base64.URLEncoding.DecodeString(cipherBase64) // Giải mã chuỗi base64 thành dữ liệu nhị phân
	if err != nil {
		return nil, err
	}
	block,err :=aes.NewCipher(key) 
	if err != nil {
		return nil, err
	}
	aesGCM,err:=cipher.NewGCM(block) // Tạo một chế độ mã hóa GCM mới
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize() // Lấy kích thước nonce của GCM
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:] // Tách nonce và dữ liệu mã hóa

	return aesGCM.Open(nil, nonce, cipherText, nil) // Giải mã dữ liệu
}