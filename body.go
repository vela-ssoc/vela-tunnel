package tunnel

import (
	"bytes"
	"io"
)

type jsonBody struct {
	err error
	buf *bytes.Buffer
}

func (jb *jsonBody) Read(p []byte) (int, error) {
	if jb.err != nil {
		return 0, jb.err
	}
	if jb.buf == nil {
		return 0, io.EOF
	}
	return jb.buf.Read(p)
}

func (jb *jsonBody) Close() error { return nil }

func (jb *jsonBody) Len() int {
	if jb.err != nil || jb.buf == nil {
		return 0
	}
	return jb.buf.Len()
}
