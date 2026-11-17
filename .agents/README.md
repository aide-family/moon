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

### 前端动效（外部技能）

七个外部技能，装成仓库内普通目录，`git` 里只有 `.agents/skills/` 这一份：

| 技能 | 来源 | 用于 |
| --- | --- | --- |
| `animate` | emilkowalski/skills | 从零造一个动效：该不该动、目的是什么、选 CSS / `@starting-style` / WAAPI / Motion、曲线与时长、打断与退出 |
| `improve-animations` | emilkowalski/skills | 全库动效体检，产出分优先级的整改计划（只读，不改代码） |
| `review-animations` | emilkowalski/skills | 审一段动效 diff，按同一套标准挑问题 |
| `animation-performance` | emilkowalski/skills | 掉帧、合成属性、CSS 与 JS 动画的取舍、`will-change` |
| `animation-vocabulary` | emilkowalski/skills | 把"那个弹弹的效果"翻译成准确术语，便于描述给 agent 或设计师 |
| `vercel-react-view-transitions` | vercel-labs/agent-skills | React 19 View Transition API：`<ViewTransition>`、`addTransitionType`、CSS 伪元素，覆盖路由切换、共享元素、列表重排、进出场 |
| `fixing-motion-performance` | ibelick/ui-skills | 动效性能审计：layout thrashing、合成属性、滚动驱动动画、blur |

`emilkowalski/skills` 出自 Emil Kowalski（`sonner` / `vaul` 作者）。`vercel-react-view-transitions`
上游目录名是 `react-view-transitions`、frontmatter 的 `name` 才是前者，这里把目录名对齐 `name`，
并且只留 `SKILL.md` 与 `references/`（`AGENTS.md`、`README.md`、`metadata.json`、`.zip` 是上游打包用的）。

动效方案本身用这些技能，落地约束仍归 `alioth-web-design` 与 [web/AGENTS.md](../../web/AGENTS.md)：
色值走语义 token、对话框结构不许重写、键盘高频操作不留在动画里。`vercel-react-view-transitions`
是 React + `document.startViewTransition` 的方案，不装第三方库；要 springs / layout / exit 动画时
才需要给 `web/` 加 `motion` 依赖，那是独立改动。

升级上游时按下面的命令重装，装完删掉 CLI 顺手生成的 `.claude/`、`.hermes/`、`.lingma/`、`.trae/`、
根目录 `skills/` 与 `skills-lock.json`，再看 `git diff .agents/skills` 决定是否提交。
本机 `git clone https://github.com` 经常超时，克隆失败时改用 codeload 打包：
`curl -L https://codeload.github.com/<owner>/<repo>/tar.gz/refs/heads/main`，解压后取对应技能目录。

```sh
npx skills add emilkowalski/skills -s animate -s improve-animations -s review-animations \
  -s animation-performance -s animation-vocabulary --copy -y

npx skills add vercel-labs/agent-skills -s react-view-transitions --copy -y
npx skills add ibelick/ui-skills -s fixing-motion-performance --copy -y
```

## rules/

给编辑器与 agent 共用的规则摘要（`.mdc`）：平台边界、后端分层、Web 实现约定、前端设计标准。
**权威版本是 [docs/](../docs/) 与 [web/AGENTS.md](../web/AGENTS.md)，冲突时以那里为准。**

## worktrees/

[git-workflow.md](../docs/git-workflow.md) 里 `git worktree add` 用的临时工作区（已被 git 忽略），
不要提交其中的内容。
