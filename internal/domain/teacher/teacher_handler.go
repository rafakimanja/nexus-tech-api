package teacher

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TeacherHandler struct {
	service TeacherService
	logger  *log.Logger
}

func NewHandler(service TeacherService, logger *log.Logger) *TeacherHandler {
	return &TeacherHandler{service: service, logger: logger}
}

func (h *TeacherHandler) GetTeachers(c *gin.Context) {
	professors, err := h.service.GetAll()
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, professors)
}

func (h *TeacherHandler) GetTeacherByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	professor, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, professor)
}

func (h *TeacherHandler) CreateTeacher(c *gin.Context) {
	var professor Professor
	if err := c.ShouldBindJSON(&professor); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Professor' invalido")
		return
	}

	id, err := h.service.Create(professor)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if id == 0 {
		h.respondError(c, http.StatusInternalServerError, "falha ao criar professor")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"teacherID": id})
}

func (h *TeacherHandler) UpdateTeacher(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	var professor Professor
	if err := c.ShouldBindJSON(&professor); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Professor' invalido")
		return
	}

	if err := h.service.Update(id, professor); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TeacherHandler) DeleteTeacher(c *gin.Context) {
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

func (h *TeacherHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
