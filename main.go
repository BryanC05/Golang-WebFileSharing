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
	"strings" // <--- ADDED THIS LINE
	"sync"
)

// This struct will hold our file information.
// We use a map and a mutex to make it concurrency-safe.
type FileStore struct {
	mu    sync.RWMutex
	files map[string]string // Key: Share Code, Value: File Path
}

// Global variable for our file store
var store = FileStore{
	files: make(map[string]string),
}

// generateShareCode creates a simple 6-char random string for the URL.
func generateShareCode(length int) (string, error) {
	bytes := make([]byte, length/2) // Each byte becomes 2 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ## 1. Upload Handler
// This handles the file upload from the sender.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form (for file uploads)
	// 32 << 20 = 32MB max memory. Larger files go to disk.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the "file" form field
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary file in our "uploads" directory
	tempFile, err := os.CreateTemp("./uploads", "share-*-"+header.Filename)
	if err != nil {
		http.Error(w, "Could not create temp file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Copy the uploaded file's content to the temp file
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}

	// Generate a unique share code
	shareCode, err := generateShareCode(6)
	if err != nil {
		http.Error(w, "Could not generate code", http.StatusInternalServerError)
		return
	}

	// Store the file path in our map, protected by the mutex
	store.mu.Lock()
	store.files[shareCode] = tempFile.Name()
	store.mu.Unlock()

	log.Printf("File uploaded: %s (Code: %s)", header.Filename, shareCode)

	// Return the share code to the user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"share_code": shareCode,
	})
}

// ## 2. Download Handler
// This handles the file download for the receiver.
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get the share code from the URL, e.g., /download/abc123
	code := filepath.Base(r.URL.Path)
	if code == "" {
		http.Error(w, "Share code missing", http.StatusBadRequest)
		return
	}

	// Look up the file path, protected by a read-lock
	store.mu.RLock()
	filePath, ok := store.files[code]
	store.mu.RUnlock()

	if !ok {
		http.Error(w, "File not found. The code may be invalid or expired.", http.StatusNotFound)
		return
	}

	log.Printf("File requested: %s (Code: %s)", filePath, code)

	// Set the correct headers to make the browser download the file
	// We use the original filename for the download prompt
	originalFilename := strings.SplitN(filepath.Base(filePath), "-", 3)[2]
	w.Header().Set("Content-Disposition", "attachment; filename="+originalFilename)
	w.Header().Set("Content-Type", "application/octet-stream")

	// http.ServeFile is a helper that streams the file
	http.ServeFile(w, r, filePath)
}

// ## 3. Homepage Handler
// Serves the main HTML page
func homepageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	// Create the uploads directory if it doesn't exist
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