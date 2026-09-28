package uauth

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UAuthHandler struct {
	service UAuthService
	logger *log.Logger
}

func NewHandler(service UAuthService, logger *log.Logger) *UAuthHandler {
	return &UAuthHandler{
		service: service,
		logger: logger,
	}
}

func (h *UAuthHandler) GetUserAuthByID(c *gin.Context){
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

	uAuth, err := h.service.GetByID(id)

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

	c.JSON(http.StatusOK, uAuth)
}

func (h *UAuthHandler) CreateUserAuth(c *gin.Context){
	var uAuth UserAuth

	if err := c.ShouldBindJSON(&uAuth); err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code": http.StatusBadRequest,
				"message": "objeto 'User Auth' invalido",
			},
		})
		return
	}

	id, err := h.service.Create(uAuth)

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

	c.JSON(http.StatusOK, gin.H{"UserAuthID": id})
}