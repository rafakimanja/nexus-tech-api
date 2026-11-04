package applicant

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicantHandler struct {
	service ApplicantService
	logger  *log.Logger
}

func NewHandler(service ApplicantService, logger *log.Logger) *ApplicantHandler {
	return &ApplicantHandler{service: service, logger: logger}
}

func (h *ApplicantHandler) GetApplicants(c *gin.Context) {
	applicants, err := h.service.GetAll()
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	dtos := make([]DemandanteDTO, 0, len(applicants))
	for _, a := range applicants {
		dtos = append(dtos, a.NewDTO())
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *ApplicantHandler) GetApplicantByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	applicant, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.respondError(c, http.StatusNotFound, "applicant not found")
			return
		}
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, applicant.NewDTO())
}

func (h *ApplicantHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
