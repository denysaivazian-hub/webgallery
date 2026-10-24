package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileUpload(t *testing.T) {
	var pngData bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatal(err)
	}

	req := multipartRequest(t, "image", "test.png", pngData.Bytes())
	res := httptest.NewRecorder()
	uploadDir := t.TempDir()
	newHandlerWithStorage(uploadDir).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", res.Code, http.StatusOK, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	if !bytes.Equal(res.Body.Bytes(), pngData.Bytes()) {
		t.Error("response body does not match uploaded image")
	}
	storedName := res.Header().Get("X-Stored-Filename")
	if storedName == "" {
		t.Fatal("X-Stored-Filename header is empty")
	}
	storedData, err := os.ReadFile(filepath.Join(uploadDir, storedName))
	if err != nil {
		t.Fatalf("read stored image: %v", err)
	}
	if !bytes.Equal(storedData, pngData.Bytes()) {
		t.Error("stored image does not match uploaded image")
	}
}

func TestFileUploadMissingImage(t *testing.T) {
	req := multipartRequest(t, "other", "test.txt", []byte("test"))
	res := httptest.NewRecorder()
	newHandler().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestFileUploadMalformedMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/file", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "multipart/form-data")
	res := httptest.NewRecorder()
	newHandler().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestFileUploadTooLarge(t *testing.T) {
	req := multipartRequest(t, "image", "large.png", bytes.Repeat([]byte("x"), maxUploadSize))
	res := httptest.NewRecorder()
	newHandler().ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestCORSPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/file", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	res := httptest.NewRecorder()
	newHandler().ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestUnknownRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	res := httptest.NewRecorder()
	newHandler().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

func TestServePNGFile(t *testing.T) {
	uploadDir := t.TempDir()
	want := []byte("png contents")
	if err := os.WriteFile(filepath.Join(uploadDir, "example.png"), want, 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/files/example.png", nil)
	res := httptest.NewRecorder()
	newHandlerWithStorage(uploadDir).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
	if !bytes.Equal(res.Body.Bytes(), want) {
		t.Errorf("body = %q, want %q", res.Body.Bytes(), want)
	}
}

func TestServeFileRejectsNonPNG(t *testing.T) {
	uploadDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(uploadDir, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/files/secret.txt", nil)
	res := httptest.NewRecorder()
	newHandlerWithStorage(uploadDir).ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

func multipartRequest(t *testing.T, field, filename string, data []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/file", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
