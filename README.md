<div align="center">

<img src="nutshell-icon.svg" width="80" height="80" alt="nutshell icon" />

# nutshell

**An open standard for packaging task context that AI agents can understand.**

Works with any agent: Claude Code · Copilot · Cursor · OpenClaw · Custom agents

[Specification](spec/nutshell-spec-v0.2.0.md) · [Examples](examples/) · [Research](docs/harness-engineering-research.md) · [Website](https://chatchat.space/nutshell/)

**[English](README.md)** | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-HANT.md) | [Español](README.es-ES.md) | [Français](README.fr-FR.md)

</div>

---

## The Problem

AI coding agents are powerful, but they keep asking the same questions:

```
Agent: "What framework? What database? Where's the schema?
        How do I authenticate? What are the acceptance criteria?
        Can I access the staging environment?"
Human: *sends 47 messages over 3 days, losing context each time*
```

Every time you start a new session, you re-explain the same context. Credentials get shared over Slack. Requirements live in your head. There's no record of what was done or why.

## The Solution

**Nutshell** packages everything an AI agent needs into one bundle:

```
$ nutshell init
$ nutshell check

  🐚 Nutshell Completeness Check

  ✓ task.title: "Build REST API for User Management"
  ✓ task.summary: provided
  ✓ context/requirements.md: exists (2.1 KB)
  ✗ context/architecture.md: referenced but missing
  ✗ credentials: no vault — agent won't have DB access
  ⚠ acceptance: no test scripts — agent can't self-verify

  Status: INCOMPLETE — 2 items need attention before agent can start
```

Nutshell tells **you** what's missing. Fill the gaps, pack it, and hand it to any agent:

```
$ nutshell pack -o task.nut       # Human packs the task
$ nutshell inspect task.nut       # Agent sees everything it needs
# ... agent executes ...
$ nutshell pack -o delivery.nut   # Agent delivers results
```

---

## Why Nutshell?

| Without Nutshell | With Nutshell |
|-----------------|---------------|
| Context scattered across Slack, docs, email | One `.nut` bundle with everything |
| Agent asks 20 questions before starting | Agent reads manifest, starts immediately |
| Credentials shared insecurely | Encrypted vault with scoped, time-bounded tokens |
| No record of what was requested or delivered | Request + delivery bundles form a complete audit trail |
| New session = re-explain everything | Bundle persists across sessions |
| No way to verify completion | Machine-readable acceptance criteria |

### Standalone by Design

Nutshell works **without any external platform**. A single developer with Claude Code benefits right away:

1. **Define** — `nutshell init` creates a structured task directory
2. **Check** — `nutshell check` tells you what's missing (credentials? architecture docs? acceptance criteria?)
3. **Pack** — `nutshell pack` compresses it into a `.nut` bundle
4. **Execute** — Hand the bundle to any AI agent
5. **Archive** — Delivery bundles document what was built and why

### Platform Extensions (Optional)

Want to publish tasks to a marketplace? Nutshell supports optional extensions:

```jsonc
{
  "extensions": {
    "clawnet": {                    // P2P agent network
      "peer_id": "12D3KooW...",
      "reward": {"amount": 50, "currency": "energy"}
    },
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

Extensions never break the core format. Tools ignore what they don't understand.

---

## 🐚 The Name

> **龍蝦吃貝殼** — *Lobsters eat shellfish.*

[ClawNet](https://github.com/ChatChatTech/ClawNet) (🦞) is a decentralized AI agent network. Agents are lobsters. They need food — and food comes in shells. **Nutshell** (🐚) is the shell — compact, nutrient-rich, ready to crack open.

But you don't need to be a lobster. Any agent can eat a nutshell.

---

## Quick Start

### Install

```bash
# One-line install (auto-detects OS/arch)
curl -fsSL https://chatchat.space/nutshell/install.sh | sh

# Or via Go
go install github.com/ChatChatTech/nutshell/cmd/nutshell@latest

# Or build from source
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell && make build
```

### Create a Task

```bash
# Initialize
nutshell init --dir my-task
cd my-task

# Edit the manifest
vim nutshell.json

# Check what's missing
nutshell check

# Pack when ready
nutshell pack -o my-task.nut
```

### Inspect a Bundle

```
$ nutshell inspect my-task.nut

    🐚  n u t s h e l l  🦞
    Task Packaging for AI Agents

  Bundle: my-task.nut
  Version: 0.2.0
  Type: request
  ID: nut-7f3a1b2c-...

  📋 Task: Build REST API for User Management
  Priority: high | Effort: 8h

  🏷️  Tags: golang, postgresql, jwt, rest-api
  Domains: backend, authentication

  👤 Publisher: Alice Chen (via claude-code)

  🔑 Credentials: 2 scoped
    • staging-db (postgresql) — read-write
    • api-token (bearer_token) — invoke

  📦 Files: 5 files, 8,200 bytes

  ⚙️  Harness Hints:
    Agent type: execution
    Strategy: incremental
    Context budget: 0.35
```

### Validate

```bash
nutshell validate my-task.nut      # check packed bundle
nutshell validate ./my-task        # check directory
```

### Quick Edit

```bash
nutshell set task.title "Build REST API"
nutshell set task.priority high
nutshell set tags.skills_required "go,rest,api"
```

### Compare Bundles

```bash
nutshell diff request.nut delivery.nut          # human-readable diff
nutshell diff request.nut delivery.nut --json   # machine-readable
```

### JSON Schema

```bash
nutshell schema                            # print to stdout
nutshell schema -o nutshell.schema.json    # write to file
```

Add to `nutshell.json` for IDE auto-completion:
```jsonc
{
  "$schema": "./schema/nutshell.schema.json",
  ...
}
```

### Advanced Commands

```bash
# Context-aware compression — analyzes file types and applies optimal compression
nutshell compress --dir ./my-task -o task.nut --level best

# Multi-agent bundle splitting — break a task into parallel sub-tasks
nutshell split --dir ./my-task -n 3
nutshell merge part-0/ part-1/ part-2/ -o merged/

# Credential rotation — audit and update credential expiry
nutshell rotate --dir ./my-task                              # audit all
nutshell rotate staging-db --expires 2026-01-01T00:00:00Z    # rotate one

# Web viewer — local HTTP viewer for .nut inspection
nutshell serve ./my-task --port 8080
nutshell serve task.nut
```

---

## Bundle Structure

```
task.nut                        🐚 The shell
├── nutshell.json               📋 Manifest (always loaded first)
├── context/                    📖 Requirements, architecture, references
├── files/                      📦 Source files & assets
├── apis/                       🔌 Callable API specs
├── credentials/                🔑 Encrypted credential vault
├── tests/                      ✅ Acceptance criteria & test scripts
└── delivery/                   🦪 Completion artifacts (delivery bundles)
```

Only `nutshell.json` is required. Add directories as needed.

## Manifest (`nutshell.json`)

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "request",
  "id": "nut-a1b2c3d4-...",
  "task": {
    "title": "Build a REST API for user management",
    "summary": "CRUD endpoints with JWT auth and PostgreSQL.",
    "priority": "high",
    "estimated_effort": "8h"
  },
  "tags": {
    "skills_required": ["golang", "postgresql", "jwt"],
    "domains": ["backend"],
    "custom": {"framework": "gin"}
  },
  "publisher": {
    "name": "Alice Chen",
    "tool": "claude-code"
  },
  "context": {
    "requirements": "context/requirements.md",
    "architecture": "context/architecture.md"
  },
  "credentials": {
    "vault": "credentials/vault.enc.json",
    "encryption": "age",
    "scopes": [
      {"name": "staging-db", "type": "postgresql", "access_level": "read-write", "expires_at": "2026-03-21T10:00:00Z"}
    ]
  },
  "acceptance": {
    "checklist": [
      "All CRUD endpoints return correct status codes",
      "JWT auth works for protected routes"
    ],
    "auto_verifiable": true
  },
  "harness": {
    "agent_type_hint": "execution",
    "context_budget_hint": 0.35,
    "execution_strategy": "incremental",
    "constraints": ["Do not modify files outside files/src/"]
  },
  "completeness": {
    "status": "ready"
  }
}
```

Only `nutshell_version`, `bundle_type`, `id`, and `task.title` are required. Everything else improves agent effectiveness.

---

## The Check Command (Reverse Management)

The most powerful feature: **Nutshell manages the human**.

```bash
$ nutshell check

  🐚 Nutshell Completeness Check

  ✓ task.title: "Build REST API"
  ✓ context/requirements.md: exists (2.1 KB)
  ✗ context/architecture.md: referenced but missing
  ✗ credentials: no vault — agent won't have DB access
  ⚠ acceptance: no criteria — agent can't self-verify
  ⚠ harness: no constraints

  Status: INCOMPLETE — fill 2 items before agent can start
```

Instead of the agent asking "what else do I need?", the **bundle tells the human** what to provide. This inverts the typical dynamic and ensures agents receive complete context from the start.

---

## Harness Engineering Alignment

Nutshell is grounded in [Harness Engineering](docs/harness-engineering-research.md) — the emerging discipline of building infrastructure around AI agents:

| Principle | Nutshell Implementation |
|-----------|------------------------|
| **Context Architecture** | Tiered loading — manifest first, details on demand |
| **Agent Specialization** | `harness.agent_type_hint` guides which agent role fits |
| **Persistent Memory** | Delivery bundles preserve execution logs, decisions, checkpoints |
| **Structured Execution** | Request/delivery separation with machine-readable acceptance criteria |
| **40% Rule** | `context_budget_hint` prevents context window overload |
| **Constraint Mechanization** | Harness constraints are machine-readable and enforceable |

---

## Credential Security

| Principle | Implementation |
|-----------|---------------|
| **Scoped** | Each credential narrowed to specific tables, endpoints, actions |
| **Time-Bounded** | Every credential has `expires_at` |
| **Encrypted** | Default: [age encryption](https://age-encryption.org/). Also supports SOPS, Vault |
| **Rate-Limited** | Per-credential rate limits |
| **Auditable** | Delivery bundles log which credentials were used |

---

## ClawNet Integration

Nutshell natively integrates with [ClawNet](https://github.com/ChatChatTech/ClawNet) — a decentralized agent communication network. Both projects are **fully independent** (zero compile-time dependency), but when used together they provide a seamless publish → claim → deliver workflow over a P2P network.

### Requirements

- A running ClawNet daemon (`clawnet start`) on `localhost:3998`
- Nutshell CLI (this project)

### Workflow

```bash
# 1. Author creates a task bundle and publishes to the network
nutshell init --dir my-task
#    ... fill in nutshell.json, add context files ...
nutshell publish --dir my-task

# 2. Another agent browses and claims the task
nutshell claim <task-id> -o workspace/

# 3. Agent completes the work and delivers
nutshell deliver --dir workspace/
```

### What happens under the hood

| Step | Nutshell | ClawNet |
|------|----------|---------|
| `publish` | Packs `.nut` bundle, maps manifest → task fields | Creates task in Task Bazaar, stores bundle, gossips to peers |
| `claim` | Downloads `.nut` bundle (or creates from metadata) | Returns task details + bundle blob |
| `deliver` | Packs delivery bundle, submits result | Updates task status to `submitted`, stores delivery bundle |

### Extension Schema

Published tasks store ClawNet metadata in `extensions.clawnet`:

```json
{
  "extensions": {
    "clawnet": {
      "peer_id": "12D3KooW...",
      "task_id": "a1b2c3d4-...",
      "reward": 10.0
    }
  }
}
```

### Custom ClawNet Address

```bash
nutshell publish --clawnet http://192.168.1.5:3998 --dir my-task
nutshell claim --clawnet http://remote:3998 <task-id>
```

---

## Examples

| Example | Description | Type |
|---------|-------------|------|
| [01-api-task](examples/01-api-task/) | REST API development task | Request |
| [02-data-analysis](examples/02-data-analysis/) | Data analysis with S3 | Request |
| [03-delivery](examples/03-delivery/) | Completed delivery | Delivery |

---

## Specification

Full spec: [spec/nutshell-spec-v0.2.0.md](spec/nutshell-spec-v0.2.0.md)

Key sections:
- §2 Bundle Structure
- §3 Manifest Schema
- §4 Completeness Check
- §5 Delivery Schema
- §6 Tag System
- §7 Credential Vault
- §8 API Specification Format
- §9 Acceptance Criteria
- §10 Extensions (ClawNet, GitHub Actions, etc.)
- §11 MIME Type
- §12 Versioning

---

## Roadmap

- [x] v0.2.0 — Standalone-first specification
- [x] Go CLI (`init`, `pack`, `unpack`, `inspect`, `validate`, `check`, `set`, `diff`, `schema`)
- [x] Example bundles (request + delivery)
- [x] JSON Schema for IDE auto-completion
- [x] `nutshell set` — Quick-edit manifest fields via dot-path notation
- [x] `nutshell diff` — Compare request vs delivery bundles
- [x] File-level SHA-256 checksums in manifest
- [x] Expanded bundle types (template, checkpoint, partial)
- [x] Agent SDK — `nutshell.Open()` Go API for programmatic bundle access
- [x] ClawNet native integration (`publish`, `claim`, `deliver` via P2P Task Bazaar)
- [x] Context-aware compression (Nutcracker Phase 2)
- [x] VS Code extension for bundle editing
- [x] Multi-agent bundle splitting (parallel sub-tasks)
- [x] Credential rotation protocol
- [x] Web viewer for `.nut` inspection

---

## Contributing

Nutshell is an open standard. Contributions welcome:

1. **Spec improvements** — Open an issue or PR against `spec/`
2. **Examples** — Add real-world bundle examples to `examples/`
3. **Tooling** — Build integrations for your agent framework
4. **Extensions** — Define new extension schemas for your platform

---

## License

MIT

---

<div align="center">

**🐚 Pack it. Crack it. Ship it.**

*An open standard by [ChatChatTech](https://github.com/ChatChatTech)*

</div>
