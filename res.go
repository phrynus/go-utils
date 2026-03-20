package utils

import (
	"net/http"
	"time"

	lzstring "github.com/daku10/go-lz-string"
	"github.com/gin-gonic/gin"
	"github.com/phrynus/go-utils/crypto"
	"github.com/phrynus/go-utils/unknown"
)

// ResData 统一响应格式
type ResData[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
	Time int64  `json:"time"`
	Sign string `json:"sign,omitempty"` // 响应签名
}

// ErrorMessages 错误码映射消息
var ErrorMessages = map[int]string{
	400: "参数错误",
	401: "未授权",
	403: "禁止访问",
	404: "资源不存在",
	408: "请求超时",
	429: "请求过于频繁",
	500: "服务器内部错误",
	502: "网关错误",
	503: "服务不可用",
	504: "网关超时",
}

type ToolMap struct {
	AppKey     string `json:"app_key"`  // 应用密钥
	Compress   bool   `json:"compress"` // 是否启用压缩
	Encryption *struct {
		Key string `json:"key"` // 加密密钥 aes
		IV  string `json:"iv"`  // 加密向量 aes
	} `json:"encryption"`
}

// Success 发送成功响应并设置HTTP状态码
func S(c *gin.Context, data interface{}, msg string) {
	tool, found := c.Get("tool")
	toolMap := ToolMap{}
	unknown.NewUnknown(tool).SmartUnmarshal(&toolMap)
	if found {
		jsonBytes := unknown.NewUnknown(data).JsonString()
		if toolMap.Compress && toolMap.Encryption != nil { // 压缩并加密
			compressed, _ := lzstring.CompressToUint8Array(jsonBytes)
			encrypted, _ := crypto.AesEncrypt(toolMap.Encryption.Key, string(compressed), toolMap.Encryption.IV)
			data = encrypted
		} else if toolMap.Compress { // 只压缩
			compressed, _ := lzstring.CompressToBase64(jsonBytes)
			data = compressed
		} else if toolMap.Encryption != nil { // 只加密
			encrypted, _ := crypto.AesEncrypt(toolMap.Encryption.Key, jsonBytes, toolMap.Encryption.IV)
			data = encrypted
		}
	}

	if msg == "" {
		msg = "ok"
	}

	now := time.Now().UnixMilli()

	Sign := unknown.NewUnknown(200, msg, now, data, toolMap.AppKey).JsonString()
	Res := ResData[interface{}]{
		Code: 200,
		Msg:  msg,
		Time: now,
		Data: data,
		Sign: crypto.MD5(Sign),
	}

	c.JSON(http.StatusOK, Res)
}

// SuccessHTML 发送 HTML 响应
func SuccessHTML(c *gin.Context, html string) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// SuccessFile 发送文件下载响应
func SuccessFile(c *gin.Context, filename string, data []byte, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

// Error 发送错误响应并设置HTTP状态码
func E(c *gin.Context, code int, msg string) {
	if msg == "" {
		if defaultMsg, exists := ErrorMessages[code]; exists {
			msg = defaultMsg
		} else {
			msg = "ServerError"
		}
	}

	now := time.Now().UnixMilli()

	Res := ResData[interface{}]{
		Code: code,
		Msg:  msg,
		Time: now,
		Data: gin.H{},
	}

	c.JSON(code, Res)
}
