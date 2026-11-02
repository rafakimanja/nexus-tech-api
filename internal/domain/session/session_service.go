package session

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"api/internal/domain/user"
	"google.golang.org/api/idtoken"
)

var ErrIdentity = errors.New("invalid Google identity")
var ErrBlocked = errors.New("account blocked")

type TokenValidator func(context.Context, string, string) (*idtoken.Payload, error)
type Service struct {
	repo     Repository
	audience string
	validate TokenValidator
}

func NewService(repo Repository, audience string, validate TokenValidator) *Service {
	return &Service{repo: repo, audience: audience, validate: validate}
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	ExpiresAt   time.Time `json:"expires_at"`
	UserID      uint      `json:"user_id"`
	User        Profile   `json:"user"`
}

func (s *Service) Login(ctx context.Context, token string) (LoginResponse, error) {
	var result LoginResponse
	if s.audience == "" || strings.TrimSpace(token) == "" {
		return result, ErrIdentity
	}
	payload, err := s.validate(ctx, token, s.audience)
	if err != nil || payload == nil {
		return result, ErrIdentity
	}
	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || !verified || payload.Subject == "" {
		return result, ErrIdentity
	}
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	account, err := s.repo.FindOrCreate(ctx, Identity{Subject: payload.Subject, Email: email, Name: name, Picture: picture})
	if err != nil {
		return result, err
	}
	profile, err := s.repo.Current(ctx, account.ID)
	if err != nil {
		return result, err
	}
	if profile.Blocked {
		return result, ErrBlocked
	}
	account.Role = profile.Role
	signed, expires, err := account.GenerateSessionToken()
	if err != nil {
		return result, err
	}
	return LoginResponse{AccessToken: signed, TokenType: "Bearer", ExpiresIn: int64(user.SessionDuration / time.Second), ExpiresAt: expires, UserID: account.ID, User: profile}, nil
}

func (s *Service) Current(ctx context.Context, id uint) (Profile, error) {
	return s.repo.Current(ctx, id)
}
