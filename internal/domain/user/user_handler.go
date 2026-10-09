package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service UserService
	logger  *log.Logger
}

func NewHandler(service UserService, logger *log.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetAll()
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "invalid ID")
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		h.respondError(c, http.StatusBadRequest, "objeto 'User' invalido")
		return
	}

	id, err := h.service.Create(user)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if id == 0 {
		h.respondError(c, http.StatusInternalServerError, "falha ao criar usuário")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"userID": id})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "invalid ID")
		return
	}

	if err := h.service.DeleteByID(id); err != nil {
		h.respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (h *UserHandler) respondError(c *gin.Context, status int, message string) {
	if h.logger != nil {
		h.logger.Println(message)
	}
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
