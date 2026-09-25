package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	url_callback = "http://localhost:3333/auth/google/callback"
	providerName = "google"
)

// NewAuth configures the Google OAuth provider. It must be called before the
// server starts accepting requests.
func NewAuth() error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	sessionSecret := os.Getenv("SESSION_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}
	if len(sessionSecret) < 32 {
		return fmt.Errorf("SESSION_SECRET must contain at least 32 characters")
	}

	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	}
	gothic.Store = store

	goth.UseProviders(google.New(clientID, clientSecret, url_callback, "openid", "email", "profile"))
	return nil
}
