package netutil

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"os"
)

// Attachment 文件附件下载
type Attachment struct {
	Filename string        // filename
	Checksum string        // 中心端计算的文件校验码一般是 SHA-1
	code     int           // http statusCode
	rc       io.ReadCloser // http 响应 body
}

// NotModified 文件是否未改变
func (att Attachment) NotModified() bool {
	return att.code == http.StatusNotModified
}

// Copy 写入到指定的流中，返回写入流的 SHA-1
func (att Attachment) Copy(dst io.Writer) (string, error) {
	hash := sha1.New()
	rd := io.TeeReader(att.rc, dst)
	//goland:noinspection GoUnhandledErrorResult
	defer att.rc.Close()

	if _, err := io.Copy(dst, rd); err != nil {
		return "", err
	}
	dat := hash.Sum(nil)
	sum := hex.EncodeToString(dat)

	return sum, nil
}

// File 保存为文件，返回写入流的 SHA-1
func (att Attachment) File(dst string) (string, error) {
	file, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	return att.Copy(file)
}
