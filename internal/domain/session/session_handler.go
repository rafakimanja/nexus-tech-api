package session

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
	logger  *log.Logger
}

func NewHandler(service *Service, logger *log.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Login(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var req struct {
		IDToken string `json:"idToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, 400, "idToken is required")
		return
	}
	result, err := h.service.Login(c.Request.Context(), req.IDToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrIdentity), errors.Is(err, gorm.ErrRecordNotFound):
			respondError(c, 401, "invalid Google identity")
		case errors.Is(err, ErrAccountConflict):
			respondError(c, 409, "account requires identity linking")
		case errors.Is(err, ErrBlocked):
			respondError(c, 403, "account blocked")
		default:
			if h.logger != nil {
				h.logger.Printf("login failed: %v", err)
			}
			respondError(c, 500, "could not complete login")
		}
		return
	}
	c.Header("Authorization", result.AccessToken)
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Me(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	profile, ok := c.Get("currentUser")
	if !ok {
		respondError(c, 401, "authentication required")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profile})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
}
