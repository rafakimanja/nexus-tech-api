package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/internal/database"
	"api/internal/domain/applicant"
	"api/internal/domain/areas"
	"api/internal/domain/areas_teacher"
	"api/internal/domain/demand"
	"api/internal/domain/teacher"
	"api/internal/domain/uauth"
	"api/internal/domain/user"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
)

type Server struct {
	port                int
	userService         user.UserService
	userHandler         *user.UserHandler
	userAuthService     uauth.UAuthService
	demandService       demand.DemandaService
	demandHandler       *demand.DemandaHandler
	applicantService    applicant.ApplicantService
	applicantHandler    *applicant.ApplicantHandler
	areasService        areas.AreasService
	areasHandler        *areas.AreasHandler
	areasTeacherHandler *areas_teacher.AreasTeacherHandler
	teacherService      teacher.TeacherService
	teacherHandler      *teacher.TeacherHandler
	db                  database.Service
	logger              *log.Logger
}

func NewServer(
	db database.Service,
	logger *log.Logger,
	userService user.UserService,
	userHandler *user.UserHandler,
	userAuthService uauth.UAuthService,
	demandService demand.DemandaService,
	demandHandler *demand.DemandaHandler,
	applicantService applicant.ApplicantService,
	applicantHandler *applicant.ApplicantHandler,
	areasService areas.AreasService,
	areasHandler *areas.AreasHandler,
	areasTeacherHandler *areas_teacher.AreasTeacherHandler,
	teacherService teacher.TeacherService,
	teacherHandler *teacher.TeacherHandler,
) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:                port,
		db:                  db,
		logger:              logger,
		userService:         userService,
		userHandler:         userHandler,
		userAuthService:     userAuthService,
		demandService:       demandService,
		demandHandler:       demandHandler,
		applicantService:    applicantService,
		applicantHandler:    applicantHandler,
		areasService:        areasService,
		areasHandler:        areasHandler,
		areasTeacherHandler: areasTeacherHandler,
		teacherService:      teacherService,
		teacherHandler:      teacherHandler,
	}

	NewServer.db.DB().AutoMigrate(&user.User{})
	NewServer.db.DB().AutoMigrate(&uauth.UserAuth{})
	NewServer.db.DB().AutoMigrate(&areas.Area{})
	areaRelationsReady := migrateProfessorAreas(NewServer.db.DB())
	NewServer.db.DB().AutoMigrate(&areas_teacher.ProfessorArea{})
	NewServer.db.DB().AutoMigrate(&teacher.Professor{})
	NewServer.db.DB().AutoMigrate(&applicant.Demandante{})
	NewServer.db.DB().AutoMigrate(&demand.Demanda{})
	if areaRelationsReady {
		migrator := NewServer.db.DB().Migrator()
		if migrator.HasColumn("areas", "type") {
			migrator.DropColumn("areas", "type")
		}
		if migrator.HasColumn("areas", "key") {
			migrator.DropColumn("areas", "key")
		}
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// migrateProfessorAreas converts the previous GORM many-to-many join table to
// the explicit association schema while preserving links and their old type.
func migrateProfessorAreas(db *gorm.DB) bool {
	migrator := db.Migrator()
	if !migrator.HasTable("professor_areas") {
		return true
	}
	if migrator.HasColumn("professor_areas", "area_id") && !migrator.HasColumn("professor_areas", "areas_id") {
		if err := migrator.RenameColumn("professor_areas", "area_id", "areas_id"); err != nil {
			log.Printf("could not migrate professor_areas.area_id: %v", err)
			return false
		}
	}
	if !migrator.HasColumn("professor_areas", "type") {
		if err := db.Exec(`ALTER TABLE professor_areas ADD COLUMN "type" text`).Error; err != nil {
			log.Printf("could not add professor_areas.type: %v", err)
			return false
		}
	}
	if migrator.HasColumn("areas", "type") {
		if err := db.Exec(`UPDATE professor_areas pa SET "type" = a."type" FROM areas a WHERE a.id = pa.areas_id AND (pa."type" IS NULL OR pa."type" = '')`).Error; err != nil {
			log.Printf("could not copy existing area types: %v", err)
			return false
		}
	}
	if err := db.Exec(`UPDATE professor_areas SET "type" = 'interesse' WHERE "type" IS NULL OR "type" = ''`).Error; err != nil {
		log.Printf("could not initialize professor area types: %v", err)
		return false
	}
	if err := db.Exec(`ALTER TABLE professor_areas ALTER COLUMN "type" SET NOT NULL`).Error; err != nil {
		log.Printf("could not enforce professor_areas.type: %v", err)
		return false
	}
	return true
}
