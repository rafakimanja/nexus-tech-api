package applicant

import (
	"api/internal/access"
	"github.com/gin-gonic/gin"
)

type ApplicantRouter struct {
	handler *ApplicantHandler
}

func NewRouter(handler *ApplicantHandler) *ApplicantRouter {
	return &ApplicantRouter{handler: handler}
}

func (r *ApplicantRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", access.Admin(), r.handler.GetApplicants)
	router.GET("/:id", access.OwnOrAdmin(), r.handler.GetApplicantByID)
}
