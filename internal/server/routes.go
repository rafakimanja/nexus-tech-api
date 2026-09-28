package server

import (
	"api/internal/domain/uauth"
	"api/internal/domain/user"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	userRoutes := user.NewRouter(s.userHandler)

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)
	r.GET("/login", s.loginHandler)
	r.GET("/auth/google", s.GetAuthGoogle)
	r.GET("/auth/google/callback", s.GetAuthGoogleCallback)

	userRoutes.RegisterRoutes(r.Group("/users"))

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

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

	idAuth, err := s.userAuthService.Create(newUserAuth)

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

	c.JSON(http.StatusOK, gin.H{
		"userID":     id,
		"userAuthID": idAuth,
	})
}
