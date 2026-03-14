package nutshell

import (
	"encoding/json"
	"testing"
)

func TestSchemaIsValidJSON(t *testing.T) {
	s := Schema()
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		t.Fatalf("Schema() is not valid JSON: %v", err)
	}

	// Check required top-level fields
	if raw["$schema"] == nil {
		t.Fatal("missing $schema")
	}
	if raw["title"] == nil {
		t.Fatal("missing title")
	}
	if raw["properties"] == nil {
		t.Fatal("missing properties")
	}

	props := raw["properties"].(map[string]interface{})
	expectedFields := []string{
		"nutshell_version", "bundle_type", "id", "task", "tags",
		"publisher", "context", "files", "apis", "credentials",
		"acceptance", "harness", "resources", "completeness",
		"compression", "extensions", "parent_id", "created_at", "expires_at",
	}
	for _, f := range expectedFields {
		if props[f] == nil {
			t.Errorf("missing property: %s", f)
		}
	}
}

func TestSchemaMatchesManifestFields(t *testing.T) {
	// Verify the schema covers the bundle types we support
	s := Schema()
	var raw map[string]interface{}
	json.Unmarshal([]byte(s), &raw)

	props := raw["properties"].(map[string]interface{})
	bt := props["bundle_type"].(map[string]interface{})
	enums := bt["enum"].([]interface{})

	expected := map[string]bool{
		"request": false, "delivery": false,
		"template": false, "checkpoint": false, "partial": false,
	}
	for _, e := range enums {
		expected[e.(string)] = true
	}
	for k, found := range expected {
		if !found {
			t.Errorf("bundle_type enum missing: %s", k)
		}
	}
}
