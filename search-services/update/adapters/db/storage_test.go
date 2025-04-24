package db

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"yadro.com/course/update/core"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

func TestNew(t *testing.T) {
	t.Run("connection error", func(t *testing.T) {
		// Create a test logger
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		// Replace the actual sqlx.Connect with a function that returns error
		originalConnect := sqlxConnect
		sqlxConnect = func(driverName, dataSourceName string) (*sqlx.DB, error) {
			return nil, errors.New("connection failed")
		}
		defer func() { sqlxConnect = originalConnect }()

		// Test the New function
		d, err := New(logger, "test_connection_string")
		assert.Error(t, err)
		assert.Nil(t, d)
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
	}

	t.Run("successful stats", func(t *testing.T) {
		expectedStats := core.DBStats{
			WordsTotal:    100,
			WordsUnique:   50,
			ComicsFetched: 10,
		}

		// Mock WordsTotal query
		mock.ExpectQuery("SELECT COALESCE\\(SUM\\(array_length\\(words, 1\\)\\), 0\\) FROM comics").
			WillReturnRows(sqlxmock.NewRows([]string{"coalesce"}).AddRow(expectedStats.WordsTotal))

		// Mock WordsUnique query
		mock.ExpectQuery("SELECT COALESCE\\(COUNT\\(DISTINCT word\\), 0\\)").
			WillReturnRows(sqlxmock.NewRows([]string{"coalesce"}).AddRow(expectedStats.WordsUnique))

		// Mock ComicsFetched query
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM comics").
			WillReturnRows(sqlxmock.NewRows([]string{"count"}).AddRow(expectedStats.ComicsFetched))

		stats, err := d.Stats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedStats, stats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error in words total query", func(t *testing.T) {
		mock.ExpectQuery("SELECT COALESCE\\(SUM\\(array_length\\(words, 1\\)\\), 0\\) FROM comics").
			WillReturnError(errors.New("query failed"))

		_, err := d.Stats(context.Background())
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestIDs(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
	}

	t.Run("successful IDs retrieval", func(t *testing.T) {
		expectedIDs := []int{1, 2, 3}

		mock.ExpectQuery("SELECT id FROM comics").
			WillReturnRows(sqlxmock.NewRows([]string{"id"}).
				AddRow(expectedIDs[0]).
				AddRow(expectedIDs[1]).
				AddRow(expectedIDs[2]))

		ids, err := d.IDs(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedIDs, ids)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error in IDs retrieval", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM comics").
			WillReturnError(errors.New("query failed"))

		_, err := d.IDs(context.Background())
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDrop(t *testing.T) {
	db, mock, err := sqlxmock.Newx()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := &DB{
		conn: db,
	}

	t.Run("successful drop", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM comics").
			WillReturnResult(sqlxmock.NewResult(0, 0))

		err := d.Drop(context.Background())
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error in drop", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM comics").
			WillReturnError(errors.New("delete failed"))

		err := d.Drop(context.Background())
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

var sqlxConnect = sqlx.Connect
