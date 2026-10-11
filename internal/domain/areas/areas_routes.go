package areas

import "github.com/gin-gonic/gin"

type AreasRouter struct {
	handler *AreasHandler
}

func NewRouter(handler *AreasHandler) *AreasRouter {
	return &AreasRouter{handler: handler}
}

func (r *AreasRouter) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", r.handler.CreateArea)
	router.GET("/", r.handler.GetAreas)
	router.GET("/:id", r.handler.GetAreaByID)
	router.PUT("/:id", r.handler.UpdateArea)
	router.DELETE("/:id", r.handler.DeleteArea)
}
