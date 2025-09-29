package tunnel

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// ZipFile 判断文件是否是 zip 文件
func (att *Attachment) ZipFile() bool {
	ext := filepath.Ext(att.Filename())
	return strings.ToLower(ext) == ".zip"
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

// Unzip 将文件解压到指定路径
//
// Deprecated: 该方法还未完善，应该存在一些问题请勿使用。
func (att *Attachment) Unzip(path string) error {
	if !att.ZipFile() {
		return zip.ErrFormat
	}

	stat, err := os.Stat(path)
	if err != nil {
		// 目录不存在就创建
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err = os.MkdirAll(path, 0o666); err != nil {
			return err
		}
	}
	if !stat.IsDir() {
		return os.ErrExist
	}

	temp := filepath.Join(path, att.Filename())
	raw, err := os.Create(temp)
	if err != nil {
		return err
	}
	// 记得删除临时文件
	//goland:noinspection GoUnhandledErrorResult
	defer os.Remove(temp)

	size, err := att.WriteTo(raw)
	if err != nil {
		_ = raw.Close()
		return err
	}

	zr, err := zip.NewReader(raw, size)
	if err != nil {
		return err
	}

	for _, zf := range zr.File {
		zfp := filepath.Join(path, zf.Name)
		fi := zf.FileInfo()
		if fi.IsDir() {
			if err = os.MkdirAll(zfp, fi.Mode()); err != nil {
				return err
			}
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		if err = att.unzipTo(zfp, rc); err != nil {
			return err
		}
	}

	return nil
}

func (*Attachment) unzipTo(path string, rc io.ReadCloser) error {
	//goland:noinspection GoUnhandledErrorResult
	defer rc.Close()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	_, err = io.Copy(file, rc)

	return err
}

type ThirdInfo struct {
	ID         int64  // 三方文件 ID
	MD5        string // MD5
	Desc       string // 说明
	Customized string // 分类
	Extension  string // 扩展名
}
