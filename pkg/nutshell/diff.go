package nutshell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiffEntry represents a single field difference between two manifests.
type DiffEntry struct {
	Field string `json:"field"`
	A     string `json:"a"`
	B     string `json:"b"`
}

// Diff compares two bundles (either .nut files or directories) and returns differences.
func Diff(pathA, pathB string) ([]DiffEntry, error) {
	mA, err := loadManifestFrom(pathA)
	if err != nil {
		return nil, fmt.Errorf("loading A (%s): %w", pathA, err)
	}
	mB, err := loadManifestFrom(pathB)
	if err != nil {
		return nil, fmt.Errorf("loading B (%s): %w", pathB, err)
	}

	var diffs []DiffEntry

	cmp := func(field, a, b string) {
		if a != b {
			diffs = append(diffs, DiffEntry{Field: field, A: a, B: b})
		}
	}

	cmp("bundle_type", mA.BundleType, mB.BundleType)
	cmp("nutshell_version", mA.NutshellVersion, mB.NutshellVersion)
	cmp("task.title", mA.Task.Title, mB.Task.Title)
	cmp("task.summary", mA.Task.Summary, mB.Task.Summary)
	cmp("task.priority", mA.Task.Priority, mB.Task.Priority)
	cmp("task.estimated_effort", mA.Task.EstimatedEffort, mB.Task.EstimatedEffort)
	cmp("publisher.name", mA.Publisher.Name, mB.Publisher.Name)
	cmp("publisher.tool", mA.Publisher.Tool, mB.Publisher.Tool)
	cmp("context.requirements", mA.Context.Requirements, mB.Context.Requirements)
	cmp("context.architecture", mA.Context.Architecture, mB.Context.Architecture)
	cmp("context.references", mA.Context.References, mB.Context.References)

	cmp("tags.skills_required",
		strings.Join(mA.Tags.SkillsRequired, ", "),
		strings.Join(mB.Tags.SkillsRequired, ", "))
	cmp("tags.domains",
		strings.Join(mA.Tags.Domains, ", "),
		strings.Join(mB.Tags.Domains, ", "))

	cmp("files.total_count",
		fmt.Sprintf("%d", mA.Files.TotalCount),
		fmt.Sprintf("%d", mB.Files.TotalCount))
	cmp("files.total_size_bytes",
		fmt.Sprintf("%d", mA.Files.TotalSizeBytes),
		fmt.Sprintf("%d", mB.Files.TotalSizeBytes))

	// File tree diff — show added/removed files
	filesA := make(map[string]int64)
	filesB := make(map[string]int64)
	for _, f := range mA.Files.Tree {
		filesA[f.Path] = f.Size
	}
	for _, f := range mB.Files.Tree {
		filesB[f.Path] = f.Size
	}
	for p := range filesA {
		if _, ok := filesB[p]; !ok {
			diffs = append(diffs, DiffEntry{Field: "file", A: p, B: "(removed)"})
		}
	}
	for p := range filesB {
		if _, ok := filesA[p]; !ok {
			diffs = append(diffs, DiffEntry{Field: "file", A: "(added)", B: p})
		}
	}

	// Harness
	hA := mA.Harness
	hB := mB.Harness
	if hA == nil {
		hA = &Harness{}
	}
	if hB == nil {
		hB = &Harness{}
	}
	cmp("harness.agent_type_hint", hA.AgentTypeHint, hB.AgentTypeHint)
	cmp("harness.execution_strategy", hA.ExecutionStrategy, hB.ExecutionStrategy)
	cmp("harness.context_budget_hint",
		fmt.Sprintf("%.2f", hA.ContextBudgetHint),
		fmt.Sprintf("%.2f", hB.ContextBudgetHint))

	// Acceptance
	aA := mA.Acceptance
	aB := mB.Acceptance
	if aA == nil {
		aA = &Acceptance{}
	}
	if aB == nil {
		aB = &Acceptance{}
	}
	cmp("acceptance.checklist",
		strings.Join(aA.Checklist, "; "),
		strings.Join(aB.Checklist, "; "))

	cmp("parent_id", mA.ParentID, mB.ParentID)

	return diffs, nil
}

func loadManifestFrom(path string) (*Manifest, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		data, err := os.ReadFile(filepath.Join(path, "nutshell.json"))
		if err != nil {
			return nil, fmt.Errorf("no nutshell.json: %w", err)
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	}

	if filepath.Ext(path) == ".nut" {
		m, _, err := Inspect(path)
		return m, err
	}

	// Plain JSON
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
