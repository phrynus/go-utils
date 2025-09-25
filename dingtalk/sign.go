package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
)

// Validate 验证签名
// timestamp 当前时间戳，单位是毫秒
func Validate(signature, timestamp, secret string) (bool, error) {
	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, err
	}

	dur := time.Since(time.UnixMilli(t))
	if math.Abs(dur.Hours()) > 1 {
		return false, fmt.Errorf("timestamp is expired")
	}

	signature2, err := Sign(timestamp, secret)
	if err != nil {
		return false, err
	}
	return signature2 == signature, nil
}

// Sign 生成签名
func Sign(timestamp string, secret string) (string, error) {
	stringToSign := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	if _, err := io.WriteString(h, stringToSign); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
