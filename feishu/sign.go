package feishu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"time"
)

// Validate 验证签名
// timestamp 当前时间戳，单位是秒
func Validate(signature, timestamp, secret string) (bool, error) {
	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, err
	}

	dur := time.Since(time.Unix(t, 0))
	if math.Abs(dur.Hours()) > 1 {
		return false, fmt.Errorf("timestamp is expired")
	}

	signature2, err := GenSign(secret, t)
	if err != nil {
		return false, err
	}
	return signature2 == signature, nil
}

// GenSign 生成签名
// timestamp 当前时间戳，单位是秒
func GenSign(secret string, timestamp int64) (string, error) {
	// timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
