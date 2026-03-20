package crypto

import "crypto/rc4"

// Rc4Encrypt 使用提供的密钥加密数据并返回编码后的文本
func Rc4Encrypt(key, data, mode string) (string, error) {
	block, err := rc4.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	plain := []byte(data)
	out := make([]byte, len(plain))
	block.XORKeyStream(out, plain)
	return encode(mode, out)
}

// Rc4Decrypt 执行 Rc4Encrypt 的逆操作并返回明文
func Rc4Decrypt(key, data, mode string) (string, error) {
	block, err := rc4.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	cipherBytes, err := decode(mode, data)
	if err != nil {
		return "", err
	}
	out := make([]byte, len(cipherBytes))
	block.XORKeyStream(out, cipherBytes)
	return string(out), nil
}
