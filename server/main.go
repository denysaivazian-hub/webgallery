package main

import (
	"encoding/json"
	"errors"
	"image"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const maxUploadSize = 10 << 20 // 10 MiB

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "files"
	}

	addr := ":" + port
	log.Printf("listening on %s", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           newHandlerWithStorage(uploadDir),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func newHandler() http.Handler {
	return newHandlerWithStorage("files")
}

func newHandlerWithStorage(uploadDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "test"}); err != nil {
			log.Printf("failed to encode response: %v", err)
		}
	})

	mux.HandleFunc("POST /file", func(w http.ResponseWriter, req *http.Request) {
		req.Body = http.MaxBytesReader(w, req.Body, maxUploadSize)
		if err := req.ParseMultipartForm(1 << 20); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				http.Error(w, "file is too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}
		defer req.MultipartForm.RemoveAll()

		file, _, err := req.FormFile("image")
		if err != nil {
			http.Error(w, "image is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if _, format, err := image.DecodeConfig(file); err != nil || format != "png" {
			http.Error(w, "image must be a valid PNG", http.StatusUnsupportedMediaType)
			return
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			http.Error(w, "cannot read image", http.StatusInternalServerError)
			return
		}

		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			log.Printf("failed to create upload directory: %v", err)
			http.Error(w, "cannot store image", http.StatusInternalServerError)
			return
		}

		storedFile, err := os.CreateTemp(uploadDir, "upload-*.png")
		if err != nil {
			log.Printf("failed to create image: %v", err)
			http.Error(w, "cannot store image", http.StatusInternalServerError)
			return
		}
		storedPath := storedFile.Name()
		stored := false
		defer func() {
			if !stored {
				_ = os.Remove(storedPath)
			}
		}()

		if _, err := io.Copy(storedFile, file); err != nil {
			_ = storedFile.Close()
			log.Printf("failed to store image: %v", err)
			http.Error(w, "cannot store image", http.StatusInternalServerError)
			return
		}
		if err := storedFile.Close(); err != nil {
			log.Printf("failed to close stored image: %v", err)
			http.Error(w, "cannot store image", http.StatusInternalServerError)
			return
		}
		stored = true

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			http.Error(w, "cannot read image", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("X-Stored-Filename", filepath.Base(storedPath))
		if _, err := io.Copy(w, file); err != nil {
			log.Printf("failed to write image response: %v", err)
		}
	})

	mux.HandleFunc("GET /files/{name}", func(w http.ResponseWriter, req *http.Request) {
		name := req.PathValue("name")
		if filepath.Ext(name) != ".png" || filepath.Base(name) != name {
			http.NotFound(w, req)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, req, filepath.Join(uploadDir, name))
	})

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		if r.Header.Get("Origin") == "http://localhost:5173" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
