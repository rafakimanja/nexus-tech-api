package server

import (
	"api/internal/domain/uauth"
	"api/internal/domain/user"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
) 

func (s *Server) loginHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Entrar</title>
</head>
<body>
  <main>
    <h1>Entrar</h1>
    <a href="/auth/google">Entrar com Google</a>
  </main>
</body>
</html>`))
}

func (s *Server) GetAuthGoogle(c *gin.Context) {
	request := c.Request.Clone(c.Request.Context())
	query := request.URL.Query()
	query.Set("provider", "google")
	request.URL.RawQuery = query.Encode()
	gothic.BeginAuthHandler(c.Writer, request)
}

func (s *Server) GetAuthGoogleCallback(c *gin.Context) {
	var newUser user.User
	var newUserAuth uauth.UserAuth

	request := c.Request.Clone(c.Request.Context())
	query := request.URL.Query()
	query.Set("provider", "google")
	request.URL.RawQuery = query.Encode()

	user, err := gothic.CompleteUserAuth(c.Writer, request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Google authentication failed"})
		return
	}

	fmt.Println("ID TOKEN:", user.IDToken)

	newUser.Name = user.Name
	newUser.Email = user.Email
	newUser.Foto_url = &user.AvatarURL

	id, err := s.userService.Create(newUser)

	if err != nil || id == 0 {
		s.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}

	newUserAuth.UserID = id
	newUserAuth.ProviderUserID = user.UserID
	newUserAuth.Provider = user.Provider

	_, err = s.userAuthService.Create(newUserAuth)

	if err != nil {
		s.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}

	token, err := newUser.GenerateToken()
	if err != nil {
		s.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}

	c.Header("Authorization", token)
	c.JSON(http.StatusOK, newUser)
}