package demand

import (
	"api/internal/server"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DemandaHandler struct {
	service DemandaService
	embed server.Emb
	logger  *log.Logger
}

func NewHandler(service DemandaService, embed server.Emb, logger *log.Logger) *DemandaHandler {
	return &DemandaHandler{service: service, embed: embed, logger: logger}
}

func (h *DemandaHandler) GetDemandas(c *gin.Context) {
	demandas, err := h.service.GetAll()
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, demandas)
}

func (h *DemandaHandler) GetDemandaByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	demanda, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, demanda)
}

func (h *DemandaHandler) CreateDemanda(c *gin.Context) {
	var demanda Demanda
	if err := c.ShouldBindJSON(&demanda); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Demanda' invalido")
		return
	}

	id, err := h.service.Create(demanda)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"demandaID": id})
}

func (h *DemandaHandler) UpdateDemanda(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	var demanda Demanda
	if err := c.ShouldBindJSON(&demanda); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'Demanda' invalido")
		return
	}

	if err := h.service.Update(id, demanda); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DemandaHandler) DeleteDemanda(c *gin.Context) {
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

func (h *DemandaHandler) GetSuggestDemand(c *gin.Context){
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "invalid ID")
		return
	}

	demanda, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if demanda.IsCancelled == true {
		h.respondError(c, http.StatusBadRequest, "Esta demanda foi cancelada")
		return
	}

	h.embed.CompareEmbeddings(demanda.Embedding)

	c.Status(http.StatusNoContent)
}

func (h *DemandaHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
