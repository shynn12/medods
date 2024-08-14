package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewClient(ctx context.Context, dbURL string) (pool *pgxpool.Pool, err error) {
	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot ping DB due to error: %v", err)
	}

	return dbpool, err
}
