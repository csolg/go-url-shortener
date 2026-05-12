package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/csolg/go-url-shortener/internal/config"
	"github.com/csolg/go-url-shortener/internal/repository"
	"github.com/csolg/go-url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type urlRepository interface {
	Save(ctx context.Context, originalURL string) (string, error)
	Get(ctx context.Context, shortID string) (string, error)
}

type app struct {
	repo    urlRepository
	baseURL string
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
	w.Write([]byte(strings.TrimRight(a.baseURL, "/") + "/" + id))
}

func (a *app) redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path == "/" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
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

func badRequest(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func newRouterWithBaseURL(repo urlRepository, baseURL string) http.Handler {
	app := &app{
		repo:    repo,
		baseURL: baseURL,
	}
	router := chi.NewRouter()
	router.Post("/", app.createShortURL)
	router.Get("/{id}", app.redirectToOriginalURL)
	router.NotFound(badRequest)
	router.MethodNotAllowed(badRequest)

	return router
}

func databaseDSN() string {
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		return dsn
	}

	return "shortener.db"
}

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	db, err := storage.OpenSQLite(ctx, databaseDSN())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	mux := newRouterWithBaseURL(repository.NewURLRepository(db), cfg.BaseURL)

	err = http.ListenAndServe(cfg.ServerAddress, mux)
	if err != nil {
		panic(err)
	}
}
