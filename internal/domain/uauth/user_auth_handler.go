package uauth

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UAuthHandler struct {
	service UAuthService
	logger  *log.Logger
}

func NewHandler(service UAuthService, logger *log.Logger) *UAuthHandler {
	return &UAuthHandler{service: service, logger: logger}
}

func (h *UAuthHandler) GetUserAuthByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "invalid ID")
		return
	}

	uAuth, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, uAuth)
}

func (h *UAuthHandler) CreateUserAuth(c *gin.Context) {
	var uAuth UserAuth
	if err := c.ShouldBindJSON(&uAuth); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'User Auth' invalido")
		return
	}

	id, err := h.service.Create(uAuth)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if id == 0 {
		h.respondError(c, http.StatusInternalServerError, "falha ao criar autenticação do usuário")
		return
	}
	c.JSON(http.StatusOK, gin.H{"UserAuthID": id})
}

func (h *UAuthHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
