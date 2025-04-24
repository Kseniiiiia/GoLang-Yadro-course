package db

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
)

func TestMigrate(t *testing.T) {
	t.Run("failed to create database driver", func(t *testing.T) {
		db, mock, err := sqlxmock.Newx(sqlxmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to create mock: %v", err)
		}
		defer db.Close()

		// Setup expectations
		mock.ExpectPing().WillReturnError(errors.New("ping failed"))

		d := &DB{
			conn: db,
			log:  slog.New(slog.NewTextHandler(nil, nil)),
		}

		// Run test
		err = d.Migrate()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "ping failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}
