package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Open("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Search(ctx context.Context, keys []string, limit int) ([]core.Comic, error) {
	query := `SELECT id, url,
    (SELECT COUNT(DISTINCT w) FROM unnest(words) w WHERE w = ANY($1))::float
		/ cardinality($1) AS score
	FROM comics
	WHERE words && $1
	ORDER BY score DESC
	LIMIT $2
	`
	rows, err := db.conn.QueryContext(ctx, query, pq.Array(keys), limit)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			db.log.Error("failed to close rows", "error", err)
		}
	}()

	var res []core.Comic
	for rows.Next() {
		var c core.Comic
		var score float64
		if err := rows.Scan(&c.ID, &c.URL, &score); err != nil {
			return nil, err
		}
		res = append(res, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (db *DB) AllComics(ctx context.Context) ([]core.StoredComic, error) {
	query := `SELECT id, url, words FROM comics`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			db.log.Error("failed to close rows", "error", err)
		}
	}()

	var res []core.StoredComic
	for rows.Next() {
		var c core.StoredComic
		if err := rows.Scan(&c.ID, &c.URL, pq.Array(&c.Words)); err != nil {
			return nil, err
		}
		res = append(res, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}