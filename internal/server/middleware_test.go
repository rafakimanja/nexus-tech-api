package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api/internal/domain/session"
	"api/internal/domain/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type accountRepository struct {
	profile session.Profile
	err     error
}

func (r *accountRepository) FindOrCreate(context.Context, session.Identity) (user.User, error) {
	return user.User{}, nil
}
func (r *accountRepository) Current(context.Context, uint) (session.Profile, error) {
	return r.profile, r.err
}

func TestMiddlewareUsesCurrentAccountState(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", strings.Repeat("s", 32))
	account := user.User{Model: gorm.Model{ID: 1}, Role: "professor"}
	token, err := account.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path string
		repo       *accountRepository
		want       int
	}{
		{"active", "/protected", &accountRepository{profile: session.Profile{ID: 1, Role: "demandante"}}, 200},
		{"blocked", "/protected", &accountRepository{profile: session.Profile{ID: 1, Blocked: true}}, 403},
		{"blocked me", "/me", &accountRepository{profile: session.Profile{ID: 1, Blocked: true, BlockReason: "strike_limit"}}, 200},
		{"deleted", "/me", &accountRepository{err: gorm.ErrRecordNotFound}, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{sessionService: session.NewService(tc.repo, "", nil)}
			r := gin.New()
			r.GET("/me", s.AuthMiddleware(), session.NewHandler(s.sessionService, nil).Me)
			r.GET("/protected", s.AuthMiddleware(), func(c *gin.Context) {
				if role, _ := c.Get("role"); role != "demandante" {
					t.Errorf("stale JWT role used: %v", role)
				}
				c.Status(200)
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.name == "blocked me" && !strings.Contains(w.Body.String(), `"blocked":true`) {
				t.Fatal(w.Body.String())
			}
		})
	}
}
