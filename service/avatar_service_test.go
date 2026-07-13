package service

import (
	"bytes"
	"flower-lottery-backend/common"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestAvatarServiceSaveNormalizesImage(t *testing.T) {
	root := t.TempDir()
	service := NewAvatarService(root)
	source := image.NewRGBA(image.Rect(0, 0, 900, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 900; x++ {
			source.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 180, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, source, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}

	header := avatarFileHeader(t, "avatar.jpeg", "image/jpeg", encoded.Bytes())
	url, err := service.Save(header)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	savedPath := filepath.Join(root, "avatars", path.Base(url))
	file, err := os.Open(savedPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	saved, format, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" || saved.Bounds().Dx() != avatarSize || saved.Bounds().Dy() != avatarSize {
		t.Fatalf("unexpected normalized avatar: format=%s bounds=%v", format, saved.Bounds())
	}

	service.DeleteLocal(url)
	if _, err = os.Stat(savedPath); !os.IsNotExist(err) {
		t.Fatalf("DeleteLocal() left file behind: %v", err)
	}
}

func TestAvatarServiceRejectsUnsupportedType(t *testing.T) {
	service := NewAvatarService(t.TempDir())
	header := avatarFileHeader(t, "avatar.gif", "image/gif", []byte("GIF89a"))
	if _, err := service.Save(header); err != common.ErrAvatarType {
		t.Fatalf("Save() error = %v, want %v", err, common.ErrAvatarType)
	}
}

func avatarFileHeader(t *testing.T, filename, contentType string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Disposition", `form-data; name="avatar"; filename="`+filename+`"`)
	headers.Set("Content-Type", contentType)
	part, err := writer.CreatePart(headers)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })
	return form.File["avatar"][0]
}
