package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// RSAEncrypt 使用提供的 PEM 公钥执行 PKCS#1 v1.5 加密
func RSAEncrypt(publicKey, data string) (string, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		return "", errors.New("无效的公钥 PEM 块")
	}
	keyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub := keyAny.(*rsa.PublicKey)

	chunkSize := pub.Size() - 11
	plain := []byte(data)
	var encrypted []byte
	for i := 0; i < len(plain); i += chunkSize {
		end := i + chunkSize
		if end > len(plain) {
			end = len(plain)
		}
		chunk, err := rsa.EncryptPKCS1v15(rand.Reader, pub, plain[i:end])
		if err != nil {
			return "", err
		}
		encrypted = append(encrypted, chunk...)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// RSADecrypt 使用提供的 PEM 私钥执行 PKCS#1 v1.5 解密
func RSADecrypt(privateKey, data string) (string, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", errors.New("无效的私钥 PEM 块")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	chunkSize := priv.Size()
	var decrypted []byte
	for i := 0; i < len(cipherBytes); i += chunkSize {
		end := i + chunkSize
		if end > len(cipherBytes) {
			end = len(cipherBytes)
		}
		chunk, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipherBytes[i:end])
		if err != nil {
			return "", err
		}
		decrypted = append(decrypted, chunk...)
	}
	return string(decrypted), nil
}
