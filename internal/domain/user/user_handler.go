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

func NewService(service UserService, logger *log.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetAll()
	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(c *gin.Context){
	strID := c.Param("id")
	id, err := strconv.Atoi(strID)

	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": http.StatusInternalServerError,
				"message": "invalid ID",
			},
		})
		return
	}

	user, err := h.service.GetByID(id)

	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateUser(c *gin.Context){
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": http.StatusBadRequest,
				"message": "objeto 'User' invalido",
			},
		})
		return
	}

	id, err := h.service.Create(user)

	if err != nil || id == 0 {
		h.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"userID": id})
}

func (h *UserHandler) DeleteUser(c *gin.Context){
	strID := c.Param("id")
	id, err := strconv.Atoi(strID)

	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": http.StatusInternalServerError,
				"message": "invalid ID",
			},
		})
		return
	}

	err = h.service.DeleteByID(id)

	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code": http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}