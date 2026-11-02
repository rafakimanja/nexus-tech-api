package teacher

import (
	"api/internal/access"
	"github.com/gin-gonic/gin"
)

type TeacherRouter struct {
	handler *TeacherHandler
}

func NewRouter(handler *TeacherHandler) *TeacherRouter {
	return &TeacherRouter{handler: handler}
}

func (r *TeacherRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", access.Admin(), r.handler.GetTeachers)
	router.GET("/:id", access.OwnOrAdmin(), r.handler.GetTeacherByID)
}
