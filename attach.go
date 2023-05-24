package tunnel

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Attachment struct {
	filename string
	hash     string
	code     int
	rc       io.ReadCloser
}

func (att *Attachment) Filename() string {
	return att.filename
}

func (att *Attachment) Hash() string {
	return att.hash
}

// NotModified 文件是否未改变
func (att *Attachment) NotModified() bool {
	return att.code == http.StatusNotModified
}

// ZipFile 判断文件是否是 zip 文件
func (att *Attachment) ZipFile() bool {
	ext := filepath.Ext(att.filename)
	return strings.ToLower(ext) == ".zip"
}

func (att *Attachment) WriteTo(w io.Writer) (n int64, err error) {
	//goland:noinspection GoUnhandledErrorResult
	defer att.rc.Close()
	return io.Copy(w, att.rc)
}

func (att *Attachment) File(path string) (string, error) {
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer func() {
		_ = file.Close()
		_ = att.rc.Close()
	}()

	h := md5.New()
	r := io.TeeReader(att.rc, h)
	if _, err = io.Copy(file, r); err != nil {
		return "", err
	}

	dat := h.Sum(nil)
	sum := hex.EncodeToString(dat)

	return sum, nil
}

// Unzip 将文件解压到指定路径
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

	temp := filepath.Join(path, att.filename)
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
