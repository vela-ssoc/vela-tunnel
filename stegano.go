package tunnel

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/fs"
)

// ManifestFile 为系统约定（规定）的隐写配置文件名字，不要随意改变。
const ManifestFile = "manifest.json"

// ReadManifest 读取元配置信息。
//
//goland:noinspection GoUnhandledErrorResult
func ReadManifest(f string, v any) error {
	zrc, err := zip.OpenReader(f)
	if err != nil {
		return err
	}
	defer zrc.Close()

	mf, err := zrc.Open(ManifestFile)
	if err != nil {
		return err
	}
	defer mf.Close()

	dec := json.NewDecoder(mf)

	return dec.Decode(v)
}

// AddFS 向流中追加隐写文件系统。
// 一个流中最好只追加一个隐写流，
// 追加多个会导致读取的不确定性。
//
// offset 非常重要，即：
// 隐写数据（zip 文件）追加时，前面已写入的数据长度，
// 这个 offset，决定了程序能否正确识别最终输出的隐写文件。
// tips: 输出的最终文件，将后缀改成 .zip 可以直接打开。
//
// FIXME 与 AddManifest 方法只能选择一个使用。
//
//goland:noinspection GoUnhandledErrorResult
func AddFS(w io.Writer, fsys fs.FS, offset int64) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	if offset > 0 {
		zw.SetOffset(offset)
	}

	return zw.AddFS(fsys)
}

// AddManifest 向流中追加隐写元数据。
// 一个流中最好只追加一个隐写流，
// 追加多个会导致读取的不确定性。
//
// offset 非常重要，即：
// 隐写数据（zip 文件）追加时，前面已写入的数据长度，
// 这个 offset，决定了程序能否正确识别最终输出的隐写文件。
// tips: 输出的最终文件，将后缀改成 .zip 可以直接打开。
//
// FIXME 与 AddFS 方法只能选择一个使用。
//
//goland:noinspection GoUnhandledErrorResult
func AddManifest(w io.Writer, manifest any, offset int64) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	if offset > 0 {
		zw.SetOffset(offset)
	}
	zc, err := zw.Create(ManifestFile)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(zc)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err = enc.Encode(manifest)

	return err
}
