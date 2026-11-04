package session

import (
	"context"
	"errors"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestIdentityTransactionPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	for _, q := range []string{
		`CREATE TEMP TABLE users (id bigserial PRIMARY KEY, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz, name text, role text, email text UNIQUE, foto_url text)`,
		`CREATE TEMP TABLE user_auths (id bigserial PRIMARY KEY, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz, user_id bigint REFERENCES users(id), provider text, provider_user_id text CHECK (provider_user_id <> 'reject'), UNIQUE(provider, provider_user_id))`,
	} {
		if err := db.Exec(q).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(db)
	ctx := context.Background()
	identity := Identity{Subject: "subject", Email: "one@example.com", Name: "One"}
	first, err := repo.FindOrCreate(ctx, identity)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.FindOrCreate(ctx, identity)
	if err != nil || first.ID != second.ID {
		t.Fatalf("identity not stable: %v", err)
	}
	identity.Email = "changed@example.com"
	second, err = repo.FindOrCreate(ctx, identity)
	if err != nil || first.ID != second.ID {
		t.Fatalf("subject not used: %v", err)
	}
	if _, err := repo.FindOrCreate(ctx, Identity{Subject: "other", Email: "one@example.com"}); !errors.Is(err, ErrAccountConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err := repo.FindOrCreate(ctx, Identity{Subject: "reject", Email: "rollback@example.com"}); err == nil {
		t.Fatal("expected identity insert failure")
	}
	var count int64
	if err := db.Table("users").Where("email = ?", "rollback@example.com").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("account was not rolled back: count=%d err=%v", count, err)
	}
}
