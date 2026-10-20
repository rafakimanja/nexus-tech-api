package areas

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AreasHandler struct {
	service AreasService
	logger  *log.Logger
}

func NewHandler(service AreasService, logger *log.Logger) *AreasHandler {
	return &AreasHandler{service: service, logger: logger}
}

func (h *AreasHandler) GetAreas(c *gin.Context) {
	areas, err := h.service.GetAll()
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, areas)
}

func (h *AreasHandler) GetAreaByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	area, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, area)
}

func (h *AreasHandler) CreateArea(c *gin.Context) {
	var area Area
	if err := c.ShouldBindJSON(&area); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Area' invalido")
		return
	}

	id, err := h.service.Create(area)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"areaID": id})
}

func (h *AreasHandler) UpdateArea(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	var area Area
	if err := c.ShouldBindJSON(&area); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Area' invalido")
		return
	}

	if err := h.service.Update(id, area); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AreasHandler) DeleteArea(c *gin.Context) {
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

func (h *AreasHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
