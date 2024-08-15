package item

import (
	"context"
	"time"
)

type Storage interface {
	Create(ctx context.Context, item ItemDTO) (string, error)
	FindOne(ctx context.Context, id string) (*Item, error)
	UpdateToken(ctx context.Context, id string, cryptedtoken string, time time.Time) error
	Clear(ctx context.Context) error
}
