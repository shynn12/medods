package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shynn12/medods/internal/item"
)

type db struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) item.Storage {
	return &db{
		pool: pool,
	}
}

func (d *db) FindOne(ctx context.Context, id string) (i *item.Item, err error) {
	i = &item.Item{}
	err = d.pool.QueryRow(ctx, "SELECT uid, email, token, expiresat from users where uid = $1;", id).Scan(&i.ID, &i.Email, &i.Refresh, &i.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (d *db) Create(ctx context.Context, item item.ItemDTO) (id string, err error) {
	err = d.pool.QueryRow(ctx, "INSERT INTO USERS (email, token, expiresat) VALUES($1, $2, $3) RETURNING uid", item.Email, item.Refresh, item.ExpiresAt).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (d *db) UpdateToken(ctx context.Context, id string, cryptedtoken string, time time.Time) error {
	row, err := d.pool.Exec(ctx, "UPDATE users SET token = $1, expiresat = $2 WHERE uid = $3;", cryptedtoken, time, id)
	if row.RowsAffected() == 0 {
		return fmt.Errorf("no such uid in the db")
	}
	return err
}

func (d *db) Clear(ctx context.Context) error {
	_, err := d.pool.Exec(ctx, "DELETE FROM USERS;")
	return err
}
