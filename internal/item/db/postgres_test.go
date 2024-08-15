package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/shynn12/medods/internal/item"
	"github.com/shynn12/medods/internal/item/db"
	"github.com/shynn12/medods/pkg/client/postgres"
)

const dbURL = "postgres://postgres:psql@localhost:5432/medods"

func TestDB(t *testing.T) {
	pool, err := postgres.NewClient(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}

	db := db.NewStorage(pool)

	nTime := time.Now()

	id, err := db.Create(context.Background(), item.ItemDTO{Email: "2@mail.ru", Refresh: "", ExpiresAt: nTime})
	if err != nil {
		t.Fatal(err)
	}

	unit, err := db.FindOne(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	want := item.Item{
		ID:        id,
		Email:     "2@mail.ru",
		ExpiresAt: nTime,
	}
	t.Log(*unit, want)
	if want.Email != unit.Email || want.ExpiresAt.Unix() != unit.ExpiresAt.Unix() || want.ID != unit.ID || want.Refresh != unit.Refresh {
		t.Fatal("Uncorrect reading")
	}

	_ = db.Clear(context.Background())

	unit, _ = db.FindOne(context.Background(), id)

	if unit != nil {
		t.Fatal()
	}
}
