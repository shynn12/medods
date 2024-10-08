package item

import (
	"context"
)

type Storage interface {
	Create(ctx context.Context, item ItemDTO) (string, error)
	FindOne(ctx context.Context, id string) (*Item, error)
	UpdateToken(ctx context.Context, id string, cryptedtoken string) error
	Clear(ctx context.Context) error
	Close(ctx context.Context)
}
