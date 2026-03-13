# Harness Engineering 深度调研报告

> 调研日期: 2026-03-14
> 来源: OpenAI 官方报告, 知乎 LLM应用技术指北专栏

---

## 一、定义

**Harness Engineering** 是围绕 AI Agent（尤其是 Coding Agent）设计和构建约束机制、反馈回路、工作流控制和持续改进循环的系统工程实践。

"Harness" 本意是马具（缰绳、鞍具），把马的力气引到正确方向。LLM 就像一匹蛮力十足但方向感不太行的马。

### 三层工程概念嵌套

```
Harness Engineering ⊇ Context Engineering ⊇ Prompt Engineering
```

- **Prompt Engineering**: 单次输入优化
- **Context Engineering**: "给 Agent 看什么" — 上下文的组织与投递
- **Harness Engineering**: "系统怎么防崩、怎么量化、怎么修" — 完整的工程体系

> Phil Schmid: "模型是 CPU，Harness 是操作系统 — CPU 再强，OS 拉胯也白搭。"

### 术语起源 (2026.02)

| 人物 | 贡献 |
|------|------|
| Mitchell Hashimoto | 首次命名, 提出"Agent犯错→设计方案→永不再犯" |
| OpenAI | 发布百万行代码实验报告 |
| Ethan Mollick | "Models, Apps, Harnesses" 框架 |
| Martin Fowler | 深度分析, 分为 Context Engineering / Architecture Constraints / Garbage Collection |

---

## 二、为什么需要 Harness Engineering

### 2.1 瓶颈在基础设施，不在模型智能

量化证据:
- **Can.ac**: 仅改变工具格式，Grok Code Fast 1 从 6.7% → **68.3%**（无修改权重）
- **LangChain**: Terminal Bench 2.0 从第 30 名 → 第 5 名（同一模型，+13.7 分）

> "Five independent teams. Same conclusion: the bottleneck is infrastructure, not intelligence." — Alex Lavaee

### 2.2 Agent 典型失败模式 (Anthropic 总结)

| 模式 | 描述 |
|------|------|
| 一步到位 | 上下文窗口耗尽, 下一会话面对半成品 |
| 过早宣布胜利 | 看到部分进展就声称完成 |
| 过早标记完成 | 不做端到端测试就标 done |
| 环境启动困难 | 花大量 token 搞清怎么运行应用 |

### 2.3 上下文窗口甜蜜区间

> 约 **40%** 填充率为性能拐点。超过后进入 "Dumb Zone"：幻觉、循环、格式错误。

---

## 三、四大支柱

### 3.1 上下文架构 (Context Architecture)

Agent 应当恰好获得当前任务所需的上下文——不多不少。

**三层上下文体系:**

| 层级 | 加载时机 | 内容 | 占用 |
|------|----------|------|------|
| Tier 1: 会话常驻 | 每次自动 | AGENTS.md, 项目结构 | 最小 |
| Tier 2: 按需加载 | 特定子Agent调用时 | 专业知识, 领域上下文 | 中等 |
| Tier 3: 持久化知识库 | 主动查询时 | 规格说明, 历史会话 | 按需 |

### 3.2 Agent 专业化 (Agent Specialization)

专注特定领域、拥有受限工具的 Agent 优于拥有全部权限的通用 Agent。

| Agent 角色 | 职责 | 工具权限 |
|-----------|------|---------|
| 研究 Agent | 探索代码库 | 只读 |
| 规划 Agent | 分解任务 | 只读 |
| 执行 Agent | 实现功能 | 限定读写 |
| 审查 Agent | 审计工作 | 只读+标记 |
| 调试 Agent | 修复问题 | 限定修复 |
| 清理 Agent | 对抗熵 | 读写 |

### 3.3 持久化记忆 (Persistent Memory)

进度持久化在文件系统上，而非上下文窗口中。

- Anthropic: `claude-progress.txt` + git log + feature list JSON
- JSON > Markdown 用于状态追踪（Agent 不容易误改结构化数据）

### 3.4 结构化执行 (Structured Execution)

将思考与执行分离: **理解 → 规划 → 执行 → 验证**

> Boris Tane: "永远不要让 Agent 在你审查和批准书面计划之前写代码。"

---

## 四、先进团队实战案例

### 4.1 OpenAI: 百万行代码零手写

| 指标 | 数值 |
|------|------|
| 团队 | 3 名工程师 |
| 时长 | 5 个月 |
| 代码 | ~100 万行 |
| 手写代码 | 0 行 |
| PR 数 | ~1,500 |
| 日均 PR/人 | 3.5 |

**五大原则:**
1. 设计环境，而非编写代码
2. 机械化执行架构约束 (`Types → Config → Repo → Service → Runtime → UI`)
3. 代码仓库作为唯一事实源
4. 将可观测性连接到 Agent
5. 对抗熵 (20% 时间 → 自动化垃圾回收)

### 4.2 Anthropic: 16 Agent 构建 C 编译器

| 指标 | 数值 |
|------|------|
| 并行 Agent | 16 个 Claude Opus 4.6 |
| 会话数 | ~2,000 |
| Rust 代码 | 100,000 行 |
| GCC torture test | 99% 通过 |
| 可编译项目 | 150+ (PostgreSQL, Redis, Linux Kernel...) |
| 成本 | ~$20,000 |

### 4.3 Stripe: Minions 系统

开发者在 Slack 发任务 → Agent 全程包办 → 人只审查 PR。
- Toolshed MCP: ~500 个工具
- 隔离的预热 Devbox
- Agent 是一等公民，不是事后补上的集成

---

## 五、六大共识

| # | 共识 | 共识度 |
|---|------|--------|
| 1 | 瓶颈在基础设施，不在模型智能 | ★★★★★ |
| 2 | 文档必须是活的反馈循环 | ★★★★☆ |
| 3 | 思考与执行必须分离 | ★★★★★ |
| 4 | 上下文不是越多越好 (~40%) | ★★★★☆ |
| 5 | 约束必须机械化执行 | ★★★★☆ |
| 6 | 工程师角色从"写代码"→"设计环境" | ★★★★☆ |

## 六、四大分歧

1. **Harness 复杂化 vs 简化**: 通用产品倾向简化, 定制项目需要精细化
2. **单 Agent vs 多 Agent**: 规模决定选择
3. **人类介入程度**: 取决于 Harness 成熟度
4. **术语边界**: 嵌套 vs 互补，尚无定论

## 七、三大空白区（Nutshell 的机会）

1. **棕地项目改造**: 零成功案例, 最大实践缺口
2. **功能/行为验证系统化**: 被指出但无解决方案
3. **AI 生成代码长期可维护性**: 问题已提出, 无人回答

---

## 八、对 Nutshell 设计的启示

### 来自 Harness Engineering 的核心洞察:

1. **Context Architecture → 任务打包应分层**: 不是把所有信息塞进一个文件，而是分层渐进式披露
2. **Agent Specialization → 包需要标记目标 Agent 类型**: 不同角色的 Agent 需要不同的上下文切面
3. **Persistent Memory → 包应支持状态追踪**: 任务进度、交接记录、决策日志
4. **Structured Execution → 包应分离需求与执行计划**: Publisher 提需求，Agent 生成计划，验证基于自动化
5. **约束机械化 → 包应内嵌验收标准**: 可执行的测试用例，不是模糊的文字描述
6. **可观测性 → 包应包含 API 端点、凭据、监控接入点**: Agent 需要看到运行时状态
7. **40% 上下文规则 → 包需要压缩和分层加载**: 不能一股脑全给，需要按需展开

### Nutshell 独有的创新空间:

- **Credential Sharing**: 安全共享 API 凭据，Harness Engineering 未明确覆盖
- **Cross-Repo Context**: 跨仓库的上下文打包, 超越单仓库 AGENTS.md
- **Task Marketplace Integration**: 与 ClawNet 供需匹配对接
- **Compression**: 智能压缩，控制上下文窗口占用

---

## 参考文献

1. OpenAI — "Harness engineering: leveraging Codex in an agent-first world"
2. 知乎 LLM应用技术指北 — "Harness Engineering 深度解析：AI Agent 时代的工程范式革命"
3. Anthropic — "Effective harnesses for long-running agents"
4. Nicholas Carlini — "Building a C Compiler with Claude"
5. Martin Fowler — "Harness Engineering"
6. Mitchell Hashimoto — "My AI Adoption Journey"
7. Stripe — "Minions: Stripe's one-shot, end-to-end coding agents"
8. Vasilopoulos et al. (2026) — "Codified Context: Three-Tier Context Infrastructure"
