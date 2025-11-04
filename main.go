package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileStore struct {
	mu    sync.RWMutex
	files map[string]string
}

var store = FileStore{
	files: make(map[string]string),
}

func generateShareCode(length int) (string, error) {
	bytes := make([]byte, length/2) 
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("./uploads", "share-*-"+header.Filename)
	if err != nil {
		http.Error(w, "Could not create temp file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}

	shareCode, err := generateShareCode(6)
	if err != nil {
		http.Error(w, "Could not generate code", http.StatusInternalServerError)
		return
	}

	store.mu.Lock()
	store.files[shareCode] = tempFile.Name()
	store.mu.Unlock()

	log.Printf("File uploaded: %s (Code: %s)", header.Filename, shareCode)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"share_code": shareCode,
	})
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	code := filepath.Base(r.URL.Path)
	if code == "" {
		http.Error(w, "Share code missing", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	filePath, ok := store.files[code]
	store.mu.RUnlock()

	if !ok {
		http.Error(w, "File not found. The code may be invalid or expired.", http.StatusNotFound)
		return
	}

	log.Printf("File requested: %s (Code: %s)", filePath, code)

	originalFilename := strings.SplitN(filepath.Base(filePath), "-", 3)[2]
	w.Header().Set("Content-Disposition", "attachment; filename="+originalFilename)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

func homepageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatalf("Could not create uploads directory: %v", err)
	}

	http.HandleFunc("/", homepageHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)

	log.Println("Web share server starting on :8080...")
	log.Println("Open http://localhost:8080 to share a file.")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
