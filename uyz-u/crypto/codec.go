package crypto

import (
	"encoding/base64"
	"encoding/hex"
)

func encode(mode string, data []byte) (string, error) {
	if mode == "base64" {
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return hex.EncodeToString(data), nil
}

func decode(mode string, data string) ([]byte, error) {
	if mode == "base64" {
		return base64.StdEncoding.DecodeString(data)
	}
	return hex.DecodeString(data)
}
