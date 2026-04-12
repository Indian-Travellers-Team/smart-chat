package apidocs

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"

	"smart-chat/auth"
	"smart-chat/internal/routes"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var ginPathParamPattern = regexp.MustCompile(`:([^/]+)`)

type openAPIDocument struct {
	Paths map[string]map[string]any `yaml:"paths"`
}

func TestOpenAPISpecCoversRegisteredRoutes(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/v1")
	auth.RegisterAuthRoutes(v1.Group("/auth"), nil)
	routes.RegisterRoutes(v1.Group("/chat"), nil)

	v2 := router.Group("/v2")
	auth.RegisterV2AuthRoutes(v2.Group("/auth"), nil)
	routes.RegisterV2Routes(v2.Group("/chat"), nil, nil, nil)
	routes.ClientRoutes(v2.Group("/client"), nil, nil, nil, nil, nil, nil, nil, nil)

	var document openAPIDocument
	if err := yaml.Unmarshal(Spec(), &document); err != nil {
		t.Fatalf("failed to parse embedded OpenAPI spec: %v", err)
	}

	registered := registeredRouteSet(router)
	documented := documentedRouteSet(document)

	missingInSpec := diffKeys(registered, documented)
	extraInSpec := diffKeys(documented, registered)

	if len(missingInSpec) > 0 || len(extraInSpec) > 0 {
		parts := make([]string, 0, 2)
		if len(missingInSpec) > 0 {
			parts = append(parts, "missing in spec: "+strings.Join(missingInSpec, ", "))
		}
		if len(extraInSpec) > 0 {
			parts = append(parts, "extra in spec: "+strings.Join(extraInSpec, ", "))
		}
		t.Fatalf("OpenAPI route coverage mismatch: %s", strings.Join(parts, " | "))
	}
}

func registeredRouteSet(router *gin.Engine) map[string]struct{} {
	routesByMethodPath := make(map[string]struct{})
	for _, routeInfo := range router.Routes() {
		if routeInfo.Method == http.MethodHead || routeInfo.Method == http.MethodOptions {
			continue
		}
		key := fmt.Sprintf("%s %s", strings.ToUpper(routeInfo.Method), normalizeRoutePath(routeInfo.Path))
		routesByMethodPath[key] = struct{}{}
	}
	return routesByMethodPath
}

func documentedRouteSet(document openAPIDocument) map[string]struct{} {
	documented := make(map[string]struct{})
	for path, operations := range document.Paths {
		for method := range operations {
			key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
			documented[key] = struct{}{}
		}
	}
	return documented
}

func diffKeys(left, right map[string]struct{}) []string {
	missing := make([]string, 0)
	for key := range left {
		if _, ok := right[key]; ok {
			continue
		}
		missing = append(missing, key)
	}
	sort.Strings(missing)
	return missing
}

func normalizeRoutePath(path string) string {
	return ginPathParamPattern.ReplaceAllString(path, `{$1}`)
}
