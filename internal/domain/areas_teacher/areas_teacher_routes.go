package areas_teacher

import "github.com/gin-gonic/gin"

type AreasTeacherRouter struct {
	handler *AreasTeacherHandler
}

func NewRouter(handler *AreasTeacherHandler) *AreasTeacherRouter {
	return &AreasTeacherRouter{handler: handler}
}

func (r *AreasTeacherRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/professor/:id", r.handler.GetByProfessorID)
	router.PUT("/professor/:id", r.handler.ReplaceByProfessorID)
}
