package user

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

func TestSessionTokenTwelveHours(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", strings.Repeat("s", 32))
	u := User{Model: gorm.Model{ID: 7}, Role: "professor"}
	value, expires, err := u.GenerateSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if delta := time.Until(expires); delta < 12*time.Hour-2*time.Second || delta > 12*time.Hour {
		t.Fatalf("unexpected expiry: %v", expires)
	}
	id, err := VerifyToken("Bearer " + value)
	if err != nil || id != 7 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestVerifyRejectsInvalidClaimsAndAlgorithm(t *testing.T) {
	secret := strings.Repeat("s", 32)
	t.Setenv("JWT_SECRET_KEY", secret)
	for _, tc := range []struct {
		name   string
		method jwt.SigningMethod
		claims jwt.MapClaims
	}{
		{"expired", jwt.SigningMethodHS256, jwt.MapClaims{"id": 1, "exp": time.Now().Add(-time.Hour).Unix()}},
		{"missing expiry", jwt.SigningMethodHS256, jwt.MapClaims{"id": 1}},
		{"fractional id", jwt.SigningMethodHS256, jwt.MapClaims{"id": 1.5, "exp": time.Now().Add(time.Hour).Unix()}},
		{"negative id", jwt.SigningMethodHS256, jwt.MapClaims{"id": -1, "exp": time.Now().Add(time.Hour).Unix()}},
		{"wrong algorithm", jwt.SigningMethodHS512, jwt.MapClaims{"id": 1, "exp": time.Now().Add(time.Hour).Unix()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := jwt.NewWithClaims(tc.method, tc.claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := VerifyToken(value); err == nil {
				t.Fatal("token accepted")
			}
		})
	}
}

func TestRejectsEmptyOrShortSecret(t *testing.T) {
	for _, secret := range []string{"", "short", strings.Repeat(" ", 32)} {
		t.Setenv("JWT_SECRET_KEY", secret)
		if ValidateJWTConfig() == nil {
			t.Fatal("invalid secret accepted")
		}
		if _, err := (&User{}).GenerateToken(); err == nil {
			t.Fatal("signed with invalid secret")
		}
	}
}
