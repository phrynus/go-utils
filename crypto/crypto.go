package crypto

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
)

// encode 编码
func encode(mode string, data []byte) (string, error) {
	if mode == "base64" {
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return hex.EncodeToString(data), nil
}

// decode 解码
func decode(mode string, data string) ([]byte, error) {
	if mode == "base64" {
		return base64.StdEncoding.DecodeString(data)
	}
	return hex.DecodeString(data)
}

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, pad...)
}

// pkcs7Unpadding 去除填充
func pkcs7Unpadding(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding > len(data) {
		return data
	}
	return data[:len(data)-padding]
}
