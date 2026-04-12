package apidocs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed openapi.yaml
	openAPISpec []byte

	//go:embed redoc.html
	redocHTML []byte
)

func RegisterRoutes(router gin.IRoutes) {
	router.GET("/openapi.yaml", serveOpenAPISpec)
	router.GET("/docs", redirectToDocsIndex)
	router.GET("/docs/", serveDocsUI)
}

func serveOpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
}

func redirectToDocsIndex(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, "/docs/")
}

func serveDocsUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", redocHTML)
}

func Spec() []byte {
	return openAPISpec
}
