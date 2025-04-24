package db

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"yadro.com/course/search/core"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

func TestNew(t *testing.T) {
	t.Run("connection error", func(t *testing.T) {
		originalConnect := sqlxConnect
		defer func() { sqlxConnect = originalConnect }()

		sqlxConnect = func(driverName, dataSourceName string) (*sqlx.DB, error) {
			return nil, errors.New("connection failed")
		}

		db, err := New(slog.Default(), "test_connection_string")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestPing(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
		log:  slog.Default(),
	}

	t.Run("successful ping", func(t *testing.T) {
		mock.ExpectPing()
		err := d.Ping(context.Background())
		assert.NoError(t, err)
	})
}

func TestSearchComics(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
		log:  slog.Default(),
	}

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, url FROM comics WHERE words && \$1.*`).
			WithArgs(pq.Array([]string{"test"})).
			WillReturnError(errors.New("query failed"))

		_, err := d.SearchComics(context.Background(), []string{"test"}, 10)
		assert.Error(t, err)
	})
}

func TestStats(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
		log:  slog.Default(),
	}

	t.Run("successful stats", func(t *testing.T) {
		expected := core.DBStats{
			WordsTotal:    100,
			WordsUnique:   50,
			ComicsFetched: 10,
		}

		mock.ExpectQuery(`SELECT COALESCE\(SUM\(array_length\(words, 1\)\), 0\) FROM comics`).
			WillReturnRows(sqlxmock.NewRows([]string{"coalesce"}).AddRow(100))

		mock.ExpectQuery(`SELECT COALESCE\(COUNT\(DISTINCT word\), 0\) FROM comics, unnest\(words\) AS word`).
			WillReturnRows(sqlxmock.NewRows([]string{"coalesce"}).AddRow(50))

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM comics`).
			WillReturnRows(sqlxmock.NewRows([]string{"count"}).AddRow(10))

		result, err := d.Stats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("error in words total query", func(t *testing.T) {
		mock.ExpectQuery(`SELECT COALESCE\(SUM\(array_length\(words, 1\)\), 0\) FROM comics`).
			WillReturnError(errors.New("query failed"))

		_, err := d.Stats(context.Background())
		assert.Error(t, err)
	})
}

func TestAllComics(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
		log:  slog.Default(),
	}

	t.Run("successful fetch", func(t *testing.T) {
		expected := []core.Comics{
			{ID: 1, URL: "http://example.com/1", Words: []string{"test", "comic"}},
			{ID: 2, URL: "http://example.com/2", Words: []string{"example"}},
		}

		rows := sqlxmock.NewRows([]string{"id", "url", "words"}).
			AddRow(1, "http://example.com/1", pq.Array([]string{"test", "comic"})).
			AddRow(2, "http://example.com/2", pq.Array([]string{"example"}))

		mock.ExpectQuery(`SELECT id, url, words FROM comics ORDER BY id`).
			WillReturnRows(rows)

		result, err := d.AllComics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlxmock.NewRows([]string{"id", "url", "words"})
		mock.ExpectQuery(`SELECT id, url, words FROM comics ORDER BY id`).
			WillReturnRows(rows)

		result, err := d.AllComics(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestGetComicsByIDs(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
		log:  slog.Default(),
	}

	t.Run("successful fetch", func(t *testing.T) {
		expected := []core.Comics{
			{ID: 1, URL: "http://example.com/1", Words: []string{"test"}},
		}

		rows := sqlxmock.NewRows([]string{"id", "url", "words"}).
			AddRow(1, "http://example.com/1", pq.Array([]string{"test"}))

		mock.ExpectQuery(`SELECT id, url, words FROM comics WHERE id = ANY\(\$1\)`).
			WithArgs(pq.Array([]int{1})).
			WillReturnRows(rows)

		result, err := d.GetComicsByIDs(context.Background(), []int{1})
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("empty ids", func(t *testing.T) {
		result, err := d.GetComicsByIDs(context.Background(), []int{})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, url, words FROM comics WHERE id = ANY\(\$1\)`).
			WithArgs(pq.Array([]int{1})).
			WillReturnError(errors.New("query failed"))

		_, err := d.GetComicsByIDs(context.Background(), []int{1})
		assert.Error(t, err)
	})
}

var sqlxConnect = sqlx.Connect
