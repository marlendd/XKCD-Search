package db

import (
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	query := `INSERT INTO comics VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	_, err := db.conn.ExecContext(ctx, query, comics.ID, comics.URL, pq.Array(comics.Words))
	return err
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	queries := []struct {
		q   string
		dst *int
	}{
		{`SELECT COALESCE(SUM(array_length(words, 1)), 0) FROM comics`, &stats.WordsTotal},
		{`SELECT COUNT(DISTINCT word) FROM comics, unnest(words) AS word`, &stats.WordsUnique},
		{`SELECT COUNT(id) FROM comics`, &stats.ComicsFetched},
	}

	for _, query := range queries {
		if err := db.conn.QueryRowContext(ctx, query.q).Scan(query.dst); err != nil {
			return core.DBStats{}, err
		}
	}

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, `SELECT id FROM comics`)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, `TRUNCATE TABLE comics`)
	return err
}
