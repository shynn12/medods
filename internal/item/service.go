package item

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	GetOneByGuid(ctx context.Context, guid string) (*Item, error)
	// CreateItem(ctx context.Context, dto ItemDTO) (string, error)
	UpdateToken(ctx context.Context, id string, token string, time time.Time) error
}

type service struct {
	storage Storage
}

func NewService(storage Storage) Service {
	return &service{storage: storage}
}

func (s *service) GetOneByGuid(ctx context.Context, guid string) (*Item, error) {
	item, err := s.storage.FindOne(ctx, guid)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// func (s *service) CreateItem(ctx context.Context, dto ItemDTO) (string, error) {
// 	id, err := s.storage.Create(ctx, dto)
// 	if err != nil {
// 		return "", err
// 	}
// 	return id, nil
// }

func (s *service) UpdateToken(ctx context.Context, id string, token string, time time.Time) error {
	cryptedtoken, err := bcrypt.GenerateFromPassword([]byte(token), 5)
	if err != nil {
		return err
	}

	err = s.storage.UpdateToken(ctx, id, string(cryptedtoken), time)

	return err
}
