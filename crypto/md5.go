package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 返回输入的MD5值
func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
