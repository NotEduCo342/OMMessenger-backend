package storage

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestProcessAvatarImage_PNG_ToJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 120, 60))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	out, ct, _, err := ProcessAvatarImage(bytes.NewReader(pngBuf.Bytes()), DefaultAvatarOptions())
	if err != nil {
		t.Fatalf("ProcessAvatarImage: %v", err)
	}
	if ct != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", ct)
	}

	decoded, err := jpeg.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("jpeg decode: %v", err)
	}
	if decoded.Bounds().Dx() != 120 || decoded.Bounds().Dy() != 60 {
		t.Fatalf("dims = %dx%d, want 120x60", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestProcessAvatarImage_DownscalesToFit(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 50))

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	opts := DefaultAvatarOptions()
	opts.MaxDim = 100
	out, _, _, err := ProcessAvatarImage(bytes.NewReader(pngBuf.Bytes()), opts)
	if err != nil {
		t.Fatalf("ProcessAvatarImage: %v", err)
	}

	decoded, err := jpeg.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("jpeg decode: %v", err)
	}
	// 200x50 scaled to fit MaxDim=100 => 100x25
	if decoded.Bounds().Dx() != 100 || decoded.Bounds().Dy() != 25 {
		t.Fatalf("dims = %dx%d, want 100x25", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestProcessAvatarImage_TooLarge(t *testing.T) {
	opts := DefaultAvatarOptions()
	opts.MaxBytes = 10
	payload := bytes.Repeat([]byte{0x00}, 11)
	_, _, _, err := ProcessAvatarImage(bytes.NewReader(payload), opts)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrTooLarge {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
}

func TestProcessAvatarImage_UnsupportedMagic(t *testing.T) {
	payload := bytes.Repeat([]byte{0x01}, 128)
	_, _, _, err := ProcessAvatarImage(bytes.NewReader(payload), DefaultAvatarOptions())
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrUnsupported {
		t.Fatalf("err = %v, want ErrUnsupported", err)
	}
}

func TestSafeJoinAvatarPath(t *testing.T) {
	if _, err := SafeJoinAvatarPath("", "../x"); err == nil {
		t.Fatalf("expected error for traversal")
	}
	if _, err := SafeJoinAvatarPath("", "..\\x"); err == nil {
		t.Fatalf("expected error for backslash")
	}
	key, err := SafeJoinAvatarPath("", "/avatars/1/a.jpg")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if key != "avatars/1/a.jpg" {
		t.Fatalf("key = %q", key)
	}
}
