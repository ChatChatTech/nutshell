package nutshell

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if !strings.HasPrefix(id, "nut-") {
		t.Fatalf("expected nut- prefix, got %s", id)
	}
	// nut-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx = 40 chars
	if len(id) != 40 {
		t.Fatalf("expected 40 chars, got %d: %s", len(id), id)
	}
	// Uniqueness
	id2 := GenerateID()
	if id == id2 {
		t.Fatal("two generated IDs should differ")
	}
}

func TestNewManifest(t *testing.T) {
	m := NewManifest()
	if m.NutshellVersion != SpecVersion {
		t.Fatalf("expected version %s, got %s", SpecVersion, m.NutshellVersion)
	}
	if m.BundleType != "request" {
		t.Fatalf("expected request, got %s", m.BundleType)
	}
	if !strings.HasPrefix(m.ID, "nut-") {
		t.Fatalf("expected nut- prefix in ID, got %s", m.ID)
	}
	if m.Task.Priority != "medium" {
		t.Fatalf("expected priority medium, got %s", m.Task.Priority)
	}
	if m.Harness == nil {
		t.Fatal("expected harness to be set")
	}
	if m.Harness.ContextBudgetHint != 0.35 {
		t.Fatalf("expected context budget 0.35, got %f", m.Harness.ContextBudgetHint)
	}
	if m.Completeness == nil || m.Completeness.Status != "draft" {
		t.Fatal("expected completeness status draft")
	}
}

func TestIsIgnored(t *testing.T) {
	patterns := []string{"*.log", "tmp/", "secret.json"}

	tests := []struct {
		path    string
		ignored bool
	}{
		{"app.log", true},
		{"logs/debug.log", true},
		{"main.go", false},
		{"tmp/cache", true},
		{"tmp", true},
		{"secret.json", true},
		{"context/data.md", false},
	}

	for _, tc := range tests {
		got := IsIgnored(tc.path, patterns)
		if got != tc.ignored {
			t.Errorf("IsIgnored(%q) = %v, want %v", tc.path, got, tc.ignored)
		}
	}
}

func TestIsIgnoredEmptyPatterns(t *testing.T) {
	if IsIgnored("anything.go", nil) {
		t.Error("nil patterns should not ignore anything")
	}
	if IsIgnored("anything.go", []string{}) {
		t.Error("empty patterns should not ignore anything")
	}
}
