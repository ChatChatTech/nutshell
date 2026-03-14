# Nutshell 打包规范 v0.2.0

> **Nutshell** — 一种开放标准，用于打包 AI Agent 能够理解、消费和执行的任务上下文。

---

## 1. 概述

Nutshell 是一种**独立的、Agent 无关的**任务打包格式。任何 AI 编码 Agent — Claude Code、GitHub Copilot、Cursor、Aider 或自定义 Agent — 都可以使用 Nutshell 包来接收结构化任务、管理执行和归档结果。

Nutshell 在两个方向上工作：
- **面向 Agent**：接收完整、结构化的任务，包含执行所需的一切
- **面向人类**：提示缺失的上下文 — Nutshell 在 Agent 开始之前告诉你还需要什么

### 核心设计原则

1. **独立优先** — 无需任何外部平台即可运行。单个开发者配合 Claude Code 即可立即受益。
2. **分层上下文** — 分层加载尊重上下文窗口限制。清单优先，细节按需加载。
3. **Agent 可读** — 每个字段都有机器可解析的 schema。无歧义。
4. **凭证优先** — 共享的 API 访问和凭证是一等公民，而非事后补丁。
5. **双向的** — 涵盖**任务发布**（request）和**任务完成**（delivery）。
6. **完整性感知** — 包知道自己缺少什么，并能提示人类填补空白。
7. **可扩展** — 可选的平台扩展，如 ClawNet、GitHub Actions 等。

### 使用场景

| 场景 | Nutshell 如何帮助 |
|----------|-------------------|
| **个人开发者 + Claude Code** | 创建一个定义任务的 `.nut` 包。Claude Code 读取它、执行、生成一个 delivery `.nut`。归档供未来参考。 |
| **团队交接** | 工程师将任务打包为 `.nut`，交给另一位工程师（或他们的 Agent）。所有上下文随包一起传递。 |
| **反向管理** | `nutshell check` 告诉人类："你缺少数据库凭证和架构文档。在 Agent 开始之前请填写这些。" |
| **任务归档** | 完成的 delivery 包作为结构化记录：请求了什么、做了什么、做了哪些决策。 |
| **P2P 市场** | （扩展）将 `.nut` 包发布到 ClawNet 进行去中心化 Agent 匹配和执行。 |

---

## 2. 包结构

Nutshell 包（`.nut`）是一个压缩归档：

```
task-name.nut
├── nutshell.json          # 清单 — 最先加载
├── context/               # 详细上下文文件
│   ├── requirements.md    # 详细需求
│   ├── architecture.md    # 系统架构文档
│   ├── references.md      # 外部参考和链接
│   └── ...                # 任何额外的上下文文档
├── files/                 # 相关源文件和资产
│   ├── src/               # 代码文件
│   ├── data/              # 数据文件
│   └── assets/            # 图片、图表等
├── apis/                  # API 规范
│   ├── endpoints.json     # 可调用的 API 定义
│   └── schemas/           # 请求/响应 Schema
├── credentials/           # 范围受限的凭证
│   └── vault.enc.json     # 加密的凭证库
├── tests/                 # 验收标准
│   ├── criteria.json      # 机器可读的验收测试
│   └── scripts/           # 测试脚本
└── delivery/              # 完成产物（delivery 包）
    ├── result.json        # 完成清单
    ├── artifacts/         # 交付文件
    └── logs/              # 执行日志和决策记录
```

并非所有目录都是必需的。最小化的包只需要 `nutshell.json`。

---

## 3. 清单 Schema（`nutshell.json`）

清单是入口点 — 始终最先加载，保持精简。

```jsonc
{
  // === 身份标识 ===
  "nutshell_version": "0.2.0",
  "bundle_type": "request",              // "request" | "delivery"
  "id": "nut-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "created_at": "2026-03-14T10:00:00Z",
  "expires_at": "2026-03-21T10:00:00Z",  // 可选截止日期

  // === 任务元数据 ===
  "task": {
    "title": "构建用户管理 REST API",
    "summary": "创建带 JWT 认证、PostgreSQL 存储和速率限制的 CRUD 端点。",
    "priority": "high",                   // "critical" | "high" | "medium" | "low"
    "estimated_effort": "8h"              // ISO 8601 时长或人类可读的字符串
  },

  // === 标签 ===
  "tags": {
    "skills_required": ["golang", "postgresql", "jwt", "rest-api"],
    "domains": ["backend", "authentication"],
    "data_sources": [],
    "custom": {
      "framework": "gin",
      "go_version": "1.22+"
    }
  },

  // === 发布者（谁创建了这个包） ===
  "publisher": {
    "name": "Alice Chen",                 // 人类或 Agent 名称
    "contact": "alice@example.com",       // 可选
    "tool": "claude-code"                 // 使用哪个工具创建（可选）
  },

  // === 上下文清单（指向详细文件的指针） ===
  "context": {
    "requirements": "context/requirements.md",
    "architecture": "context/architecture.md",
    "references": "context/references.md",
    "additional": []
  },

  // === 文件清单 ===
  "files": {
    "total_count": 12,
    "total_size_bytes": 45000,
    "tree": [
      {"path": "files/src/main.go", "size": 2400, "role": "scaffold"},
      {"path": "files/src/models/user.go", "size": 800, "role": "reference"},
      {"path": "files/data/schema.sql", "size": 1200, "role": "specification"}
    ]
  },

  // === API 访问 ===
  "apis": {
    "endpoints_spec": "apis/endpoints.json",
    "base_urls": {
      "staging": "https://api-staging.example.com",
      "docs": "https://docs.example.com/api"
    },
    "auth_method": "bearer_token",
    "credential_ref": "credentials/vault.enc.json"
  },

  // === 凭证 ===
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

  // === 验收标准 ===
  "acceptance": {
    "criteria_file": "tests/criteria.json",
    "test_scripts": ["tests/scripts/test_api.sh"],
    "auto_verifiable": true,
    "checklist": [
      "所有 CRUD 端点返回正确的状态码",
      "JWT 认证对受保护路由生效",
      "速率限制在超出阈值后返回 429"
    ]
  },

  // === 执行提示（Harness 工程指导） ===
  "harness": {
    "agent_type_hint": "execution",        // "research" | "planning" | "execution" | "review"
    "context_budget_hint": 0.35,           // 目标上下文窗口利用率 (0.0-1.0)
    "execution_strategy": "incremental",   // "one-shot" | "incremental" | "parallel"
    "checkpoints": true,
    "constraints": [
      "不要修改 files/src/ 之外的文件",
      "所有新代码必须有对应的测试",
      "遵循参考文件中的现有代码风格"
    ]
  },

  // === 相关资源 ===
  "resources": {
    "repos": [
      {"url": "https://github.com/example/user-service", "branch": "develop", "relevance": "primary"}
    ],
    "docs": [
      {"url": "https://docs.example.com/api-guide", "title": "API 设计指南"}
    ],
    "images": [
      {"path": "files/assets/architecture.png", "description": "系统架构图"}
    ],
    "links": [
      {"url": "https://jira.example.com/PROJ-123", "title": "原始工单"}
    ]
  },

  // === 完整性（还缺少什么？） ===
  "completeness": {
    "status": "ready",                     // "draft" | "incomplete" | "ready"
    "missing": [],                         // 例如 ["credentials.vault", "context.architecture"]
    "warnings": []                         // 例如 ["未定义验收标准"]
  },

  // === 压缩元数据 ===
  "compression": {
    "algorithm": "gzip",                   // "gzip" | "zstd" | "none"
    "original_size_bytes": 128000,
    "compressed_size_bytes": 45000,
    "context_tokens_estimate": 12000
  },

  // === 扩展（可选的平台集成） ===
  "extensions": {}
}
```

### 必需字段

仅以下字段严格必需：
- `nutshell_version`
- `bundle_type`
- `id`
- `task.title`

其他所有字段都是可选的，但能提高 Agent 的效率。

---

## 4. 完整性检查

Nutshell 的关键特性：**包知道自己缺少什么。**

当人类运行 `nutshell check` 时，工具会检查清单和目录，识别空白，并提示：

```
$ nutshell check ./my-task

  🐚 Nutshell 完整性检查

  ✓ task.title: "构建用户管理 REST API"
  ✓ task.summary: 已提供
  ✓ context/requirements.md: 存在 (2.1 KB)
  ✗ context/architecture.md: 已引用但缺失
  ✗ credentials: 未配置 vault — Agent 将无法访问数据库
  ⚠ acceptance: 无测试脚本 — Agent 无法自我验证
  ⚠ harness.constraints: 为空 — Agent 没有防护栏

  状态: 不完整 — 开始执行前有 2 项需要处理

  修复方法:
    1. 创建 context/architecture.md 包含系统架构
    2. 添加凭证: nutshell add-credential --name staging-db --type postgresql
```

这反转了典型的动态：不是 Agent 问人类"我还需要什么？"，而是**包告诉人类**需要提供什么。

---

## 5. Delivery 包 Schema

当 Agent 完成工作后，它会生成一个 delivery 包。

### `delivery/result.json`

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "delivery",
  "request_id": "nut-a1b2c3d4-...",        // 引用原始请求
  "id": "nut-d4c3b2a1-...",
  "completed_at": "2026-03-15T18:30:00Z",

  "deliverer": {
    "name": "claude-code",                  // Agent 或人类名称
    "model": "claude-sonnet-4-20260514",    // 可选：使用了哪个模型
    "sessions_used": 3,
    "total_tokens": 45000
  },

  "status": "completed",                    // "completed" | "partial" | "blocked" | "failed"
  "completion_percentage": 100,
  "summary": "已实现所有 4 个 CRUD 端点，包含 JWT 认证、速率限制和 PostgreSQL 迁移。",

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
      {"item": "所有 CRUD 端点返回正确的状态码", "status": "passed"},
      {"item": "JWT 认证对受保护路由生效", "status": "passed"},
      {"item": "速率限制在超出阈值后返回 429", "status": "passed"}
    ]
  },

  "execution_log": {
    "strategy_used": "incremental",
    "checkpoints": [
      {"at": "2026-03-15T10:00:00Z", "description": "搭建项目结构"},
      {"at": "2026-03-15T12:30:00Z", "description": "实现 CRUD 处理器"},
      {"at": "2026-03-15T15:00:00Z", "description": "添加 JWT 中间件"}
    ],
    "decisions": [
      {"decision": "使用 gin-jwt/v2 而非自定义 JWT 实现", "reason": "维护更好，覆盖边界情况"},
      {"decision": "使用 golang-migrate 进行数据库迁移", "reason": "architecture.md 中推荐"}
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

## 6. 标签系统

标签是 Agent 无关的标记，用于分类任务和能力。

### 标准分类

| 分类 | 键 | 格式 | 示例 |
|----------|-----|--------|---------|
| 所需技能 | `skills_required` | string[] | `["golang", "postgresql"]` |
| 领域 | `domains` | string[] | `["backend", "devops"]` |
| 数据源 | `data_sources` | string[] | `["postgresql://host"]` |
| 自定义 | `custom.*` | any | `{"framework": "gin"}` |

### 自定义标签

`tags.custom` 下的任何键都是有效的。这使得领域特定扩展不会污染核心 schema：

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

## 7. 凭证库

凭证是 Nutshell 中的一等公民。Agent 需要与人类工程师相同的访问权限。

### 库结构（解密后）

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

### 加密

默认加密使用 [age](https://age-encryption.org/)：

```bash
# 加密 vault
age -r age1recipient... -o vault.enc.json vault.json

# 解密
age -d -i identity.key vault.enc.json > vault.json
```

支持的后端：
- **age** — 默认。轻量级，无需基础设施。
- **sops** — Mozilla SOPS，用于云 KMS 集成。
- **vault** — HashiCorp Vault transit 引擎。
- **none** — 明文（仅限开发环境）。

### 安全模型

1. **范围受限的访问**：凭证缩小到任务所需的范围
2. **时间限定**：每个凭证都有过期时间
3. **限制标签**：数据库表、API 端点、速率限制
4. **轮换**：发布者可以在不重新发布包的情况下轮换凭证
5. **审计轨迹**：Delivery 包记录使用了哪些凭证

---

## 8. API 规范格式

`apis/endpoints.json` 描述可调用的 API：

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
      "description": "列出所有用户，支持分页",
      "params": {
        "page": {"type": "integer", "default": 1},
        "per_page": {"type": "integer", "default": 20, "max": 100}
      },
      "response_schema": "schemas/user_list.json"
    },
    {
      "method": "POST",
      "path": "/api/v1/users",
      "description": "创建新用户",
      "body_schema": "schemas/user_create.json"
    }
  ]
}
```

---

## 9. 验收标准格式

`tests/criteria.json` 中的机器可读测试：

```jsonc
{
  "criteria_version": "0.2.0",
  "test_framework": "shell",               // "shell" | "pytest" | "go-test" | "jest" | "custom"
  "auto_verifiable": true,
  "criteria": [
    {
      "id": "AC-001",
      "description": "GET /api/v1/users 返回 200 和用户列表",
      "type": "api_test",
      "script": "tests/scripts/test_api.sh",
      "expected": {"status_code": 200, "body_contains": "users"}
    },
    {
      "id": "AC-002",
      "description": "数据库迁移创建 users 表",
      "type": "sql_check",
      "query": "SELECT count(*) FROM information_schema.tables WHERE table_name = 'users'",
      "expected": {"result": 1}
    }
  ]
}
```

---

## 10. 扩展

扩展允许可选的平台集成，不污染核心 schema。所有扩展都位于清单的 `extensions` 键下。

### 10.1 ClawNet 扩展

用于在 [ClawNet](https://github.com/ChatChatTech/ClawNet) P2P Agent 网络上发布任务：

```jsonc
{
  "extensions": {
    "clawnet": {
      "peer_id": "12D3KooWAbCdEf...",
      "reputation": 85.0,
      "reward": {
        "amount": 50.0,
        "currency": "energy"
      },
      "gossip_topic": "/clawnet/tasks"
    }
  }
}
```

**发布到 ClawNet 时的标签映射：**

| Nutshell 字段 | ClawNet 字段 |
|----------------|---------------|
| `tags.skills_required` | `Task.Tags` |
| `tags.domains` | `AgentResume.Skills` |
| `tags.data_sources` | `AgentResume.DataSources` |
| `extensions.clawnet.reward` | `Task.Reward` |
| `extensions.clawnet.peer_id` | `Task.AuthorID` |

**发布流程：**

```
发布者                         ClawNet 网络                    Agent
   │                              │                           │
   ├── nutshell pack task.nut     │                           │
   ├── POST /api/tasks ──────────►│ 存储 Task                 │
   │   (附带: task.nut 哈希)       │ 向节点广播                 │
   │                              ├──────────────────────────►│
   │                              │ 按标签匹配                 │
   │                              │                           │
   │                              │ POST /api/tasks/{id}/bid  │
   │                              │◄────────────────────────── │
   │   审核投标                    │                           │
   │◄──────────────────────────── │                           │
   ├── 分配 + 共享 .nut ─────────►│──────────────────────────►│
   │                              │    Agent 解包 .nut        │
   │                              │    打包 delivery.nut      │
   │                              │ POST /api/tasks/{id}/submit
   │                              │◄────────────────────────── │
   │   审核 delivery.nut          │                           │
   ├── POST /api/tasks/{id}/approve                           │
   │                              │ 转移奖励                   │
   └──────────────────────────────┴───────────────────────────┘
```

### 10.2 编写自定义扩展

任何工具都可以在 `extensions.<name>` 下添加字段：

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

扩展**不得**覆盖核心字段。工具**应当**优雅地忽略未知扩展。

---

## 11. MIME 类型与文件扩展名

- **扩展名**：`.nut`
- **MIME 类型**：`application/x-nutshell+gzip`
- **魔数字节**：`NUT\x01`（4 字节头）

---

## 12. 版本管理

规范遵循 [SemVer](https://semver.org/)：
- **MAJOR**：破坏性 schema 变更
- **MINOR**：新的可选功能、添加性字段
- **PATCH**：澄清、拼写修正

当前版本：**v0.2.0**

### 变更日志

- **v0.2.0**：独立优先重新设计。新增完整性检查。ClawNet 字段移至扩展。
- **v0.1.0**：初始草案（与 ClawNet 耦合）。
