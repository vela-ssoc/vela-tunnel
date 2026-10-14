package stegano

import (
	"io"
	"os"
	"testing"
)

type ExampleManifest struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Offset  int64             `json:"offset"`
	Payload map[string]string `json:"payload"`
	// ... 其它的字段
}

func TestAddManifest(t *testing.T) {
	src, err := os.Open("base.png") // 原始干净的文件
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	dst, err := os.Create("dest.png")
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()

	offset, err := io.Copy(dst, src)
	if err != nil {
		t.Fatal(err)
	}

	manifest := &ExampleManifest{
		Name:    "manifest",
		Version: "1.0.0",
		Offset:  offset,
		Payload: map[string]string{
			"foo": "bar",
		},
	}

	if err = AddManifest(dst, manifest, offset); err != nil {
		t.Fatal(err)
	}

	t.Log("隐写成功")
}

func TestReadManifest(t *testing.T) {
	manifest := new(ExampleManifest)
	if err := ReadManifest("dest.png", manifest); err != nil {
		t.Fatal(err)
	}

	t.Logf("manifest: %+v\n", manifest)
}
