package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const specPath = "api/openapi.json"

var nullableProperties = map[string][]string{
	"apidoc.PeerSigningKeyList": {"last_fetched_at", "last_attempt_at", "last_error"},
	"apidoc.ErrorDetail":        {"field"},
	"apidoc.Subscription":       {"activated_at", "terminated_at"},
}

func main() {
	if err := run(specPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fixed, err := fix(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(path, fixed, 0o644)
}

func fix(raw []byte) ([]byte, error) {
	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", specPath, err)
	}
	collapseObjectOneOf(spec)
	if err := markNullable(spec); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(spec); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func collapseObjectOneOf(node any) {
	switch value := node.(type) {
	case map[string]any:
		if reference, ok := referenceOfObjectOneOf(value["oneOf"]); ok {
			delete(value, "oneOf")
			value["$ref"] = reference
		}
		for _, child := range value {
			collapseObjectOneOf(child)
		}
	case []any:
		for _, child := range value {
			collapseObjectOneOf(child)
		}
	}
}

func referenceOfObjectOneOf(node any) (string, bool) {
	variants, ok := node.([]any)
	if !ok || len(variants) != 2 {
		return "", false
	}
	reference := ""
	plainObject := false
	for _, variant := range variants {
		schema, ok := variant.(map[string]any)
		if !ok {
			return "", false
		}
		if ref, ok := schema["$ref"].(string); ok {
			reference = ref
			continue
		}
		if schema["type"] == "object" && len(schema) == 1 {
			plainObject = true
		}
	}
	return reference, reference != "" && plainObject
}

func markNullable(spec map[string]any) error {
	components, _ := spec["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	for schemaName, properties := range nullableProperties {
		schema, ok := schemas[schemaName].(map[string]any)
		if !ok {
			return fmt.Errorf("schema %s not found in %s", schemaName, specPath)
		}
		declared, _ := schema["properties"].(map[string]any)
		for _, name := range properties {
			property, ok := declared[name].(map[string]any)
			if !ok {
				return fmt.Errorf("property %s.%s not found in %s", schemaName, name, specPath)
			}
			if kind, ok := property["type"].(string); ok {
				property["type"] = []any{kind, "null"}
			}
		}
	}
	return nil
}
