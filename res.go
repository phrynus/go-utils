package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
	500: "服务器错误",
}

// Success 发送成功响应并设置HTTP状态码
func S(c *gin.Context, data interface{}, msg string) {
	if msg == "" {
		msg = "ok"
	}

	now := time.Now().UnixMilli()

	Res := ResData[interface{}]{
		Code: 200,
		Msg:  msg,
		Time: now,
		Data: data,
		Sign: Sign(fmt.Sprintf("%d", now), 200, msg, now, data),
	}

	c.JSON(http.StatusOK, Res)
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
		Sign: Sign(fmt.Sprintf("%d", now), code, msg, now, gin.H{}),
	}

	// 根据错误码设置正确的HTTP状态码
	httpStatus := http.StatusInternalServerError
	switch code {
	case 400:
		httpStatus = http.StatusBadRequest
	case 401:
		httpStatus = http.StatusUnauthorized
	case 403:
		httpStatus = http.StatusForbidden
	case 404:
		httpStatus = http.StatusNotFound
	case 500:
		httpStatus = http.StatusInternalServerError
	}

	c.JSON(httpStatus, Res)
}
