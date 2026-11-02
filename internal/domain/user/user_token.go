package user

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const SessionDuration = 12 * time.Hour

func jwtSecret() ([]byte, error) {
	secret, configured := os.LookupEnv("JWT_SECRET_KEY")
	if !configured {
		secret = os.Getenv("JWT_SECRECT_KEY")
	}
	if len(strings.TrimSpace(secret)) < 32 {
		return nil, fmt.Errorf("JWT_SECRET_KEY must contain at least 32 characters")
	}
	return []byte(secret), nil
}

func ValidateJWTConfig() error { _, err := jwtSecret(); return err }

func (u *User) GenerateSessionToken() (string, time.Time, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(SessionDuration)

	claims := jwt.MapClaims{
		"id":   u.ID,
		"role": u.Role,
		"iat":  now.Unix(),
		"exp":  expires.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", time.Time{}, fmt.Errorf("Erro ao gerar token JWT, err=%s", err.Error())
	}

	return tokenStr, expires, nil
}

func (u *User) GenerateToken() (string, error) {
	token, _, err := u.GenerateSessionToken()
	return token, err
}

func VerifyToken(tokenValue string) (uint, error) {
	secret, err := jwtSecret()
	if err != nil {
		return 0, err
	}

	token, err := jwt.Parse(RemoveBearerPrefix(tokenValue), func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method")
		}

		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	id, ok := claims["id"].(float64)
	if !ok || id < 1 || id > math.MaxUint32 || id != math.Trunc(id) {
		return 0, fmt.Errorf("invalid id claim")
	}

	exp, ok := claims["exp"].(float64)
	if !ok || exp <= float64(time.Now().Unix()) {
		return 0, fmt.Errorf("invalid or expired token")
	}
	return uint(id), nil
}

func RemoveBearerPrefix(token string) string {
	return strings.TrimPrefix(token, "Bearer ")
}
