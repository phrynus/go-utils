package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ContentTypeJSON = "application/json"
)

// PostJSON 发送JSON格式的POST请求
func PostJSON(url string, reqBody, respBody any) (http.Header, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, ContentTypeJSON, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 检查响应内容类型（Content-Type 可能包含 charset 等参数）
	contentType := resp.Header.Get("content-type")
	if !strings.Contains(strings.ToLower(contentType), ContentTypeJSON) {
		return nil, fmt.Errorf("http.response.header.content-type != %s, got: %s", ContentTypeJSON, contentType)
	}

	// 解析响应内容
	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, fmt.Errorf("http.response.body json decode failed, %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return resp.Header, fmt.Errorf("invalid http.response.status: %s", resp.Status)
	}

	return resp.Header, nil
}
