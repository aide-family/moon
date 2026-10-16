# 已确认的样式修订

只记录用户确认后、之后每次做界面都要遵守的决定。新条目加在最上面。

## 2026-10-16 多列表单行控件铺满

适用：全局新建/编辑表单（含 Dialog、StepForm、行内多列栅格）。

- 同一行多个输入/选择时，每个控件必须铺满自己的栅格单元格，禁止彼此之间留下大块空白。
- `FormField` 根节点用 `min-w-0`，并对子树强制 `[&_[data-slot=select-trigger]]:w-full`（`Input` 本身已是 `w-full`）。
- 未包在 `FormField` 内的多列 `Select`（例如告警规则级别行）必须给 `SelectTrigger` 显式 `className="w-full"`。不要依赖 `SelectTrigger` 默认的 `w-fit`。

## 2026-10-16 Alioth 视觉体系（蓝主色与玻璃壳）

适用：全局 `web/` 视觉与壳。

- 主色 / ring / sidebar-primary / chart-1 为 `#1677ff`；背景 `--app-gradient`；列表卡片 `.glass-card`；字体 sans Inter、heading IBM Plex Sans、mono JetBrains Mono（Geist 作回退）。token 在 `web/src/index.css`。
- 壳：`Sidebar variant="inset"`；顶栏 `h-12` + `bg-background/80 backdrop-blur-sm`；导航 active 左侧 `border-primary`；主内容区透出渐变。组件栈为 radix-nova + shadcn + lucide，不引入 `antd`，不换 base-nova。
- 交互：领域子菜单、登录/注册分路由与互跳、表单 placeholder + 必填校验、多字段 Dialog `sm:max-w-2xl`、列表总数左/翻页右、页内容 `gap-6`。

## 2026-10-16 表单校验、placeholder 与登录注册入口

适用：全局，包括以后新做的页面。

- 每个 `Input` / `Textarea` / `Select`（`SelectValue` 的 placeholder）都必须有 i18n placeholder；可选字段也要（可写「可选…」）。
- 提交前按后端 proto `buf.validate` 与 biz 必填做前端校验；空值不发请求。不引入 zod/RHF：文本用 `required` + submit 时 `trim`；必填 Select 禁止 `NONE`/空串；`StepForm` 点「下一步」先校验当前步。
- 客户端校验失败与服务端错误共用表单错误展示位。
- 登录与注册保持分路由；底部必须有互跳链接，文案为「没有账号？去注册」/「已有账号？去登录」（及对应英文）。

## 2026-10-16 多字段对话框宽度与栅格

适用：全局新建/编辑 `Dialog`，包括以后新做的页面。

- 1–2 个字段的简单对话框：保持默认 `sm:max-w-md`，表单单列。
- 多字段新建/编辑：`DialogContent` 用 `sm:max-w-2xl`（可叠 `max-h-[80vh]`），表单用 `grid gap-4 sm:grid-cols-2`。
- 内容较短的控件（端口、类型、状态、数字阈值等）可同排最多 4 列：`sm:grid-cols-3` 或 `sm:grid-cols-4`，不超过 4。
- 长文本（备注、私钥、endpoint、大段命令等）用 `sm:col-span-full`（或跨满当前栅格）。
- 字段很多时仍可分步；步内也可多列。页眉和步骤按钮固定，当前步在框内滚动。
- 不改 `DialogContent` 默认宽度，由各多字段表单显式传 `className`。

## 2026-10-15 菜单、对话框和列表

适用：全局，包括以后新做的页面。

- 侧栏第一级只放功能领域，用可折叠子菜单。页面放在对应领域下面。第一级不放直接进入功能的链接。个人资料仍只从顶栏进入。
- 现在的领域：告警（数据源、告警规则、告警级别），系统管理（租户、用户、成员、角色）。当前页所在领域保持展开。
- 新建和编辑仍用 `Dialog`。对话框 `flex max-h-[80vh] flex-col overflow-hidden`。1–2 个字段用默认 `sm:max-w-md`。多字段用 `sm:max-w-2xl` 与多列栅格（见 2026-10-16）。字段很多就分步，页眉和步骤按钮固定，当前步在框内滚动。
- 列表在表格上方：左侧总数，右侧上一页和下一页。没有数据、加载中或失败时不渲染翻页。表头 `sticky top-0`。滚动表体，高度用页头和筛选之外的剩余视口。筛选和分页仍进 URL。

## 2026-10-15 交互约束

适用：全局，包括以后新做的页面。

- 系统管理菜单放在侧栏一个可折叠子菜单里。个人资料不进这个子菜单，只从顶栏用户图标进入。
- 每个输入框和选择框都有 placeholder。标签本身说不清约束或后果时，在 label 旁放 `CircleHelp` tooltip。
- 语言、主题、个人使用图标按钮。有多个选项时用下拉菜单，不用来回切换的文字按钮。
- 新建和编辑用 `Dialog`。字段少的详情用右侧 `Sheet`。需要独立地址或分成多块的详情，用当前页的子路由。
- 一行操作超过两项时，操作区只留「详情」和「更多」。更多是下拉菜单。两项及以内直接摆出来，其中可查看的行仍保留「详情」。
- 删除、封禁、停用、移交负责人是危险操作：先弹出确认框，确认按钮用 destructive。删除、封禁、停用在菜单里也用 destructive。启用和解封不走这套。
