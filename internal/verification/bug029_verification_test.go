package verification

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"upcycle-hub/pkg/utils"
)

func TestBug029VerificationUploadPathSafety(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "../escape.jpg")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("image"))
	mw.Close()
	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.ParseMultipartForm(1024 * 1024)
	fh := req.MultipartForm.File["file"][0]
	fh.Filename = "../escape.jpg"
	got, err := utils.SaveUploadedFile(fh, dir, 1024)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(dir, got)
	if err != nil {
		t.Fatal(err)
	}
	if rel == ".." || len(rel) >= 3 && rel[:3] == "../" {
		t.Fatalf("escaped upload path: %s", got)
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatal(err)
	}
}
func TestBug029VerificationUploadPathRegression(t *testing.T) {
	if err := utils.EnsureDir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
