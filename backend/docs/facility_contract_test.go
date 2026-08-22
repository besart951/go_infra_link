package docs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type contractSpec struct {
	Paths       map[string]map[string]contractOperation `json:"paths"`
	Definitions map[string]contractSchema               `json:"definitions"`
}

type contractOperation struct {
	Parameters []contractParameter `json:"parameters"`
}

type contractParameter struct {
	Name     string         `json:"name"`
	In       string         `json:"in"`
	Required bool           `json:"required"`
	Schema   contractSchema `json:"schema"`
}

type contractSchema struct {
	Ref        string                    `json:"$ref"`
	Required   []string                  `json:"required"`
	Properties map[string]contractSchema `json:"properties"`
	Items      *contractSchema           `json:"items"`
}

func TestFacilityMutationContractsRequireBaseVersion(t *testing.T) {
	spec := readContractSpec(t)
	checked := 0
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			if !isVersionedMutation(path, method) {
				continue
			}
			checked++
			if !operationRequiresBaseVersion(operation, spec.Definitions) {
				t.Errorf("%s %s does not require base_version", strings.ToUpper(method), path)
			}
		}
	}
	if checked < 30 {
		t.Fatalf("only %d versioned mutation routes were inventoried", checked)
	}
}

func readContractSpec(t *testing.T) contractSpec {
	t.Helper()
	encoded, err := os.ReadFile("swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec contractSpec
	if err := json.Unmarshal(encoded, &spec); err != nil {
		t.Fatal(err)
	}
	return spec
}

func isVersionedMutation(path, method string) bool {
	if method != "put" && method != "patch" && method != "delete" {
		return false
	}
	if strings.HasPrefix(path, "/api/v1/facility/") {
		return true
	}
	return isVersionedProjectPath(path)
}

func isVersionedProjectPath(path string) bool {
	if path == "/api/v1/projects/{id}" {
		return true
	}
	for _, resource := range []string{"control-cabinets", "sps-controllers", "field-devices"} {
		if strings.HasPrefix(path, "/api/v1/projects/{id}/"+resource+"/") ||
			strings.HasPrefix(path, "/api/v1/projects/{id}/facility/"+resource+"/") {
			return true
		}
	}
	return strings.HasPrefix(path, "/api/v1/projects/{id}/facility/sps-controller-system-types/")
}

func operationRequiresBaseVersion(operation contractOperation, definitions map[string]contractSchema) bool {
	for _, parameter := range operation.Parameters {
		if parameter.In == "query" && parameter.Name == "base_version" && parameter.Required {
			return true
		}
		if parameter.In == "body" && schemaRequiresBaseVersion(parameter.Schema, definitions, map[string]bool{}) {
			return true
		}
	}
	return false
}

func schemaRequiresBaseVersion(schema contractSchema, definitions map[string]contractSchema, seen map[string]bool) bool {
	if schema.Ref != "" {
		name := schema.Ref[strings.LastIndex(schema.Ref, "/")+1:]
		if seen[name] {
			return false
		}
		seen[name] = true
		return schemaRequiresBaseVersion(definitions[name], definitions, seen)
	}
	if containsString(schema.Required, "base_version") {
		return true
	}
	if schema.Items != nil && schemaRequiresBaseVersion(*schema.Items, definitions, seen) {
		return true
	}
	for _, property := range schema.Properties {
		if schemaRequiresBaseVersion(property, definitions, seen) {
			return true
		}
	}
	return false
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
