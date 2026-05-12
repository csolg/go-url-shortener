package storage

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const urlSchema = `
CREATE TABLE IF NOT EXISTS urls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	original_url TEXT NOT NULL
);`

func OpenSQLite(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.ExecContext(ctx, urlSchema); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
