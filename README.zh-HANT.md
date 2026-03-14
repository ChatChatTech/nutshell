<div align="center">

<img src="nutshell-icon.svg" width="80" height="80" alt="nutshell 圖示" />

# nutshell

**一個開放標準，用於打包 AI 代理可理解的任務上下文。**

相容任意代理：Claude Code · Copilot · Cursor · OpenClaw · 自訂代理

[規範](spec/nutshell-spec-v0.2.0.md) · [範例](examples/) · [研究](docs/harness-engineering-research.md) · [網站](https://chatchat.space/nutshell/)

[English](README.md) | [简体中文](README.zh-CN.md) | **[繁體中文](README.zh-HANT.md)** | [Español](README.es-ES.md) | [Français](README.fr-FR.md)

</div>

---

## 問題

AI 程式設計代理很強大，但它們總是反覆詢問相同的問題：

```
代理: "用什麼框架？什麼資料庫？Schema 在哪？
       怎麼認證？驗收標準是什麼？
       能存取預發佈環境嗎？"
人類: *三天內發了 47 則訊息，每次都遺失上下文*
```

每次啟動新會話，你都要重新解釋同樣的上下文。憑證透過 Slack 傳遞。需求只存在你腦子裡。沒有任何紀錄說明做了什麼或為什麼做。

## 解決方案

**Nutshell** 把 AI 代理所需的一切打包到一個套件中：

```
$ nutshell init
$ nutshell check

  🐚 Nutshell 完整性檢查

  ✓ task.title: "建立使用者管理 REST API"
  ✓ task.summary: 已提供
  ✓ context/requirements.md: 存在 (2.1 KB)
  ✗ context/architecture.md: 已參照但缺失
  ✗ credentials: 無金鑰庫 — 代理無法存取資料庫
  ⚠ acceptance: 無測試腳本 — 代理無法自我驗證

  狀態: 不完整 — 2 個項目需要處理後代理才能開始
```

Nutshell 告訴**你**哪些還缺。填補空白，打包，交給任何代理：

```
$ nutshell pack -o task.nut       # 人類打包任務
$ nutshell inspect task.nut       # 代理看到所有需要的內容
# ... 代理執行 ...
$ nutshell pack -o delivery.nut   # 代理交付結果
```

---

## 為什麼選擇 Nutshell？

| 沒有 Nutshell | 有 Nutshell |
|-------------|-----------|
| 上下文散落在 Slack、文件、郵件中 | 一個 `.nut` 套件包含一切 |
| 代理在開始前要問 20 個問題 | 代理讀取清單，立即開始 |
| 憑證以不安全的方式共享 | 加密金鑰庫，帶作用域和時間限制的令牌 |
| 沒有請求或交付的紀錄 | 請求 + 交付套件形成完整的稽核追蹤 |
| 新會話 = 重新解釋一切 | 套件跨會話持久化 |
| 無法驗證完成情況 | 機器可讀的驗收標準 |

### 獨立設計

Nutshell **無需任何外部平台**即可運作。單個開發者使用 Claude Code 就能立即受益：

1. **定義** — `nutshell init` 建立結構化的任務目錄
2. **檢查** — `nutshell check` 告訴你缺少什麼（憑證？架構文件？驗收標準？）
3. **打包** — `nutshell pack` 壓縮成 `.nut` 套件
4. **執行** — 將套件交給任何 AI 代理
5. **歸檔** — 交付套件記錄建構了什麼以及為什麼

### 平台擴充（選用）

想把任務發佈到市場？Nutshell 支援選用擴充：

```jsonc
{
  "extensions": {
    "clawnet": {                    // P2P 代理網路
      "peer_id": "12D3KooW...",
      "reward": {"amount": 50, "currency": "energy"}
    },
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

擴充不會破壞核心格式。工具會忽略它們不理解的內容。

---

## 🐚 名字的由來

> **龍蝦吃貝殼** — *龍蝦吃貝殼。*

[ClawNet](https://github.com/ChatChatTech/ClawNet)（🦞）是一個去中心化的 AI 代理網路。代理就是龍蝦，它們需要食物 — 食物裝在殼裡。**Nutshell**（🐚）就是那個殼 — 緊湊、營養豐富、隨時可以打開。

但你不需要是一隻龍蝦。任何代理都能吃 nutshell。

---

## 快速開始

### 安裝

```bash
# 一鍵安裝（自動偵測作業系統和架構）
curl -fsSL https://chatchat.space/nutshell/install.sh | sh

# 或透過 Go 安裝
go install github.com/ChatChatTech/nutshell/cmd/nutshell@latest

# 或從原始碼建置
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell && make build
```

### 建立任務

```bash
# 初始化
nutshell init --dir my-task
cd my-task

# 編輯清單
vim nutshell.json

# 檢查缺少什麼
nutshell check

# 準備好後打包
nutshell pack -o my-task.nut
```

### 檢視套件內容

```
$ nutshell inspect my-task.nut

    🐚  n u t s h e l l  🦞
    AI 代理任務打包

  Bundle: my-task.nut
  Version: 0.2.0
  Type: request
  ID: nut-7f3a1b2c-...

  📋 任務: 建立使用者管理 REST API
  優先級: high | 工作量: 8h

  🏷️  標籤: golang, postgresql, jwt, rest-api
  領域: backend, authentication

  👤 發佈者: Alice Chen (via claude-code)

  🔑 憑證: 2 個（帶作用域）
    • staging-db (postgresql) — read-write
    • api-token (bearer_token) — invoke

  📦 檔案: 5 個檔案, 8,200 位元組

  ⚙️  Harness 提示:
    代理類型: execution
    策略: incremental
    上下文預算: 0.35
```

### 驗證

```bash
nutshell validate my-task.nut      # 檢查打包後的套件
nutshell validate ./my-task        # 檢查目錄
```

### 快速編輯

```bash
nutshell set task.title "Build REST API"
nutshell set task.priority high
nutshell set tags.skills_required "go,rest,api"
```

### 比較套件

```bash
nutshell diff request.nut delivery.nut          # 人類可讀的差異
nutshell diff request.nut delivery.nut --json   # 機器可讀的差異
```

### JSON Schema

```bash
nutshell schema                            # 輸出到標準輸出
nutshell schema -o nutshell.schema.json    # 寫入檔案
```

加入 `nutshell.json` 以啟用 IDE 自動補全：
```jsonc
{
  "$schema": "./schema/nutshell.schema.json",
  ...
}
```

### 進階命令

```bash
# 上下文感知壓縮 — 分析檔案類型並套用最佳壓縮
nutshell compress --dir ./my-task -o task.nut --level best

# 多代理套件拆分 — 將任務拆分為並行子任務
nutshell split --dir ./my-task -n 3
nutshell merge part-0/ part-1/ part-2/ -o merged/

# 憑證輪換 — 稽核並更新憑證過期時間
nutshell rotate --dir ./my-task                              # 稽核全部
nutshell rotate staging-db --expires 2026-01-01T00:00:00Z    # 輪換單個

# Web 檢視器 — 本地 HTTP 檢視器用於 .nut 檢查
nutshell serve ./my-task --port 8080
nutshell serve task.nut
```

---

## 套件結構

```
task.nut                        🐚 殼
├── nutshell.json               📋 清單（始終最先載入）
├── context/                    📖 需求、架構、參考資料
├── files/                      📦 原始檔案和資源
├── apis/                       🔌 可呼叫的 API 規範
├── credentials/                🔑 加密憑證庫
├── tests/                      ✅ 驗收標準和測試腳本
└── delivery/                   🦪 完成產物（交付套件）
```

只有 `nutshell.json` 是必需的。根據需要添加目錄。

## 清單（`nutshell.json`）

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "request",
  "id": "nut-a1b2c3d4-...",
  "task": {
    "title": "建立使用者管理 REST API",
    "summary": "帶 JWT 認證和 PostgreSQL 的 CRUD 端點。",
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
      "所有 CRUD 端點回傳正確的狀態碼",
      "JWT 認證在受保護路由上正常運作"
    ],
    "auto_verifiable": true
  },
  "harness": {
    "agent_type_hint": "execution",
    "context_budget_hint": 0.35,
    "execution_strategy": "incremental",
    "constraints": ["不要修改 files/src/ 之外的檔案"]
  },
  "completeness": {
    "status": "ready"
  }
}
```

只有 `nutshell_version`、`bundle_type`、`id` 和 `task.title` 是必需的。其他欄位都能提升代理的效率。

---

## Check 命令（反向管理）

最強大的功能：**Nutshell 管理人類**。

```bash
$ nutshell check

  🐚 Nutshell 完整性檢查

  ✓ task.title: "Build REST API"
  ✓ context/requirements.md: 存在 (2.1 KB)
  ✗ context/architecture.md: 已參照但缺失
  ✗ credentials: 無金鑰庫 — 代理無法存取資料庫
  ⚠ acceptance: 無標準 — 代理無法自我驗證
  ⚠ harness: 無約束

  狀態: 不完整 — 填補 2 個項目後代理才能開始
```

不是代理問「我還需要什麼？」，而是**套件告訴人類**該提供什麼。這顛覆了傳統模式，確保代理從一開始就獲得完整的上下文。

---

## Harness Engineering 對齊

Nutshell 基於 [Harness Engineering](docs/harness-engineering-research.md) — 圍繞 AI 代理建構基礎設施的新興學科：

| 原則 | Nutshell 實現 |
|------|-------------|
| **上下文架構** | 分層載入 — 先載入清單，按需載入細節 |
| **代理專業化** | `harness.agent_type_hint` 指導適合哪種代理角色 |
| **持久記憶** | 交付套件保留執行日誌、決策、檢查點 |
| **結構化執行** | 請求/交付分離，帶機器可讀的驗收標準 |
| **40% 規則** | `context_budget_hint` 防止上下文視窗溢出 |
| **約束機械化** | Harness 約束是機器可讀且可執行的 |

---

## 憑證安全

| 原則 | 實現 |
|------|------|
| **作用域限定** | 每個憑證縮小到特定的表、端點、操作 |
| **時間限定** | 每個憑證都有 `expires_at` |
| **加密** | 預設：[age 加密](https://age-encryption.org/)。也支援 SOPS、Vault |
| **速率限制** | 每個憑證的速率限制 |
| **可稽核** | 交付套件記錄使用了哪些憑證 |

---

## ClawNet 整合

Nutshell 原生整合 [ClawNet](https://github.com/ChatChatTech/ClawNet) — 一個去中心化的代理通訊網路。兩個專案**完全獨立**（零編譯時依賴），但搭配使用時可透過 P2P 網路提供無縫的 發佈 → 認領 → 交付 工作流。

### 前提條件

- 在 `localhost:3998` 執行的 ClawNet 守護程式（`clawnet start`）
- Nutshell CLI（本專案）

### 工作流

```bash
# 1. 作者建立任務套件並發佈到網路
nutshell init --dir my-task
#    ... 填寫 nutshell.json，加入上下文檔案 ...
nutshell publish --dir my-task

# 2. 另一個代理瀏覽並認領任務
nutshell claim <task-id> -o workspace/

# 3. 代理完成工作並交付
nutshell deliver --dir workspace/
```

### 底層機制

| 步驟 | Nutshell | ClawNet |
|------|----------|---------|
| `publish` | 打包 `.nut` 套件，將清單對映到任務欄位 | 在 Task Bazaar 中建立任務，儲存套件，向對等節點廣播 |
| `claim` | 下載 `.nut` 套件（或從中繼資料建立） | 回傳任務詳情 + 套件資料 |
| `deliver` | 打包交付套件，提交結果 | 更新任務狀態為 `submitted`，儲存交付套件 |

### 擴充 Schema

發佈的任務在 `extensions.clawnet` 中儲存 ClawNet 中繼資料：

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

### 自訂 ClawNet 地址

```bash
nutshell publish --clawnet http://192.168.1.5:3998 --dir my-task
nutshell claim --clawnet http://remote:3998 <task-id>
```

---

## 範例

| 範例 | 描述 | 類型 |
|------|------|------|
| [01-api-task](examples/01-api-task/) | REST API 開發任務 | 請求 |
| [02-data-analysis](examples/02-data-analysis/) | 帶 S3 的資料分析 | 請求 |
| [03-delivery](examples/03-delivery/) | 已完成的交付 | 交付 |

---

## 規範

完整規範：[spec/nutshell-spec-v0.2.0.md](spec/nutshell-spec-v0.2.0.md)

主要章節：
- §2 套件結構
- §3 清單 Schema
- §4 完整性檢查
- §5 交付 Schema
- §6 標籤系統
- §7 憑證庫
- §8 API 規範格式
- §9 驗收標準
- §10 擴充（ClawNet、GitHub Actions 等）
- §11 MIME 類型
- §12 版本管理

---

## 路線圖

- [x] v0.2.0 — 獨立優先規範
- [x] Go CLI（`init`、`pack`、`unpack`、`inspect`、`validate`、`check`、`set`、`diff`、`schema`）
- [x] 範例套件（請求 + 交付）
- [x] JSON Schema，支援 IDE 自動補全
- [x] `nutshell set` — 透過點路徑符號快速編輯清單欄位
- [x] `nutshell diff` — 比較請求與交付套件
- [x] 檔案級 SHA-256 校驗和
- [x] 擴充套件類型（template、checkpoint、partial）
- [x] Agent SDK — `nutshell.Open()` Go API，用於程式化套件存取
- [x] ClawNet 原生整合（透過 P2P Task Bazaar 的 `publish`、`claim`、`deliver`）
- [x] 上下文感知壓縮（Nutcracker 第二階段）
- [x] VS Code 擴充，用於套件編輯
- [x] 多代理套件拆分（並行子任務）
- [x] 憑證輪換協定
- [x] Web 檢視器，用於 `.nut` 檢查

---

## 貢獻

Nutshell 是一個開放標準。歡迎貢獻：

1. **規範改進** — 針對 `spec/` 提交 issue 或 PR
2. **範例** — 向 `examples/` 加入真實世界的套件範例
3. **工具** — 為你的代理框架建構整合
4. **擴充** — 為你的平台定義新的擴充 schema

---

## 授權條款

MIT

---

<div align="center">

**🐚 打包。破殼。交付。**

*由 [ChatChatTech](https://github.com/ChatChatTech) 制定的開放標準*

</div>
