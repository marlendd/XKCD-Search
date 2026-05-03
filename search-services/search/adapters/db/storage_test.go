package db_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"yadro.com/course/search/adapters/db"
	"yadro.com/course/search/core"
)

func newDB(t *testing.T) (*db.DB, sqlmock.Sqlmock) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = conn.Close()
	})

	sqlxDB := sqlx.NewDb(conn, "pgx")
	return db.NewFromConn(slog.Default(), sqlxDB), mock
}

func TestSearch(t *testing.T) {
	d, mock := newDB(t)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "score"}).
			AddRow(1, "http://example.com", 0.9).
			AddRow(2, "http://example2.com", 0.5)

		mock.ExpectQuery("SELECT id, url").
			WithArgs(pq.Array([]string{"golang"}), 10).
			WillReturnRows(rows)

		comics, err := d.Search(context.Background(), []string{"golang"}, 10)
		assert.NoError(t, err)
		assert.Len(t, comics, 2)
		assert.Equal(t, core.Comic{ID: 1, URL: "http://example.com"}, comics[0])
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, url").
			WillReturnError(errors.New("db error"))

		_, err := d.Search(context.Background(), []string{"golang"}, 10)
		assert.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "score"}).
			AddRow("not-int", "http://example.com", 0.9) // невалидный id

		mock.ExpectQuery("SELECT id, url").
			WillReturnRows(rows)

		_, err := d.Search(context.Background(), []string{"golang"}, 10)
		assert.Error(t, err)
	})
}

func TestAllComics(t *testing.T) {
	d, mock := newDB(t)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "words"}).
			AddRow(1, "http://example.com", pq.Array([]string{"golang", "test"}))

		mock.ExpectQuery("SELECT id, url, words").
			WillReturnRows(rows)

		comics, err := d.AllComics(context.Background())
		assert.NoError(t, err)
		assert.Len(t, comics, 1)
		assert.Equal(t, core.StoredComic{
			ID:    1,
			URL:   "http://example.com",
			Words: []string{"golang", "test"},
		}, comics[0])
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, url, words").
			WillReturnError(errors.New("db error"))

		_, err := d.AllComics(context.Background())
		assert.Error(t, err)
	})
}