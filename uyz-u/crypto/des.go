package crypto

import (
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"errors"
	"io"
)

// DesEncrypt 镜像参考实现的 DES CBC，使用 PKCS7 填充
func DesEncrypt(key, data, mode string) (string, error) {
	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plain := pkcs7Padding([]byte(data), block.BlockSize())
	ciphertext := make([]byte, des.BlockSize+len(plain))
	iv := ciphertext[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext[des.BlockSize:], plain)
	return encode(mode, ciphertext)
}

// DesDecrypt 对提供的密钥和模式执行 DesEncrypt 的逆操作
func DesDecrypt(key, data, mode string) (string, error) {
	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	cipherBytes, err := decode(mode, data)
	if err != nil {
		return "", err
	}
	if len(cipherBytes) < des.BlockSize {
		return "", errors.New("密文太短")
	}
	iv := cipherBytes[:des.BlockSize]
	cipherBytes = cipherBytes[des.BlockSize:]
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(cipherBytes, cipherBytes)
	plain := pkcs7Unpadding(cipherBytes)
	return string(plain), nil
}
