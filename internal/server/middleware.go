package server

import (
	"api/internal/domain/user"
	"net/http"

	"github.com/gin-gonic/gin"
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
					"message": err.Error(),
				},
			})
			return
		}

		ctx.Set("userID", userID)
		ctx.Next()
	}
}
