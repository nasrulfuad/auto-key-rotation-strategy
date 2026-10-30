package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixCollapsesObjectOneOfAndMarksNullable(t *testing.T) {
	raw := []byte(`{
		"paths": {"/x": {"put": {"requestBody": {"content": {"application/json": {"schema": {"oneOf": [{"type": "object"}, {"$ref": "#/components/schemas/apidoc.UpsertEntitlementRequest"}]}}}}}}},
		"components": {"schemas": {
			"apidoc.PeerSigningKeyList": {"properties": {"last_fetched_at": {"type": "string"}, "last_attempt_at": {"type": "string"}, "last_error": {"type": "string"}}},
			"apidoc.ErrorDetail": {"properties": {"field": {"type": "string"}}},
			"apidoc.Entitlement": {"properties": {"terminated_at": {"type": "string"}}}
		}}
	}`)

	fixed, err := fix(raw)
	require.NoError(t, err)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(fixed, &spec))
	schema := spec["paths"].(map[string]any)["/x"].(map[string]any)["put"].(map[string]any)["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	assert.Equal(t, map[string]any{"$ref": "#/components/schemas/apidoc.UpsertEntitlementRequest"}, schema)
	lastError := spec["components"].(map[string]any)["schemas"].(map[string]any)["apidoc.PeerSigningKeyList"].(map[string]any)["properties"].(map[string]any)["last_error"]
	assert.Equal(t, map[string]any{"type": []any{"string", "null"}}, lastError)
}

func TestFixFailsWhenANullablePropertyDisappears(t *testing.T) {
	_, err := fix([]byte(`{"components": {"schemas": {}}}`))
	assert.Error(t, err)
}
