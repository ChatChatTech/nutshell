package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChatChatTech/nutshell/pkg/nutshell"
)

const (
	cyan   = "\033[96m"
	green  = "\033[92m"
	yellow = "\033[93m"
	red    = "\033[91m"
	dim    = "\033[2m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

const shellArt = `
    🐚  n u t s h e l l  🦞
    Task Packaging for AI Agents
`

func usage() {
	fmt.Println(shellArt)
	fmt.Println("Usage:")
	fmt.Println("  nutshell init    [--dir <path>]                  Initialize a new bundle directory")
	fmt.Println("  nutshell pack    [--dir <path>] [-o <file>]      Pack directory into .nut bundle")
	fmt.Println("  nutshell unpack  <file> [-o <path>]              Unpack a .nut bundle")
	fmt.Println("  nutshell inspect <file|-> [--json]               Inspect bundle without unpacking")
	fmt.Println("  nutshell validate <file|dir> [--json]            Validate bundle against spec")
	fmt.Println("  nutshell check   [--dir <path>] [--json]         Completeness check — what's missing?")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		cmdInit(args)
	case "pack":
		cmdPack(args)
	case "unpack":
		cmdUnpack(args)
	case "inspect":
		cmdInspect(args)
	case "validate":
		cmdValidate(args)
	case "check":
		cmdCheck(args)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "%s✗%s Unknown command: %s\n", red, reset, cmd)
		usage()
		os.Exit(1)
	}
}

func getFlag(args []string, flags ...string) (string, []string) {
	for i, a := range args {
		for _, f := range flags {
			if a == f && i+1 < len(args) {
				val := args[i+1]
				rest := make([]string, 0, len(args)-2)
				rest = append(rest, args[:i]...)
				rest = append(rest, args[i+2:]...)
				return val, rest
			}
		}
	}
	return "", args
}

func hasFlag(args []string, flags ...string) (bool, []string) {
	for i, a := range args {
		for _, f := range flags {
			if a == f {
				rest := make([]string, 0, len(args)-1)
				rest = append(rest, args[:i]...)
				rest = append(rest, args[i+1:]...)
				return true, rest
			}
		}
	}
	return false, args
}

func getPositional(args []string) string {
	for _, a := range args {
		if a == "-" || !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

func cmdInit(args []string) {
	dir, _ := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}

	// Only create the minimal directories — user adds more as needed
	dirs := []string{
		"context",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(dir, d), 0755)
	}

	manifestPath := filepath.Join(dir, "nutshell.json")
	if _, err := os.Stat(manifestPath); err == nil {
		fmt.Printf("%s⚠%s nutshell.json already exists at %s/\n", yellow, reset, dir)
		return
	}

	m := nutshell.NewManifest()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(manifestPath, data, 0644)

	// Create template requirements.md
	reqPath := filepath.Join(dir, "context", "requirements.md")
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		os.WriteFile(reqPath, []byte("# Requirements\n\n## Objective\n\n(Describe the task objective)\n\n## Functional Requirements\n\n- FR-1: ...\n\n## Non-Functional Requirements\n\n- NFR-1: ...\n"), 0644)
	}

	fmt.Printf("%s✓%s Initialized nutshell bundle at %s%s/%s\n", green, reset, bold, dir, reset)
	fmt.Printf("  Edit %snutshell.json%s to configure your task.\n", cyan, reset)
	fmt.Printf("  Run %snutshell check%s to see what's still needed.\n", cyan, reset)
}

func cmdPack(args []string) {
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}
	output, _ := getFlag(args, "--output", "-o")

	// Determine output path
	if output == "" {
		data, err := os.ReadFile(filepath.Join(dir, "nutshell.json"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s No nutshell.json found in %s\n", red, reset, dir)
			os.Exit(1)
		}
		var m nutshell.Manifest
		json.Unmarshal(data, &m)
		slug := strings.ToLower(m.Task.Title)
		slug = strings.ReplaceAll(slug, " ", "-")
		if len(slug) > 40 {
			slug = slug[:40]
		}
		if slug == "" {
			slug = "bundle"
		}
		output = slug + ".nut"
	}

	manifest, err := nutshell.Pack(dir, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	info, _ := os.Stat(output)
	compSize := info.Size()
	origSize := manifest.Files.TotalSizeBytes
	ratio := float64(0)
	if origSize > 0 {
		ratio = (1 - float64(compSize)/float64(origSize)) * 100
	}

	fmt.Printf("%s✓%s Packed %s%d%s files into %s%s%s\n",
		green, reset, cyan, manifest.Files.TotalCount, reset, bold, output, reset)
	fmt.Printf("  %sOriginal: %d bytes → Compressed: %d bytes (%.1f%% reduction)%s\n",
		dim, origSize, compSize, ratio, reset)
	fmt.Printf("  %sBundle ID: %s%s\n", dim, manifest.ID, reset)

	// Show content hash for content-addressing
	if hash, err := nutshell.HashBundle(output); err == nil {
		fmt.Printf("  %sHash: %s%s\n", dim, hash, reset)
	}
}

func cmdUnpack(args []string) {
	file := getPositional(args)
	if file == "" {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell unpack <file> [-o <path>]\n", red, reset)
		os.Exit(1)
	}
	output, _ := getFlag(args, "--output", "-o")
	if output == "" {
		base := filepath.Base(file)
		output = strings.TrimSuffix(base, filepath.Ext(base))
	}

	manifest, err := nutshell.Unpack(file, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s✓%s Unpacked to %s%s/%s\n", green, reset, bold, output, reset)
	if manifest != nil {
		fmt.Printf("  %sTask: %s%s\n", dim, manifest.Task.Title, reset)
		fmt.Printf("  %sType: %s%s\n", dim, manifest.BundleType, reset)
	}
}

func cmdInspect(args []string) {
	jsonMode, args := hasFlag(args, "--json")
	file := getPositional(args)
	if file == "" {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell inspect <file|->\n", red, reset)
		os.Exit(1)
	}

	var manifest *nutshell.Manifest
	var entries []string
	var err error

	if file == "-" {
		manifest, entries, err = nutshell.InspectReader(os.Stdin)
	} else {
		manifest, entries, err = nutshell.Inspect(file)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	if jsonMode {
		out := map[string]interface{}{
			"manifest": manifest,
			"entries":  entries,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Println(shellArt)
	fmt.Printf("  %sBundle:%s %s\n", bold, reset, filepath.Base(file))
	fmt.Printf("  %sVersion:%s %s\n", bold, reset, manifest.NutshellVersion)
	fmt.Printf("  %sType:%s %s\n", bold, reset, manifest.BundleType)
	fmt.Printf("  %sID:%s %s\n", bold, reset, manifest.ID)
	fmt.Println()

	fmt.Printf("  %s📋 Task:%s %s\n", cyan, reset, manifest.Task.Title)
	if manifest.Task.Summary != "" {
		fmt.Printf("  %s%s%s\n", dim, manifest.Task.Summary, reset)
	}
	fmt.Printf("  Priority: %s | Effort: %s\n",
		manifest.Task.Priority, manifest.Task.EstimatedEffort)
	fmt.Println()

	if len(manifest.Tags.SkillsRequired) > 0 {
		fmt.Printf("  %s🏷️  Tags:%s %s\n", cyan, reset, strings.Join(manifest.Tags.SkillsRequired, ", "))
	}
	if len(manifest.Tags.Domains) > 0 {
		fmt.Printf("  %sDomains: %s%s\n", dim, strings.Join(manifest.Tags.Domains, ", "), reset)
	}
	fmt.Println()

	if manifest.Publisher.Name != "" {
		fmt.Printf("  %s👤 Publisher:%s %s", cyan, reset, manifest.Publisher.Name)
		if manifest.Publisher.Tool != "" {
			fmt.Printf(" (via %s)", manifest.Publisher.Tool)
		}
		fmt.Println()
		fmt.Println()
	}

	if manifest.Credentials != nil && len(manifest.Credentials.Scopes) > 0 {
		fmt.Printf("  %s🔑 Credentials:%s %d scoped\n", cyan, reset, len(manifest.Credentials.Scopes))
		for _, s := range manifest.Credentials.Scopes {
			fmt.Printf("    • %s (%s) — %s\n", s.Name, s.Type, s.AccessLevel)
		}
		fmt.Println()
	}

	fmt.Printf("  %s📦 Files:%s %d files, %d bytes\n",
		cyan, reset, manifest.Files.TotalCount, manifest.Files.TotalSizeBytes)
	fmt.Printf("  %sArchive entries: %d%s\n", dim, len(entries), reset)

	if manifest.Compression != nil && manifest.Compression.ContextTokensEstimate > 0 {
		fmt.Printf("  %sEst. tokens: ~%d%s\n", dim, manifest.Compression.ContextTokensEstimate, reset)
	}

	if manifest.Harness != nil {
		fmt.Printf("\n  %s⚙️  Harness Hints:%s\n", cyan, reset)
		fmt.Printf("    Agent type: %s\n", manifest.Harness.AgentTypeHint)
		fmt.Printf("    Strategy: %s\n", manifest.Harness.ExecutionStrategy)
		fmt.Printf("    Context budget: %.2f\n", manifest.Harness.ContextBudgetHint)
		if len(manifest.Harness.Constraints) > 0 {
			fmt.Printf("    Constraints: %d\n", len(manifest.Harness.Constraints))
			for _, c := range manifest.Harness.Constraints {
				fmt.Printf("      • %s\n", c)
			}
		}
	}

	if manifest.Extensions != nil && len(manifest.Extensions) > 0 {
		fmt.Printf("\n  %s🔌 Extensions:%s", cyan, reset)
		for name := range manifest.Extensions {
			fmt.Printf(" %s", name)
		}
		fmt.Println()
	}
}

func cmdValidate(args []string) {
	jsonMode, args := hasFlag(args, "--json")
	file := getPositional(args)
	if file == "" {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell validate <file|dir>\n", red, reset)
		os.Exit(1)
	}

	_, result, err := nutshell.ValidateFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	if jsonMode {
		out := map[string]interface{}{
			"valid":    result.IsValid(),
			"errors":   result.Errors,
			"warnings": result.Warnings,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		if !result.IsValid() {
			os.Exit(1)
		}
		return
	}

	fmt.Printf("\n  %sValidating:%s %s\n\n", bold, reset, file)

	for _, e := range result.Errors {
		fmt.Printf("  %s✗ ERROR:%s %s\n", red, reset, e)
	}
	for _, w := range result.Warnings {
		fmt.Printf("  %s⚠ WARN:%s  %s\n", yellow, reset, w)
	}

	if result.IsValid() && len(result.Warnings) == 0 {
		fmt.Printf("  %s✓ All checks passed%s\n", green, reset)
	} else if result.IsValid() {
		fmt.Printf("\n  %s✓ Valid%s with %d warning(s)\n", green, reset, len(result.Warnings))
	} else {
		fmt.Printf("\n  %s✗ Invalid%s — %d error(s), %d warning(s)\n",
			red, reset, len(result.Errors), len(result.Warnings))
		os.Exit(1)
	}
}

func cmdCheck(args []string) {
	jsonMode, args := hasFlag(args, "--json")
	dir, _ := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = getPositional(args)
	}
	if dir == "" {
		dir = "."
	}

	manifest, result, err := nutshell.Check(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	if jsonMode {
		out := map[string]interface{}{
			"status":   manifest.Completeness.Status,
			"missing":  result.Errors,
			"warnings": result.Warnings,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Println()
	fmt.Printf("  %s🐚 Nutshell Completeness Check%s\n\n", bold, reset)

	// Show present items
	if manifest.Task.Title != "" {
		fmt.Printf("  %s✓%s task.title: \"%s\"\n", green, reset, manifest.Task.Title)
	}
	if manifest.Task.Summary != "" {
		fmt.Printf("  %s✓%s task.summary: provided\n", green, reset)
	}
	if len(manifest.Tags.SkillsRequired) > 0 {
		fmt.Printf("  %s✓%s tags: %s\n", green, reset, strings.Join(manifest.Tags.SkillsRequired, ", "))
	}

	// Check context files
	checkFilePresent(dir, "context.requirements", manifest.Context.Requirements)
	checkFilePresent(dir, "context.architecture", manifest.Context.Architecture)
	checkFilePresent(dir, "context.references", manifest.Context.References)
	for _, a := range manifest.Context.Additional {
		checkFilePresent(dir, "context.additional", a)
	}

	fmt.Println()

	// Show errors (missing items)
	for _, e := range result.Errors {
		fmt.Printf("  %s✗%s %s\n", red, reset, e)
	}
	for _, w := range result.Warnings {
		fmt.Printf("  %s⚠%s %s\n", yellow, reset, w)
	}

	fmt.Println()
	if manifest.Completeness != nil {
		switch manifest.Completeness.Status {
		case "ready":
			fmt.Printf("  Status: %sREADY%s — bundle is complete, agent can start\n", green, reset)
		case "incomplete":
			fmt.Printf("  Status: %sINCOMPLETE%s — %d items need attention before agent can start\n",
				red, reset, len(result.Errors))
		default:
			fmt.Printf("  Status: %sDRAFT%s\n", yellow, reset)
		}
	}
	fmt.Println()
}

func checkFilePresent(dir, label, path string) {
	if path == "" {
		return
	}
	full := filepath.Join(dir, path)
	info, err := os.Stat(full)
	if err == nil {
		fmt.Printf("  %s✓%s %s: exists (%s)\n", green, reset, label, humanSize(info.Size()))
	}
}

func humanSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	kb := float64(bytes) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.1f MB", mb)
}
