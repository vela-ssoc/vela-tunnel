package netutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPError 返回错误
type HTTPError struct {
	Code   int         `json:"code"`   // http 返回的 状态码
	Header http.Header `json:"header"` // Header
	Body   []byte      `json:"body"`   // http body，最多 1024 字节
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("httpclient 请求成功但是响应错误，状态码：%d, 报文：%s", e.Code, e.Body)
}

func (e *HTTPError) NotAcceptable() bool {
	return e.Code == http.StatusNotAcceptable
}

func (e *HTTPError) JSON(v any) error {
	return json.Unmarshal(e.Body, v)
}
