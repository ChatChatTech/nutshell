package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
	fmt.Print(shellArt)
	fmt.Println("Usage:")
	fmt.Println("  nutshell init     [--dir <path>]                  Initialize a new bundle directory")
	fmt.Println("  nutshell pack     [--dir <path>] [-o <file>]      Pack directory into .nut bundle")
	fmt.Println("  nutshell unpack   <file> [-o <path>]              Unpack a .nut bundle")
	fmt.Println("  nutshell inspect  <file|-> [--json]               Inspect bundle without unpacking")
	fmt.Println("  nutshell validate <file|dir> [--json]             Validate bundle against spec")
	fmt.Println("  nutshell check    [--dir <path>] [--json]         Completeness check — what's missing?")
	fmt.Println("  nutshell set      <key> <value> [--dir <path>]    Quick-edit a manifest field")
	fmt.Println("  nutshell diff     <a.nut> <b.nut> [--json]        Compare two bundles")
	fmt.Println("  nutshell schema   [-o <file>]                     Output JSON Schema for nutshell.json")
	fmt.Println()
	fmt.Println("  Advanced:")
	fmt.Println("  nutshell compress [--dir <path>] [-o <file>] [--level fast|best]  Context-aware compression")
	fmt.Println("  nutshell split    [--dir <path>] [-n <count>]     Split task into parallel sub-tasks")
	fmt.Println("  nutshell merge    <dir1> <dir2> ... [-o <dir>]    Merge delivery sub-bundles")
	fmt.Println("  nutshell rotate   <scope> [--expires <date>] [--dir <path>]  Rotate credential expiry")
	fmt.Println("  nutshell serve    [<file|dir>] [--port <port>]    Web viewer for .nut inspection")
	fmt.Println()
	fmt.Println("  ClawNet Integration (optional — requires running ClawNet daemon):")
	fmt.Println("  nutshell publish  [--dir <path>] [--clawnet <addr>]  Pack & publish task to ClawNet network")
	fmt.Println("  nutshell claim    <task-id> [--clawnet <addr>]       Claim a task from ClawNet, create local dir")
	fmt.Println("  nutshell deliver  [--dir <path>] [--clawnet <addr>]  Pack delivery & submit to ClawNet task")
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
	case "set":
		cmdSet(args)
	case "diff":
		cmdDiff(args)
	case "schema":
		cmdSchema(args)
	case "compress":
		cmdCompress(args)
	case "split":
		cmdSplit(args)
	case "merge":
		cmdMerge(args)
	case "rotate":
		cmdRotate(args)
	case "serve":
		cmdServe(args)
	case "publish":
		cmdPublish(args)
	case "claim":
		cmdClaim(args)
	case "deliver":
		cmdDeliver(args)
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

	fmt.Print(shellArt)
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

func cmdSet(args []string) {
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}

	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell set <key> <value> [--dir <path>]\n", red, reset)
		fmt.Fprintf(os.Stderr, "  Example: nutshell set task.title \"Build REST API\"\n")
		os.Exit(1)
	}

	key := args[0]
	value := args[1]

	if err := nutshell.Set(dir, key, value); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s✓%s Set %s%s%s = %s\n", green, reset, cyan, key, reset, value)
}

func cmdDiff(args []string) {
	jsonMode, args := hasFlag(args, "--json")

	// Collect non-flag positional args
	var positional []string
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
		}
	}

	if len(positional) < 2 {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell diff <a.nut|dir> <b.nut|dir> [--json]\n", red, reset)
		os.Exit(1)
	}

	diffs, err := nutshell.Diff(positional[0], positional[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	if jsonMode {
		data, _ := json.MarshalIndent(diffs, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Println()
	fmt.Printf("  %s🐚 Nutshell Diff%s\n", bold, reset)
	fmt.Printf("  %sA:%s %s\n", dim, reset, positional[0])
	fmt.Printf("  %sB:%s %s\n\n", dim, reset, positional[1])

	if len(diffs) == 0 {
		fmt.Printf("  %s✓ No differences found.%s\n\n", green, reset)
		return
	}

	for _, d := range diffs {
		fmt.Printf("  %s•%s %s%s%s\n", yellow, reset, bold, d.Field, reset)
		if d.A != "" {
			fmt.Printf("    %s- %s%s\n", red, d.A, reset)
		}
		if d.B != "" {
			fmt.Printf("    %s+ %s%s\n", green, d.B, reset)
		}
	}
	fmt.Println()
}

func cmdSchema(args []string) {
	output, _ := getFlag(args, "--output", "-o")
	schema := nutshell.Schema()
	if output != "" {
		if err := os.WriteFile(output, []byte(schema+"\n"), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
			os.Exit(1)
		}
		fmt.Printf("%s✓%s Schema written to %s%s%s\n", green, reset, bold, output, reset)
		return
	}
	fmt.Println(schema)
}

// ── Context-aware compression ──

func cmdCompress(args []string) {
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}
	output, args := getFlag(args, "--output", "-o")
	levelStr, _ := getFlag(args, "--level", "-l")

	level := nutshell.CompressBest
	switch levelStr {
	case "none":
		level = nutshell.CompressNone
	case "fast":
		level = nutshell.CompressFast
	case "default", "":
		level = nutshell.CompressDefault
	case "best":
		level = nutshell.CompressBest
	}

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

	manifest, plan, err := nutshell.PackWithCompression(dir, output, level)
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

	fmt.Printf("%s✓%s Packed %s%d%s files with context-aware compression\n",
		green, reset, cyan, manifest.Files.TotalCount, reset)
	fmt.Printf("  %sOutput: %s%s\n", dim, output, reset)
	fmt.Printf("  %sOriginal: %d bytes → Compressed: %d bytes (%.1f%% reduction)%s\n",
		dim, origSize, compSize, ratio, reset)

	// Show compression analysis
	var textCount, precompCount, binaryCount int
	for _, f := range plan.Files {
		switch f.Category {
		case "text":
			textCount++
		case "precompressed", "media":
			precompCount++
		default:
			binaryCount++
		}
	}
	fmt.Printf("  %sAnalysis: %d text, %d pre-compressed, %d binary%s\n",
		dim, textCount, precompCount, binaryCount, reset)
	if plan.EstimatedTokens > 0 {
		fmt.Printf("  %sEstimated context tokens: ~%d%s\n", dim, plan.EstimatedTokens, reset)
	}
}

// ── Multi-agent splitting ──

func cmdSplit(args []string) {
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}
	nStr, _ := getFlag(args, "-n", "--count")

	var plan *nutshell.SplitPlan

	// If -n is specified, create N equal sub-tasks
	if nStr != "" {
		n := 2
		fmt.Sscanf(nStr, "%d", &n)
		if n < 2 {
			n = 2
		}

		data, err := os.ReadFile(filepath.Join(dir, "nutshell.json"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s No nutshell.json found in %s\n", red, reset, dir)
			os.Exit(1)
		}
		var m nutshell.Manifest
		json.Unmarshal(data, &m)

		subs := make([]nutshell.SubTask, n)
		for i := range subs {
			subs[i] = nutshell.SubTask{
				Title: fmt.Sprintf("%s (Part %d)", m.Task.Title, i+1),
			}
		}
		plan = &nutshell.SplitPlan{SubTasks: subs}
	}

	results, err := nutshell.Split(dir, plan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s✓%s Split into %s%d%s sub-tasks\n", green, reset, cyan, len(results), reset)
	for _, r := range results {
		fmt.Printf("  %s[%d]%s %s → %s%s/%s\n", dim, r.Index, reset, r.Title, bold, r.Directory, reset)
	}
	fmt.Printf("\n  %sEach sub-task has parent_id linking back to the original bundle.%s\n", dim, reset)
	fmt.Printf("  %sMerge deliveries with: nutshell merge <dir1> <dir2> ... -o merged/%s\n", dim, reset)
}

func cmdMerge(args []string) {
	output, args := getFlag(args, "--output", "-o")
	if output == "" {
		output = "merged"
	}

	// Collect all positional args as directories
	var dirs []string
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			dirs = append(dirs, a)
		}
	}

	if len(dirs) < 2 {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell merge <dir1> <dir2> ... [-o <dir>]\n", red, reset)
		os.Exit(1)
	}

	manifest, err := nutshell.Merge(dirs, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s✓%s Merged %s%d%s sub-bundles into %s%s/%s\n",
		green, reset, cyan, len(dirs), reset, bold, output, reset)
	fmt.Printf("  %sBundle ID: %s%s\n", dim, manifest.ID, reset)
	fmt.Printf("  %sType: %s%s\n", dim, manifest.BundleType, reset)
}

// ── Credential rotation ──

func cmdRotate(args []string) {
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}
	expires, args := getFlag(args, "--expires", "-e")
	scope := getPositional(args)

	if scope == "" {
		// No scope — run audit
		statuses, err := nutshell.AuditCredentials(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
			os.Exit(1)
		}
		if len(statuses) == 0 {
			fmt.Printf("  %sNo credentials defined in manifest.%s\n", dim, reset)
			return
		}

		fmt.Println()
		fmt.Printf("  %s🔑 Credential Audit%s\n\n", bold, reset)
		for _, s := range statuses {
			switch s.Status {
			case "expired":
				fmt.Printf("  %s✗%s %s (%s) — %sEXPIRED%s", red, reset, s.Name, s.Type, red, reset)
				if s.ExpiresAt != "" {
					fmt.Printf(" (was %s)", s.ExpiresAt)
				}
				fmt.Println()
			case "expiring_soon":
				fmt.Printf("  %s⚠%s %s (%s) — %sexpires in %d days%s (%s)\n",
					yellow, reset, s.Name, s.Type, yellow, s.DaysLeft, reset, s.ExpiresAt)
			case "valid":
				fmt.Printf("  %s✓%s %s (%s) — valid (%d days left)\n",
					green, reset, s.Name, s.Type, s.DaysLeft)
			case "no_expiry":
				fmt.Printf("  %s⚠%s %s (%s) — %sno expiry set%s\n",
					yellow, reset, s.Name, s.Type, yellow, reset)
			}
		}
		fmt.Printf("\n  %sTo rotate: nutshell rotate <scope> [--expires 2026-06-01T00:00:00Z]%s\n\n", dim, reset)
		return
	}

	result, err := nutshell.RotateCredential(dir, scope, expires)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s✓%s Rotated credential %s%s%s\n", green, reset, cyan, result.Scope, reset)
	if result.OldExpiry != "" {
		fmt.Printf("  %sOld expiry: %s%s\n", dim, result.OldExpiry, reset)
	}
	fmt.Printf("  %sNew expiry: %s%s\n", dim, result.NewExpiry, reset)
}

// ── Web viewer ──

func cmdServe(args []string) {
	portStr, args := getFlag(args, "--port", "-p")
	target := getPositional(args)
	if target == "" {
		target = "."
	}

	port := 8080
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	addr, _, err := nutshell.ServeViewer(target, port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Print(shellArt)
	fmt.Printf("  %s🌐 Web Viewer%s\n", bold, reset)
	fmt.Printf("  %sServing:%s %s\n", dim, reset, target)
	fmt.Printf("  %sOpen:%s http://%s\n\n", dim, reset, addr)
	fmt.Printf("  Press Ctrl+C to stop.\n")

	// Block until interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Printf("\n%s✓%s Server stopped.\n", green, reset)
}

// ── ClawNet integration commands ──

func getClawNetClient(args []string) (*nutshell.ClawNetClient, []string) {
	addr, rest := getFlag(args, "--clawnet", "-c")
	return nutshell.NewClawNetClient(addr), rest
}

func cmdPublish(args []string) {
	client, args := getClawNetClient(args)
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}

	// Verify ClawNet is reachable
	status, err := client.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		fmt.Fprintf(os.Stderr, "  %sHint: Is ClawNet daemon running? (clawnet start)%s\n", dim, reset)
		os.Exit(1)
	}

	// Read manifest
	data, err := os.ReadFile(filepath.Join(dir, "nutshell.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s No nutshell.json in %s\n", red, reset, dir)
		os.Exit(1)
	}
	var manifest nutshell.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s Invalid nutshell.json: %s\n", red, reset, err)
		os.Exit(1)
	}

	if manifest.Task.Title == "" {
		fmt.Fprintf(os.Stderr, "%s✗%s task.title is required to publish\n", red, reset)
		os.Exit(1)
	}

	// Pack the bundle
	slug := strings.ToLower(manifest.Task.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	nutFile := slug + ".nut"
	_, err = nutshell.Pack(dir, nutFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s Pack failed: %s\n", red, reset, err)
		os.Exit(1)
	}

	// Hash the bundle
	nutHash, _ := nutshell.HashBundle(nutFile)

	// Publish task to ClawNet
	task, err := client.PublishTask(&manifest, nutHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	// Upload the .nut bundle
	if err := client.UploadBundle(task.ID, nutFile); err != nil {
		// Non-fatal: task was created, bundle upload is optional
		fmt.Printf("  %s⚠%s Bundle upload skipped: %s\n", yellow, reset, err)
	}

	// Write ClawNet extension back into manifest
	if manifest.Extensions == nil {
		manifest.Extensions = make(map[string]json.RawMessage)
	}
	manifest.Extensions["clawnet"] = nutshell.ManifestToClawNetExtension(status.PeerID, task.ID, task.Reward)
	updated, _ := json.MarshalIndent(manifest, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), updated, 0644)

	fmt.Printf("%s✓%s Published to ClawNet network\n", green, reset)
	fmt.Printf("  %sTask ID:  %s%s\n", dim, task.ID, reset)
	fmt.Printf("  %sPeer:     %s (%s)%s\n", dim, status.AgentName, status.PeerID[:16]+"...", reset)
	fmt.Printf("  %sReward:   %.1f energy%s\n", dim, task.Reward, reset)
	fmt.Printf("  %sBundle:   %s%s\n", dim, nutFile, reset)
	fmt.Printf("  %sHash:     %s%s\n", dim, nutHash, reset)
}

func cmdClaim(args []string) {
	client, args := getClawNetClient(args)
	outDir, args := getFlag(args, "--output", "-o")
	taskID := getPositional(args)

	if taskID == "" {
		fmt.Fprintf(os.Stderr, "%s✗%s Usage: nutshell claim <task-id> [--clawnet <addr>] [-o <dir>]\n", red, reset)
		os.Exit(1)
	}

	// Verify ClawNet is reachable
	_, err := client.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	// Fetch task details
	task, err := client.GetTask(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	if outDir == "" {
		slug := strings.ToLower(task.Title)
		slug = strings.ReplaceAll(slug, " ", "-")
		if len(slug) > 40 {
			slug = slug[:40]
		}
		if slug == "" {
			slug = taskID[:8]
		}
		outDir = slug
	}

	// Try to download the .nut bundle
	nutFile := filepath.Join(os.TempDir(), taskID+".nut")
	err = client.DownloadBundle(taskID, nutFile)
	if err == nil {
		// Unpack the bundle
		_, err := nutshell.Unpack(nutFile, outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s✗%s Unpack failed: %s\n", red, reset, err)
			os.Exit(1)
		}
		os.Remove(nutFile) // clean up temp file
		fmt.Printf("%s✓%s Claimed task and unpacked bundle to %s%s/%s\n", green, reset, bold, outDir, reset)
	} else {
		// No bundle attached — create from task metadata
		os.MkdirAll(filepath.Join(outDir, "context"), 0755)
		m := nutshell.NewManifest()
		m.Task.Title = task.Title
		m.Task.Summary = task.Description
		if task.Deadline != "" {
			m.ExpiresAt = task.Deadline
		}
		// Parse tags
		var tags []string
		json.Unmarshal([]byte(task.Tags), &tags)
		m.Tags.SkillsRequired = tags
		// Add ClawNet extension
		if m.Extensions == nil {
			m.Extensions = make(map[string]json.RawMessage)
		}
		m.Extensions["clawnet"] = nutshell.ManifestToClawNetExtension(task.AuthorID, task.ID, task.Reward)
		data, _ := json.MarshalIndent(m, "", "  ")
		os.WriteFile(filepath.Join(outDir, "nutshell.json"), data, 0644)

		reqPath := filepath.Join(outDir, "context", "requirements.md")
		os.WriteFile(reqPath, []byte(fmt.Sprintf("# %s\n\n%s\n", task.Title, task.Description)), 0644)

		fmt.Printf("%s✓%s Claimed task and created bundle directory %s%s/%s\n", green, reset, bold, outDir, reset)
		fmt.Printf("  %s(No .nut bundle was attached — created from task metadata)%s\n", dim, reset)
	}

	fmt.Printf("  %sTask: %s%s\n", dim, task.Title, reset)
	fmt.Printf("  %sReward: %.1f energy%s\n", dim, task.Reward, reset)
	fmt.Printf("  %sRun 'nutshell check --dir %s' to see what's needed%s\n", dim, outDir, reset)
}

func cmdDeliver(args []string) {
	client, args := getClawNetClient(args)
	dir, args := getFlag(args, "--dir", "-d")
	if dir == "" {
		dir = "."
	}

	// Read manifest to find ClawNet task ID
	data, err := os.ReadFile(filepath.Join(dir, "nutshell.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s No nutshell.json in %s\n", red, reset, dir)
		os.Exit(1)
	}
	var manifest nutshell.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s Invalid nutshell.json: %s\n", red, reset, err)
		os.Exit(1)
	}

	// Extract ClawNet task ID from extensions
	taskID := ""
	if ext, ok := manifest.Extensions["clawnet"]; ok {
		var clawExt map[string]interface{}
		json.Unmarshal(ext, &clawExt)
		if tid, ok := clawExt["task_id"].(string); ok {
			taskID = tid
		}
	}
	if taskID == "" {
		fmt.Fprintf(os.Stderr, "%s✗%s No extensions.clawnet.task_id found in manifest\n", red, reset)
		fmt.Fprintf(os.Stderr, "  %sHint: Was this task claimed from ClawNet? (nutshell claim <task-id>)%s\n", dim, reset)
		os.Exit(1)
	}

	// Verify ClawNet is reachable
	_, err = client.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	// Switch to delivery bundle type
	manifest.BundleType = "delivery"
	manifest.ParentID = manifest.ID
	manifest.ID = nutshell.GenerateID()
	updated, _ := json.MarshalIndent(manifest, "", "  ")
	os.WriteFile(filepath.Join(dir, "nutshell.json"), updated, 0644)

	// Pack the delivery bundle
	slug := strings.ToLower(manifest.Task.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	nutFile := slug + "-delivery.nut"
	_, err = nutshell.Pack(dir, nutFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s Pack failed: %s\n", red, reset, err)
		os.Exit(1)
	}

	nutHash, _ := nutshell.HashBundle(nutFile)

	// Upload delivery bundle
	if err := client.UploadBundle(taskID, nutFile); err != nil {
		fmt.Printf("  %s⚠%s Bundle upload skipped: %s\n", yellow, reset, err)
	}

	// Submit result to ClawNet
	result := fmt.Sprintf("Delivery bundle: %s (hash: %s)", nutFile, nutHash)
	if err := client.SubmitDelivery(taskID, result, nutHash); err != nil {
		fmt.Fprintf(os.Stderr, "%s✗%s %s\n", red, reset, err)
		os.Exit(1)
	}

	fmt.Printf("%s✓%s Delivered to ClawNet\n", green, reset)
	fmt.Printf("  %sTask ID: %s%s\n", dim, taskID, reset)
	fmt.Printf("  %sBundle:  %s%s\n", dim, nutFile, reset)
	fmt.Printf("  %sHash:    %s%s\n", dim, nutHash, reset)
}
