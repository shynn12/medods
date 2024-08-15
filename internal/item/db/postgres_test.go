package db_test

import (
	"context"
	"testing"

	"github.com/shynn12/medods/internal/item"
	"github.com/shynn12/medods/internal/item/db"
	"github.com/shynn12/medods/pkg/client/postgres"
)

const dbURL = "postgres://postgres:psql@localhost:5432/medods1"

func TestDB(t *testing.T) {
	pool, err := postgres.NewClient(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}

	db := db.NewStorage(pool)

	id, err := db.Create(context.Background(), item.ItemDTO{Email: "2@mail.ru", Refresh: ""})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(context.Background())
	unit, err := db.FindOne(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	want := item.Item{
		ID:    id,
		Email: "2@mail.ru",
	}
	t.Log(*unit, want)
	if want.Email != unit.Email || want.ID != unit.ID || want.Refresh != unit.Refresh {
		t.Fatal("Uncorrect reading")
	}
}
