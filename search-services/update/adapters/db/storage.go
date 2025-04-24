package db

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	_, err := db.conn.ExecContext(ctx, `
		INSERT INTO comics (id, url, words) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, comics.ID, comics.URL, comics.Words)
	if err != nil {
		return fmt.Errorf("failed to insert comic: %w", err)
	}

	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	if err := db.conn.GetContext(ctx, &stats.WordsTotal, `
        SELECT COALESCE(SUM(array_length(words, 1)), 0) FROM comics
    `); err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get words total: %w", err)
	}

	if err := db.conn.GetContext(ctx, &stats.WordsUnique, `
        SELECT COALESCE(COUNT(DISTINCT word), 0)
        FROM comics, unnest(words) AS word
    `); err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get unique words: %w", err)
	}

	if err := db.conn.GetContext(ctx, &stats.ComicsFetched, `SELECT COUNT(*) FROM comics`); err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get comics fetched: %w", err)
	}

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	if err := db.conn.SelectContext(ctx, &ids, `SELECT id FROM comics`); err != nil {
		return nil, fmt.Errorf("failed to get comic IDs: %w", err)
	}

	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, `DELETE FROM comics`); err != nil {
		return fmt.Errorf("failed to delete comics: %w", err)
	}

	return nil
}
