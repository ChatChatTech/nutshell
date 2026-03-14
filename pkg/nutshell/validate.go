package nutshell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ValidationResult holds errors and warnings from validation.
type ValidationResult struct {
	Errors   []string
	Warnings []string
}

func (v *ValidationResult) IsValid() bool {
	return len(v.Errors) == 0
}

// Validate checks a manifest against the spec.
func Validate(manifest *Manifest) *ValidationResult {
	r := &ValidationResult{}

	// Required fields
	if manifest.NutshellVersion == "" {
		r.Errors = append(r.Errors, "Missing required field: nutshell_version")
	}
	if manifest.BundleType == "" {
		r.Errors = append(r.Errors, "Missing required field: bundle_type")
	} else {
		validTypes := map[string]bool{
			"request": true, "delivery": true,
			"template": true, "checkpoint": true, "partial": true,
		}
		if !validTypes[manifest.BundleType] {
			r.Errors = append(r.Errors, fmt.Sprintf(
				"Invalid bundle_type: '%s' (must be request|delivery|template|checkpoint|partial)",
				manifest.BundleType))
		}
	}
	if manifest.ID == "" {
		r.Errors = append(r.Errors, "Missing required field: id")
	}
	if manifest.BundleType == "request" && manifest.Task.Title == "" {
		r.Errors = append(r.Errors, "task.title is required")
	}

	// Warnings
	if manifest.BundleType == "request" {
		if manifest.Task.Summary == "" {
			r.Warnings = append(r.Warnings, "task.summary is empty")
		}
		if len(manifest.Tags.SkillsRequired) == 0 {
			r.Warnings = append(r.Warnings, "No skills_required tags — matching will be broad")
		}
	}

	// Credentials security
	if manifest.Credentials != nil {
		if manifest.Credentials.Encryption == "none" {
			r.Warnings = append(r.Warnings, "Credentials are unencrypted — not recommended for production")
		}
		for _, scope := range manifest.Credentials.Scopes {
			if scope.ExpiresAt == "" {
				r.Warnings = append(r.Warnings, fmt.Sprintf("Credential '%s' has no expiration", scope.Name))
			}
		}
	}

	// Harness hints
	if manifest.Harness != nil && manifest.Harness.ContextBudgetHint > 0.5 {
		r.Warnings = append(r.Warnings, fmt.Sprintf(
			"Context budget hint %.2f exceeds recommended 0.4 (40%% rule)",
			manifest.Harness.ContextBudgetHint,
		))
	}

	return r
}

// ValidateFile validates a .nut file or directory.
func ValidateFile(path string) (*Manifest, *ValidationResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	var manifest Manifest

	if info.IsDir() {
		// Directory mode
		data, err := os.ReadFile(filepath.Join(path, "nutshell.json"))
		if err != nil {
			return nil, nil, fmt.Errorf("no nutshell.json found: %w", err)
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, nil, fmt.Errorf("invalid nutshell.json: %w", err)
		}
	} else if filepath.Ext(path) == ".nut" {
		// Bundle mode
		m, _, err := Inspect(path)
		if err != nil {
			return nil, nil, err
		}
		manifest = *m
	} else {
		// Plain JSON file
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, nil, fmt.Errorf("invalid JSON: %w", err)
		}
	}

	result := Validate(&manifest)
	return &manifest, result, nil
}

// Check performs a completeness check on a bundle directory.
// Returns what's present, what's missing, and warnings.
func Check(dir string) (*Manifest, *ValidationResult, error) {
	data, err := os.ReadFile(filepath.Join(dir, "nutshell.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("no nutshell.json found in %s: %w", dir, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, nil, fmt.Errorf("invalid nutshell.json: %w", err)
	}

	r := &ValidationResult{}

	// Check task basics
	if manifest.Task.Title == "" {
		r.Errors = append(r.Errors, "task.title is empty — agent won't know what to do")
	}
	if manifest.Task.Summary == "" {
		r.Warnings = append(r.Warnings, "task.summary is empty — agent needs context")
	}

	// Check referenced files exist
	checkFile := func(field, path string) {
		if path == "" {
			return
		}
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: referenced but missing (%s)", field, path))
		}
	}

	checkFile("context.requirements", manifest.Context.Requirements)
	checkFile("context.architecture", manifest.Context.Architecture)
	checkFile("context.references", manifest.Context.References)
	for _, a := range manifest.Context.Additional {
		checkFile("context.additional", a)
	}

	if manifest.APIs != nil {
		checkFile("apis.endpoints_spec", manifest.APIs.EndpointsSpec)
	}

	// Credential check
	if manifest.Credentials != nil && manifest.Credentials.Vault != "" {
		full := filepath.Join(dir, manifest.Credentials.Vault)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			r.Errors = append(r.Errors, fmt.Sprintf("credentials.vault: referenced but missing (%s)", manifest.Credentials.Vault))
		}
	} else if manifest.APIs != nil && manifest.APIs.AuthMethod != "" && manifest.APIs.AuthMethod != "none" {
		r.Warnings = append(r.Warnings, "APIs require auth but no credentials configured — agent won't have access")
	}

	// Acceptance check
	if manifest.Acceptance == nil || len(manifest.Acceptance.Checklist) == 0 {
		r.Warnings = append(r.Warnings, "No acceptance criteria — agent can't self-verify completion")
	}

	// Harness check
	if manifest.Harness == nil || len(manifest.Harness.Constraints) == 0 {
		r.Warnings = append(r.Warnings, "No harness constraints — agent has no guardrails")
	}

	// Tags check
	if len(manifest.Tags.SkillsRequired) == 0 {
		r.Warnings = append(r.Warnings, "No skills_required tags — task scope is unclear")
	}

	// Update completeness in manifest
	missing := r.Errors
	warnings := r.Warnings
	status := "ready"
	if len(missing) > 0 {
		status = "incomplete"
	} else if len(warnings) > 0 {
		status = "ready" // ready but with warnings
	}
	manifest.Completeness = &Completeness{
		Status:   status,
		Missing:  missing,
		Warnings: warnings,
	}

	return &manifest, r, nil
}
