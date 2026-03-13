# Nutshell — Task Packaging for AI Agents

> 🐚 Skill for creating, inspecting, validating, and managing Nutshell task bundles.

## When to Use This Skill

Use Nutshell when you need to:
- **Package a task** with all its context, files, credentials, and acceptance criteria into a single `.nut` bundle
- **Hand off work** between humans and AI agents (or between agents) with no context loss
- **Archive completed work** as structured delivery bundles with decisions, artifacts, and execution logs
- **Check what's missing** before starting — let the bundle tell the human what to provide (`nutshell check`)
- **Inspect a received bundle** to understand what's being asked before executing

Do NOT use Nutshell for:
- Simple one-line requests that don't need context packaging
- Tasks where all context is already in the conversation
- Real-time streaming communication (use messaging/DM instead)

## Install

```bash
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell
make build

# or install globally
go install ./cmd/nutshell/
```

## Core Concepts

**Bundle types:**
- `request` — A task to be done. Contains requirements, context files, credentials, acceptance criteria.
- `delivery` — A completed task. Contains artifacts, execution log, decisions made, acceptance results.

**Bundle format:** A `.nut` file is a gzip-compressed tar archive with `NUT\x01` magic bytes. The entry point is always `nutshell.json` (the manifest).

**Standalone-first:** Nutshell works without any external platform. One developer + one AI agent is the base use case. Platform integrations (ClawNet, GitHub Actions, Linear) are optional extensions.

**Reverse management:** The `check` command tells the *human* what's missing, inverting the typical dynamic where agents have to ask for context.

## CLI Commands

### Initialize a bundle directory
```bash
nutshell init [--dir <path>]
```
Creates `nutshell.json` manifest and `context/` directory. Edit the manifest to define your task.

### Check completeness (reverse management)
```bash
nutshell check [--dir <path>] [--json]
```
Inspects the manifest and directory to identify what's missing before an agent can start. Checks:
- Required fields (title, summary)
- Referenced files exist (context docs, credential vault, API specs)
- Acceptance criteria defined
- Harness constraints set
- Skills/domain tags present

The `--json` flag outputs machine-readable results for programmatic use.

### Pack a bundle
```bash
nutshell pack [--dir <path>] [-o <file>]
```
Compresses the directory into a `.nut` bundle. Respects `.nutignore` for excluding files. Shows content hash (SHA-256) for integrity verification.

### Inspect a bundle
```bash
nutshell inspect <file|-> [--json]
```
Reads the manifest and file list without extracting. Supports stdin (`-`) for piping:
```bash
cat task.nut | nutshell inspect --json - | jq '.manifest.task.title'
```

### Unpack a bundle
```bash
nutshell unpack <file> [-o <path>]
```
Extracts a `.nut` bundle to a directory. Path traversal protection is enforced.

### Validate against spec
```bash
nutshell validate <file|dir> [--json]
```
Checks the manifest against the Nutshell v0.2.0 specification. Validates required fields, credential security, context budget limits.

## Manifest Structure (`nutshell.json`)

Key fields an agent should understand:

| Field | Purpose |
|-------|---------|
| `task.title` | What to do (required) |
| `task.summary` | Detailed description |
| `task.priority` | critical / high / medium / low |
| `context.requirements` | Path to requirements doc |
| `context.architecture` | Path to architecture doc |
| `credentials.vault` | Encrypted credential vault |
| `acceptance.checklist` | What "done" looks like |
| `harness.constraints` | What the agent must NOT do |
| `harness.agent_type_hint` | research / planning / execution / review |
| `harness.context_budget_hint` | Target context window fill ratio (0.0–1.0) |
| `completeness.status` | draft / incomplete / ready |
| `parent_id` | ID of parent bundle (for chaining) |
| `extensions` | Optional platform integrations |

## Bundle Directory Layout

```
my-task/
├── nutshell.json          # Manifest (always read first)
├── context/               # Requirements, architecture, references
├── files/                 # Source code, data, assets
├── apis/                  # API endpoint specifications
├── credentials/           # Encrypted credential vault
├── tests/                 # Test scripts & acceptance criteria
├── delivery/              # Completion artifacts (delivery bundles)
└── .nutignore             # Files to exclude from packing
```

Only `nutshell.json` is required. Add directories as needed — `nutshell init` only creates `context/` to avoid bloat.

## Typical Workflow

### As a task publisher (human)
```bash
nutshell init --dir my-task
# Edit nutshell.json — set title, summary, tags
# Add context docs to context/
# Add source files to files/
# Add credentials to credentials/
nutshell check --dir my-task     # See what's still missing
nutshell pack --dir my-task -o my-task.nut
# Hand my-task.nut to an agent
```

### As a task executor (agent)
```bash
nutshell unpack task.nut -o workspace
# Read nutshell.json first
# Load context/ files as needed
# Execute according to harness constraints
# Run acceptance tests
nutshell pack --dir workspace -o delivery.nut
# Submit delivery.nut back
```

### Piping and composability
```bash
# Extract just the task title
cat task.nut | nutshell inspect --json - | jq -r '.manifest.task.title'

# Validate in CI
nutshell validate --json ./bundle-dir || exit 1

# Check completeness programmatically
nutshell check --json --dir . | jq '.status'
```

## .nutignore

Works like `.gitignore`. Place in the bundle directory root to exclude files from packing:

```
# Scratch files
notes.txt
*.tmp

# Build artifacts
delivery/
*.log
```

## Extensions

Platform-specific fields live under `extensions` in the manifest. They never break the core format:

```json
{
  "extensions": {
    "clawnet": {
      "peer_id": "12D3KooW...",
      "reward": {"amount": 50, "currency": "energy"}
    }
  }
}
```

Tools should ignore unknown extensions gracefully.

## Specification

Full spec: `spec/nutshell-spec-v0.2.0.md`

MIME type: `application/x-nutshell+gzip`
Magic bytes: `NUT\x01`
