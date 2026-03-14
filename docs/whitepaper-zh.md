# Nutshell：面向 AI Agent 的任务打包标准

**上下文架构与 Harness 工程白皮书**

*版本 0.2.0 — 2026 年 3 月*

---

## 摘要

随着 AI Agent 成为软件工程工作流的核心，一个基础设施层面的缺口逐渐暴露：目前不存在一种标准方式来打包 Agent 执行任务所需的上下文。需求散落在 Slack 消息中，凭证以不安全的方式共享，架构知识留存在工程师的脑中，每次新的 Agent 会话都从零开始。

**Nutshell** 通过开放标准和 CLI 工具链来解决这一问题，将任务上下文打包成自包含的 `.nut` 包。单个包承载一切 — 需求、源文件、凭证、验收标准和执行约束 — 让任何 Agent 都能立即开始工作，无需反复询问 20 个澄清问题。

Nutshell 基于 **Harness 工程** 理念：一个新兴的学科，专注于构建连接、保护和编排 AI Agent 的基础设施层 — 而非替代 Agent 本身的工作。

---

## 1. 问题：上下文脆弱性

现代 AI 编码 Agent（Claude Code、Cursor、Copilot、Aider）能力出众。但它们始终在同一件事上失败：获取完整上下文。

### 散落问题

| 上下文组件 | 当前存储位置 |
|---|---|
| 任务需求 | Slack 消息、Notion 文档、邮件线程 |
| 架构决策 | 工程师记忆、零散的 README |
| 数据库模式 | Migration 文件、Wiki 页面 |
| API 凭证 | 环境变量、共享的 `.env` 文件 |
| 验收标准 | Jira 工单、口头约定 |
| 执行约束 | 团队隐性知识 |

每次 Agent 会话都以同样的仪式开始：人类重新解释上下文，Agent 提出澄清问题，15 条消息之后，实际工作才开始。当会话重置时，上下文也随之重置。

### 代价

- **时间浪费**：30–60% 的 Agent 交互用于上下文传递，而非执行。
- **上下文漂移**：信息在反复转述中变异。Agent 的理解悄然偏离现实。
- **安全暴露**：凭证以明文在聊天中共享。无审计轨迹，无过期机制。
- **不可复现**：Agent 失败后，无法用完全相同的上下文重放任务。

---

## 2. Harness 工程：理论基础

### 什么是 Harness？

> *"在每个工程学科中，harness 都是同样的东西：连接、保护和编排组件的层 — 它本身不做具体工作。"*

Agent 核心循环看似简单：

```
while (模型返回工具调用):
    执行工具 → 捕获结果 → 追加到上下文 → 再次调用模型
```

每个生产级 Agent — Claude Code、Cursor、Codex — 都实现了这个循环。差异化不在于循环本身，而在于其周围的一切：上下文如何结构化、凭证如何注入、约束如何执行、结果如何验证。

这些周围的基础设施就是 **harness**。

### 大模型 vs 大 Harness 之争

AI 行业正在辩论价值归属于模型还是 harness：

**大模型派**："秘密武器全在模型里。Harness 应该是最薄的包装层。" — Boris Cherny (Anthropic)，描述 Claude Code 架构。

**大 Harness 派**："从 AI 获取价值的最大障碍是你自己对模型进行上下文和工作流工程化的能力。" — Jerry Liu (LlamaIndex)。

研究表明两者都重要。Scale AI 的 SWE-Atlas 发现 harness 选择会产生可测量但适度的性能差异。然而，结构化的上下文 — 在正确时间以正确格式提供正确信息 — 始终能提升每个模型。

### Nutshell 的立场：结构化上下文是最大杠杆

Nutshell 不替代模型或编排循环。它标准化进入上下文窗口的**内容**。这是最高杠杆的干预，因为：

1. **模型无关**：好的上下文同等提升 GPT、Claude、Gemini 和开源模型。
2. **Harness 无关**：任何编排器都能消费 `.nut` 包。无厂商锁定。
3. **可累积**：上下文一旦打包，就能跨会话、跨 Agent、跨团队持久保存。
4. **可审计**：Request 和 delivery 包形成完整的请求-交付记录。

---

## 3. Nutshell 标准

### 3.1 包格式

`.nut` 文件是以 `NUT\x01` 魔数字节为前缀的 gzip 压缩 tar 归档。MIME 类型为 `application/x-nutshell+gzip`。

```
task.nut (二进制)
├── NUT\x01          ← 4 字节魔数头
└── gzip(tar(
    ├── nutshell.json    ← 清单（最先加载）
    ├── context/         ← 需求、架构、参考文档
    ├── files/           ← 源代码、数据、资产
    ├── credentials/     ← 加密凭证库
    ├── tests/           ← 验收测试脚本
    └── delivery/        ← 交付产物
))
```

仅 `nutshell.json` 为必需。其他目录均可选 — 按任务需要添加。

### 3.2 清单 Schema

清单（`nutshell.json`）是入口点。Agent 首先读取它来理解需要做什么。

**必需字段：**
- `nutshell_version` — 规范版本（`"0.2.0"`）
- `bundle_type` — `request` | `delivery` | `template` | `checkpoint` | `partial`
- `id` — 唯一标识符（`nut-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`）
- `task.title` — 需要做什么

**上下文字段：**
- `task.summary` — 详细描述
- `task.priority` — `critical` | `high` | `medium` | `low`
- `task.estimated_effort` — 人类可读的工作量估计（`"2h"`、`"1d"`）
- `context.requirements` — 需求文档路径
- `context.architecture` — 架构文档路径
- `tags.skills_required` — Agent 需要的技能（`["golang", "postgresql"]`）
- `tags.domains` — 领域分类（`["backend", "auth"]`）

**安全字段：**
- `credentials.vault` — 加密库文件路径
- `credentials.encryption` — `age` | `sops` | `vault` | `none`
- `credentials.scopes[]` — 命名、类型化、范围受限的凭证，带过期时间

**执行控制：**
- `harness.agent_type_hint` — `execution` | `research` | `review` | `creative`
- `harness.context_budget_hint` — 目标上下文窗口填充比（0.0–1.0）
- `harness.execution_strategy` — `incremental` | `batch` | `streaming`
- `harness.constraints` — 机器可读的约束条件
- `acceptance.checklist` — "完成"的标准
- `acceptance.test_scripts` — 自动化验证脚本

**完整性：**
- `files.tree[].hash` — 每文件 SHA-256 哈希（`sha256:<hex>`）
- `completeness.status` — `draft` | `incomplete` | `ready`

**扩展性：**
- `extensions` — 任意平台特定元数据（ClawNet、Linear、GitHub Actions）

### 3.3 包类型

| 类型 | 方向 | 用途 |
|---|---|---|
| `request` | 人类 → Agent | "这是我需要你做的事" |
| `delivery` | Agent → 人类 | "这是我做了什么，以及怎么做的" |
| `template` | 可复用 | 通用模式的空白任务结构 |
| `checkpoint` | Agent → 自身 | 长时间任务的中间状态 |
| `partial` | Agent → Agent | 多 Agent 拆分中的子任务结果 |

`request → delivery` 对形成完整的审计轨迹。`parent_id` 字段将 delivery 包链接回其发起的 request。

---

## 4. 反向管理

### 反转

传统工作流：Agent 向人类索要缺失的上下文。人类零散地提供。上下文在每次交换中退化。

**Nutshell 工作流**：包在任何 Agent 介入**之前**就告诉人类缺少什么。

```
$ nutshell check

  🐚 Nutshell 完整性检查

  ✓ task.title: "构建用户管理 REST API"
  ✓ task.summary: 已提供
  ✓ context/requirements.md: 存在 (2.1 KB)
  ✗ context/architecture.md: 已引用但缺失
  ✗ credentials: 无凭证库 — Agent 将无法访问数据库
  ⚠ acceptance: 无测试脚本 — Agent 无法自我验证

  状态: 不完整 — 2 项需要处理
```

这就是**反向管理**：工具来管理人类，确保 Agent 从一开始就获得完整上下文。`check` 命令是 Nutshell 中最强大的功能 — 它将完整性的责任从 Agent 转移到基础设施。

### 检查项

| 检查项 | 通过条件 |
|---|---|
| 任务标题 | 非空 |
| 任务摘要 | 非空 |
| 上下文文件 | 所有引用的路径在磁盘上存在 |
| 凭证 | 配置了 vault，scope 有过期时间 |
| 验收标准 | 定义了 checklist 或测试脚本 |
| Harness 约束 | 至少设置一个约束 |
| 技能/领域标签 | 至少有一个标签 |
| 上下文预算 | ≤ 0.5（过高时发出警告） |

---

## 5. 凭证安全架构

Nutshell 将凭证视为一等基础设施，而非事后补丁。

### 原则

| 原则 | 实现方式 |
|---|---|
| **范围受限** | 每个凭证缩小到特定的表、端点、操作 |
| **时间限定** | 每个凭证都有 `expires_at` |
| **加密存储** | 默认：[age 加密](https://age-encryption.org/)。也支持 SOPS、HashiCorp Vault |
| **速率限制** | 每凭证速率限制 |
| **可审计** | Delivery 包记录使用了哪些凭证 |
| **可轮换** | `nutshell rotate` 审计过期状态并更新过期日期 |

### 凭证生命周期

```
1. 作者在 nutshell.json 中定义 scopes
2. `nutshell check` 在缺少过期时间时发出警告
3. `nutshell rotate` 审计过期状态
4. `nutshell pack` 在包中包含加密的 vault
5. Agent 在运行时使用提供的密钥解密
6. Delivery 包记录凭证使用情况
```

---

## 6. 上下文感知压缩

不同文件的压缩效果不同。Nutshell 的压缩引擎按类型分类文件并应用最优策略：

| 分类 | 示例 | 策略 |
|---|---|---|
| 文本 | `.go`、`.py`、`.md`、`.json` | 积极压缩（高 gzip 级别） |
| 已压缩 | `.jpg`、`.png`、`.zip`、`.gz` | 存储而不重新压缩 |
| 二进制 | 其他二进制文件 | 默认级别压缩 |

`nutshell compress` 命令还会估算文本文件的 **token 数量**（约 0.25 token/字节），帮助工程师了解包将消耗多少上下文窗口。

---

## 7. 多 Agent 拆分

复杂任务可以通过 `nutshell split` 分解为并行子任务：

```
$ nutshell split --dir my-task -n 3

  创建了 3 个子任务：
    part-0/  →  "Backend: src/"
    part-1/  →  "Tests: tests/"
    part-2/  →  "Docs: docs/"
```

每个子任务：
- 拥有独立的 `nutshell.json`，`bundle_type: "partial"`
- 在 `parent_id` 中存储原始任务的 ID
- 接收共享上下文（`context/` 目录复制到所有子任务）
- 在 `extensions.split` 中记录拆分元数据

Agent 完成子任务后，`nutshell merge` 将 delivery 包合并为统一结果。

---

## 8. 工具链

### CLI 命令

| 命令 | 用途 |
|---|---|
| `init` | 创建新的任务目录脚手架 |
| `check` | 完整性验证（反向管理） |
| `pack` | 创建带 SHA-256 完整性的 `.nut` 包 |
| `unpack` | 提取包（防止路径穿越攻击） |
| `inspect` | 查看清单而不解压 |
| `validate` | 根据规范检查 |
| `set` | 通过点路径快速编辑清单字段 |
| `diff` | 比较 request 和 delivery 包 |
| `schema` | 输出 JSON Schema 用于 IDE 集成 |
| `compress` | 上下文感知压缩与 token 估算 |
| `split` | 将任务分解为并行子任务 |
| `merge` | 合并 delivery 子包 |
| `rotate` | 审计和更新凭证过期时间 |
| `serve` | 本地 Web 查看器用于包检查 |

### Agent SDK (Go)

```go
import "github.com/ChatChatTech/nutshell/pkg/nutshell"

b, _ := nutshell.Open("task.nut")
m := b.Manifest()
content, _ := b.ReadFile("context/requirements.md")
files := b.FilesByPrefix("context/")
```

### VS Code 扩展

- `nutshell.json` 的 JSON Schema 验证和自动补全
- 清单各部分的代码片段
- 通过命令面板执行 pack/unpack/inspect/validate
- 保存时自动验证并显示诊断信息
- `.nut` 文件的右键菜单集成

### Web 查看器

`nutshell serve` 启动本地 Web 查看器，提供：
- 结构化字段的清单展示
- 文件浏览器与内容预览
- 完整性状态仪表板
- 用于编程访问的 REST API

---

## 9. 平台扩展

Nutshell 以独立为先但可扩展。`extensions` 字段支持任意平台集成，不破坏核心格式。

### ClawNet 集成

[ClawNet](https://github.com/ChatChatTech/ClawNet) 是去中心化 AI Agent 通信网络。Nutshell 原生集成：

```bash
nutshell publish --dir my-task    # 打包并发布到 P2P 网络
nutshell claim <task-id>          # 认领并解包远程任务
nutshell deliver --dir workspace  # 打包并提交交付物
```

### 其他扩展

```json
{
  "extensions": {
    "clawnet": {"peer_id": "12D3KooW...", "reward": 50.0},
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

扩展永不破坏核心格式。工具会忽略它们不理解的内容。

---

## 10. 技术细节

| 属性 | 值 |
|---|---|
| 语言 | Go 1.26 |
| 依赖 | 零（仅标准库） |
| 包格式 | `NUT\x01` + gzip + tar |
| 哈希算法 | SHA-256（文件级 + 包级） |
| 凭证加密 | age、SOPS、HashiCorp Vault |
| 规范版本 | 0.2.0 |
| MIME 类型 | `application/x-nutshell+gzip` |
| 测试覆盖 | 44 个单元测试，70%+ 覆盖率 |
| 许可证 | MIT |

---

## 11. 结论

AI Agent 领域正在快速演变。模型每季度在进步，Harness 每周在变化。但对结构化、安全、可移植上下文的需求是永恒的。

Nutshell 押注于一个简单理念：如果你给 Agent 更好的上下文，它们就会产生更好的结果 — 无论你使用哪个模型或 harness。通过标准化上下文的打包方式，我们创建的基础设施能跨会话、跨 Agent、跨团队和跨平台累积价值。

**打包它。破壳它。交付它。** 🐚

---

*Nutshell 是 [ChatChatTech](https://github.com/ChatChatTech) 的开放标准。欢迎贡献。*
