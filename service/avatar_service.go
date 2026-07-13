package service

import (
	"crypto/rand"
	"encoding/hex"
	"flower-lottery-backend/common"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	maxAvatarUploadBytes = 8 << 20
	maxAvatarPixels      = 40_000_000
	avatarSize           = 512
)

type AvatarService struct{ root string }

func NewAvatarService(root string) *AvatarService {
	if strings.TrimSpace(root) == "" {
		root = "storage/uploads"
	}
	return &AvatarService{root: root}
}
func (s *AvatarService) MaxUploadBytes() int64 { return maxAvatarUploadBytes }

func (s *AvatarService) Save(file *multipart.FileHeader) (string, error) {
	if file == nil || file.Size <= 0 {
		return "", common.ErrAvatarRequired
	}
	if file.Size > maxAvatarUploadBytes {
		return "", common.ErrAvatarTooLarge
	}
	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if contentType != "image/jpeg" && contentType != "image/png" {
		return "", common.ErrAvatarType
	}
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()
	limited := io.LimitReader(source, maxAvatarUploadBytes+1)
	img, _, err := image.Decode(limited)
	if err != nil {
		return "", common.ErrAvatarDecode
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 || int64(width)*int64(height) > maxAvatarPixels {
		return "", common.ErrAvatarDimensions
	}
	output := cropSquare(img, avatarSize)
	name, err := randomName()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(s.root, "avatars")
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".jpg")
	target, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	if err = jpeg.Encode(target, output, &jpeg.Options{Quality: 85}); err != nil {
		target.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err = target.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return "/uploads/avatars/" + name + ".jpg", nil
}

func (s *AvatarService) DeleteLocal(url string) {
	const prefix = "/uploads/avatars/"
	if !strings.HasPrefix(url, prefix) {
		return
	}
	name := path.Base(strings.TrimPrefix(url, prefix))
	if name == "." || name == "" {
		return
	}
	_ = os.Remove(filepath.Join(s.root, "avatars", name))
}

func cropSquare(src image.Image, size int) *image.RGBA {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	cropSize := width
	if height < cropSize {
		cropSize = height
	}
	startX := bounds.Min.X + (width-cropSize)/2
	startY := bounds.Min.Y + (height-cropSize)/2
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	for y := 0; y < size; y++ {
		sy := startY + y*cropSize/size
		for x := 0; x < size; x++ {
			sx := startX + x*cropSize/size
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func randomName() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
