package user

import "github.com/gin-gonic/gin"

type UserRouter struct {
	handler *UserHandler
}

func NewRouter(handler *UserHandler) *UserRouter {
	return &UserRouter{
		handler: handler,
	}
}

func(r *UserRouter) RegisterRoutes(router *gin.RouterGroup){
	router.POST("/", r.handler.CreateUser)
	router.GET("/", r.handler.GetUsers)
}