package http

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	myapp "github.com/nasrulfuad/auto-key-rotation-strategy/my_app"
)

var echoPathParameter = regexp.MustCompile(`:([a-z_]+)`)

func TestOpenAPISpecDocumentsEveryRouteAndNothingElse(t *testing.T) {
	raw, err := myapp.OpenAPIFS.ReadFile("api/openapi.json")
	require.NoError(t, err)
	var spec struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &spec))

	documented := map[string]bool{}
	for path, operations := range spec.Paths {
		for method := range operations {
			documented[strings.ToUpper(method)+" "+path] = true
		}
	}
	routed := map[string]bool{}
	for _, route := range (&Router{}).Table() {
		routed[route.Method+" "+echoPathParameter.ReplaceAllString(route.Path, "{$1}")] = true
	}

	assert.Equal(t, routed, documented)
}
