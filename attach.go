package tunnel

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Attachment struct {
	dispositions map[string]string
	code         int
	body         io.ReadCloser
	cancel       context.CancelFunc
}

func (att *Attachment) Read(p []byte) (int, error) {
	return att.body.Read(p)
}

func (att *Attachment) Close() error {
	err := att.body.Close()
	att.cancel()
	return err
}

func (att *Attachment) Filename() string {
	return att.dispositions["filename"]
}

func (att *Attachment) Hash() string {
	return att.dispositions["hash"]
}

func (att *Attachment) ThirdInfo() ThirdInfo {
	dis := att.dispositions
	str := dis["id"]
	num, _ := strconv.ParseInt(str, 10, 64)

	return ThirdInfo{
		ID:         num,
		MD5:        dis["hash"],
		Desc:       dis["desc"],
		Customized: dis["customized"],
		Extension:  dis["extension"],
	}
}

// NotModified 文件未发生变化
func (att *Attachment) NotModified() bool {
	return att.code == http.StatusNotModified
}

func (att *Attachment) WriteTo(w io.Writer) (int64, error) {
	//goland:noinspection GoUnhandledErrorResult
	defer att.Close()
	return io.Copy(w, att.body)
}

func (att *Attachment) File(path string) (string, error) {
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer func() {
		_ = file.Close()
		_ = att.Close()
	}()

	h := md5.New()
	r := io.TeeReader(att.body, h)
	if _, err = io.Copy(file, r); err != nil {
		return "", err
	}

	dat := h.Sum(nil)
	sum := hex.EncodeToString(dat)

	return sum, nil
}

type ThirdInfo struct {
	ID         int64  // 三方文件 ID
	MD5        string // MD5
	Desc       string // 说明
	Customized string // 分类
	Extension  string // 扩展名
}
