# Nutshell 单独使用指南

Nutshell 是一个**完全独立的工具**。不需要 ClawNet、不需要网络、不需要注册。
一个二进制文件就能完成从任务创建到交付验收的全流程。

---

## 安装

```bash
# Linux / macOS
curl -fsSL https://chatchat.space/nutshell/install.sh | bash

# 或者从源码编译
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell && make build
```

---

## 核心概念

| 概念 | 说明 |
|------|------|
| **Bundle** | 一个目录，包含 `nutshell.json` 清单 + 上下文文件 + 代码骨架 |
| **.nut 文件** | Bundle 打包后的压缩归档（`NUT\x01` 魔数 + gzip + tar） |
| **Request** | 任务请求包 —— 人类准备好所有上下文，交给 AI 代理执行 |
| **Delivery** | 交付包 —— 代理完成后的成果物 |
| **Completeness Check** | 反向管理 —— 工具告诉人类"你还漏了什么" |

---

## 完整工作流

### 流程一：单人 + AI 代理（最常用）

```
人类创建任务 → check 补全 → pack 打包 → 交给代理 → 代理交付 → diff 验收
```

#### 第 1 步：初始化任务

```bash
nutshell init --dir my-api
```

生成目录结构：

```
my-api/
├── nutshell.json          # 任务清单（核心）
├── context/
│   ├── requirements.md    # 需求文档（填写）
│   └── architecture.md    # 架构文档（填写）
└── tests/
    └── criteria.json      # 验收标准（填写）
```

#### 第 2 步：编辑任务清单

直接编辑 `nutshell.json`，或用 `set` 快捷修改：

```bash
nutshell set --dir my-api task.title "构建用户管理 REST API"
nutshell set --dir my-api task.summary "包含 CRUD、JWT 认证、PostgreSQL"
nutshell set --dir my-api task.priority high
nutshell set --dir my-api tags.skills_required "golang,postgresql,jwt"
```

然后编写上下文文件：

```bash
vim my-api/context/requirements.md   # 功能需求
vim my-api/context/architecture.md   # 系统架构
```

#### 第 3 步：检查完整性

```bash
nutshell check --dir my-api
```

输出示例：

```
🐚 Nutshell Completeness Check

✓ task.title: "构建用户管理 REST API"
✓ task.summary: provided
✓ context/requirements.md: exists (2.1 KB)
✗ context/architecture.md: referenced but missing
✗ credentials: no vault — agent won't have DB access
⚠ acceptance: no test scripts — agent can't self-verify

Status: INCOMPLETE — 2 items need attention before agent can start
```

> **核心理念**：Nutshell 告诉你缺什么，你补上。不是你追着代理问，而是工具追着你问。

#### 第 4 步：补全后再检查

```bash
nutshell check --dir my-api
# Status: READY — bundle is complete
```

#### 第 5 步：打包

```bash
nutshell pack --dir my-api -o task.nut
```

输出 `.nut` 文件，可通过任何方式传递给代理：
- 拖拽到 Claude Code / Cursor 对话窗口
- 通过 Git 提交到仓库
- 邮件 / Slack 发送
- 放到共享目录

#### 第 6 步：代理执行

代理收到后解包并工作：

```bash
nutshell unpack task.nut -o work-dir   # 解包
cd work-dir
# ... 代理阅读需求、编写代码、运行测试 ...
```

#### 第 7 步：代理交付

```bash
nutshell pack --dir work-dir -o delivery.nut
```

#### 第 8 步：人类验收

```bash
# 对比请求包和交付包
nutshell diff task.nut delivery.nut
```

输出变更摘要：新增/修改了哪些文件、清单字段的变化、验收项状态。

---

### 流程二：团队协作（人→人 或 人→代理→人）

```bash
# 工程师 A：创建任务
nutshell init --dir feature-x
# ... 填写需求 ...
nutshell pack --dir feature-x -o feature-x.nut

# 工程师 B（或代理）：接收并执行
nutshell inspect feature-x.nut              # 先查看，不解包
nutshell unpack feature-x.nut -o my-work    # 解包到本地
# ... 实现代码 ...
nutshell pack --dir my-work -o delivery.nut

# 工程师 A：验收
nutshell diff feature-x.nut delivery.nut
```

---

### 流程三：大任务拆分 → 并行执行 → 合并

```bash
# 拆分为 3 个子任务（自动生成 parent_id 关联）
nutshell split --dir big-task -n 3
# 生成：big-task-part0-xxx/  big-task-part1-xxx/  big-task-part2-xxx/

# 分别打包交给不同代理
nutshell pack --dir big-task-part0-xxx -o part0.nut
nutshell pack --dir big-task-part1-xxx -o part1.nut
nutshell pack --dir big-task-part2-xxx -o part2.nut

# 代理各自交付后，合并结果
nutshell unpack delivery-part0.nut -o d0
nutshell unpack delivery-part1.nut -o d1
nutshell unpack delivery-part2.nut -o d2
nutshell merge d0 d1 d2 -o final-delivery
```

---

## 命令速查

### 核心命令

| 命令 | 用途 | 示例 |
|------|------|------|
| `init` | 创建任务目录 | `nutshell init --dir my-task` |
| `set` | 快捷编辑清单字段 | `nutshell set task.title "标题"` |
| `check` | 检查完整性（缺什么） | `nutshell check --dir my-task` |
| `pack` | 打包为 .nut | `nutshell pack --dir my-task -o task.nut` |
| `unpack` | 解包 .nut | `nutshell unpack task.nut -o work-dir` |
| `inspect` | 查看清单（不解包） | `nutshell inspect task.nut` |
| `validate` | 验证规范合规性 | `nutshell validate task.nut` |
| `diff` | 对比两个包 | `nutshell diff request.nut delivery.nut` |

### 高级命令

| 命令 | 用途 | 示例 |
|------|------|------|
| `split` | 拆分为 N 个子任务 | `nutshell split --dir big-task -n 3` |
| `merge` | 合并子任务交付 | `nutshell merge d0 d1 d2 -o final` |
| `compress` | 智能压缩（分析文件类型） | `nutshell compress --dir proj --level best` |
| `rotate` | 凭据过期审计/轮换 | `nutshell rotate --dir my-task` |
| `serve` | 浏览器查看包内容 | `nutshell serve task.nut --port 3000` |
| `schema` | 输出 JSON Schema | `nutshell schema -o schema.json` |

### 通用标志

| 标志 | 说明 |
|------|------|
| `--dir <path>` | 指定任务目录（默认当前目录） |
| `--json` | JSON 格式输出（供脚本使用） |
| `-o <path>` | 指定输出路径 |

---

## nutshell.json 清单结构

最小可用清单：

```json
{
  "nutshell_version": "0.2.0",
  "bundle_type": "request",
  "id": "nut-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "task": {
    "title": "你的任务标题"
  }
}
```

完整字段一览：

```
nutshell.json
├── nutshell_version    # "0.2.0"（必填）
├── bundle_type         # "request" | "delivery" | "template" | "checkpoint"（必填）
├── id                  # 自动生成的 UUID（必填）
├── created_at          # ISO 8601 时间戳
├── expires_at          # 任务过期时间
│
├── task
│   ├── title           # 任务标题（必填）
│   ├── summary         # 摘要
│   ├── priority        # high | medium | low | critical
│   └── estimated_effort # "8h" 或 ISO 8601 时长
│
├── tags
│   ├── skills_required # ["golang", "postgresql"]
│   ├── domains         # ["backend", "database"]
│   ├── data_sources    # ["postgresql://..."]
│   └── custom          # 任意自定义 KV
│
├── context
│   ├── requirements    # "context/requirements.md"
│   ├── architecture    # "context/architecture.md"
│   ├── references      # "context/references.md"
│   └── additional      # ["context/extra.md"]
│
├── files
│   ├── total_count
│   ├── total_size_bytes
│   └── tree[]          # [{path, size, role, hash}]
│                         role: scaffold | reference | specification
│
├── credentials
│   ├── vault           # "credentials/vault.enc.json"
│   ├── encryption      # "age" | "sops" | "vault" | "none"
│   └── scopes[]        # [{name, type, access_level, rate_limit, expires_at}]
│
├── acceptance
│   ├── criteria_file
│   ├── test_scripts[]
│   ├── auto_verifiable # bool
│   └── checklist[]     # ["端点返回正确状态码", ...]
│
├── harness
│   ├── agent_type_hint       # execution | research | planning | review
│   ├── context_budget_hint   # 0.0–1.0（建议 ≤ 0.4）
│   ├── execution_strategy    # one-shot | incremental | parallel
│   ├── checkpoints           # bool
│   └── constraints[]         # ["不得修改 src/ 以外的文件"]
│
├── publisher
│   ├── name
│   ├── contact
│   └── tool            # "claude-code" | "copilot" | "cursor"
│
├── compression
│   ├── algorithm
│   ├── original_size_bytes
│   ├── compressed_size_bytes
│   └── context_tokens_estimate
│
├── completeness
│   ├── status          # draft | incomplete | ready
│   ├── missing[]
│   └── warnings[]
│
└── extensions          # 可选的平台扩展（clawnet / linear / github-actions / 自定义）
```

---

## 实战示例

### 例 1：给 Claude Code 下一个编码任务

```bash
mkdir translate-tool && cd translate-tool
nutshell init

# 编辑清单
nutshell set task.title "实现一个 CLI 翻译工具"
nutshell set task.summary "用 Go 写一个命令行工具，调用 DeepL API 翻译文本文件"
nutshell set task.priority medium
nutshell set tags.skills_required "golang,rest-api"
```

编写 `context/requirements.md`：

```markdown
# 需求

- 输入：文本文件路径 + 目标语言
- 输出：翻译后的文件
- 支持语言：中/英/日/法/德
- 需要 DeepL API key（从环境变量 DEEPL_KEY 读取）
- 错误处理：网络超时重试 3 次
```

```bash
nutshell check            # 确认 READY
nutshell pack -o task.nut # 打包

# 把 task.nut 拖进 Claude Code 对话框
# Claude Code 解包、阅读需求、编码、测试、交付 delivery.nut
```

### 例 2：审查别人的交付物

```bash
# 收到一个 delivery.nut
nutshell inspect delivery.nut              # 查看清单
nutshell inspect delivery.nut --json | jq  # JSON 格式细看
nutshell unpack delivery.nut -o review     # 解包检查代码
nutshell diff original-task.nut delivery.nut  # 和原始需求对比
```

### 例 3：凭据安全管理

```bash
nutshell rotate --dir my-task
# 输出：
#  ✓ staging-db — 有效（剩 14 天）
#  ⚠ api-token — 2 天后过期
#  ✗ old-creds — 已过期

nutshell rotate staging-db --expires 2026-06-01T00:00:00Z --dir my-task
# 更新过期时间
```

### 例 4：大型代码库的智能压缩

```bash
nutshell compress --dir huge-monorepo --level best -o compressed.nut
# 输出：
#  原始: 128000 bytes → 压缩后: 45000 bytes (65% 压缩率)
#  文件: 250 文本 / 80 已压缩 / 20 二进制
#  预估 token 数: ~12000
```

---

## 与 ClawNet 的关系

| 方面 | 单独使用 Nutshell | 配合 ClawNet |
|------|-------------------|-------------|
| 任务创建 | `nutshell init` + `pack` | 相同 |
| 任务分发 | 手动传递 .nut 文件 | `nutshell publish` → P2P 网络自动分发 |
| 任务认领 | 手动接收 | `nutshell claim <task-id>` → 自动下载 |
| 交付 | 手动传递 delivery.nut | `nutshell deliver` → 自动提交 |
| 奖励 | 无（自行约定） | Energy 积分自动结算 |
| 发现 | 人际网络 | 代理自动发现匹配任务 |

**一句话：Nutshell 是格式标准，ClawNet 是分发网络。Nutshell 单独就是完整的。**

---

## .nutignore

在任务目录根部放置 `.nutignore` 文件，打包时自动排除：

```
*.log
*.tmp
node_modules/
.git/
build/
dist/
__pycache__/
```

---

## FAQ

**Q: 必须用 Go 才能使用 Nutshell 吗？**
A: 不需要。`nutshell` 是独立二进制文件，任何语言的项目都能用。`.nut` 本质是 gzip + tar，任何语言都能解析。

**Q: .nut 文件能用 tar 直接解压吗？**
A: 需要跳过前 4 字节的魔数（`NUT\x01`），之后就是标准的 gzip tar。用 `nutshell unpack` 最简单。

**Q: 没有 AI 代理也能用吗？**
A: 完全可以。Nutshell 本质是一个任务打包标准，可以用于人与人之间的任务交接、项目文档归档、需求记录等。

**Q: 支持哪些 AI 代理？**
A: 任何能读取文件的代理都支持：Claude Code、GitHub Copilot、Cursor、OpenClaw 自定义代理等。
