package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/csolg/go-url-shortener/internal/repository"
	"github.com/csolg/go-url-shortener/internal/storage"
)

const shortURLPrefix = "http://localhost:8080/"

type urlRepository interface {
	Save(ctx context.Context, originalURL string) (string, error)
	Get(ctx context.Context, shortID string) (string, error)
}

type app struct {
	repo urlRepository
}

func Encode(num uint64) string {
	return repository.Encode(num)
}

func Decode(s string) uint64 {
	num, _ := repository.Decode(s)
	return num
}

func (a *app) createShortURL(w http.ResponseWriter, r *http.Request) {
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

	id, err := a.repo.Save(r.Context(), url)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURLPrefix + id))
}

func (a *app) redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path == "/" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.Error(w, "Missing short URL id", http.StatusBadRequest)
		return
	}

	originalURL, err := a.repo.Get(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "Short URL not found", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *app) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/":
		a.createShortURL(w, r)
	case r.Method == http.MethodGet && r.URL.Path != "/":
		a.redirectToOriginalURL(w, r)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func newRouter(repo urlRepository) *http.ServeMux {
	app := &app{repo: repo}
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, app.handleRequest)

	return mux
}

func databaseDSN() string {
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		return dsn
	}

	return "shortener.db"
}

func main() {
	ctx := context.Background()
	db, err := storage.OpenSQLite(ctx, databaseDSN())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	mux := newRouter(repository.NewURLRepository(db))

	err = http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
