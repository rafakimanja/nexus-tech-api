package teacher

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeacherHandler struct {
	service TeacherService
	logger  *log.Logger
}

func NewHandler(service TeacherService, logger *log.Logger) *TeacherHandler {
	return &TeacherHandler{service: service, logger: logger}
}

func (h *TeacherHandler) GetTeachers(c *gin.Context) {
	professores, err := h.service.GetAll()
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]ProfessorDTO, 0, len(professores))
	for _, professor := range professores {
		dtos = append(dtos, professor.NewDTO())
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *TeacherHandler) GetTeacherByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	professor, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.respondError(c, http.StatusNotFound, "professor not found")
			return
		}
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, professor.NewDTO())
}

func (h *TeacherHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
