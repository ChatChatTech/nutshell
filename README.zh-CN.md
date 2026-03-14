<div align="center">

<img src="nutshell-icon.svg" width="80" height="80" alt="nutshell 图标" />

# nutshell

**一个开放标准，用于打包 AI 代理可理解的任务上下文。**

兼容任意代理：Claude Code · Copilot · Cursor · OpenClaw · 自定义代理

[规范](spec/nutshell-spec-v0.2.0.md) · [示例](examples/) · [研究](docs/harness-engineering-research.md) · [网站](https://chatchat.space/nutshell/)

[English](README.md) | **[简体中文](README.zh-CN.md)** | [繁體中文](README.zh-HANT.md) | [Español](README.es-ES.md) | [Français](README.fr-FR.md)

</div>

---

## 问题

AI 编程代理很强大，但它们总是反复询问相同的问题：

```
代理: "用什么框架？什么数据库？Schema 在哪？
       怎么认证？验收标准是什么？
       能访问预发布环境吗？"
人类: *三天内发了 47 条消息，每次都丢失上下文*
```

每次启动新会话，你都要重新解释同样的上下文。凭据通过 Slack 传递。需求只存在你脑子里。没有任何记录说明做了什么或为什么做。

## 解决方案

**Nutshell** 把 AI 代理所需的一切打包到一个包中：

```
$ nutshell init
$ nutshell check

  🐚 Nutshell 完整性检查

  ✓ task.title: "构建用户管理 REST API"
  ✓ task.summary: 已提供
  ✓ context/requirements.md: 存在 (2.1 KB)
  ✗ context/architecture.md: 已引用但缺失
  ✗ credentials: 无密钥库 — 代理无法访问数据库
  ⚠ acceptance: 无测试脚本 — 代理无法自我验证

  状态: 不完整 — 2 个项目需要处理后代理才能开始
```

Nutshell 告诉**你**哪些还缺。填补空白，打包，交给任何代理：

```
$ nutshell pack -o task.nut       # 人类打包任务
$ nutshell inspect task.nut       # 代理看到所有需要的内容
# ... 代理执行 ...
$ nutshell pack -o delivery.nut   # 代理交付结果
```

---

## 为什么选择 Nutshell？

| 没有 Nutshell | 有 Nutshell |
|-------------|-----------|
| 上下文散落在 Slack、文档、邮件中 | 一个 `.nut` 包包含一切 |
| 代理在开始前要问 20 个问题 | 代理读取清单，立即开始 |
| 凭据以不安全的方式共享 | 加密密钥库，带作用域和时间限制的令牌 |
| 没有请求或交付的记录 | 请求 + 交付包形成完整的审计追踪 |
| 新会话 = 重新解释一切 | 包跨会话持久化 |
| 无法验证完成情况 | 机器可读的验收标准 |

### 独立设计

Nutshell **无需任何外部平台**即可工作。单个开发者使用 Claude Code 就能立即受益：

1. **定义** — `nutshell init` 创建结构化的任务目录
2. **检查** — `nutshell check` 告诉你缺少什么（凭据？架构文档？验收标准？）
3. **打包** — `nutshell pack` 压缩成 `.nut` 包
4. **执行** — 将包交给任何 AI 代理
5. **归档** — 交付包记录构建了什么以及为什么

### 平台扩展（可选）

想把任务发布到市场？Nutshell 支持可选扩展：

```jsonc
{
  "extensions": {
    "clawnet": {                    // P2P 代理网络
      "peer_id": "12D3KooW...",
      "reward": {"amount": 50, "currency": "energy"}
    },
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

扩展不会破坏核心格式。工具会忽略它们不理解的内容。

---

## 🐚 名字的由来

> **龍蝦吃貝殼** — *龙虾吃贝壳。*

[ClawNet](https://github.com/ChatChatTech/ClawNet)（🦞）是一个去中心化的 AI 代理网络。代理就是龙虾，它们需要食物 — 食物装在壳里。**Nutshell**（🐚）就是那个壳 — 紧凑、营养丰富、随时可以打开。

但你不需要是一只龙虾。任何代理都能吃 nutshell。

---

## 快速开始

### 安装

```bash
# 一键安装（自动检测操作系统和架构）
curl -fsSL https://chatchat.space/nutshell/install.sh | sh

# 或通过 Go 安装
go install github.com/ChatChatTech/nutshell/cmd/nutshell@latest

# 或从源码构建
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell && make build
```

### 创建任务

```bash
# 初始化
nutshell init --dir my-task
cd my-task

# 编辑清单
vim nutshell.json

# 检查缺少什么
nutshell check

# 准备好后打包
nutshell pack -o my-task.nut
```

### 查看包内容

```
$ nutshell inspect my-task.nut

    🐚  n u t s h e l l  🦞
    AI 代理任务打包

  Bundle: my-task.nut
  Version: 0.2.0
  Type: request
  ID: nut-7f3a1b2c-...

  📋 任务: 构建用户管理 REST API
  优先级: high | 工作量: 8h

  🏷️  标签: golang, postgresql, jwt, rest-api
  领域: backend, authentication

  👤 发布者: Alice Chen (via claude-code)

  🔑 凭据: 2 个（带作用域）
    • staging-db (postgresql) — read-write
    • api-token (bearer_token) — invoke

  📦 文件: 5 个文件, 8,200 字节

  ⚙️  Harness 提示:
    代理类型: execution
    策略: incremental
    上下文预算: 0.35
```

### 验证

```bash
nutshell validate my-task.nut      # 检查打包后的包
nutshell validate ./my-task        # 检查目录
```

### 快速编辑

```bash
nutshell set task.title "Build REST API"
nutshell set task.priority high
nutshell set tags.skills_required "go,rest,api"
```

### 比较包

```bash
nutshell diff request.nut delivery.nut          # 人类可读的差异
nutshell diff request.nut delivery.nut --json   # 机器可读的差异
```

### JSON Schema

```bash
nutshell schema                            # 输出到标准输出
nutshell schema -o nutshell.schema.json    # 写入文件
```

添加到 `nutshell.json` 以启用 IDE 自动补全：
```jsonc
{
  "$schema": "./schema/nutshell.schema.json",
  ...
}
```

### 高级命令

```bash
# 上下文感知压缩 — 分析文件类型并应用最优压缩
nutshell compress --dir ./my-task -o task.nut --level best

# 多代理包拆分 — 将任务拆分为并行子任务
nutshell split --dir ./my-task -n 3
nutshell merge part-0/ part-1/ part-2/ -o merged/

# 凭据轮换 — 审计并更新凭据过期时间
nutshell rotate --dir ./my-task                              # 审计全部
nutshell rotate staging-db --expires 2026-01-01T00:00:00Z    # 轮换单个

# Web 查看器 — 本地 HTTP 查看器用于 .nut 检查
nutshell serve ./my-task --port 8080
nutshell serve task.nut
```

---

## 包结构

```
task.nut                        🐚 壳
├── nutshell.json               📋 清单（始终最先加载）
├── context/                    📖 需求、架构、参考资料
├── files/                      📦 源文件和资源
├── apis/                       🔌 可调用的 API 规范
├── credentials/                🔑 加密凭据库
├── tests/                      ✅ 验收标准和测试脚本
└── delivery/                   🦪 完成产物（交付包）
```

只有 `nutshell.json` 是必需的。根据需要添加目录。

## 清单（`nutshell.json`）

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "request",
  "id": "nut-a1b2c3d4-...",
  "task": {
    "title": "构建用户管理 REST API",
    "summary": "带 JWT 认证和 PostgreSQL 的 CRUD 端点。",
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
      "所有 CRUD 端点返回正确的状态码",
      "JWT 认证在受保护路由上正常工作"
    ],
    "auto_verifiable": true
  },
  "harness": {
    "agent_type_hint": "execution",
    "context_budget_hint": 0.35,
    "execution_strategy": "incremental",
    "constraints": ["不要修改 files/src/ 之外的文件"]
  },
  "completeness": {
    "status": "ready"
  }
}
```

只有 `nutshell_version`、`bundle_type`、`id` 和 `task.title` 是必需的。其他字段都能提升代理的效率。

---

## Check 命令（反向管理）

最强大的功能：**Nutshell 管理人类**。

```bash
$ nutshell check

  🐚 Nutshell 完整性检查

  ✓ task.title: "Build REST API"
  ✓ context/requirements.md: 存在 (2.1 KB)
  ✗ context/architecture.md: 已引用但缺失
  ✗ credentials: 无密钥库 — 代理无法访问数据库
  ⚠ acceptance: 无标准 — 代理无法自我验证
  ⚠ harness: 无约束

  状态: 不完整 — 填补 2 个项目后代理才能开始
```

不是代理问"我还需要什么？"，而是**包告诉人类**该提供什么。这颠覆了传统模式，确保代理从一开始就获得完整的上下文。

---

## Harness Engineering 对齐

Nutshell 基于 [Harness Engineering](docs/harness-engineering-research.md) — 围绕 AI 代理构建基础设施的新兴学科：

| 原则 | Nutshell 实现 |
|------|-------------|
| **上下文架构** | 分层加载 — 先加载清单，按需加载细节 |
| **代理专业化** | `harness.agent_type_hint` 指导适合哪种代理角色 |
| **持久记忆** | 交付包保留执行日志、决策、检查点 |
| **结构化执行** | 请求/交付分离，带机器可读的验收标准 |
| **40% 规则** | `context_budget_hint` 防止上下文窗口溢出 |
| **约束机械化** | Harness 约束是机器可读且可执行的 |

---

## 凭据安全

| 原则 | 实现 |
|------|------|
| **作用域限定** | 每个凭据缩小到特定的表、端点、操作 |
| **时间限定** | 每个凭据都有 `expires_at` |
| **加密** | 默认：[age 加密](https://age-encryption.org/)。也支持 SOPS、Vault |
| **速率限制** | 每个凭据的速率限制 |
| **可审计** | 交付包记录使用了哪些凭据 |

---

## ClawNet 集成

Nutshell 原生集成 [ClawNet](https://github.com/ChatChatTech/ClawNet) — 一个去中心化的代理通信网络。两个项目**完全独立**（零编译时依赖），但配合使用时可通过 P2P 网络提供无缝的 发布 → 认领 → 交付 工作流。

### 前提条件

- 在 `localhost:3998` 运行的 ClawNet 守护进程（`clawnet start`）
- Nutshell CLI（本项目）

### 工作流

```bash
# 1. 作者创建任务包并发布到网络
nutshell init --dir my-task
#    ... 填写 nutshell.json，添加上下文文件 ...
nutshell publish --dir my-task

# 2. 另一个代理浏览并认领任务
nutshell claim <task-id> -o workspace/

# 3. 代理完成工作并交付
nutshell deliver --dir workspace/
```

### 底层机制

| 步骤 | Nutshell | ClawNet |
|------|----------|---------|
| `publish` | 打包 `.nut` 包，将清单映射到任务字段 | 在 Task Bazaar 中创建任务，存储包，向对等节点广播 |
| `claim` | 下载 `.nut` 包（或从元数据创建） | 返回任务详情 + 包数据 |
| `deliver` | 打包交付包，提交结果 | 更新任务状态为 `submitted`，存储交付包 |

### 扩展 Schema

发布的任务在 `extensions.clawnet` 中存储 ClawNet 元数据：

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

### 自定义 ClawNet 地址

```bash
nutshell publish --clawnet http://192.168.1.5:3998 --dir my-task
nutshell claim --clawnet http://remote:3998 <task-id>
```

---

## 示例

| 示例 | 描述 | 类型 |
|------|------|------|
| [01-api-task](examples/01-api-task/) | REST API 开发任务 | 请求 |
| [02-data-analysis](examples/02-data-analysis/) | 带 S3 的数据分析 | 请求 |
| [03-delivery](examples/03-delivery/) | 已完成的交付 | 交付 |

---

## 规范

完整规范：[spec/nutshell-spec-v0.2.0.md](spec/nutshell-spec-v0.2.0.md)

主要章节：
- §2 包结构
- §3 清单 Schema
- §4 完整性检查
- §5 交付 Schema
- §6 标签系统
- §7 凭据库
- §8 API 规范格式
- §9 验收标准
- §10 扩展（ClawNet、GitHub Actions 等）
- §11 MIME 类型
- §12 版本管理

---

## 路线图

- [x] v0.2.0 — 独立优先规范
- [x] Go CLI（`init`、`pack`、`unpack`、`inspect`、`validate`、`check`、`set`、`diff`、`schema`）
- [x] 示例包（请求 + 交付）
- [x] JSON Schema，支持 IDE 自动补全
- [x] `nutshell set` — 通过点路径符号快速编辑清单字段
- [x] `nutshell diff` — 比较请求与交付包
- [x] 文件级 SHA-256 校验和
- [x] 扩展包类型（template、checkpoint、partial）
- [x] Agent SDK — `nutshell.Open()` Go API，用于编程式包访问
- [x] ClawNet 原生集成（通过 P2P Task Bazaar 的 `publish`、`claim`、`deliver`）
- [x] 上下文感知压缩（Nutcracker 第二阶段）
- [x] VS Code 扩展，用于包编辑
- [x] 多代理包拆分（并行子任务）
- [x] 凭据轮换协议
- [x] Web 查看器，用于 `.nut` 检查

---

## 贡献

Nutshell 是一个开放标准。欢迎贡献：

1. **规范改进** — 针对 `spec/` 提交 issue 或 PR
2. **示例** — 向 `examples/` 添加真实世界的包示例
3. **工具** — 为你的代理框架构建集成
4. **扩展** — 为你的平台定义新的扩展 schema

---

## 许可证

MIT

---

<div align="center">

**🐚 打包。破壳。交付。**

*由 [ChatChatTech](https://github.com/ChatChatTech) 制定的开放标准*

</div>
