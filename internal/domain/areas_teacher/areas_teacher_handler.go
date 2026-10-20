package areas_teacher

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AreasTeacherHandler struct {
	service AreasTeacherService
	logger  *log.Logger
}

func NewHandler(service AreasTeacherService, logger *log.Logger) *AreasTeacherHandler {
	return &AreasTeacherHandler{service: service, logger: logger}
}

func (h *AreasTeacherHandler) GetByProfessorID(c *gin.Context) {
	professorID, ok := parseProfessorID(c)
	if !ok {
		return
	}

	associations, err := h.service.GetByProfessorID(professorID)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]ProfessorAreaDTO, 0, len(associations))
	for _, association := range associations {
		dtos = append(dtos, association.NewDTO())
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *AreasTeacherHandler) ReplaceByProfessorID(c *gin.Context) {
	professorID, ok := parseProfessorID(c)
	if !ok {
		return
	}
	var request struct {
		Areas []AssociationInput `json:"areas"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Areas == nil {
		h.respondError(c, http.StatusBadRequest, "invalid object; provide an areas array")
		return
	}

	if err := h.service.ReplaceByProfessorID(professorID, request.Areas); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidAssociation) {
			status = http.StatusBadRequest
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		h.respondError(c, status, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func parseProfessorID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": http.StatusBadRequest, "message": "invalid professor ID"}})
		return 0, false
	}
	return id, true
}

func (h *AreasTeacherHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
