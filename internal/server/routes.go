package server

import (
	"api/internal/domain/uauth"
	"api/internal/domain/user"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"google.golang.org/api/idtoken"
)

var (
	GOOGLE_CLIENT_ID = "GOOGLE_CLIENT_ID"
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
	r.POST("/login/v2", s.Login)
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

func (s *Server) Login(c *gin.Context){
	googleClientID := os.Getenv(GOOGLE_CLIENT_ID)

	var req struct {
		IDToken string `json:"idToken"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": http.StatusBadRequest,
				"message": err.Error(),
			},
		})
		return
	}

	payload, err := idtoken.Validate(
		c.Request.Context(),
		req.IDToken,
		googleClientID,
	)

	if err != nil {
		s.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    http.StatusUnauthorized,
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email": payload.Claims["email"],
		"name": payload.Claims["name"],
		"foto_url": payload.Claims["picture"],
	})
}