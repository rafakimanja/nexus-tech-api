package server

import (
	"api/internal/domain/user"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    http.StatusUnauthorized,
					"message": "missing authorization header",
				},
			})
			return
		}

		userID, err := user.VerifyToken(authHeader)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    http.StatusUnauthorized,
					"message": "invalid or expired token",
				},
			})
			return
		}

		ctx.Set("userID", userID)
		profile, err := s.sessionService.Current(ctx.Request.Context(), userID)
		if err != nil {
			status, message := http.StatusInternalServerError, "could not verify account"
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status, message = http.StatusUnauthorized, "account not found"
			}
			ctx.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": status, "message": message}})
			return
		}
		if profile.Blocked && ctx.FullPath() != "/me" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": http.StatusForbidden, "message": "account blocked"}})
			return
		}
		ctx.Set("currentUser", profile)
		ctx.Set("role", profile.Role)
		ctx.Next()
	}
}
