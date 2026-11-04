package access

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func Deny(c *gin.Context) {
	c.AbortWithStatusJSON(403, gin.H{"error": gin.H{"code": 403, "message": "access denied"}})
}
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != "admin" {
			Deny(c)
			return
		}
		c.Next()
	}
}
func OwnOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil || id == 0 {
			c.AbortWithStatusJSON(400, gin.H{"error": gin.H{"code": 400, "message": "invalid ID"}})
			return
		}
		if c.GetString("role") != "admin" && c.GetUint("userID") != uint(id) {
			Deny(c)
			return
		}
		c.Next()
	}
}
