package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/csolg/go-url-shortener/internal/repository"
	"github.com/csolg/go-url-shortener/internal/storage"
)

func TestCreateAndRedirectShortURL(t *testing.T) {
	router := newTestRouter(t)
	originalURL := "https://practicum.yandex.ru/"

	createReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	createRec := httptest.NewRecorder()

	router.ServeHTTP(createRec, createReq)

	createRes := createRec.Result()
	defer createRes.Body.Close()

	if createRes.StatusCode != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createRes.StatusCode)
	}

	shortURL := createRec.Body.String()
	if !strings.HasPrefix(shortURL, shortURLPrefix) {
		t.Fatalf("expected short URL with prefix %q, got %q", shortURLPrefix, shortURL)
	}

	id := strings.TrimPrefix(shortURL, shortURLPrefix)
	redirectReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	redirectRec := httptest.NewRecorder()

	router.ServeHTTP(redirectRec, redirectReq)

	redirectRes := redirectRec.Result()
	defer redirectRes.Body.Close()

	if redirectRes.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect status %d, got %d", http.StatusTemporaryRedirect, redirectRes.StatusCode)
	}

	if location := redirectRes.Header.Get("Location"); location != originalURL {
		t.Fatalf("expected Location %q, got %q", originalURL, location)
	}
}

func TestBadRequests(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "GET root",
			method: http.MethodGet,
			path:   "/",
		},
		{
			name:   "POST non-root path",
			method: http.MethodPost,
			path:   "/unknown",
			body:   "https://practicum.yandex.ru/",
		},
		{
			name:   "PUT root",
			method: http.MethodPut,
			path:   "/",
			body:   "https://practicum.yandex.ru/",
		},
		{
			name:   "GET unknown id",
			method: http.MethodGet,
			path:   "/unknown",
		},
		{
			name:   "POST empty body",
			method: http.MethodPost,
			path:   "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
			}
		})
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	db, err := storage.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "shortener.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	})

	return newRouter(repository.NewURLRepository(db))
}
