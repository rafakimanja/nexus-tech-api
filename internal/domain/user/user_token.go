package user

import (
	"fmt"
	"os"
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