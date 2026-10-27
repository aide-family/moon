## 2026-10-27 壳层分割线、资料页第三方账号与登录备案

适用：全局壳、个人资料页、登录/注册/OAuth 回调页。

- 主内容区不再用 inset 卡片：`Sidebar` 改回默认 `variant="sidebar"`，`SidebarInset` 不保留外边距、圆角或阴影，内容区与视口齐平；菜单与内容之间只用一条 1px 竖线分隔，不再出现圆角浮层和投影。
- 菜单右侧分割线与顶栏下边框保持同一颜色，统一用 `border-border/60`；不要另造更亮的分割线 token（试过 `--sidebar-divider`，用户反馈太亮已回退）。
- 个人资料页 `xl` 及以上两栏：个人资料卡在左、第三方账号卡在右，容器 `xl:max-w-5xl xl:grid-cols-2`；更窄屏幕保持单栏 `max-w-2xl`。已绑定的第三方账号必须逐项展示明细（第三方账号、OpenID、绑定时间），不能只拼一行。
- 登录页底部居中放备案号链接，外链 `https://beian.miit.gov.cn/`；目前没有真实备案号，先用占位号（`login.icp`），拿到后只改文案。

## 2026-10-27 Kubernetes 集群详情默认视图、筛选框与命名空间选择

适用：Kubernetes 集群列表，以及集群详情页的资源导航、筛选条和命名空间选择。

- 从集群列表点「详情」进入时默认落在「集群概况」，URL 不写 `tab`；其他资源视图一律显式写 `tab`，没有 `tab` 的旧链接也按概况处理。
- 清单筛选只保留输入框与 placeholder，不再渲染可见 label；无障碍名称改由 `aria-label` 承担。节点、Pod、工作负载等所有资源视图共用同一套无 label 筛选条，集群列表的名称搜索同样处理。
- 命名空间选择使用共享 `SearchableSelect`：弹出面板提供「搜索命名空间」输入框和可勾选选项，第一项为「全部命名空间」（空值）；选中后写回 URL 的 `namespace`，窄屏下弹层必须完整落在视口内。

## 2026-10-27 对话框固定头尾与 GitHub 代码块编辑器

适用：全局新建/编辑对话框，以及需要编辑 kubeconfig、YAML/JSON 等配置的字段。

- 对话框固定三段式：`DialogContent` 保持 `flex max-h-[80vh] flex-col overflow-hidden`；`DialogHeader` / `DialogFooter` 加 `shrink-0` 固定，只有中间内容区 `min-h-0 flex-1 overflow-y-auto` 纵向滚动。页脚操作按钮不随内容滚动，内容再多也不能把「取消/保存」挤出视口。
- 表单级提交错误放在滚动区之外、页脚之上；字段级错误仍留在对应字段下方（滚动区内）。
- 可编辑配置（kubeconfig、YAML/JSON 等）不用 `Textarea`，用共享 `CodeEditor`：GitHub light/dark 配色（与 `CodeBlock` 同一套 `--code-bg` / `--code-fg`）、等宽字体、无行号、无行高亮、无 overview/glyph/folding，内容超出时在编辑器内部滚动；外框沿用 `CodeBlock` 卡片样式（`rounded-lg border border-border/70` + 语言标题栏）。编辑器放 `sm:col-span-full`，随内容区一起滚动。

## 2026-10-22 顶栏租户切换与项目仓库入口

适用：全局顶栏。

- 顶栏的租户切换下拉触发器不加边框，背景与顶栏一致：`border-0 bg-transparent shadow-none dark:bg-transparent dark:hover:bg-transparent`，不要独立底色或描边。
- 顶栏右上角固定放项目 GitHub 图标链接，新窗口打开 `https://github.com/aide-family/Alioth`，图标颜色跟随文字颜色。

## 2026-10-22 通知页与全局操作列

适用：通知渠道、通知模板、通知组、告警订阅、消息日志，以及所有带操作列的表格。

- 所有表格操作列必须显示「操作」表头；「详情」「更多」及其他非危险行内操作统一用 Link 样式（`Button variant="link" className="h-auto px-0"`），危险操作仍用 destructive。
- 通知渠道和通知模板的筛选条中，类型和应用筛选都带对应图标；搜索输入与搜索按钮放在同一个 `flex flex-nowrap` 容器里，按钮 `shrink-0`。
- 通知组表单：邮件渠道、短信渠道各一个下拉，其余 Webhook 作为可多选接收者；只有选中邮件渠道才出现邮件接收人和成员邮箱，只有选中短信渠道才出现短信接收人和成员手机号，两渠道都没选时不渲染人员板块。每个选中渠道都要选择能覆盖它的模板，Webhook 按应用分别选模板。
- 告警订阅对话框里通知组、告警级别、数据源用横向复选框标签（`flex flex-wrap gap-2`，宽度不足自动换行）；通知组必选并放在级别和数据源之前。
- 消息日志的「投递日志 / 通知任务」改成 Tabs，移动到表格上方最右侧，与左侧筛选同一行；投递日志和通知任务都展示来源通知组和订阅名称。
- 通知模板的正文和邮件主题使用 Monaco 模板编辑器：`{{`、`$`、`.` 触发补全，补全项来自事件字段和 Go/Prometheus 模板函数；预设模板用 `/notification-templates/presets` 主从页，左侧名称、右侧可编辑内容，保存生成当前租户副本，并提供 mock 数据预览。

# 已确认的样式修订

只记录用户确认后、之后每次做界面都要遵守的决定。新条目加在最上面。

## 2026-10-21 告警规则表达式、事件图表与抑制

适用：告警规则列表与详情、实时告警事件图表、抑制告警对话框。

- 告警规则的表达式展示不再使用 `ExprBlock` 的边框/底色容器：直接展示原始表达式（等宽、保留换行、超长只在自身范围纵向滚动），紧随其后逐行展示级别条件。`ExprBlock` 继续用于实时告警详情。
- 事件图表只展示与事件标签一致的序列：序列的每个标签都必须出现在事件标签中且值相同，不再回退展示表达式的其他序列；没有命中时显示空状态。
- 事件图表的表达式下方必须紧跟这条事件命中的级别条件（级别 + 采样方式 + 比较条件 + 阈值 + 持续秒数），和表达式搭配展示；规则已删除或取不到条件时只展示表达式。
- 实时告警详情里的紧凑图表也展示规则表达式，表达式单独一行、允许换行。
- 抑制告警的时长梯度用「主色实心 + 勾选图标 + 外圈」表示选中，未选中项弱化文字；除常用梯度和自定义外提供「永久」选项，永久抑制不设到期时间，详情中抑制结束显示「永久」。
- 选择告警页面时必须显示数量上限（10 个），达到上限后未选中的复选框禁用，批量选择也只补到剩余名额。

## 2026-10-21 告警界面与通用菜单

适用：全局下拉菜单，以及告警相关页面。

- 下拉菜单里危险操作（删除、停用、抑制、恢复等 destructive 项）一律排在普通操作之后，两组之间用 `DropdownMenuSeparator` 隔开。`RowActions` 自动分组，页面内手写的菜单也要遵守。
- 选择类对话框（如「选择告警页面」）：顶部标题/说明/批量操作/筛选固定，底部操作固定，只有中间列表 `min-h-0 flex-1 overflow-y-auto` 滚动；整框 `flex max-h-[80vh] flex-col overflow-hidden`。批量操作提供全选、取消全选、按类型选择（如选私有），并带本地筛选框和已选数量。
- 只有一条数据序列的图表不显示图例，图表占满整行；数据源查询这类多序列图表仍保留图例。
- 告警表达式展示统一用 `ExprBlock`（`bg-muted/40` 边框块 + 等宽、保留换行、块内滚动）。实时告警详情继续使用；告警规则列表与详情按上面 2026-10-21 的新约定改用无边框原始表达式。
- 抑制实时告警必须先选时长：常用梯度加自定义输入，提交前校验 1 分钟到 30 天，并显示抑制到期时间。

## 2026-10-21 实时告警详情抽屉

适用：实时告警详情，以及以后同类的多字段告警详情抽屉。

- 抽屉宽度 `40vw`：`data-[side=right]:sm:w-[40vw]! data-[side=right]:sm:max-w-[40vw]!`；窄屏保持默认宽度。
- 头部固定，内容区 `min-h-0 flex-1 overflow-y-auto overflow-x-hidden`。只允许纵向滚动，任何字段和标签都 `min-w-0` + `break-words`，表达式用 `whitespace-pre-wrap break-words`，不允许横向滚动。
- 规则、级别、数据源显示名称并链接到对应管理页：`/alert-rules?groupUid=..&ruleUid=..`、`/alert-dictionaries?type=ALERT&uid=..`、`/datasources?uid=..`。名称由事件接口随事件返回，规则或字典软删除后仍显示历史名称。
- 标签用 `key=value` 的 tag（`Badge` 变体 `secondary`）展示，按 key 排序；长值在 tag 内换行，不撑宽抽屉。事件标签由后端附加数据源名称（`datasource=<名称>`）和数据源自身标签。
- 表达式和命中的级别条件放在同一个 `bg-muted/40` 块里：上面是表达式原文，下面用 `border-t` 分隔一行 `级别条件 <级别>：<时间方式><比较条件> <阈值> · <秒> 秒`。条件按事件的 `ruleUid` + `levelUid` 从规则取，取不到时只显示表达式。
- 抽屉底部固定展示事件图表（复用事件图表组件，`compact` 模式高度 `h-56`，不再重复表达式元信息）。

## 2026-10-21 实时告警 Tab 栏

适用：实时告警页顶部按告警页面切换的 Tab 栏。

- 「我的告警」固定第一，不参与拖动。其余告警页面 Tab 整块可拖动（`cursor-grab`，左侧 `GripVertical` 手柄），拖放后先本地生效再调 `PUT /v1/alert-pages/my-tabs` 持久化，失败回滚并在 Tab 下方显示错误。顺序存在 `user_alert_pages.sort`。
- 数量用 `Badge` 标注，与 Tab 文案颜色区分：有告警时 `destructive`（红色），为 0 时 `secondary`。`font-mono tabular-nums`。

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

适用：机器、SSH 命令、Web Shell，以及全局滚动条。以后新页面沿用滚动条和互斥审核。

- 已通过的命令执行时选择多台在线机器，同一条命令分别下发。未通过不能执行。
- 同一条命令上「通过」和「拒绝」不同时出现。未通过只显示通过；已通过只显示拒绝。
- 机器列表展示已采集的系统、CPU、内存、磁盘。机器详情用宽抽屉 `sm:max-w-3xl`，字段两列。其他详情仍用窄抽屉。
- 在线机器且具备 `ssh:exec` 时提供登录，进入 `/agents/:uid/shell`。离线不提供登录。同一台机器同时只开一个终端。
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

## 2026-10-25 数据源查询模板、上次语句与查询历史

适用：`/datasources` 查询 Tab 的全部驱动分支（指标 / Loki / Vlogs / Elasticsearch / TRACE 代理）。

- 模板与历史入口放进编辑器 label 行右侧（`FormField` 的 `actions` 插槽），不新增常驻工具条，结果区高度零变化。label 行允许换行，窄屏只留图标按钮 + `aria-label`。
- 模板用 `Popover`（可搜索、可分组、能放语句预览），不用 `Select`（放不下预览和分组）也不用 `DropdownMenu`（会拦截搜索输入）。全量历史用右侧 `Sheet`（与详情/元数据同一套语言），最近 8 条用 `Popover`。
- 选中模板只填入编辑器、不自动执行；选中历史条目回填参数并立即重新执行。两者都复用面板既有的 `runQuery`，不新增执行入口。
- 缓存与历史只写本机 `localStorage`，键为 `alioth:q:v1:*`，按用户 + 租户 + 数据源 + 方言隔离；不恢复自定义绝对时间，一律回落 `1h`。
- 行内次级操作（仅填充 / 复制 / 删除）用 `opacity-0 group-hover:opacity-100 group-focus-within:opacity-100`，禁止 `hidden`，保证键盘可达；列表支持 ↑/↓/Home/End。
