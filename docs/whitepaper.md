# Nutshell: Task Packaging for AI Agents

**A Whitepaper on Context Architecture and Harness Engineering**

*Version 0.2.0 — March 2026*

---

## Abstract

As AI agents become central to software engineering workflows, a fundamental infrastructure gap has emerged: there is no standard way to package the context an agent needs to execute a task. Requirements scatter across Slack threads, credentials are shared insecurely, architecture knowledge lives in engineers' heads, and every new agent session starts from scratch.

**Nutshell** addresses this gap with an open standard and CLI toolchain for packaging task context into self-contained `.nut` bundles. A single bundle carries everything — requirements, source files, credentials, acceptance criteria, and execution constraints — so any agent can start immediately without asking 20 clarifying questions.

Nutshell is grounded in **Harness Engineering**: the emerging discipline of building the infrastructure layer that connects, protects, and orchestrates AI agents without doing the work itself.

---

## 1. The Problem: Context Fragility

Modern AI coding agents (Claude Code, Cursor, Copilot, Aider) are remarkably capable. Yet they consistently fail at the same thing: obtaining complete context.

### The Scatter Problem

| Context Component | Where It Lives Today |
|---|---|
| Task requirements | Slack messages, Notion docs, email threads |
| Architecture decisions | Engineer's memory, scattered READMEs |
| Database schemas | Migration files, wiki pages |
| API credentials | Environment variables, shared `.env` files |
| Acceptance criteria | Jira tickets, verbal agreements |
| Execution constraints | Implicit team knowledge |

Every agent session begins with the same ritual: the human re-explains context, the agent asks clarifying questions, and 15 messages later, actual work begins. When the session resets, so does the context.

### The Cost

- **Time waste**: 30–60% of agent interaction is context transfer, not execution.
- **Context drift**: Information mutates across re-tellings. The agent's understanding diverges from reality.
- **Security exposure**: Credentials shared in plaintext chat. No audit trail. No expiration.
- **No reproducibility**: If the agent fails, there's no way to replay the task with identical context.

---

## 2. Harness Engineering: The Theoretical Foundation

### What Is a Harness?

> *"In every engineering discipline, a harness is the same thing: the layer that connects, protects, and orchestrates components — without doing the work itself."*

The core agent loop is deceptively simple:

```
while (model returns tool calls):
    execute tool → capture result → append to context → call model again
```

Every production agent — Claude Code, Cursor, Codex — implements this loop. The differentiation is not in the loop itself, but in what surrounds it: how context is structured, how credentials are injected, how constraints are enforced, and how results are verified.

This surrounding infrastructure is the **harness**.

### The Big Model vs. Big Harness Debate

The AI industry debates whether value accrues to the model or the harness:

**Big Model position**: "The secret sauce is all in the model. The harness should be the thinnest possible wrapper." — Boris Cherny (Anthropic), describing Claude Code's architecture.

**Big Harness position**: "The biggest barrier to getting value from AI is your own ability to context and workflow engineer the models." — Jerry Liu (LlamaIndex).

Research suggests both matter. Scale AI's SWE-Atlas found that harness choice produces measurable but modest differences in agent performance. However, structured context — the right information, in the right format, at the right time — consistently improves every model.

### Nutshell's Position: Structured Context Is the Leverage Point

Nutshell does not replace the model or the orchestration loop. It standardizes what goes *into* the context window. This is the highest-leverage intervention because:

1. **It's model-agnostic**: Good context improves GPT, Claude, Gemini, and open-source models equally.
2. **It's harness-agnostic**: Any orchestrator can consume a `.nut` bundle. No vendor lock-in.
3. **It compounds**: Once context is packaged, it persists across sessions, agents, and teams.
4. **It's auditable**: Request and delivery bundles form a complete record of what was asked and what was done.

---

## 3. The Nutshell Standard

### 3.1 Bundle Format

A `.nut` file is a gzip-compressed tar archive prefixed with `NUT\x01` magic bytes. The MIME type is `application/x-nutshell+gzip`.

```
task.nut (binary)
├── NUT\x01          ← 4-byte magic header
└── gzip(tar(
    ├── nutshell.json    ← Manifest (always loaded first)
    ├── context/         ← Requirements, architecture, references
    ├── files/           ← Source code, data, assets
    ├── credentials/     ← Encrypted credential vault
    ├── tests/           ← Acceptance test scripts
    └── delivery/        ← Completion artifacts
))
```

Only `nutshell.json` is required. All other directories are optional — add them as the task demands.

### 3.2 Manifest Schema

The manifest (`nutshell.json`) is the entry point. Agents read this first to understand what they need to do.

**Required fields:**
- `nutshell_version` — Spec version (`"0.2.0"`)
- `bundle_type` — `request` | `delivery` | `template` | `checkpoint` | `partial`
- `id` — Unique identifier (`nut-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)
- `task.title` — What needs to be done

**Context fields:**
- `task.summary` — Detailed description
- `task.priority` — `critical` | `high` | `medium` | `low`
- `task.estimated_effort` — Human-readable estimate (`"2h"`, `"1d"`)
- `context.requirements` — Path to requirements document
- `context.architecture` — Path to architecture document
- `tags.skills_required` — Skills the agent needs (`["golang", "postgresql"]`)
- `tags.domains` — Domain categories (`["backend", "auth"]`)

**Security fields:**
- `credentials.vault` — Path to encrypted vault file
- `credentials.encryption` — `age` | `sops` | `vault` | `none`
- `credentials.scopes[]` — Named, typed, scoped credentials with expiration

**Execution control:**
- `harness.agent_type_hint` — `execution` | `research` | `review` | `creative`
- `harness.context_budget_hint` — Target context window fill ratio (0.0–1.0)
- `harness.execution_strategy` — `incremental` | `batch` | `streaming`
- `harness.constraints` — Machine-readable restrictions
- `acceptance.checklist` — What "done" looks like
- `acceptance.test_scripts` — Automated verification scripts

**Integrity:**
- `files.tree[].hash` — SHA-256 hash per file (`sha256:<hex>`)
- `completeness.status` — `draft` | `incomplete` | `ready`

**Extensibility:**
- `extensions` — Arbitrary platform-specific metadata (ClawNet, Linear, GitHub Actions)

### 3.3 Bundle Types

| Type | Direction | Purpose |
|---|---|---|
| `request` | Human → Agent | "Here's what I need you to do" |
| `delivery` | Agent → Human | "Here's what I did, and how" |
| `template` | Reusable | Blank task structure for common patterns |
| `checkpoint` | Agent → Self | Intermediate state for long-running tasks |
| `partial` | Agent → Agent | Sub-task result in multi-agent splitting |

The `request → delivery` pair forms a complete audit trail. The `parent_id` field links delivery bundles back to their originating request.

---

## 4. Reverse Management

### The Inversion

Traditional workflow: Agent asks human for missing context. Human provides it piecemeal. Context degrades with each exchange.

**Nutshell workflow**: The bundle tells the human what's missing *before* any agent is involved.

```
$ nutshell check

  🐚 Nutshell Completeness Check

  ✓ task.title: "Build REST API for User Management"
  ✓ task.summary: provided
  ✓ context/requirements.md: exists (2.1 KB)
  ✗ context/architecture.md: referenced but missing
  ✗ credentials: no vault — agent won't have DB access
  ⚠ acceptance: no test scripts — agent can't self-verify

  Status: INCOMPLETE — 2 items need attention
```

This is **reverse management**: the tooling manages the human, ensuring agents receive complete context from the start. The `check` command is the most powerful feature in Nutshell — it shifts the burden of completeness from the agent to the infrastructure.

### What Gets Checked

| Check | Pass Condition |
|---|---|
| Task title | Non-empty |
| Task summary | Non-empty |
| Context files | All referenced paths exist on disk |
| Credentials | Vault configured, scopes have expiry |
| Acceptance criteria | Checklist or test scripts defined |
| Harness constraints | At least one constraint set |
| Skills/domain tags | At least one tag present |
| Context budget | ≤ 0.5 (warn if too high) |

---

## 5. Credential Security Architecture

Nutshell treats credentials as first-class infrastructure, not afterthoughts.

### Principles

| Principle | Implementation |
|---|---|
| **Scoped** | Each credential narrowed to specific tables, endpoints, actions |
| **Time-bounded** | Every credential has `expires_at` |
| **Encrypted** | Default: [age encryption](https://age-encryption.org/). Also supports SOPS, HashiCorp Vault |
| **Rate-limited** | Per-credential rate limits |
| **Auditable** | Delivery bundles log which credentials were used |
| **Rotatable** | `nutshell rotate` audits expiry status and updates expiration dates |

### Credential Lifecycle

```
1. Author defines scopes in nutshell.json
2. `nutshell check` warns if expiry is missing
3. `nutshell rotate` audits expiration status
4. `nutshell pack` includes encrypted vault in bundle
5. Agent decrypts at runtime using provided key
6. Delivery bundle records credential usage
```

---

## 6. Context-Aware Compression

Not all files compress equally. Nutshell's compression engine classifies files by type and applies optimal strategies:

| Category | Examples | Strategy |
|---|---|---|
| Text | `.go`, `.py`, `.md`, `.json` | Compress aggressively (high gzip level) |
| Precompressed | `.jpg`, `.png`, `.zip`, `.gz` | Store without recompression |
| Binary | Other binary files | Compress at default level |

The `nutshell compress` command also estimates **token count** for text files (at ~0.25 tokens/byte), helping engineers understand how much context window a bundle will consume.

---

## 7. Multi-Agent Splitting

Complex tasks can be decomposed into parallel sub-tasks with `nutshell split`:

```
$ nutshell split --dir my-task -n 3

  Created 3 sub-tasks:
    part-0/  →  "Backend: src/"
    part-1/  →  "Tests: tests/"
    part-2/  →  "Docs: docs/"
```

Each sub-task:
- Gets its own `nutshell.json` with `bundle_type: "partial"`
- Stores the original task's ID in `parent_id`
- Receives shared context (`context/` directory copied to all)
- Records split metadata in `extensions.split`

After agents complete their sub-tasks, `nutshell merge` combines delivery bundles into a unified result.

---

## 8. Toolchain

### CLI Commands

| Command | Purpose |
|---|---|
| `init` | Scaffold a new task directory |
| `check` | Completeness verification (reverse management) |
| `pack` | Create `.nut` bundle with SHA-256 integrity |
| `unpack` | Extract bundle with path traversal protection |
| `inspect` | View manifest without extracting |
| `validate` | Check against specification |
| `set` | Quick-edit manifest fields via dot-path |
| `diff` | Compare request vs. delivery bundles |
| `schema` | Output JSON Schema for IDE integration |
| `compress` | Context-aware compression with token estimation |
| `split` | Decompose task into parallel sub-tasks |
| `merge` | Combine delivery sub-bundles |
| `rotate` | Audit and update credential expiration |
| `serve` | Local web viewer for bundle inspection |

### Agent SDK (Go)

```go
import "github.com/ChatChatTech/nutshell/pkg/nutshell"

b, _ := nutshell.Open("task.nut")
m := b.Manifest()
content, _ := b.ReadFile("context/requirements.md")
files := b.FilesByPrefix("context/")
```

### VS Code Extension

- JSON Schema validation and auto-completion for `nutshell.json`
- Snippets for manifest sections
- Commands for pack/unpack/inspect/validate via command palette
- Auto-validate on save with diagnostics
- Context menu integration for `.nut` files

### Web Viewer

`nutshell serve` launches a local web viewer with:
- Manifest display with structured fields
- File browser with content preview
- Completeness status dashboard
- REST API for programmatic access

---

## 9. Platform Extensions

Nutshell is standalone-first but extensible. The `extensions` field supports arbitrary platform integrations without breaking the core format.

### ClawNet Integration

[ClawNet](https://github.com/ChatChatTech/ClawNet) is a decentralized AI agent communication network. Nutshell integrates natively:

```bash
nutshell publish --dir my-task    # Pack & publish to P2P network
nutshell claim <task-id>          # Claim and unpack a remote task
nutshell deliver --dir workspace  # Pack & submit delivery
```

### Other Extensions

```json
{
  "extensions": {
    "clawnet": {"peer_id": "12D3KooW...", "reward": 50.0},
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

Extensions never break the core format. Tools ignore what they don't understand.

---

## 10. Technical Details

| Property | Value |
|---|---|
| Language | Go 1.26 |
| Dependencies | Zero (stdlib only) |
| Bundle format | `NUT\x01` + gzip + tar |
| Hash algorithm | SHA-256 (file-level + bundle-level) |
| Credential encryption | age, SOPS, HashiCorp Vault |
| Spec version | 0.2.0 |
| MIME type | `application/x-nutshell+gzip` |
| Test coverage | 44 unit tests, 70%+ coverage |
| License | MIT |

---

## 11. Conclusion

The AI agent landscape is evolving rapidly. Models improve quarterly. Harnesses change weekly. But the need for structured, secure, portable context is permanent.

Nutshell bets on a simple idea: if you give agents better context, they produce better results — regardless of which model or harness you use. By standardizing how context is packaged, we create infrastructure that compounds across sessions, agents, teams, and platforms.

**Pack it. Crack it. Ship it.** 🐚

---

*Nutshell is an open standard by [ChatChatTech](https://github.com/ChatChatTech). Contributions welcome.*
