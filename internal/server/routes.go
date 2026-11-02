package server

import (
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/areas_teacher"
	"api/internal/domain/demand"
	"api/internal/domain/recommendation"
	"api/internal/domain/session"
	"api/internal/domain/teacher"
	"api/internal/domain/user"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/health", s.healthHandler)
	r.GET("/", s.HelloWorldHandler)

	r.GET("/login", s.loginHandler)
	sessionHandler := session.NewHandler(s.sessionService, s.logger)
	r.POST("/login/v2", sessionHandler.Login)
	r.GET("/me", s.AuthMiddleware(), sessionHandler.Me)
	r.GET("/auth/google", s.GetAuthGoogle)
	r.GET("/auth/google/callback", s.GetAuthGoogleCallback)

	userRoutes := user.NewRouter(s.userHandler)
	usersGroup := r.Group("/users")
	usersGroup.Use(s.AuthMiddleware())
	userRoutes.RegisterRoutes(usersGroup)

	demandRoutes := demand.NewRouter(s.demandHandler)
	demandsGroup := r.Group("/demands")
	demandsGroup.Use(s.AuthMiddleware())
	demandRoutes.RegisterRoutes(demandsGroup)

	applicantRoutes := applicant.NewRouter(s.applicantHandler)
	applicantsGroup := r.Group("/applicants")
	applicantsGroup.Use(s.AuthMiddleware())
	applicantRoutes.RegisterRoutes(applicantsGroup)

	areasRoutes := areas.NewRouter(s.areasHandler)
	areasGroup := r.Group("/areas")
	areasGroup.Use(s.AuthMiddleware())
	areasRoutes.RegisterRoutes(areasGroup)

	areasTeacherRoutes := areas_teacher.NewRouter(s.areasTeacherHandler)
	areasTeacherGroup := r.Group("/areas-teachers")
	areasTeacherGroup.Use(s.AuthMiddleware())
	areasTeacherRoutes.RegisterRoutes(areasTeacherGroup)

	teacherRoutes := teacher.NewRouter(s.teacherHandler)
	teachersGroup := r.Group("/teachers")
	teachersGroup.Use(s.AuthMiddleware())
	teacherRoutes.RegisterRoutes(teachersGroup)
	teachersGroup.GET("/:id/recommendations", recommendation.NewHandler(s.db.DB()).Get)

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
