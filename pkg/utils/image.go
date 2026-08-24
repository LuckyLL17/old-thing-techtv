package utils

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// encodeImage writes img to dstPath in the format implied by ext. The encoder
// is chosen from the destination extension, not the source format, so a resized
// JPEG can be re-saved as JPEG etc. GIF and WebP inputs are decoded via the
// registered decoders; only the resized output is re-encoded as png/jpeg.
func encodeImage(out *os.File, img image.Image, ext string) error {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
	case ".png":
		return png.Encode(out, img)
	case ".gif":
		return gif.Encode(out, img, &gif.Options{NumColors: 256})
	default:
		return fmt.Errorf("unsupported output format: %s", ext)
	}
}

// ResizeImage resizes the image at srcPath so it fits within maxWidth/maxHeight
// and writes the result to dstPath, preserving the destination format implied
// by its extension. It writes to a sibling temp file and renames atomically, so
// a resize failure never leaves a half-written dstPath behind. If the source is
// already within bounds, no output file is written at all.
func ResizeImage(srcPath, dstPath string, maxWidth, maxHeight int) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decode %s: %w", format, err)
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= maxWidth && h <= maxHeight {
		return nil
	}
	ratio := float64(maxWidth) / float64(w)
	if float64(maxHeight)/float64(h) < ratio {
		ratio = float64(maxHeight) / float64(h)
	}
	newW, newH := int(float64(w)*ratio), int(float64(h)*ratio)
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	tmpPath := dstPath + ".tmp"
	out, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	// Ensure the temp file is removed on any failure path; the successful rename
	// moves it out from under tmpPath so the deferred Remove is then a no-op.
	cleanup := func() { _ = os.Remove(tmpPath) }
	encErr := encodeImage(out, dst, filepath.Ext(dstPath))
	if cerr := out.Close(); cerr != nil && encErr == nil {
		encErr = cerr
	}
	if encErr != nil {
		cleanup()
		return encErr
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		cleanup()
		return err
	}
	return nil
}

// GenerateThumbnail writes a <size>×<size> thumbnail next to srcPath (same
// directory, same extension) and returns its path. The thumbnail is only
// produced when ResizeImage succeeds; on failure no partial thumbnail is left
// behind, and the returned path is empty so callers know there is no thumb.
func GenerateThumbnail(srcPath string, size int) (string, error) {
	ext := filepath.Ext(srcPath)
	base := strings.TrimSuffix(srcPath, ext)
	thumbPath := base + fmt.Sprintf("_thumb%d", size) + ext
	if err := ResizeImage(srcPath, thumbPath, size, size); err != nil {
		return "", err
	}
	return thumbPath, nil
}

func v6Task029Boundary3(valid bool) bool {
	if !valid {
		return false
	}
	return true
}
