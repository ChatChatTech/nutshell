package nutshell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Set updates a field in nutshell.json using dot-path notation.
// Supported paths: task.title, task.summary, task.priority, task.estimated_effort,
// bundle_type, publisher.name, publisher.contact, publisher.tool,
// context.requirements, context.architecture, context.references,
// harness.agent_type_hint, harness.execution_strategy, harness.context_budget_hint,
// harness.checkpoints, parent_id, expires_at.
func Set(dir, key, value string) error {
	manifestPath := filepath.Join(dir, "nutshell.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("no nutshell.json found in %s: %w", dir, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("invalid nutshell.json: %w", err)
	}

	if err := setField(&manifest, key, value); err != nil {
		return err
	}

	out, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, out, 0644)
}

func setField(m *Manifest, key, value string) error {
	switch key {
	// Top-level
	case "bundle_type":
		m.BundleType = value
	case "parent_id":
		m.ParentID = value
	case "expires_at":
		m.ExpiresAt = value

	// Task
	case "task.title":
		m.Task.Title = value
	case "task.summary":
		m.Task.Summary = value
	case "task.priority":
		m.Task.Priority = value
	case "task.estimated_effort":
		m.Task.EstimatedEffort = value

	// Publisher
	case "publisher.name":
		m.Publisher.Name = value
	case "publisher.contact":
		m.Publisher.Contact = value
	case "publisher.tool":
		m.Publisher.Tool = value

	// Context
	case "context.requirements":
		m.Context.Requirements = value
	case "context.architecture":
		m.Context.Architecture = value
	case "context.references":
		m.Context.References = value

	// Harness
	case "harness.agent_type_hint":
		if m.Harness == nil {
			m.Harness = &Harness{}
		}
		m.Harness.AgentTypeHint = value
	case "harness.execution_strategy":
		if m.Harness == nil {
			m.Harness = &Harness{}
		}
		m.Harness.ExecutionStrategy = value
	case "harness.context_budget_hint":
		if m.Harness == nil {
			m.Harness = &Harness{}
		}
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float for %s: %w", key, err)
		}
		m.Harness.ContextBudgetHint = f
	case "harness.checkpoints":
		if m.Harness == nil {
			m.Harness = &Harness{}
		}
		m.Harness.Checkpoints = value == "true"

	// Tags (append mode)
	case "tags.skills_required":
		m.Tags.SkillsRequired = splitCSV(value)
	case "tags.domains":
		m.Tags.Domains = splitCSV(value)

	default:
		return fmt.Errorf("unknown field: %s", key)
	}
	return nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
