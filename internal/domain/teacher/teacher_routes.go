package teacher

import "github.com/gin-gonic/gin"

type TeacherRouter struct {
	handler *TeacherHandler
}

func NewRouter(handler *TeacherHandler) *TeacherRouter {
	return &TeacherRouter{handler: handler}
}

func (r *TeacherRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", r.handler.CreateTeacher)
	router.GET("/", r.handler.GetTeachers)
	router.GET("/:id", r.handler.GetTeacherByID)
	router.PUT("/:id", r.handler.UpdateTeacher)
	router.PUT("/:id/areas", r.handler.UpdateAreasTeacher)
	router.DELETE("/:id", r.handler.DeleteTeacher)
}
