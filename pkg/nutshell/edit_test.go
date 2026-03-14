package nutshell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetField(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	// Set task.title
	if err := Set(dir, "task.title", "My New Task"); err != nil {
		t.Fatalf("set task.title failed: %v", err)
	}

	// Re-read and verify
	raw, _ := os.ReadFile(filepath.Join(dir, "nutshell.json"))
	var m2 Manifest
	json.Unmarshal(raw, &m2)
	if m2.Task.Title != "My New Task" {
		t.Fatalf("expected 'My New Task', got '%s'", m2.Task.Title)
	}
}

func TestSetMultipleFields(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	fields := map[string]string{
		"task.title":              "Build API",
		"task.summary":            "Build a REST API",
		"task.priority":           "high",
		"publisher.name":          "Alice",
		"bundle_type":             "request",
		"harness.context_budget_hint": "0.25",
		"tags.skills_required":    "go,rest,api",
	}

	for k, v := range fields {
		if err := Set(dir, k, v); err != nil {
			t.Fatalf("set %s failed: %v", k, err)
		}
	}

	raw, _ := os.ReadFile(filepath.Join(dir, "nutshell.json"))
	var m2 Manifest
	json.Unmarshal(raw, &m2)

	if m2.Task.Title != "Build API" {
		t.Fatalf("expected 'Build API', got '%s'", m2.Task.Title)
	}
	if m2.Publisher.Name != "Alice" {
		t.Fatalf("expected 'Alice', got '%s'", m2.Publisher.Name)
	}
	if m2.Harness.ContextBudgetHint != 0.25 {
		t.Fatalf("expected 0.25, got %f", m2.Harness.ContextBudgetHint)
	}
	if len(m2.Tags.SkillsRequired) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(m2.Tags.SkillsRequired), m2.Tags.SkillsRequired)
	}
}

func TestSetUnknownField(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	err := Set(dir, "nonexistent.field", "value")
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestSetInvalidFloat(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), data, 0644)

	err := Set(dir, "harness.context_budget_hint", "not-a-number")
	if err == nil {
		t.Fatal("expected error for invalid float")
	}
}
