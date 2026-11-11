package server

import (
	"api/internal/access"
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/areas_teacher"
	"api/internal/domain/demand"
	"api/internal/domain/notification"
	"api/internal/domain/profile"
	"api/internal/domain/recommendation"
	"api/internal/domain/report"
	"api/internal/domain/session"
	"api/internal/domain/teacher"
	"api/internal/domain/user"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/health", s.healthHandler)
	registerDocumentation(r)
	r.GET("/", s.HelloWorldHandler)

	r.GET("/login", s.loginHandler)
	sessionHandler := session.NewHandler(s.sessionService, s.logger)
	r.POST("/login/v2", sessionHandler.Login)
	r.GET("/me", s.AuthMiddleware(), sessionHandler.Me)
	notification.NewHandler(s.db.DB()).Register(r.Group("/notifications", s.AuthMiddleware()))
	reportHandler := report.NewHandler(s.db.DB())
	reports := r.Group("/reports", s.AuthMiddleware())
	reports.POST("", reportHandler.Create)
	reports.GET("", reportHandler.List)
	reports.GET("/:id", reportHandler.Get)
	reports.PATCH("/:id/judgment", reportHandler.Judge)
	r.GET("/auth/google", s.GetAuthGoogle)
	r.GET("/auth/google/callback", s.GetAuthGoogleCallback)

	userRoutes := user.NewRouter(s.userHandler)
	usersGroup := r.Group("/users")
	usersGroup.Use(s.AuthMiddleware())
	usersGroup.Use(access.Admin())
	userRoutes.RegisterRoutes(usersGroup)

	demandRoutes := demand.NewRouter(s.demandHandler)
	demandsGroup := r.Group("/demands")
	demandsGroup.Use(s.AuthMiddleware())
	demandRoutes.RegisterRoutes(demandsGroup)

	applicantRoutes := applicant.NewRouter(s.applicantHandler)
	applicantsGroup := r.Group("/applicants")
	applicantsGroup.Use(s.AuthMiddleware())
	applicantRoutes.RegisterRoutes(applicantsGroup)
	profileHandler := profile.NewHandler(profile.NewService(profile.NewRepository(s.db.DB())), s.logger)
	applicantsGroup.POST("/", profileHandler.SaveApplicant)
	applicantsGroup.PUT("/:id", profileHandler.SaveApplicant)

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
	teachersGroup.POST("/", profileHandler.SaveTeacher)
	teachersGroup.PUT("/:id", profileHandler.SaveTeacher)
	teachersGroup.PUT("/:id/areas", s.areasTeacherHandler.ReplaceByProfessorID)
	teachersGroup.GET("/:id/recommendations", recommendation.NewHandler(s.db.DB()).Get)

	return r
}

// CORS_ALLOWED_ORIGINS is a comma-separated list of exact browser origins.
// Preserve the existing local development origin when it is not configured.
func allowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:5173"}
	}
	origins := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			origins = append(origins, value)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:5173"}
	}
	return origins
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
