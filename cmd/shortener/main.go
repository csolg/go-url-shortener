package main

import (
	"io"
	"net/http"
	"strings"
	"sync"
)

const shortURLPrefix = "http://localhost:8080/"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	mu       sync.Mutex
	nextID   uint64
	urlStore = map[string]string{}
)

func Encode(num uint64) string {
	if num == 0 {
		return string(alphabet[0])
	}

	base := uint64(len(alphabet))
	s := make([]byte, 0, 11)

	for num > 0 {
		s = append(s, alphabet[num%base])
		num /= base
	}

	// reverse
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return string(s)
}

func Decode(s string) uint64 {
	var num uint64
	base := uint64(len(alphabet))

	for i := 0; i < len(s); i++ {
		char := s[i]

		var val uint64

		switch {
		case char >= '0' && char <= '9':
			val = uint64(char - '0')
		case char >= 'a' && char <= 'z':
			val = uint64(char-'a') + 10
		case char >= 'A' && char <= 'Z':
			val = uint64(char-'A') + 36
		}

		num = num*base + val
	}

	return num
}

func createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	url := string(body)
	if strings.TrimSpace(url) == "" {
		http.Error(w, "Empty URL", http.StatusBadRequest)
		return
	}

	// create short url
	shortURL := generateShortURL(url)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func generateShortURL(url string) string {
	mu.Lock()
	defer mu.Unlock()

	nextID++
	id := Encode(nextID)
	urlStore[id] = url

	return shortURLPrefix + id
}

func redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path == "/" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.Error(w, "Missing short URL id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	originalURL, ok := urlStore[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Short URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/":
		createShortURL(w, r)
	case r.Method == http.MethodGet && r.URL.Path != "/":
		redirectToOriginalURL(w, r)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handleRequest)

	return mux
}

func main() {
	mux := newRouter()

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
