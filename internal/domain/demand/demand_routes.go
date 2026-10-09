package demand

import "github.com/gin-gonic/gin"

type DemandaRouter struct {
	handler *DemandaHandler
}

func NewRouter(handler *DemandaHandler) *DemandaRouter {
	return &DemandaRouter{handler: handler}
}

func (r *DemandaRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", r.handler.CreateDemanda)
	router.GET("/", r.handler.GetDemandas)
	router.GET("/:id", r.handler.GetDemandaByID)
	router.PUT("/:id", r.handler.UpdateDemanda)
	router.DELETE("/:id", r.handler.DeleteDemanda)
}
