package muxconn

import (
	"encoding/json"
	"io"
)

// authEOF 认证报文结束标记符，因为 JSON 字符串不能包含这种字符。
const authEOF = 0x00

// WriteAuth 写入认证报文。
func WriteAuth(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := w.Write([]byte{authEOF})

	return err
}

// ReadAuth 读取认证报文
func ReadAuth(r io.Reader, result any) error {
	dr := &delimReader{raw: r, dem: authEOF}
	return json.NewDecoder(dr).Decode(result)
}

type delimReader struct {
	raw io.Reader // 原始流
	dem byte      // 结束标志
	err error     // 错误信息
}

func (dr *delimReader) Read(p []byte) (int, error) {
	if dr.err != nil {
		return 0, dr.err
	}

	n, err := dr.raw.Read(p)
	for i := range n {
		if p[i] == dr.dem {
			dr.err = io.EOF // 标记下次读取返回 EOF
			return i, nil   // 返回分隔符之前的数据
		}
	}

	return n, err
}
