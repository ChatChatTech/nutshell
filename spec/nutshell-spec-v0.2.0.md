# Nutshell Packaging Specification v0.2.0

> **Nutshell** — An open standard for packaging task context that AI agents can understand, consume, and act upon.

---

## 1. Overview

Nutshell is a **standalone, agent-agnostic** task packaging format. Any AI coding agent — Claude Code, GitHub Copilot, Cursor, Aider, or a custom agent — can use Nutshell bundles to receive structured tasks, manage execution, and archive results.

Nutshell works in two directions:
- **For agents**: receive a complete, structured task with everything needed to execute
- **For humans**: get prompted for missing context — Nutshell tells you what's still needed before an agent can begin

### Core Design Principles

1. **Standalone-First** — Works without any external platform. A single developer with Claude Code benefits immediately.
2. **Layered Context** — Tiered loading respects context window limits. Manifest first, details on demand.
3. **Agent-Readable** — Every field has a machine-parseable schema. No ambiguity.
4. **Credential-First** — Shared API access and credentials are first-class, not afterthoughts.
5. **Bidirectional** — Covers both **task publishing** (request) and **task completion** (delivery).
6. **Completeness-Aware** — The bundle knows what's missing and can prompt humans to fill gaps.
7. **Extensible** — Optional extensions for platforms like ClawNet, GitHub Actions, etc.

### Use Cases

| Scenario | How Nutshell Helps |
|----------|-------------------|
| **Solo developer + Claude Code** | Create a `.nut` bundle defining a task. Claude Code reads it, executes, produces a delivery `.nut`. Archived for future reference. |
| **Team handoff** | Engineer packages a task as `.nut`, hands to another engineer (or their agent). All context travels with the bundle. |
| **Reverse management** | `nutshell check` tells the human: "You're missing DB credentials and the architecture doc. Fill these in before the agent can start." |
| **Task archive** | Completed delivery bundles serve as structured records: what was requested, what was done, what decisions were made. |
| **P2P marketplace** | (Extension) Publish `.nut` bundles to ClawNet for decentralized agent matching and execution. |

---

## 2. Bundle Structure

A Nutshell Bundle (`.nut`) is a compressed archive:

```
task-name.nut
├── nutshell.json          # Manifest — always loaded first
├── context/               # Detailed context files
│   ├── requirements.md    # Detailed requirements
│   ├── architecture.md    # System architecture docs
│   ├── references.md      # External references & links
│   └── ...                # Any additional context docs
├── files/                 # Related source files & assets
│   ├── src/               # Code files
│   ├── data/              # Data files
│   └── assets/            # Images, diagrams, etc.
├── apis/                  # API specifications
│   ├── endpoints.json     # Callable API definitions
│   └── schemas/           # Request/response schemas
├── credentials/           # Scoped credentials
│   └── vault.enc.json     # Encrypted credential vault
├── tests/                 # Acceptance criteria
│   ├── criteria.json      # Machine-readable acceptance tests
│   └── scripts/           # Test scripts
└── delivery/              # Completion artifacts (delivery bundles)
    ├── result.json        # Completion manifest
    ├── artifacts/         # Delivered files
    └── logs/              # Execution logs & decisions
```

Not all directories are required. A minimal bundle needs only `nutshell.json`.

---

## 3. Manifest Schema (`nutshell.json`)

The manifest is the entry point — always loaded first, kept compact.

```jsonc
{
  // === IDENTITY ===
  "nutshell_version": "0.2.0",
  "bundle_type": "request",              // "request" | "delivery"
  "id": "nut-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "created_at": "2026-03-14T10:00:00Z",
  "expires_at": "2026-03-21T10:00:00Z",  // optional deadline

  // === TASK METADATA ===
  "task": {
    "title": "Build a REST API for user management",
    "summary": "Create CRUD endpoints with JWT auth, PostgreSQL storage, and rate limiting.",
    "priority": "high",                   // "critical" | "high" | "medium" | "low"
    "estimated_effort": "8h"              // ISO 8601 duration or human string
  },

  // === TAGS ===
  "tags": {
    "skills_required": ["golang", "postgresql", "jwt", "rest-api"],
    "domains": ["backend", "authentication"],
    "data_sources": [],
    "custom": {
      "framework": "gin",
      "go_version": "1.22+"
    }
  },

  // === PUBLISHER (who created this bundle) ===
  "publisher": {
    "name": "Alice Chen",                 // human or agent name
    "contact": "alice@example.com",       // optional
    "tool": "claude-code"                 // which tool created this (optional)
  },

  // === CONTEXT MANIFEST (pointers to detail files) ===
  "context": {
    "requirements": "context/requirements.md",
    "architecture": "context/architecture.md",
    "references": "context/references.md",
    "additional": []
  },

  // === FILE MANIFEST ===
  "files": {
    "total_count": 12,
    "total_size_bytes": 45000,
    "tree": [
      {"path": "files/src/main.go", "size": 2400, "role": "scaffold"},
      {"path": "files/src/models/user.go", "size": 800, "role": "reference"},
      {"path": "files/data/schema.sql", "size": 1200, "role": "specification"}
    ]
  },

  // === API ACCESS ===
  "apis": {
    "endpoints_spec": "apis/endpoints.json",
    "base_urls": {
      "staging": "https://api-staging.example.com",
      "docs": "https://docs.example.com/api"
    },
    "auth_method": "bearer_token",
    "credential_ref": "credentials/vault.enc.json"
  },

  // === CREDENTIALS ===
  "credentials": {
    "vault": "credentials/vault.enc.json",
    "encryption": "age",                   // "age" | "sops" | "vault" | "none"
    "scopes": [
      {
        "name": "staging-db",
        "type": "postgresql",
        "access_level": "read-write",
        "expires_at": "2026-03-21T10:00:00Z"
      }
    ]
  },

  // === ACCEPTANCE CRITERIA ===
  "acceptance": {
    "criteria_file": "tests/criteria.json",
    "test_scripts": ["tests/scripts/test_api.sh"],
    "auto_verifiable": true,
    "checklist": [
      "All CRUD endpoints respond with correct status codes",
      "JWT authentication works for protected routes",
      "Rate limiting returns 429 after threshold"
    ]
  },

  // === EXECUTION HINTS (Harness Engineering guidance) ===
  "harness": {
    "agent_type_hint": "execution",        // "research" | "planning" | "execution" | "review"
    "context_budget_hint": 0.35,           // Target context window utilization (0.0-1.0)
    "execution_strategy": "incremental",   // "one-shot" | "incremental" | "parallel"
    "checkpoints": true,
    "constraints": [
      "Do not modify files outside files/src/",
      "All new code must have corresponding tests",
      "Follow existing code style in reference files"
    ]
  },

  // === RELATED RESOURCES ===
  "resources": {
    "repos": [
      {"url": "https://github.com/example/user-service", "branch": "develop", "relevance": "primary"}
    ],
    "docs": [
      {"url": "https://docs.example.com/api-guide", "title": "API Design Guide"}
    ],
    "images": [
      {"path": "files/assets/architecture.png", "description": "System architecture diagram"}
    ],
    "links": [
      {"url": "https://jira.example.com/PROJ-123", "title": "Original ticket"}
    ]
  },

  // === COMPLETENESS (what's still missing?) ===
  "completeness": {
    "status": "ready",                     // "draft" | "incomplete" | "ready"
    "missing": [],                         // e.g. ["credentials.vault", "context.architecture"]
    "warnings": []                         // e.g. ["No acceptance criteria defined"]
  },

  // === COMPRESSION METADATA ===
  "compression": {
    "algorithm": "gzip",                   // "gzip" | "zstd" | "none"
    "original_size_bytes": 128000,
    "compressed_size_bytes": 45000,
    "context_tokens_estimate": 12000
  },

  // === EXTENSIONS (optional platform integrations) ===
  "extensions": {}
}
```

### Required Fields

Only these fields are strictly required:
- `nutshell_version`
- `bundle_type`
- `id`
- `task.title`

Everything else is optional but improves agent effectiveness.

---

## 4. Completeness Check

A key feature of Nutshell: **the bundle knows what it's missing.**

When a human runs `nutshell check`, the tool inspects the manifest and directory, identifies gaps, and prompts:

```
$ nutshell check ./my-task

  🐚 Nutshell Completeness Check

  ✓ task.title: "Build REST API for User Management"
  ✓ task.summary: provided
  ✓ context/requirements.md: exists (2.1 KB)
  ✗ context/architecture.md: referenced but missing
  ✗ credentials: no vault configured — agent won't have DB access
  ⚠ acceptance: no test scripts — agent can't self-verify
  ⚠ harness.constraints: empty — agent has no guardrails

  Status: INCOMPLETE — 2 items need attention before agent can start

  To fix:
    1. Create context/architecture.md with system architecture
    2. Add credentials with: nutshell add-credential --name staging-db --type postgresql
```

This reverses the typical dynamic: instead of the agent asking the human "what else do I need?", the **bundle tells the human** what to provide.

---

## 5. Delivery Bundle Schema

When an agent completes work, it produces a delivery bundle.

### `delivery/result.json`

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "delivery",
  "request_id": "nut-a1b2c3d4-...",        // Reference to original request
  "id": "nut-d4c3b2a1-...",
  "completed_at": "2026-03-15T18:30:00Z",

  "deliverer": {
    "name": "claude-code",                  // agent or human name
    "model": "claude-sonnet-4-20260514",    // optional: which model
    "sessions_used": 3,
    "total_tokens": 45000
  },

  "status": "completed",                    // "completed" | "partial" | "blocked" | "failed"
  "completion_percentage": 100,
  "summary": "All 4 CRUD endpoints implemented with JWT auth, rate limiting, and PostgreSQL migrations.",

  "artifacts": {
    "files_created": [
      {"path": "delivery/artifacts/src/main.go", "lines": 150},
      {"path": "delivery/artifacts/src/handlers/user.go", "lines": 280}
    ],
    "files_modified": [
      {"path": "delivery/artifacts/src/models/user.go", "diff_lines": 30}
    ]
  },

  "acceptance_results": {
    "tests_passed": 12,
    "tests_failed": 0,
    "tests_skipped": 1,
    "checklist": [
      {"item": "All CRUD endpoints respond with correct status codes", "status": "passed"},
      {"item": "JWT authentication works for protected routes", "status": "passed"},
      {"item": "Rate limiting returns 429 after threshold", "status": "passed"}
    ]
  },

  "execution_log": {
    "strategy_used": "incremental",
    "checkpoints": [
      {"at": "2026-03-15T10:00:00Z", "description": "Scaffolded project structure"},
      {"at": "2026-03-15T12:30:00Z", "description": "Implemented CRUD handlers"},
      {"at": "2026-03-15T15:00:00Z", "description": "Added JWT middleware"}
    ],
    "decisions": [
      {"decision": "Used gin-jwt/v2 instead of custom JWT implementation", "reason": "Better maintained, covers edge cases"},
      {"decision": "Used golang-migrate for DB migrations", "reason": "Recommended in architecture.md"}
    ],
    "issues_encountered": [],
    "full_log": "delivery/logs/execution.log"
  },

  "tags": {
    "skills_used": ["golang", "postgresql", "jwt", "gin", "rest-api"],
    "tools_used": ["go-test", "curl", "psql"],
    "complexity_actual": "medium"
  }
}
```

---

## 6. Tag System

Tags are agent-agnostic labels for categorizing tasks and capabilities.

### Standard Categories

| Category | Key | Format | Example |
|----------|-----|--------|---------|
| Required Skills | `skills_required` | string[] | `["golang", "postgresql"]` |
| Domains | `domains` | string[] | `["backend", "devops"]` |
| Data Sources | `data_sources` | string[] | `["postgresql://host"]` |
| Custom | `custom.*` | any | `{"framework": "gin"}` |

### Custom Tags

Any key under `tags.custom` is valid. This enables domain-specific extensions without polluting the core schema:

```json
{
  "custom": {
    "gpu_required": true,
    "region": "us-east-1",
    "max_cost_usd": 10.00,
    "security_clearance": "internal"
  }
}
```

---

## 7. Credential Vault

Credentials are first-class citizens in Nutshell. Agents need the same access as human engineers.

### Vault Structure (Decrypted)

```jsonc
{
  "vault_version": "0.2.0",
  "created_at": "2026-03-14T10:00:00Z",
  "expires_at": "2026-03-21T10:00:00Z",
  "credentials": [
    {
      "name": "staging-db",
      "type": "postgresql",
      "value": {
        "host": "staging-db.internal",
        "port": 5432,
        "database": "userservice",
        "username": "agent_writer",
        "password": "rotated-token-abc123"
      },
      "scopes": ["read", "write"],
      "restrictions": {
        "tables": ["users", "sessions"],
        "max_rows_per_query": 1000
      }
    },
    {
      "name": "api-token",
      "type": "bearer_token",
      "value": {
        "token": "sk-agent-scoped-xyz789"
      },
      "scopes": ["invoke"],
      "restrictions": {
        "endpoints": ["/api/v1/users*"],
        "rate_limit": "100/min"
      }
    }
  ]
}
```

### Encryption

Default encryption uses [age](https://age-encryption.org/):

```bash
# Encrypt vault
age -r age1recipient... -o vault.enc.json vault.json

# Decrypt
age -d -i identity.key vault.enc.json > vault.json
```

Supported backends:
- **age** — Default. Lightweight, no infrastructure.
- **sops** — Mozilla SOPS for cloud KMS integration.
- **vault** — HashiCorp Vault transit engine.
- **none** — Plaintext (development only).

### Security Model

1. **Scoped Access**: Credentials narrowed to what the task requires
2. **Time-Bounded**: Every credential has an expiration
3. **Restriction Tags**: Database tables, API endpoints, rate limits
4. **Rotation**: Publisher can rotate without re-publishing the bundle
5. **Audit Trail**: Delivery bundles log which credentials were used

---

## 8. API Specification Format

`apis/endpoints.json` describes callable APIs:

```jsonc
{
  "api_version": "0.2.0",
  "base_url": "https://api-staging.example.com",
  "auth": {
    "type": "bearer_token",
    "credential_ref": "api-token"
  },
  "endpoints": [
    {
      "method": "GET",
      "path": "/api/v1/users",
      "description": "List all users with pagination",
      "params": {
        "page": {"type": "integer", "default": 1},
        "per_page": {"type": "integer", "default": 20, "max": 100}
      },
      "response_schema": "schemas/user_list.json"
    },
    {
      "method": "POST",
      "path": "/api/v1/users",
      "description": "Create a new user",
      "body_schema": "schemas/user_create.json"
    }
  ]
}
```

---

## 9. Acceptance Criteria Format

Machine-readable tests in `tests/criteria.json`:

```jsonc
{
  "criteria_version": "0.2.0",
  "test_framework": "shell",               // "shell" | "pytest" | "go-test" | "jest" | "custom"
  "auto_verifiable": true,
  "criteria": [
    {
      "id": "AC-001",
      "description": "GET /api/v1/users returns 200 with user list",
      "type": "api_test",
      "script": "tests/scripts/test_api.sh",
      "expected": {"status_code": 200, "body_contains": "users"}
    },
    {
      "id": "AC-002",
      "description": "Database migration creates users table",
      "type": "sql_check",
      "query": "SELECT count(*) FROM information_schema.tables WHERE table_name = 'users'",
      "expected": {"result": 1}
    }
  ]
}
```

---

## 10. Extensions

Extensions allow optional platform integrations without polluting the core schema. All extensions live under the `extensions` key in the manifest.

### 10.1 ClawNet Extension

For publishing tasks on the [ClawNet](https://github.com/ChatChatTech/ClawNet) P2P agent network:

```jsonc
{
  "extensions": {
    "clawnet": {
      "peer_id": "12D3KooWAbCdEf...",
      "reputation": 85.0,
      "reward": {
        "amount": 500,
        "currency": "shells"
      },
      "gossip_topic": "/clawnet/tasks"
    }
  }
}
```

**Tag mapping** when published to ClawNet:

| Nutshell Field | ClawNet Field |
|----------------|---------------|
| `tags.skills_required` | `Task.Tags` |
| `tags.domains` | `AgentResume.Skills` |
| `tags.data_sources` | `AgentResume.DataSources` |
| `extensions.clawnet.reward` | `Task.Reward` |
| `extensions.clawnet.peer_id` | `Task.AuthorID` |

**Publishing flow:**

```
Publisher                     ClawNet Network                Agent
   │                              │                           │
   ├── nutshell pack task.nut     │                           │
   ├── POST /api/tasks ──────────►│ store Task                │
   │   (attach: task.nut hash)    │ gossip to peers           │
   │                              ├──────────────────────────►│
   │                              │ matched by tags            │
   │                              │                           │
   │                              │ POST /api/tasks/{id}/bid  │
   │                              │◄────────────────────────── │
   │   review bids                │                           │
   │◄──────────────────────────── │                           │
   ├── assign + share .nut ──────►│──────────────────────────►│
   │                              │    agent unpacks .nut     │
   │                              │    packs delivery.nut     │
   │                              │ POST /api/tasks/{id}/submit
   │                              │◄────────────────────────── │
   │   review delivery.nut       │                           │
   ├── POST /api/tasks/{id}/approve                           │
   │                              │ transfer reward           │
   └──────────────────────────────┴───────────────────────────┘
```

### 10.2 Writing Custom Extensions

Any tool can add fields under `extensions.<name>`:

```jsonc
{
  "extensions": {
    "github-actions": {
      "workflow": ".github/workflows/agent-task.yml",
      "runner": "ubuntu-latest"
    },
    "linear": {
      "issue_id": "ENG-1234",
      "project": "Backend"
    }
  }
}
```

Extensions MUST NOT override core fields. Tools SHOULD ignore unknown extensions gracefully.

---

## 11. MIME Type & File Extension

- **Extension**: `.nut`
- **MIME Type**: `application/x-nutshell+gzip`
- **Magic Bytes**: `NUT\x01` (4 bytes header)

---

## 12. Versioning

The spec follows [SemVer](https://semver.org/):
- **MAJOR**: Breaking schema changes
- **MINOR**: New optional features, additive fields
- **PATCH**: Clarifications, typo fixes

Current: **v0.2.0**

### Changelog

- **v0.2.0**: Standalone-first redesign. Added completeness check. ClawNet fields moved to extensions.
- **v0.1.0**: Initial draft (ClawNet-coupled).
