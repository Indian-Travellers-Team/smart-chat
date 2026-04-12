package apidocs

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed openapi.yaml
	openAPISpec []byte

	//go:embed redoc.html
	redocHTML []byte
)

func RegisterRoutes(router gin.IRouter, username, password string) error {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return fmt.Errorf("swagger docs credentials are required")
	}

	docs := router.Group("", gin.BasicAuth(gin.Accounts{username: password}))
	docs.GET("/openapi.yaml", serveOpenAPISpec)
	docs.GET("/docs", redirectToDocsIndex)
	docs.GET("/docs/", serveDocsUI)

	return nil
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
