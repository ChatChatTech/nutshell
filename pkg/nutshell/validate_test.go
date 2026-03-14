package nutshell

import (
	"testing"
)

func TestValidateValidRequest(t *testing.T) {
	m := NewManifest()
	m.Task.Title = "Test Task"
	m.Task.Summary = "A test summary"
	m.Tags.SkillsRequired = []string{"go"}

	r := Validate(m)
	if !r.IsValid() {
		t.Fatalf("expected valid, got errors: %v", r.Errors)
	}
	if len(r.Warnings) != 0 {
		t.Fatalf("expected no warnings, got: %v", r.Warnings)
	}
}

func TestValidateMissingRequiredFields(t *testing.T) {
	m := &Manifest{}
	r := Validate(m)
	if r.IsValid() {
		t.Fatal("expected invalid for empty manifest")
	}
	// Should have errors for nutshell_version, bundle_type, id, task.title (since type is empty, no title check though)
	if len(r.Errors) < 3 {
		t.Fatalf("expected at least 3 errors, got %d: %v", len(r.Errors), r.Errors)
	}
}

func TestValidateInvalidBundleType(t *testing.T) {
	m := NewManifest()
	m.BundleType = "invalid"
	r := Validate(m)
	found := false
	for _, e := range r.Errors {
		if contains(e, "Invalid bundle_type") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected bundle_type error, got: %v", r.Errors)
	}
}

func TestValidateExpandedBundleTypes(t *testing.T) {
	for _, bt := range []string{"request", "delivery", "template", "checkpoint", "partial"} {
		m := NewManifest()
		m.BundleType = bt
		m.Task.Title = "Test"
		r := Validate(m)
		for _, e := range r.Errors {
			if contains(e, "Invalid bundle_type") {
				t.Fatalf("bundle_type '%s' should be valid, got error: %s", bt, e)
			}
		}
	}
}

func TestValidateDeliveryNoTitleRequired(t *testing.T) {
	m := NewManifest()
	m.BundleType = "delivery"
	m.Task.Title = "" // delivery doesn't require title
	r := Validate(m)
	for _, e := range r.Errors {
		if contains(e, "task.title") {
			t.Fatalf("delivery should not require task.title, got: %s", e)
		}
	}
}

func TestValidateCredentialWarnings(t *testing.T) {
	m := NewManifest()
	m.Task.Title = "Test"
	m.Credentials = &Credentials{
		Encryption: "none",
		Scopes: []CredentialScope{
			{Name: "api-key", Type: "api_key"},
		},
	}
	r := Validate(m)
	if len(r.Warnings) < 2 {
		t.Fatalf("expected >=2 warnings (unencrypted + no expiry), got %d: %v", len(r.Warnings), r.Warnings)
	}
}

func TestValidateContextBudgetWarning(t *testing.T) {
	m := NewManifest()
	m.Task.Title = "Test"
	m.Tags.SkillsRequired = []string{"go"}
	m.Harness = &Harness{ContextBudgetHint: 0.8}
	r := Validate(m)
	found := false
	for _, w := range r.Warnings {
		if contains(w, "Context budget") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected context budget warning for 0.8, got: %v", r.Warnings)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
