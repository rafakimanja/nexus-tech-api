package server

import (
	"api/internal/domain/uauth"
	"api/internal/domain/user"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	r.GET("/health", s.healthHandler)
	r.GET("/", s.HelloWorldHandler)

	r.GET("/login", s.loginHandler)
	r.POST("/login/v2", s.Login)
	r.GET("/auth/google", s.GetAuthGoogle)
	r.GET("/auth/google/callback", s.GetAuthGoogleCallback)

	userRoutes := user.NewRouter(s.userHandler)
	usersGroup := r.Group("/users")
	usersGroup.Use(s.AuthMiddleware())
	userRoutes.RegisterRoutes(usersGroup)

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

func (s *Server) Login(c *gin.Context) {
	var newUser user.User
	var newUserAuth uauth.UserAuth
	googleClientID := os.Getenv(GOOGLE_CLIENT_ID)

	var req struct {
		IDToken string `json:"idToken"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    http.StatusBadRequest,
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

	newUser.Name = payload.Claims["name"].(string)
	newUser.Email = payload.Claims["email"].(string)
	*newUser.Foto_url = payload.Claims["picture"].(string)

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
	newUserAuth.ProviderUserID = req.IDToken
	newUserAuth.Provider = "google"

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