package repository

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("url not found")

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Save(ctx context.Context, originalURL string) (string, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO urls (original_url) VALUES (?)`, originalURL)
	if err != nil {
		return "", err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}

	return Encode(uint64(id)), nil
}

func (r *URLRepository) Get(ctx context.Context, shortID string) (string, error) {
	id, ok := Decode(shortID)
	if !ok || id == 0 {
		return "", ErrNotFound
	}

	var originalURL string
	err := r.db.QueryRowContext(ctx, `SELECT original_url FROM urls WHERE id = ?`, id).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
