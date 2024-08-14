package db

import (
	"context"
	"log"
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
	err = d.pool.QueryRow(ctx, "SELECT uid, email, token from users where uid = $1", id).Scan(&i.ID, &i.Email, &i.Refresh)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// func (d *db) Create(ctx context.Context, item item.ItemDTO) (id string, err error) {
// 	row := d.pool.QueryRow(ctx, "INSERT INTO USERS (email, ) ")

// 	err = row.Scan(id)
// 	if err != nil {
// 		return "", err
// 	}

// 	return id, nil
// }

func (d *db) UpdateToken(ctx context.Context, id string, cryptedtoken string, time time.Time) error {
	log.Println(len(cryptedtoken))
	_, err := d.pool.Exec(ctx, "UPDATE users SET token = $1, expiresat = $2 WHERE uid = $3", cryptedtoken, time, id)
	return err
}
