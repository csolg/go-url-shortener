package repository

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/csolg/go-url-shortener/internal/storage"
)

func TestURLRepositorySaveAndGet(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "shortener.db")
	db, err := storage.OpenSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	repo := NewURLRepository(db)
	originalURL := "https://practicum.yandex.ru/"

	id, err := repo.Save(context.Background(), originalURL)
	if err != nil {
		t.Fatalf("save url: %v", err)
	}

	if id != "1" {
		t.Fatalf("expected first id %q, got %q", "1", id)
	}

	got, err := repo.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get url: %v", err)
	}

	if got != originalURL {
		t.Fatalf("expected URL %q, got %q", originalURL, got)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	db, err = storage.OpenSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("reopen sqlite: %v", err)
	}
	defer db.Close()

	got, err = NewURLRepository(db).Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get url after reopen: %v", err)
	}

	if got != originalURL {
		t.Fatalf("expected persisted URL %q, got %q", originalURL, got)
	}
}

func TestURLRepositoryGetNotFound(t *testing.T) {
	db, err := storage.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "shortener.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	repo := NewURLRepository(db)

	_, err = repo.Get(context.Background(), "unknown")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
