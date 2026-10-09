package applicant

import "github.com/gin-gonic/gin"

type ApplicantRouter struct {
	handler *ApplicantHandler
}

func NewRouter(handler *ApplicantHandler) *ApplicantRouter {
	return &ApplicantRouter{handler: handler}
}

func (r *ApplicantRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", r.handler.CreateApplicant)
	router.GET("/", r.handler.GetApplicants)
	router.GET("/:id", r.handler.GetApplicantByID)
	router.PUT("/:id", r.handler.UpdateApplicant)
	router.DELETE("/:id", r.handler.DeleteApplicant)
}
