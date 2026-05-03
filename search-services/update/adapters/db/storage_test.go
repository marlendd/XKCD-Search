package db_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"yadro.com/course/update/adapters/db"
	"yadro.com/course/update/core"
)

func newDB(t *testing.T) (*db.DB, sqlmock.Sqlmock) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = conn.Close()
	})
	return db.NewFromConn(slog.Default(), sqlx.NewDb(conn, "pgx")), mock
}

func TestAdd(t *testing.T) {
	d, mock := newDB(t)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO comics").
			WithArgs(1, "http://example.com", pq.Array([]string{"golang"})).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := d.Add(context.Background(), core.Comics{
			ID:    1,
			URL:   "http://example.com",
			Words: []string{"golang"},
		})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO comics").
			WillReturnError(errors.New("db error"))

		err := d.Add(context.Background(), core.Comics{ID: 1})
		assert.Error(t, err)
	})
}

func TestStats(t *testing.T) {
	d, mock := newDB(t)

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT COALESCE").
			WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(100))
		mock.ExpectQuery("SELECT COUNT.DISTINCT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))
		mock.ExpectQuery("SELECT COUNT.id").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		stats, err := d.Stats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, core.DBStats{
			WordsTotal:    100,
			WordsUnique:   50,
			ComicsFetched: 10,
		}, stats)
	})

	t.Run("error on first query", func(t *testing.T) {
		mock.ExpectQuery("SELECT COALESCE").
			WillReturnError(errors.New("db error"))

		_, err := d.Stats(context.Background())
		assert.Error(t, err)
	})
}

func TestIDs(t *testing.T) {
	d, mock := newDB(t)

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM comics").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1).AddRow(2).AddRow(3))

		ids, err := d.IDs(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, ids)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM comics").
			WillReturnError(errors.New("db error"))

		_, err := d.IDs(context.Background())
		assert.Error(t, err)
	})
}

func TestDrop(t *testing.T) {
	d, mock := newDB(t)

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("TRUNCATE TABLE comics").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := d.Drop(context.Background())
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec("TRUNCATE TABLE comics").
			WillReturnError(errors.New("db error"))

		err := d.Drop(context.Background())
		assert.Error(t, err)
	})
}