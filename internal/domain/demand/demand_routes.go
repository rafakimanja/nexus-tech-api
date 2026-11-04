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
	// DELETE preserves the record and requires the same reason as cancelar.
	router.DELETE("/:id", r.handler.Action("cancelar"))
	for _, action := range []string{"analisar", "devolver", "assumir", "concluir", "cancelar"} {
		router.POST("/:id/"+action, r.handler.Action(action))
	}
}
