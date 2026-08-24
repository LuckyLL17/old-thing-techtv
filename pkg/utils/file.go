package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// resolveSafePath builds a destination path for an uploaded file under destDir,
// guaranteeing the result never escapes destDir. The stored filename is always a
// generated basename derived from a random token plus the validated extension,
// so a client-supplied filename (which may contain "..", "/", "\" or drive
// letters) can never reach the filesystem. The resulting absolute path is
// re-checked against destDir to defend against any residual traversal residue.
func resolveSafePath(destDir, ext string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	absDir, err := filepath.Abs(destDir)
	if err != nil {
		return "", err
	}
	for {
		token := make([]byte, 8)
		if _, err := rand.Read(token); err != nil {
			return "", err
		}
		name := hex.EncodeToString(token) + ext
		fullPath := filepath.Join(absDir, name)
		// Defense in depth: confirm the cleaned path still resolves inside absDir.
		// Rel must not start with ".." and must not be absolute.
		rel, err := filepath.Rel(absDir, fullPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
			return "", fmt.Errorf("invalid upload destination path")
		}
		return fullPath, nil
	}
}

func SaveUploadedFile(file *multipart.FileHeader, destDir string, maxSize int64) (string, error) {
	if file.Size > maxSize {
		return "", fmt.Errorf("file size exceeds limit %d bytes", maxSize)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
	// Only the extension is taken from the client; the filename is generated
	// server-side, which removes any path traversal surface from file.Filename.
	fullPath, err := resolveSafePath(destDir, ext)
	if err != nil {
		return "", err
	}
	// Write to a sibling temp file first, then rename onto the final path so a
	// partial write never leaves a half-written image under the real name.
	tmpPath := fullPath + ".tmp"
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	// Best-effort cleanup of the temp file if anything below fails. The rename
	// on success moves the file out from under tmpPath, so Remove is a no-op then.
	cleanup := func() { _ = os.Remove(tmpPath) }
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		cleanup()
		return "", err
	}
	if err := dst.Close(); err != nil {
		cleanup()
		return "", err
	}
	if err := os.Rename(tmpPath, fullPath); err != nil {
		cleanup()
		return "", err
	}
	return fullPath, nil
}

func RemoveFile(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func v6Task029Boundary2(left, right uint64) bool {
	return left > 0 && right > 0
}
