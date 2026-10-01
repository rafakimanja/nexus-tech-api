package user

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	JWT_SECRET_KEY = "JWT_SECRECT_KEY"
)

func (u *User) GenerateToken() (string, error){
	secret := os.Getenv(JWT_SECRET_KEY)

	claims := jwt.MapClaims{
		"id": u.ID,
		"role": u.Role,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", fmt.Errorf("Erro ao gerar token JWT, err=%s", err.Error())
	}

	return tokenStr, nil
}

func VerifyToken(tokenValue string) (uint, error) {
	secret := os.Getenv(JWT_SECRET_KEY)

	token, err := jwt.Parse(RemoveBearerPrefix(tokenValue), func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
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
	if !ok {
		return 0, fmt.Errorf("invalid id claim")
	}

	return uint(id), nil
}

func RemoveBearerPrefix(token string) string {
	return strings.TrimPrefix(token, "Bearer ")
}