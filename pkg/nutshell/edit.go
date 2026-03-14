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
		// extensions.<name>.<path...> — arbitrary nested JSON
		if strings.HasPrefix(key, "extensions.") {
			return setExtension(m, key, value)
		}
		return fmt.Errorf("unknown field: %s", key)
	}
	return nil
}

// setExtension handles extensions.<name>.<nested.path> = value.
// It auto-creates the extension map and nested objects, and auto-detects
// numeric vs boolean vs string values.
func setExtension(m *Manifest, key, value string) error {
	parts := strings.SplitN(key, ".", 3) // ["extensions", "<name>", "<rest>"]
	if len(parts) < 3 {
		return fmt.Errorf("extensions path must be extensions.<name>.<field>: %s", key)
	}
	extName := parts[1]
	fieldPath := strings.Split(parts[2], ".")

	if m.Extensions == nil {
		m.Extensions = make(map[string]json.RawMessage)
	}

	// Decode existing extension or start empty
	var ext map[string]interface{}
	if raw, ok := m.Extensions[extName]; ok {
		if err := json.Unmarshal(raw, &ext); err != nil {
			ext = make(map[string]interface{})
		}
	} else {
		ext = make(map[string]interface{})
	}

	// Walk/create the nested path
	cur := ext
	for i, p := range fieldPath {
		if i == len(fieldPath)-1 {
			// Leaf — set the value with type detection
			cur[p] = detectValue(value)
		} else {
			next, ok := cur[p]
			if !ok {
				child := make(map[string]interface{})
				cur[p] = child
				cur = child
			} else if child, ok := next.(map[string]interface{}); ok {
				cur = child
			} else {
				child := make(map[string]interface{})
				cur[p] = child
				cur = child
			}
		}
	}

	data, err := json.Marshal(ext)
	if err != nil {
		return err
	}
	m.Extensions[extName] = json.RawMessage(data)
	return nil
}

// detectValue converts a string to float64, bool, or keeps it as string.
func detectValue(s string) interface{} {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	return s
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
