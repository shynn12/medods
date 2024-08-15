package db

import (
	"context"
	"fmt"

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
	err = d.pool.QueryRow(ctx, "SELECT uid, email, token from users where uid = $1;", id).Scan(&i.ID, &i.Email, &i.Refresh)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (d *db) Create(ctx context.Context, item item.ItemDTO) (id string, err error) {
	err = d.pool.QueryRow(ctx, "INSERT INTO USERS (email, token) VALUES($1, $2) RETURNING uid;", item.Email, item.Refresh).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (d *db) UpdateToken(ctx context.Context, id string, cryptedtoken string) error {
	row, err := d.pool.Exec(ctx, "UPDATE users SET token = $1 WHERE uid = $2;", cryptedtoken, id)
	if err != nil {
		return err
	}

	if row.RowsAffected() == 0 {
		return fmt.Errorf("no such uid in the db")
	}

	return nil
}

func (d *db) Close(ctx context.Context) {
	d.pool.Close()
}

func (d *db) Clear(ctx context.Context) error {
	_, err := d.pool.Exec(ctx, "DELETE FROM USERS;")
	return err
}
