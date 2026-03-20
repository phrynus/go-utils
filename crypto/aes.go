package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// AesEncrypt 使用提供的密钥加密明文，并根据模式（"base64" 或 "hex"）返回 base64 或 hex 编码的密文
func AesEncrypt(key, data, mode string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plain := pkcs7Padding([]byte(data), block.BlockSize())

	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext[aes.BlockSize:], plain)

	return encode(mode, ciphertext)
}

// AesDecrypt 对提供的密钥和模式执行 AesEncrypt 的逆操作
func AesDecrypt(key, data, mode string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	cipherBytes, err := decode(mode, data)
	if err != nil {
		return "", err
	}
	if len(cipherBytes) < aes.BlockSize {
		return "", errors.New("密文太短")
	}
	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]

	cipher.NewCBCDecrypter(block, iv).CryptBlocks(cipherBytes, cipherBytes)
	plain := pkcs7Unpadding(cipherBytes)
	return string(plain), nil
}
