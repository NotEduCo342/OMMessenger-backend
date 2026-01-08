package storage

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

var (
	ErrTooLarge     = errors.New("file too large")
	ErrInvalidImage = errors.New("invalid image")
	ErrUnsupported  = errors.New("unsupported image type")
)

type ImageProcessOptions struct {
	MaxBytes    int64
	MaxDim      int
	JPEGQuality int
	// If source has alpha (e.g. PNG), flatten onto this background.
	FlattenBackground colorRGB
}

type colorRGB struct{ R, G, B uint8 }

func DefaultAvatarOptions() ImageProcessOptions {
	return ImageProcessOptions{
		MaxBytes:          5 * 1024 * 1024,
		MaxDim:            2048,
		JPEGQuality:       85,
		FlattenBackground: colorRGB{R: 255, G: 255, B: 255},
	}
}

// Detect allowed types by magic number.
func detectMagic(header []byte) (string, error) {
	if len(header) < 12 {
		return "", ErrInvalidImage
	}
	// JPEG: FF D8 FF
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "image/jpeg", nil
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 &&
		header[4] == 0x0D && header[5] == 0x0A && header[6] == 0x1A && header[7] == 0x0A {
		return "image/png", nil
	}
	// WebP: RIFF....WEBP
	if header[0] == 'R' && header[1] == 'I' && header[2] == 'F' && header[3] == 'F' &&
		header[8] == 'W' && header[9] == 'E' && header[10] == 'B' && header[11] == 'P' {
		return "image/webp", nil
	}
	return "", ErrUnsupported
}

// ProcessAvatarImage reads an uploaded image, validates, decodes, downscales to fit within MaxDim,
// and encodes as JPEG. It never upscales.
func ProcessAvatarImage(r io.Reader, opts ImageProcessOptions) ([]byte, string, int64, error) {
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 5 * 1024 * 1024
	}
	if opts.MaxDim <= 0 {
		opts.MaxDim = 2048
	}
	if opts.JPEGQuality <= 0 || opts.JPEGQuality > 100 {
		opts.JPEGQuality = 85
	}

	// Read bounded.
	limited := io.LimitReader(r, opts.MaxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", 0, err
	}
	if int64(len(data)) > opts.MaxBytes {
		return nil, "", 0, ErrTooLarge
	}
	if len(data) < 12 {
		return nil, "", 0, ErrInvalidImage
	}

	srcType, err := detectMagic(data[:12])
	if err != nil {
		return nil, "", 0, err
	}

	var img image.Image
	switch srcType {
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case "image/png":
		img, err = png.Decode(bytes.NewReader(data))
	case "image/webp":
		img, err = webp.Decode(bytes.NewReader(data))
	default:
		return nil, "", 0, ErrUnsupported
	}
	if err != nil {
		return nil, "", 0, fmt.Errorf("decode: %w", err)
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, "", 0, ErrInvalidImage
	}

	// Compute target size (fit within MaxDim, preserve aspect, never upscale).
	tw, th := w, h
	maxDim := opts.MaxDim
	if w > maxDim || h > maxDim {
		if w >= h {
			tw = maxDim
			th = int(float64(h) * (float64(maxDim) / float64(w)))
		} else {
			th = maxDim
			tw = int(float64(w) * (float64(maxDim) / float64(h)))
		}
		if tw < 1 {
			tw = 1
		}
		if th < 1 {
			th = 1
		}
	}

	dstRect := image.Rect(0, 0, tw, th)
	// Flatten onto opaque RGBA.
	dst := image.NewRGBA(dstRect)
	bg := image.NewUniform(color.RGBA{R: opts.FlattenBackground.R, G: opts.FlattenBackground.G, B: opts.FlattenBackground.B, A: 255})
	draw.Draw(dst, dst.Bounds(), bg, image.Point{}, draw.Src)

	// Scale/draw.
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	// Encode JPEG.
	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: opts.JPEGQuality}); err != nil {
		return nil, "", 0, fmt.Errorf("encode: %w", err)
	}
	return out.Bytes(), "image/jpeg", int64(out.Len()), nil
}
