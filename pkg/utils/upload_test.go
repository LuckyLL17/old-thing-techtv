package utils

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newPngBytes returns a small valid PNG so image.Decode and the resize/thumbnail
// pipeline have a real image to work with.
func newPngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// makeFileHeader builds a real *multipart.FileHeader with an attacker-controlled
// Filename by parsing a synthesized multipart/form-data body. This is exactly
// the value gin hands the handler from c.FormFile.
func makeFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="file"; filename="` + filename + `"`},
		"Content-Type":        {"image/png"},
	})
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	r := multipart.NewReader(bytes.NewReader(body.Bytes()), w.Boundary())
	form, err := r.ReadForm(int64(len(content)) + 1024)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	t.Cleanup(func() { _ = form.RemoveAll() })
	if len(form.File["file"]) != 1 {
		t.Fatalf("expected 1 file part, got %d", len(form.File["file"]))
	}
	return form.File["file"][0]
}

// TestSaveUploadedFile_PathTraversalNeutralized proves the headline bug: a client
// filename packed with ".." and separators cannot write outside destDir. The
// stored name must be a generated token, and the resulting path must stay
// within destDir.
func TestSaveUploadedFile_PathTraversalNeutralized(t *testing.T) {
	destDir := t.TempDir()
	img := newPngBytes(t, 2, 2)

	attacks := []string{
		"../../../etc/passwd.png",
		"..\\..\\windows\\evil.png",
		"subdir/../../evil.png",
		"/etc/hosts.png",
		"....//....//evil.png",
		"a/b/c/../../../escape.png",
	}
	for _, attack := range attacks {
		fh := makeFileHeader(t, attack, img)
		path, err := SaveUploadedFile(fh, destDir, 1<<20)
		if err != nil {
			t.Fatalf("attack %q: unexpected error: %v", attack, err)
		}
		// The stored filename must be a bare basename, not anything derived from
		// the attack string.
		base := filepath.Base(path)
		if base == filepath.Base(attack) || strings.Contains(base, "..") || strings.Contains(base, "/") || strings.Contains(base, "\\") {
			t.Fatalf("attack %q: stored name %q retained attacker-controlled path components", attack, base)
		}
		// Final path must resolve inside destDir.
		absDest, _ := filepath.Abs(destDir)
		absPath, _ := filepath.Abs(path)
		rel, err := filepath.Rel(absDest, absPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
			t.Fatalf("attack %q: path %q escapes destDir %q (rel=%q)", attack, absPath, absDest, rel)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("attack %q: expected written file at %q: %v", attack, path, err)
		}
		// No stray .tmp leftovers.
		if _, err := os.Stat(path + ".tmp"); err == nil {
			t.Fatalf("attack %q: leftover temp file at %q", attack, path+".tmp")
		}
	}
}

// TestSaveUploadedFile_RejectsBadExtension confirms the extension allowlist is
// the first gate and rejects e.g. .php / .html even with a traversal name.
func TestSaveUploadedFile_RejectsBadExtension(t *testing.T) {
	destDir := t.TempDir()
	img := newPngBytes(t, 2, 2)
	fh := makeFileHeader(t, "../../../shell.php", img)
	if _, err := SaveUploadedFile(fh, destDir, 1<<20); err == nil {
		t.Fatal("expected rejection of non-image extension")
	}
}

// TestSaveUploadedFile_LeavesNoTempOnSizeError ensures that when the write cannot
// proceed (e.g. maxSize exceeded), no temp/stray files are created in destDir.
func TestSaveUploadedFile_LeavesNoTempOnSizeError(t *testing.T) {
	destDir := t.TempDir()
	img := make([]byte, 64)
	fh := makeFileHeader(t, "big.png", img)
	if _, err := SaveUploadedFile(fh, destDir, 8); err == nil {
		t.Fatal("expected size-limit error")
	}
	entries, _ := os.ReadDir(destDir)
	if len(entries) != 0 {
		t.Fatalf("expected empty destDir on size error, got %d entries: %v", len(entries), entries)
	}
}

// TestResizeImage_NoPartialOutputOnDecodeError proves the temp-lifecycle fix: a
// decode failure (e.g. a non-image file named .png) must not leave a dstPath or
// a .tmp file behind.
func TestResizeImage_NoPartialOutputOnDecodeError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	if err := os.WriteFile(src, []byte("not really a png"), 0644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "out.png")
	if err := ResizeImage(src, dst, 100, 100); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := os.Stat(dst); err == nil {
		t.Fatal("partial dst file left behind after decode failure")
	}
	if _, err := os.Stat(dst + ".tmp"); err == nil {
		t.Fatal("temp file left behind after decode failure")
	}
}

// TestGenerateThumbnail_FailureLeavesNoFile confirms that when resize fails,
// no thumbnail (and no .tmp) is left in the source directory.
func TestGenerateThumbnail_FailureLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bad.png")
	if err := os.WriteFile(src, []byte("not an image"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateThumbnail(src, 100); err == nil {
		t.Fatal("expected thumbnail failure for non-image")
	}
	thumb := strings.TrimSuffix(src, ".png") + "_thumb100.png"
	if _, err := os.Stat(thumb); err == nil {
		t.Fatal("thumbnail file left behind after failure")
	}
	if _, err := os.Stat(thumb + ".tmp"); err == nil {
		t.Fatal("thumbnail temp file left behind after failure")
	}
}
