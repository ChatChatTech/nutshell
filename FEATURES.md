# Nutshell 功能清单

> **Nutshell** — AI Agent 任务打包标准 (v0.2.0)
>
> 将 AI agent 执行任务所需的一切上下文打包进一个 `.nut` 文件。

---

## CLI 命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `nutshell init` | 初始化任务目录，生成 `nutshell.json` 模板和 `context/` 目录 | `nutshell init --dir my-task` |
| `nutshell pack` | 将目录打包为 `.nut` bundle 文件（NUT magic bytes + gzip + tar） | `nutshell pack -o task.nut` |
| `nutshell unpack` | 解包 `.nut` 文件到目录 | `nutshell unpack task.nut -o output/` |
| `nutshell inspect` | 查看 bundle 内容摘要，不解包。支持 stdin（`-`）和 `--json` | `nutshell inspect task.nut --json` |
| `nutshell validate` | 校验 bundle/目录/JSON 是否符合 spec | `nutshell validate my-task.nut` |
| `nutshell check` | 完整性检查——告诉人类还缺什么（反向管理） | `nutshell check --dir my-task --json` |
| `nutshell set` | 用 dot-path 快速编辑 manifest 字段 | `nutshell set task.title "Build API"` |
| `nutshell diff` | 对比两个 bundle 的差异（request vs delivery 等） | `nutshell diff a.nut b.nut --json` |
| `nutshell schema` | 输出 JSON Schema，用于 IDE 自动补全 | `nutshell schema -o nutshell.schema.json` |

---

## 核心功能

### 1. 任务清单（Manifest）

`nutshell.json` 是 bundle 的核心，包含：

- **task**: 任务标题、摘要、优先级、预估工时
- **tags**: 技能要求、领域分类、数据源、自定义标签
- **publisher**: 发布者信息和工具来源
- **context**: 需求文档、架构文档、参考资料路径
- **credentials**: 加密凭证库，支持 age/sops/vault 加密
- **apis**: 可调用 API 配置（OpenAPI spec + base URLs + auth）
- **acceptance**: 验收标准、测试脚本、checklist
- **harness**: agent 编排提示（agent 类型、上下文预算、执行策略、约束）
- **resources**: 外部仓库、文档、图片、链接引用
- **completeness**: 机器可读的完整性状态（draft/incomplete/ready）
- **extensions**: 可扩展字段（如 ClawNet、Linear、GitHub Actions 集成）

### 2. Bundle 类型

| 类型 | 用途 |
|------|------|
| `request` | 任务请求——人类给 agent 的任务包 |
| `delivery` | 交付包——agent 完成的工作成果 |
| `template` | 模板——可复用的任务结构 |
| `checkpoint` | 检查点——agent 的中间状态快照 |
| `partial` | 部分交付——尚未完成的阶段性成果 |

### 3. 完整性检查（Reverse Management）

`nutshell check` 会检查：
- 任务标题和摘要是否填写
- 引用的文件（requirements.md、architecture.md 等）是否存在
- 凭证库是否配置
- 验收标准是否定义
- harness 约束是否设置
- 技能标签是否填写

输出 draft / incomplete / ready 三种状态。**让 bundle 反过来管理人类**——告诉你还缺什么，而不是让 agent 一遍遍问。

### 4. 文件级校验和

打包时自动为每个文件计算 **SHA-256** 哈希，写入 manifest 的 `files.tree[].hash`。确保内容完整性，支持内容寻址。

### 5. Bundle 级哈希

`nutshell pack` 完成后输出整个 `.nut` 文件的 SHA-256 哈希（`sha256:<hex>`），可用于内容寻址/去重。

### 6. JSON Schema / IDE 自动补全

内置完整的 JSON Schema（`nutshell schema` 命令），覆盖 manifest 的所有字段。在 VS Code 等编辑器中实现自动补全和实时校验。

### 7. 快速编辑（Set）

`nutshell set` 支持 dot-path 语法，无需手动编辑 JSON：

```
nutshell set task.title "Build REST API"
nutshell set task.priority high
nutshell set harness.context_budget_hint 0.25
nutshell set tags.skills_required "go,rest,api"
```

支持的路径：task.*, publisher.*, context.*, harness.*, tags.*, bundle_type, parent_id, expires_at

### 8. Bundle 对比（Diff）

`nutshell diff` 对比两个 bundle（.nut 文件、目录、JSON 文件）的差异：
- 字段级对比（标题、类型、标签、约束等）
- 文件级对比（新增/删除的文件）
- 支持 `--json` 输出，方便自动化

典型场景：对比 request bundle 和 delivery bundle，看 agent 做了什么。

### 9. .nutignore

支持 `.nutignore` 文件，语法类似 `.gitignore`：
- 通配符（`*.log`）
- 目录前缀（`delivery/`）
- 精确匹配（`secret.json`）

打包时自动排除匹配的文件。

### 10. 安全特性

- **凭证加密**: 支持 age、SOPS、HashiCorp Vault 加密方式
- **作用域凭证**: 每个凭证限定表、端点、操作
- **时间限制**: 凭证必须有过期时间（缺失会告警）
- **速率限制**: 支持 per-credential 速率限制
- **路径遍历防护**: unpack 拒绝绝对路径和 `..` 路径
- **上下文预算**: `context_budget_hint` 防止上下文窗口过载（超过 0.5 告警）

---

## Agent SDK（Go API）

`nutshell.Open()` 提供程序化 bundle 访问，让 agent 代码直接消费 `.nut` 文件：

```go
import "github.com/ChatChatTech/nutshell/pkg/nutshell"

// 打开 bundle
b, err := nutshell.Open("task.nut")

// 读取 manifest
m := b.Manifest()
fmt.Println(m.Task.Title)

// 读取文件
content, _ := b.ReadFile("context/requirements.md")

// 读取上下文（Requirements）
ctx, _ := b.ReadContext()

// 列出所有文件
files := b.ListFiles()

// 按前缀过滤
contextFiles := b.FilesByPrefix("context/")

// 检查文件是否存在
if b.HasFile("tests/acceptance.sh") { ... }
```

其他 SDK 方法：
- `ReadFileString()` — 返回 string 而非 []byte
- `ManifestJSON()` — 原始 manifest JSON
- `Repack(w)` — 重新打包写入 writer

---

## 其他能力

| 功能 | 说明 |
|------|------|
| **stdin 支持** | `cat task.nut \| nutshell inspect -` 支持管道输入 |
| **--json 输出** | inspect / validate / check / diff 都支持 JSON 输出 |
| **parent_id 链** | delivery bundle 可通过 `parent_id` 链接到原始 request |
| **Extensions** | 通过 `extensions` 字段集成任何平台（ClawNet、Linear、GitHub Actions） |
| **最小化 init** | `nutshell init` 只创建必要的 `context/` 目录和模板 |

---

## 技术栈

- **语言**: Go 1.26
- **Bundle 格式**: `NUT\x01` magic bytes + gzip + tar
- **哈希**: SHA-256（文件级 + bundle 级）
- **Spec**: nutshell-spec-v0.2.0
- **测试**: 28 个单元测试，65%+ 覆盖率
- **零依赖**: 仅使用 Go 标准库
