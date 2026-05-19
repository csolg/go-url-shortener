package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/csolg/go-url-shortener/internal/config"
	"github.com/csolg/go-url-shortener/internal/repository"
	"github.com/csolg/go-url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type app struct {
	repo    repository.URLStore
	baseURL string
}

func (a *app) createShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "Empty URL", http.StatusBadRequest)
		return
	}
	if !isAbsoluteURL(originalURL) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := a.repo.Save(r.Context(), originalURL)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err := fmt.Fprintf(w, "%s/%s", strings.TrimRight(a.baseURL, "/"), id); err != nil {
		return
	}
}

func (a *app) redirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
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

func isAbsoluteURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsed.IsAbs() && parsed.Host != ""
}

func badRequest(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func newRouterWithBaseURL(repo repository.URLStore, baseURL string) http.Handler {
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

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	db, err := storage.OpenSQLite(ctx, cfg.DatabaseDSN)
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
