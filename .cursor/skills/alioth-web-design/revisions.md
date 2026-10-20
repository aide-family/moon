# 已确认的样式修订

只记录用户确认后、之后每次做界面都要遵守的决定。新条目加在最上面。

## 2026-10-21 个人资料页与同类设置页

适用：个人资料页，以及以后同类的「只读信息 + 少量可编辑项」设置页。

- 一张 `Card`：`CardHeader` 用 `CardTitle`（16px，身份取昵称/名称/邮箱）+ `CardDescription`（14px muted 邮箱，`break-all`）。只读字段放 `CardContent` 的 `dl`（`grid gap-4 sm:grid-cols-2`，每项 `grid gap-1`，`dt` 用 `text-sm text-muted-foreground`），不要在手写 `<p>` 里拼「标签: 值」。
- 只读区与表单之间放 `Separator`，`CardContent` 用 `grid gap-6`。
- 主操作放 `CardFooter`（`justify-end gap-2`）右对齐，不拉满卡片宽度；保存状态用 `role="status"` 放按钮左侧 `mr-auto`。表单用 `id` 与按钮 `form` 关联。
- 宽度 `max-w-2xl`；内容卡用 `glass-card`；页面根节点不滚动，滚动交给内层 `min-h-0 flex-1 overflow-y-auto`。
- 加载态在同一张卡里用 `Skeleton`（身份取 session，只给可编辑项占位），失败态在卡内给 `text-sm text-destructive` + `variant="outline"` 重试。

## 2026-10-19 重复条件卡片

适用：告警规则级别条件，以及以后同类的重复条件卡片。

- 身份字段（告警级别）放在卡片第一行左侧，删除图标放在该行右上角。可以删掉任意一条，包括只剩一条时。
- 条件字段顺序为：比较条件、阈值、时间方式、时间。时间单位是秒，输入框右侧固定显示 `s`。

## 2026-10-19 工具条、主从 Tabs、PromQL 与指标结果

适用：全局，包括列表工具条、主从右侧 Tabs、PromQL 编辑器与指标查询结果。

- 同行工具条：搜索 `Input` 与搜索按钮必须 `flex-nowrap`，按钮 `shrink-0`，禁止因外层 `flex-wrap` 把按钮拆到下一行。左右两组之间可以换行。
- 主从页右侧带 Tabs：`TabsList` 固定不随内容滚动；仅 `TabsContent`（及内部表格容器）滚动。`Card` / 内容区用 `flex min-h-0 flex-1 overflow-hidden`，不要把 Tab 条放进整块 `overflow-y-auto`。
- PromQL 编辑器：无行号、无行高亮、无 overview / glyph / folding；轻 `border-input`，左右留约 10px 内边距，文字不贴边框。背景用 `bg-background`，与全局背景一致，不要透明、不要 Monaco 默认灰底。空内容保持约一行高度，多行按内容增高（有上限）。补全列表打开时 **Enter** 接受当前建议；补全关闭时 **Enter** 触发查询/提交；**Shift+Enter** 换行。补全用 `fixedOverflowWidgets`，不要被编辑器 `overflow` 截断。
- 所有表格：表头相对 `Table` 滚动容器固定，只滚表体，宽度不超出父级视口。不要在 `Table` 外套第二层 `overflow`。
- 指标查询结果（表 / 图 / JSON）用 `min-h-0 flex-1` 铺满剩余高度，内部滚动，不撑开外层。查询条件在左、结果视图切换在同一行最右。结果表横向滚动时第一列（优先 `__name__`）`sticky left-0`。图表与图例同高铺满；图例单行省略。图例和被截断的元数据说明各自悬停都等 1.5 秒再提示（`skipDelayDuration={0}`），不要上一个刚出现就让下一个立刻弹出。无选中时画全部序列，有选中时只画选中的，选空则恢复全部。JSON 用 `monaco.editor.colorize` 代码块（主题与 PromQL 编辑器相同），不是只读编辑器。复制用图标；成功后变为勾，2 秒后恢复。

## 2026-10-17 边缘控制台

适用：机器、SSH 命令、SFTP、Web Shell，以及全局滚动条。以后新页面沿用滚动条和互斥审核。

- 已通过的命令执行时选择多台在线机器，同一条命令分别下发。未通过不能执行。
- 同一条命令上「通过」和「拒绝」不同时出现。未通过只显示通过；已通过只显示拒绝。
- 机器列表展示已采集的系统、CPU、内存、磁盘。机器详情用宽抽屉 `sm:max-w-3xl`，字段两列。其他详情仍用窄抽屉。
- 在线机器且具备 `ssh:exec` 时提供登录，进入 `/agents/:uid/shell`。离线不提供登录。同一台机器同时只开一个终端。
- SFTP 不提供新建目录。文件区用图标网格，目录在前，双击进入，面包屑返回。保留上传、下载、删除。
- 全局滚动条：`scrollbar-width: thin`，滑块 8px、圆角 4px、轨道透明。暗色滑块用 `rgba(255, 255, 255, 0.2)`。

## 2026-10-17 错误提示、枚举文案、左侧列表与分步提交

适用：全局，包括以后新做的页面。

- 验证码图片用可读字号绘制，页面按原始比例展示，不要压到 `h-12`。刷新验证码时保留已经出现的错误。
- 接口错误：400 写到表单字段（对不上字段就写在表单错误区）；401 清除令牌并进入登录页，登录、注册和验证码接口的 401 留在当前表单；403 进入无权限页；404 用全局 message；5xx 用全局 notify。409 写操作留在表单错误区，列表加载显示服务端消息。
- 类型、状态、驱动、比较条件等协议常量用 `enumLabel` 显示中英文，请求体仍传常量。
- 左侧主从列表（数据源、规则组）名称前用绿点表示启用、红点表示停用。操作收进 `EllipsisVertical` 下拉：详情、编辑、删除、禁用/启用。删除和禁用先确认，菜单项用 destructive。
- `StepForm` 的「下一步」和「保存」必须是两个按钮。点下一步不能在同一次点击里提交。只有最后一步才保存并关闭。

## 2026-10-16 登录注册左右分栏

适用：登录页、注册页。

- 大屏（`lg` 及以上）左侧约 `55%` 自动轮播插画与介绍文案，小于 `lg` 隐藏左侧。
- 右侧表单放在 `glass-card w-full max-w-md rounded-2xl`，垂直居中。主按钮拉满宽度。错误在表单下方 `text-sm text-destructive`。
- 右上角固定主题与语言图标菜单，和顶栏同一套 `LocaleThemeMenu`。
- 登录仍保留密码 / 邮箱验证码切换；登录与注册仍分路由，底部互跳不变。背景透出 `--app-gradient`。

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
- 现在的领域：告警（数据源、告警规则、告警字典），系统管理（租户、用户、成员、角色）。当前页所在领域保持展开。
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
