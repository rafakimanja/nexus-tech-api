package applicant

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	c.JSON(http.StatusOK, applicants)
}

func (h *ApplicantHandler) GetApplicantByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	applicant, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, applicant)
}

func (h *ApplicantHandler) CreateApplicant(c *gin.Context) {
	var applicant Demandante
	if err := c.ShouldBindJSON(&applicant); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Demandante' invalido")
		return
	}

	id, err := h.service.Create(applicant)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"applicantID": id})
}

func (h *ApplicantHandler) UpdateApplicant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	var applicant Demandante
	if err := c.ShouldBindJSON(&applicant); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Demandante' invalido")
		return
	}

	if err := h.service.Update(id, applicant); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ApplicantHandler) DeleteApplicant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ApplicantHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
