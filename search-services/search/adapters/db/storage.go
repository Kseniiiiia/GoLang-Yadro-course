package db

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
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

func (s *DB) SearchComics(ctx context.Context, words []string, limit int) ([]core.Comics, error) {
	var comics []core.Comics
	err := s.conn.SelectContext(ctx, &comics, `
        WITH search_words AS (
            SELECT unnest($1::text[]) AS word
        ),
        comic_matches AS (
            SELECT 
                c.id,
                c.url,
                -- Количество уникальных совпадающих слов
                COUNT(DISTINCT sw.word) AS unique_matches,
                -- Общее количество совпадений (с учетом частоты)
                SUM(
                    (SELECT COUNT(*) 
                     FROM unnest(c.words) AS comic_word 
                     WHERE comic_word = sw.word)
                ) AS total_matches
            FROM 
                comics c
            CROSS JOIN 
                search_words sw
            WHERE 
                c.words && $1
            GROUP BY 
                c.id, c.url
        )
        SELECT 
            id,
            url
        FROM 
            comic_matches
        ORDER BY
            -- Приоритет 1: комиксы с наибольшим количеством уникальных совпадений
            unique_matches DESC,
            -- Приоритет 2: комиксы с наибольшим абсолютным количеством совпадений
            total_matches DESC
        LIMIT $2
    `, pq.Array(words), limit)

	if err != nil {
		return nil, fmt.Errorf("failed to search comics: %w", err)
	}
	return comics, nil
}

func (s *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	if err := s.conn.GetContext(ctx, &stats.WordsTotal, `
		SELECT COALESCE(SUM(array_length(words, 1)), 0) FROM comics
	`); err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get words total: %w", err)
	}

	if err := s.conn.GetContext(ctx, &stats.WordsUnique, `
		SELECT COALESCE(COUNT(DISTINCT word), 0)
		FROM comics, unnest(words) AS word
	`); err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get unique words: %w", err)
	}

	if err := s.conn.GetContext(ctx, &stats.ComicsFetched, `
		SELECT COUNT(*) FROM comics
	`); err != nil {
		return core.DBStats{}, fmt.Errorf("failed to get comics fetched: %w", err)
	}

	return stats, nil
}

func (s *DB) AllComics(ctx context.Context) ([]core.Comics, error) {
	var dbComics []struct {
		ID    int            `db:"id"`
		URL   string         `db:"url"`
		Words pq.StringArray `db:"words"`
	}

	err := s.conn.SelectContext(ctx, &dbComics, `
        SELECT id, url, words 
        FROM comics
        ORDER BY id
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all comics: %w", err)
	}

	comics := make([]core.Comics, len(dbComics))
	for i, c := range dbComics {
		comics[i] = core.Comics{
			ID:    c.ID,
			URL:   c.URL,
			Words: []string(c.Words),
		}
	}

	return comics, nil
}

func (s *DB) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *DB) GetComicsByIDs(ctx context.Context, ids []int) ([]core.Comics, error) {
	if len(ids) == 0 {
		return []core.Comics{}, nil
	}

	var rawComics []struct {
		ID    int            `db:"id"`
		URL   string         `db:"url"`
		Words pq.StringArray `db:"words"`
	}

	query := `
        SELECT id, url, words 
        FROM comics 
        WHERE id = ANY($1)
    `
	err := s.conn.SelectContext(ctx, &rawComics, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get comics: %w", err)
	}

	comics := make([]core.Comics, len(rawComics))
	for i, raw := range rawComics {
		comics[i] = core.Comics{
			ID:    raw.ID,
			URL:   raw.URL,
			Words: []string(raw.Words),
		}
	}

	return comics, nil
}
