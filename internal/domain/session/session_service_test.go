package session

import (
	"context"
	"errors"
	"strings"
	"testing"

	"api/internal/domain/user"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type fakeRepository struct {
	identity Identity
	calls    int
	err      error
	profile  Profile
}

func (r *fakeRepository) FindOrCreate(_ context.Context, i Identity) (user.User, error) {
	r.identity = i
	r.calls++
	return user.User{Model: gorm.Model{ID: 1}}, r.err
}
func (r *fakeRepository) Current(context.Context, uint) (Profile, error) { return r.profile, nil }

func TestLoginContract(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", strings.Repeat("s", 32))
	r := &fakeRepository{profile: Profile{ID: 1, Role: "professor"}}
	s := NewService(r, "client-id", func(_ context.Context, token, audience string) (*idtoken.Payload, error) {
		if token != "google-token" || audience != "client-id" {
			t.Fatal("incorrect validation arguments")
		}
		return &idtoken.Payload{Subject: "google-subject", Claims: map[string]interface{}{"email": "USER@example.com", "email_verified": true}}, nil
	})
	result, err := s.Login(context.Background(), "google-token")
	if err != nil {
		t.Fatal(err)
	}
	if result.ExpiresIn != 43200 || result.TokenType != "Bearer" || result.UserID != 1 || result.User.Role != "professor" {
		t.Fatalf("unexpected response: %+v", result)
	}
	if r.identity.Subject != "google-subject" || r.identity.Email != "user@example.com" {
		t.Fatalf("unexpected identity: %+v", r.identity)
	}
	if _, err := user.VerifyToken(result.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestLoginRejectsUnverifiedIdentityBeforePersistence(t *testing.T) {
	r := &fakeRepository{}
	s := NewService(r, "client", func(context.Context, string, string) (*idtoken.Payload, error) {
		return &idtoken.Payload{Subject: "subject", Claims: map[string]interface{}{"email": "user@example.com", "email_verified": false}}, nil
	})
	if _, err := s.Login(context.Background(), "token"); !errors.Is(err, ErrIdentity) || r.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, r.calls)
	}
}

func TestLoginDoesNotIssueTokenOnConflictOrBlock(t *testing.T) {
	for _, tc := range []struct {
		name     string
		repo     *fakeRepository
		expected error
	}{
		{"conflict", &fakeRepository{err: ErrAccountConflict}, ErrAccountConflict},
		{"blocked", &fakeRepository{profile: Profile{Blocked: true}}, ErrBlocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewService(tc.repo, "client", func(context.Context, string, string) (*idtoken.Payload, error) {
				return &idtoken.Payload{Subject: "subject", Claims: map[string]interface{}{"email": "user@example.com", "email_verified": true}}, nil
			})
			result, err := s.Login(context.Background(), "token")
			if !errors.Is(err, tc.expected) || result.AccessToken != "" {
				t.Fatalf("err=%v token issued=%v", err, result.AccessToken != "")
			}
		})
	}
}
