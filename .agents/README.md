# .agents

仓库内的 agent 预设、技能与规则。DSH 会自动发现 `.agents/skills/`；规则文件按需读取。

## skills/

每个子目录是一个技能：`SKILL.md`（带 `name` / `description` frontmatter），需要的细节放在同目录的
引用文件里，由 `SKILL.md` 指出什么时候读。DSH 的技能发现路径包含 `<repo>/.agents/skills`
（优先级 200），会话开始时这些技能会出现在技能清单中。

改哪一类工作，先读对应的 `SKILL.md`；样式类修订要写回 `alioth-web-design/revisions.md`。

| 技能 | 用于 |
| --- | --- |
| `alioth-platform` | 仓库整体：服务、端口、边界 |
| `alioth-backend` | Go 分层、ent、凭证 |
| `alioth-update-api` | 新增/修改接口的流程 |
| `alioth-milestone` | 里程碑推进方式 |
| `alioth-web` | Web 请求、权限、文案 |
| `alioth-web-page` | 新页面落地 |
| `alioth-web-design` | 布局、间距、色板、组件外观（修订记 `revisions.md`） |
| `alioth-web-review` | 前端改动自检与放行 |

## rules/

给编辑器与 agent 共用的规则摘要（`.mdc`）：平台边界、后端分层、Web 实现约定、前端设计标准。
**权威版本是 [docs/](../docs/) 与 [web/AGENTS.md](../web/AGENTS.md)，冲突时以那里为准。**

## worktrees/

[git-workflow.md](../docs/git-workflow.md) 里 `git worktree add` 用的临时工作区（已被 git 忽略），
不要提交其中的内容。
