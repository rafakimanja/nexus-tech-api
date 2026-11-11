package server

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed apidocs/openapi.json
var openAPISpec []byte

//go:embed apidocs/index.html
var documentationHTML []byte

func registerDocumentation(r *gin.Engine) {
	r.GET("/docs", func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", documentationHTML) })
	r.GET("/openapi.json", func(c *gin.Context) { c.Data(http.StatusOK, "application/json; charset=utf-8", openAPISpec) })
}
