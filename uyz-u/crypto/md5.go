package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5Hex 返回输入的小写 32 字符摘要
func MD5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
