# 角色管理重构 · 视觉与交互验收标准（Finish Gate）

**审查对象**：`docs/rbac-role-redesign.md` 设计契约 + 现有实现（`web/src/pages/role/`、`web/src/pages/member/`、`web/src/api/role/`、`web/src/components/ui/`、`app/alioth/pkg/perm/perm.go`、`api/alioth/v1/identity.proto`）
**审查方式**：代码与设计 token 实测（无浏览器会话，见 §9「验证限制」）
**结论**：**HOLD（针对设计契约）** — 契约本身有 12 处必须先修正的问题，修正后按 §7 清单逐项验收才可 PASS。

> 一句话：这个方案的产品判断是对的（可读、可查、可批量、备注贯穿），但**双语标注的实现路径、Tag 的对比度、系统角色备注的入口、成员分配里停用角色的处理**这四处是硬伤，照着现在的文档实现会返工。

---

## 0. 必须先修正的 12 项契约问题

| # | 问题 | 证据 | 影响 | 修法（本文件后续按此验收） |
| --- | --- | --- | --- | --- |
| C1 | `RoleTypeTag` 系统角色用 `tone="primary"`，12px 文字对比度不足 | `#1677ff` 叠在 `primary/10` 上：浅色 **3.46:1**、深色 **3.87:1**（`index.css:67`，`status-badge.tsx:27`）；WCAG AA 小字要求 4.5:1 | 系统角色标签在深浅色下都不达标 | 系统角色改用 `info` tone（sky）：浅 **5.17:1**、深 **9.25:1**；自定义角色用 `neutral`（浅 **4.50:1**，刚好达标，建议升到 `#475569`=**7.17:1**）。或新增 `primaryStrong` token（`#0958d9`/`#4096ff`，5.20/5.31） |
| C2 | 双语标注用 `t()` 实现不了「两种语言同屏」 | `t()` 只返回当前语言（`i18n/index.ts:2370-2377`），`permission.resource.*` 是单语键 | 切到英文后中文标注消失，需求 2 直接不成立 | 权限字典放 `lib/permission-catalog.ts` 的 `{ zh, en }` 字面量（**不放 i18n**），UI 文案仍走 i18n；需要另一语言时用 `i18n.getFixedT("en")` |
| C3 | 系统角色「备注可改」没有入口 | §3.1 规定系统角色只有「详情 / 权限」，§3.3 又要求系统角色可改备注 | 4 个系统角色的备注（正是 tooltip 最有价值的数据）无处录入 | 系统角色行内保持「详情 + 权限」两个内联操作；备注在**详情 Sheet** 里用「编辑备注」打开同一表单（名称只读）；不把「编辑」放进列表 |
| C4 | 成员分配把停用角色过滤掉，但文案里又定义了 `member.roleDisabled` | §9.1 要求 `status === "ACTIVE"`，§10.3 又给「已停用」文案 | 成员身上已有停用角色时，勾选状态与真实数据不一致，保存后角色被静默移除 | 二选一并写进文档：**(a)** 保留已分配的停用角色（标记已停用、可取消、不可新增）；**(b)** 保存时保留原有停用角色。推荐 (a) |
| C5 | 详情里的分组计数 `2/2`、`2/3` 语义错 | §3.2 线框图：详情用「已选/总数」 | 只读视图里「未选」没有行动意义，用户会以为漏配 | 详情组头显示 **`N 项`**，区块标题显示「权限（12 项）」；`x/y` 只用于编辑器 |
| C6 | 编辑器过滤掉 owner 专属权限后，详情会丢权限 | `role/index.tsx:81-83` 在**加载时**就过滤了 `namespace:delete/transfer`；owner 的 `permissions` 里包含这两项 | owner 详情要么少 2 项，要么组头出现 `4/2` | 保留两份目录：`displayCatalog`（22 项，详情用）与 `assignableCatalog`（20 项，编辑器用）；过滤只发生在编辑器 |
| C7 | 备注为空时 `?` 图标 `disabled` | §9.2 | disabled 按钮不可聚焦、不触发 hover/focus，tooltip 永远看不到（a11y 反模式） | 空备注：图标保留且**可用**，tooltip 文案「暂无备注」；或隐藏图标并用不可见占位保持对齐（二选一，推荐前者） |
| C8 | `< sm` 行内备注兜底覆盖不了触屏平板 | §9.2 只处理 `<sm`；iPad 768–1024 命中 `hover: none` 但不是窄屏 | 平板用户既没有 hover，也没有行内备注，备注不可达 | 增加 `@media (hover: none)` 分支：行内备注或点击展开；验收用 768×1024 + `hasTouch` 设备模拟 |
| C9 | 备注列 `w-[24rem]` 过贪 | §3.1 | 与 type/status/actions 合计 44.5rem 固定宽，5 列在 1024–1280 屏上被挤压或横滚 | 备注列 **18rem**（`TABLE_COLUMN_WIDTH.remark`），`≥lg` 显示；表格最小宽用 `TABLE_MIN_WIDTH.compact`(56rem) |
| C10 | 编辑器的 footer/toolbar 用 `sticky` 不可靠 | §3.4 写「sticky top-0 / sticky bottom-0」，但容器是 `flex-col` + 主体 `overflow-y-auto` | sticky 相对的是滚动容器，footer 在滚动容器外时 sticky 不生效；在容器内则会被长列表推走 | 三段式改为 flex 兄弟：`header shrink-0` + `toolbar shrink-0` + `body min-h-0 flex-1 overflow-y-auto` + `footer shrink-0`；不依赖 sticky |
| C11 | 权限 Sheet 在 390px 只有 292px 宽 | `sheet.tsx:63`：`data-[side=right]:w-3/4`，`sm:max-w-3xl` 只封顶不收窄 | 分组名 + 双语 + 计数 + 勾选框挤在 292px，必然换行/溢出 | `<sm` 用整宽（`w-full`，保留 1rem 圆角/边距），`sm+` 用 `w-3/4 + max-w-3xl` |
| C12 | 组头整行可点 + 内含 Checkbox；行 `<label>` 内含 `?` 按钮 | §3.4 / §9.2 | 嵌套交互元素：点 checkbox 触发折叠；点 `?` 触发勾选；键盘出现两个焦点 | 组头 = `CollapsibleTrigger`（按钮，只包标题区）+ 独立 Checkbox 兄弟节点；角色行 = 容器 div，`label` 与 `?` 按钮平级，`?` 用 `onClick` 且不嵌进 label |

---

## 1. 产品镜头（Lens）

租户管理员在「角色」页要做三件事：**看懂一个角色能做什么**（审计/交接）、**配出一个新角色**（授权）、**把角色交给对的人**（成员分配）。最高频动作是「扫一眼角色名 + 类型 + 备注 → 打开权限看差异 → 勾选保存」，低频高风险动作是「删除/停用角色」和「给 owner 之外的人放大权限」。所以：角色的**类型**（系统/自定义）是权限边界的第一信号，**备注**是「这个角色为什么存在」的唯一解释，**权限列表**必须能被非研发读懂（中英双语 + 模块分组），**授权器**必须让人在 20 项里 10 秒内找到并确认改动，而不是在裸编码里逐条猜。数据密集表格是默认形态，不做卡片画廊、不做装饰性渐变、不做自动保存。

---

## 2. 设计契约

| 项 | 决定 |
| --- | --- |
| 用户 + 任务 | 租户管理员 / owner：审计角色权限、配置自定义角色、把角色分配给成员 |
| 首读对象 | 列表：**角色名 + 类型 Tag**；详情：**权限分组**；授权器：**已选/总数 + 分组** |
| 主操作 | 列表：打开「权限」；授权器：**保存修改**（唯一主按钮） |
| 密度 | 紧凑（行高 40px，表头 40px，20 行/页）；备注单行截断 + tooltip；权限编辑器行高 44px |
| 层级 | 页面标题 → 工具栏（搜索/新建）→ 列表 → 行操作；详情内：名称/类型/状态 → 备注 → 权限分组 → UID |
| 交互模型 | 列表表格 + 右侧 Sheet（详情/授权，`sm:max-w-3xl`）+ Dialog（短表单）；**不使用**卡片网格、时间线、看板 |
| 响应式 | 表格横向滚动留在 `Table` 自己的滚动视口；`<lg` 隐藏备注列、名称下第二行；`<sm` Sheet 整宽、权限行单列、工具条换行 |
| 参考（模式，非视觉） | ① GitHub 组织角色页：角色=权限集合，类型与权限分离；② AWS IAM Policy 摘要：按 service 分组、动作用词；③ Linear 团队权限：勾选即预览影响；④ Radix/shadcn 三态组勾选；⑤ 现有 `alert` 页多选器：搜索 + 全选 + 已选计数 |
| 禁止的默认 | 类型用红/绿（会与「启用/停用」状态色冲突）；权限行用彩虹色；未保存自动提交；用 `title` 属性代替可访问 tooltip；用「(系统角色)」拼在名称后；权限模板/从角色复制；拖拽排序；表格内联编辑 |
| 完成证据 | 本文 §7 清单全勾 + 1366×768 / 390×844 / 768×1024(touch) / 深浅色 / 中英双语 / 纯键盘 全量截图 + `pnpm typecheck && pnpm lint && pnpm build` |

---

## 3. 逐条需求验收标准

### R1 角色详情展示权限列表

**视觉**
- 权限区块在「备注」之后、「UID」之前，占满 Sheet 整行（`DetailSheet` 增加 `span`，不能用两列网格把它压成半宽）。
- 区块标题：`权限（12 项）`；分组头：`成员 · Members` + 右侧 `2 项`；行：`读取 / Read` + mono `member:read`。
- 只读，**不出现勾选框**；系统角色与无 `role:write` 用户不显示「配置权限」。
- 系统角色显示一行说明：`系统角色的权限由平台预置，只能查看。`

**交互**
- 打开详情时焦点落在 Sheet 标题（Radix 默认），`Esc` 关闭后焦点回到触发行操作，不落到 `body`。
- 区块内部滚动；标题与名称/类型/状态固定；滚动到底能看到 UID。
- 详情从列表数据渲染（无 loading），但保存权限后若详情仍开着，必须刷新或关闭，禁止展示旧计数。

**文案**
- 空权限（兜底）：`该角色还没有权限` / `This role has no permissions yet`；不渲染空的分组壳。
- 目录加载失败：`权限名称暂时无法加载，以下显示权限编码` + `重试`。

**边界态（逐项验证）**
1. owner：**22 项**，`namespace:delete/transfer` 必须可见（见 C6）。
2. viewer：9 项；admin：20 项；自定义角色：实际项数。
3. 未知编码（前端字典缺失）：落「未分类 · Uncategorized」组，显示原始 code，不空白。
4. 900px 高、1366 宽：全部可达，无内容超出 Sheet 底部（`detail-sheet.tsx:44` 的 `overflow` 只在 `wide` 时存在）。
5. 中/英切换：分组头语言顺序翻转，两种语言同时可见。
6. 200% 缩放：分组头与计数不重叠，长模块名换行不溢出。

**判定**：1–6 全过 + 只读态无误 + 无「(系统角色)」这类拼接文案。

---

### R2 权限按模块 + 动作分组并中英双语

**分组规则**
- 分组键 = `resource`；组顺序 = `perm.Catalog()` 顺序：member、role、namespace、datasource、alert_rule、alert、notification、agent、ssh；未知 resource 追加到末尾「未分类」。
- 组内动作顺序：`read → write → intervene → exec → delete → transfer`，未知动作排最后（`perm.go:63-79` 的目录顺序作兜底）。
- 计数分母：编辑器 = 可分配 20 项；详情 = 目录 22 项（按 C5 显示 `N 项`）。

**双语格式（必须同时可见）**

| 位置 | 中文界面 | 英文界面 |
| --- | --- | --- |
| 组头 | `成员 · Members` | `Members · 成员` |
| 行主标 | `读取 / Read` | `Read / 读取` |
| 编码 | `member:read`（mono，`aria-hidden`） | 同左 |

**字典（唯一来源：`lib/permission-catalog.ts`，与 `perm.go` 对齐）**

| resource | zh | en | 动作 |
| --- | --- | --- | --- |
| member | 成员 | Members | read, write |
| role | 角色 | Roles | read, write |
| namespace | 租户 | Namespaces | read, write（+ owner-only: delete, transfer） |
| datasource | 数据源 | Datasources | read, write |
| alert_rule | 告警规则 | Alert rules | read, write |
| alert | 告警 | Alerts | read, write, intervene |
| notification | 通知 | Notifications | read, write |
| agent | 机器 | Agents | read, write |
| ssh | SSH 命令 | SSH commands | read, write, exec |

动作：`read 读取/Read`、`write 写入/Write`、`intervene 介入/Intervene`、`exec 执行/Execute`、`delete 删除/Delete`、`transfer 移交负责人/Transfer ownership`。
> 文案口径必须先定死：`intervene`、`transfer` 已在 `action.*` 里翻成「介入」「移交负责人」（`i18n/index.ts:62,52`），不要在这套字典里改成「干预」「移交」，否则同一权限在两个界面两种说法。`write` 若改「管理/Manage」，必须同步搜索占位符与文档。

**搜索匹配**：中文名、English 名、`resource:action` 全码、resource、action 任一命中，大小写不敏感，`trim`。

**边界态**
- 空字典 / 缺键：显示 code，禁止空白（`perm.go` 加模块时前端不能炸）。
- 长英文（`SSH commands`、`Notifications`）：组头允许两行，计数不与文字重叠。
- 100+ 项：分组仍按目录顺序，不允许乱序（后端返回的 `permissions` 是 DB link 顺序，必须前端排序，见 `role_repo.go:258-292`）。
- 语言切换：不改动选中状态、不改动滚动位置。

**判定**：任意权限在两种语言下都能被「中文、英文、code」三种方式检索到；组头/行/编码三件套齐全。

---

### R3 授权交互重设计（核心）

**交互模型**：右侧 Sheet，三段式，非模态表单堆叠。

1. **头部**：`配置权限` + 副标题 `角色名 + 类型 Tag + 备注一行`；右侧统计 `已选 12 / 20`，脏值时追加 `新增 2 · 移除 1`。
2. **工具条**（不滚动）：搜索（含 `Search` 图标）、`仅看已选 (n)`、`展开全部/收起全部`、`全选`、`清空`（5 秒内可撤销）。
3. **主体**（唯一滚动区）：分组卡，组头 = 展开箭头 + 标题 + 计数 + 三态 Checkbox；组内行 = 整行 label + Checkbox + `动作中文 / English` + 右对齐 mono code（`≥lg` 显示）。
4. **页脚**：左说明（owner-only 权限不可授予），右 `取消` + `保存修改`。

**必须实现的交互（逐条判定）**
- 单点：点行任意处 / 空格键切换；组头三态即时更新；页脚计数即时更新。
- 组三态：`indeterminate` → 全选 → 全不选；`aria-checked="mixed"`，Indicator 需补 `MinusIcon`（`checkbox.tsx:19-25` 目前只有 `CheckIcon`）。
- 折叠：状态记忆在 `Set<resource>`；搜索或「仅看已选」时强制展开且组头不可折叠。
- 搜索：实时过滤；命中组自动展开；无命中显示 `没有匹配的权限` + `清空搜索`；**筛选不得清空已选**。
- 全选：只作用于**当前可见集合**；搜索激活时按钮文案变 `全选搜索结果（n）`。
- 清空：清空全部（不看筛选）；按钮变「撤销」，5 秒内可恢复快照；期间再编辑则撤销失效。
- 脏值：`dirty = baseline △ selected`；`Esc` / 遮罩 / ✕ / `取消` 四条路径都要弹「放弃未保存的修改？」；`⌘/Ctrl+S` 仅在 `dirty && !saving` 时保存并阻止浏览器默认保存。
- 保存：成功后关闭 Sheet + toast `权限已更新` + `reload()` + **焦点回到触发行操作**；失败时保留选择与筛选，错误显示在页脚上方，可重试。
- 只读态（系统角色 / 无 `role:write`）：无 Checkbox（用 Check 图标）、无批量按钮、页脚只有 `关闭`、顶部一行 `系统角色的权限由平台预置，只能查看。`；列表里仍然给「权限」入口。
- 0 选中：保存禁用 + `至少选择一个权限，否则该角色看不到任何资源。`（对齐后端 `min_items = 1`，`identity.proto:1103`）。

**状态矩阵（实现后逐个截图）**

| 状态 | 头部 | 主体 | 页脚 |
| --- | --- | --- | --- |
| 目录加载中 | 骨架 | 3 个分组骨架 | 保存禁用 |
| 目录加载失败 | 正常 | 空态 + 重试 | 保存禁用 |
| 搜索无命中 | 已选数不变 | 空态 + 清空搜索 | 正常（可取消关闭） |
| 仅看已选为空 | — | 空态 | 正常 |
| 0 选中 | `已选 0 / 20` | 提示文案 | 保存禁用 |
| dirty | `新增 n · 移除 m` | 正常 | 保存可用 |
| saving | 保存按钮 loading | 只读 | 取消禁用 |
| failed | 错误条 | 保留选择 | 保存可重试 |
| 只读（系统角色） | 说明行 | Check 图标 | 只有关闭 |

**键盘 / a11y**
- Tab 顺序：搜索 → 工具条按钮 → 组头勾选 → 组内勾选 → 页脚；不允许出现「整行按钮里套 button」的双焦点。
- 焦点环：`focus-visible:ring-2 ring-ring ring-offset-2`（与 `Button` 一致）；不许只靠背景色变化。
- 组头 Checkbox 的 `aria-label` 含模块名（`全选 成员`），`aria-controls` 指向组内容区。
- code 的 span 加 `aria-hidden`，避免屏幕阅读器连读 `member:read`。
- 折叠动画 `duration-150`，`motion-reduce:transition-none`。

**边界态**
- 100+ 项：主体可滚动，头部/工具条/页脚固定；搜索输入不卡顿；不需要虚拟滚动，但需要 `content-visibility` 或分组折叠作为降级手段时须在文档说明。
- 超长中英文：行高 44px 允许两行，Code 在窄屏隐藏，禁止把动作文案截断成 `读取 / R…`。
- 390×844：Sheet 整宽、单列、工具条**换行**（不允许横向滚动条），页脚按钮不重叠。
- 保存前切租户/刷新：不做特殊处理，但 `useListQuery`/`reload` 不得把筛选状态清空。

**判定**：上表 9 个状态 + 键盘全流程 + 脏值四条退出路径，全部有截图。

---

### R4 备注字段（remark）

**数据/接口**
- `RoleItem.remark`（proto 7）、`CreateRoleRequest.remark`（`max_len=200`）、独立 `PUT /v1/role/{uid}/remark`；ent `MaxLen(200).Default("")`；系统角色可写备注、不可改名（`biz/role/role.go:40-53` 已拒绝改名）。
- 4 个系统角色 seed 默认备注，仅在空值时补齐（老租户可用）。
- 空备注写 `''`，不写 `null`。

**视觉/交互**
- 新建/编辑 Dialog：名称（系统角色只读 + `系统角色不能改名`）+ 备注 `Textarea` 3 行 `resize-none` + 右下 `0/200` 计数；超过 200 实时红框 + `备注最多 200 字`，保存禁用。
- 计数按 `Array.from(value).length`（emoji 不被算成 2）。
- 列表备注列：单行 `truncate`，空值 `—`，hover/focus 显示全文 tooltip；`<lg` 移到名称下第二行（`line-clamp-1`）。
- 详情：`whitespace-pre-wrap` 保留换行，`break-words`。
- 授权 Sheet 副标题：备注 `line-clamp-2`，帮助确认「这个角色为什么存在」。
- 保存：自定义角色并行 `updateRole` + `updateRoleRemark`；系统角色只发 remark；部分失败回填对应字段错误，不清空面板。

**边界态**
- 200 字中文 / 200 字符英文 / 带 emoji / 带换行 / 首尾空格（trim）各一条。
- 备注含 `{{ }}` 不被 i18n 插值吞掉（备注是数据，不经过 `t()`）。
- 空备注：列表 `—`；成员分配 tooltip `暂无备注`；详情不留空白行。
- 名称 64 字上限（`identity.proto:1002`）：前端 `maxLength=64` + 计数，错误提示字段级，不靠后端 500。

**判定**：列表/详情/编辑/授权头部/成员 tooltip 五处显示同一份数据；系统角色备注可改、名称不可改。

---

### R5 列表独立 Tag 列区分系统/自定义角色

**视觉**
- 独立「类型」列，宽 `7rem`，`whitespace-nowrap`；名称列**删除** `角色名 (系统角色)`（`role/index.tsx:353`）。
- 系统角色：`info` tone（sky）+ 锁图标`size-3`（`aria-hidden`），文案 `系统角色` / `System role`。
- 自定义角色：`neutral` tone（灰底、`text-muted-foreground`），文案 `自定义角色` / `Custom role`。
- 高度 20px、`px-2`、`text-xs`、`font-normal`、`rounded-4xl`，与 `StatusBadge` 完全同形（`badge.tsx:7`）。
- 浅色：系统 `#0369a1` on `sky-500/10` = **5.17:1**；自定义 `#64748b` on `muted/60` = **4.50:1**（临界，建议换 `#475569` = **7.17:1**）。
- 深色：系统 `#7dd3fc` = **9.25:1**；自定义 `#94a3b8` = **6.21:1**。

**交互**
- Tag 不是按钮，不进 Tab 顺序；不要把说明塞成 Tag 上的 tooltip（每行多一个 Tab 停靠点，且与备注 `?` 冲突）；说明放详情与只读授权页。
- 类型列不排序（本次不做），列头不加排序箭头，避免暗示。

**边界/响应**
- 名称 64 字：名称列 `break-words` 换行，类型列不被挤变形（不能 `truncate` 整个名称列）。
- `<sm`：仍保留类型列（Badge 形态），表格在自身视口横滚；禁止退化成名称后的字符串。
- 中英切换：Tag 文案随语言变化，宽度 7rem 容得下 `System role`。
- 深浅色：Tag 不透明背景导致层级看起来像按钮时，降低到 `bg-*/10`（沿用 toneClasses）。

**判定**：4 个系统角色 + 3 个自定义角色同屏，类型一眼可分；名称列不再出现拼接文案。

---

### R6 成员分配角色时 `?` 图标 + 备注 tooltip

**数据**
- 成员页角色来源由 `selectRoles`（只有 uid/name，`identity.proto:434-439`）改为 `listRoles({page:1,pageSize:100})`，拿到 `isSystem/remark/status`。
- 保留「排除 owner」逻辑，但用 `isSystem && name === "owner"` 判定（名称 `owner` 已被唯一索引占用，`ent/schema/role.go:31`，自定义同名角色不存在）。
- 停用角色：按 C4 决定处理，**已分配的停用角色必须可见**。
- 已知限制：>100 角色不可见，必须写进代码注释与文档。

**视觉**
- 行结构：`[Checkbox] 角色名 [类型 Tag] [?]`；名称 2 行内换行，`?` 与 Checkbox 首行对齐。
- `?` = `CircleHelp` `size-4`，`text-muted-foreground`，hover `text-foreground`，**点击热区 ≥24×24px**（WCAG 2.2 target size）。
- Tooltip：沿用 `ui/tooltip`（`bg-foreground text-background`，`text-xs`，`px-3 py-1.5`，`rounded-md`，带箭头），浅色 17.06:1、深色 15.27:1。
- 局部 `TooltipProvider delayDuration={300}`（全局是 0，`tooltip.tsx:7`，扫读时瞬间弹会很吵）。
- 内容：`备注`（12px 标签）+ 备注正文（`whitespace-pre-wrap break-words`），`max-w-xs`；不截断（200 字上限约 8 行，可接受），但要验证 768 高下不超出视口。
- 窄屏（`<sm` 或 `hover: none`）：不依赖 hover，行内第二行 `text-xs text-muted-foreground line-clamp-2`；空备注不渲染该行。

**交互 / a11y**
- hover 300ms 打开；移开关闭；Tab 聚焦打开、Tab 移开关闭；`Esc` 关闭且焦点留在 `?` 上。
- `?` 是 `<button type="button">`，`aria-label="{角色名} 的备注"`；打开时 Radix 自动挂 `aria-describedby` 指向 tooltip；tooltip `role="tooltip"`。
- 点击 `?` 绝对不能切换勾选（不得嵌在行 `<label>` 内，见 C12）。
- Tab 顺序：`?` 在 Checkbox 之后、下一个角色之前；不要把 `?` 放进 `<label>`。
- 只读/无写权限用户：可以看 tooltip，只是不能勾选。
- 备注为空：图标可用 + tooltip「暂无备注」，或隐藏图标（二选一，禁止 disabled）。

**边界态（必须逐项验证）**
1. 390×844 触屏：行内备注可见；点 `?` 不出现死路。
2. 768×1024 触屏（`hover: none`）：按 C8 有可达路径。
3. iPad/触屏点击后滚动列表：tooltip 不被裁切、不残留。
4. 最后一行：tooltip 自动翻转（`side="right"` 越界时 Radix collision 处理）。
5. Dialog 打开时 tooltip 正常显示，且不会导致 Dialog 关闭。
6. 200 字 + 换行备注：tooltip 不溢出、不遮住 `?`。
7. 切换语言：tooltip 内「备注」标签随语言变化，备注正文不变。

**判定**：鼠标、键盘、触屏三条路径都能读到备注；空备注、超长备注、最后一行均无死角。

---

## 4. 表格列宽与信息密度建议

| 列 | ≥lg（1024+） | sm–lg | <sm | 规则 |
| --- | --- | --- | --- | --- |
| 名称 | 弹性 `minmax(0,1fr)`，最小 14rem | 同左 | 同左 | `break-words`，**不 truncate**；64 字换行 |
| 类型 | `w-[7rem]` nowrap | 同左 | 同左 | `RoleTypeTag` |
| 备注 | `w-[18rem]` nowrap，单行 `truncate` + tooltip | 隐藏，名称下第二行 `line-clamp-1` | 同 sm | `TABLE_COLUMN_WIDTH.remark` 新增字面量 |
| 状态 | `w-[6.5rem]`（`TableStatusHead`） | 同左 | 同左 | `StatusBadge` |
| 操作 | `w-[7rem]`（`TableActionsHead` + `RowActions nowrap`） | 同左 | 同左 | 系统角色：详情+权限；自定义：详情+更多 |

- 表格最小宽：`TABLE_MIN_WIDTH.compact`（56rem）。1440 内容区：名称约 408px；1280：约 300px；1024：约 220px（够 14 字中文 + 换行）。
- 行高：表头 40px（`TableHead h-10`），单元格 `p-2` → 40px；备注列不得把行高撑到 48px 以上。
- 每页 20 行、无斑马纹、hover `bg-muted/50`；字段空值统一 `—`。
- 表格横滚只发生在 `Table` 的 `overflow-auto` 视口内（`table.tsx:10`），页面 `body` 不得出现横向滚动条（`index.css:158`）。
- 备注列 `w-[18rem]` 约等于 24 个中文字或 34 个英文字符；超过即截断，**必须**能用键盘触达全文 tooltip（触发元素是单元格文本，`tabIndex=0` + `aria-describedby`；或直接在详情里看全文）。
- 密度取舍：不加「权限数」列（详情已给），不加「创建时间」列（角色数量少，不是决策依据）；如需扫读覆盖度，可在类型 Tag 后追加 `12 项` 的次要文本，但默认不做。

---

## 5. Tag 与 Tooltip 视觉规范

### 5.1 类型 Tag

| 维度 | 系统角色 | 自定义角色 |
| --- | --- | --- |
| 形状 | `Badge variant="outline"`，`h-5 px-2 text-xs font-normal rounded-4xl` | 同左 |
| 浅色 | `border-sky-500/30 bg-sky-500/10 text-sky-700`（5.17:1） | `border-border bg-muted/60 text-muted-foreground`（4.50:1，临界）→ 建议 `text-slate-600`（7.17:1） |
| 深色 | `dark:text-sky-300`（9.25:1） | `dark:text-slate-400`（6.21:1） |
| 图标 | `Lock size-3`（`aria-hidden`），与文字 `gap-1` | 无 |
| 文案 | 系统角色 / System role | 自定义角色 / Custom role |
| 语义 | 「不可在此改名/删除」的**约束**，不是状态 | 「管理员自建」的**来源** |
| 禁止 | 红/绿/琥珀（与启用停用状态冲突） | 蓝 `primary`（12px 只有 3.46:1）、紫渐变 |

间距：Tag 与角色名 `gap-2`（8px）；操作列按钮 `gap-2`；备注列与状态列各留单元格 `p-2`，不再加额外 margin。

### 5.2 备注 Tooltip

| 维度 | 规格 |
| --- | --- |
| 容器 | `ui/tooltip`：`bg-foreground text-background`，`rounded-md`，`px-3 py-1.5`，`text-xs`，箭头 10px |
| 宽度 | `max-w-xs`（20rem）；内容 `whitespace-pre-wrap break-words` |
| 内容 | 第一行 `备注`（`font-medium`），正文 12px / 行高 1.5；空态 `暂无备注` |
| 触发 | hover 300ms、focus 立即；触屏按 C8 兜底 |
| 位置 | `side="right" align="start"`，越界自动翻转；`sideOffset=6` |
| 关闭 | 移开、失焦、`Esc`；点击外部（触屏兜底） |
| 动效 | 现有 `fade-in-0 zoom-in-95`，`motion-reduce` 下关闭动效 |
| 层级 | Portal 到 body（`z-50`），不被 Dialog/Sheet 的 `overflow-hidden` 裁切 |
| 禁止 | `title` 属性、纯 hover 无键盘、`disabled` 触发器、超过 8 行仍无限拉长（200 字是上限） |

---

## 6. 打回缺陷清单（会直接被退回的常见做法）

| # | 现象 | 违反 | 证据 | 修法 | 级别 |
| --- | --- | --- | --- | --- | --- |
| 1 | 系统角色 Tag 用 `primary` 蓝 | R5 / 对比度 | 3.46:1（浅）/3.87:1（深） | 改 `info` 或 `primaryStrong` | P0 |
| 2 | 双语用 `t()`，切语言后另一语言消失 | R2 | `i18n/index.ts:2370` 单语言资源 | 字典对象 + `getFixedT` | P0 |
| 3 | 详情沿用加载时过滤过的目录，owner 少 2 项权限 | R1 | `role/index.tsx:81-83` | 双目录（display/assignable） | P0 |
| 4 | 编辑器过滤逻辑和详情共用 `catalog`，组头出现 `4/2` | R1/R2 | 同上 | 分母分场景 | P0 |
| 5 | Tooltip 触发器 `disabled` | R6 a11y | disabled 不可聚焦 | 保留可用 + 空态文案 | P0 |
| 6 | 组头整行 `button` 里塞 `Checkbox`；行 `label` 里塞 `?` | R3/R6 | C12 | 拆平级节点 | P0 |
| 7 | 编辑器 footer 用 `sticky`，长列表时被推走/无效 | R3 | C10 | flex 兄弟 + `shrink-0` | P0 |
| 8 | 390px 下授权 Sheet 只有 292px 宽 | R3 响应式 | `sheet.tsx:63` `w-3/4` | `<sm` 整宽 | P1 |
| 9 | 成员分配过滤掉停用角色，成员已有停用角色凭空消失 | R6 数据一致性 | C4 | 保留已分配项 | P0 |
| 10 | 成员分配仍用 `selectRoles`，无 `remark/isSystem` | R6 | `member/index.tsx:89-96` | 换 `listRoles` | P0 |
| 11 | `?` 点击顺带勾选/取消勾选角色 | R6 | label 包 button 的默认行为 | 平级 + `preventDefault` | P0 |
| 12 | 备注列 `w-[24rem]` 导致 1280 屏横滚 | R5 密度 | C9 | 18rem + `≥lg` 才显示 | P1 |
| 13 | 名称列 `whitespace-nowrap`，64 字角色名把表格撑破 | R5 | `table.tsx:85` 默认 nowrap | 名称列 `break-words` + `min-w-0` | P1 |
| 14 | 搜索输入只有 placeholder，没有可访问名 | 全局 a11y | `role/index.tsx:302-307` | 加 `aria-label` | P2 |
| 15 | 空态复用了「还没有角色」，搜索无结果时也说「还没有角色」 | R5 边界 | `role/index.tsx:336-338` | 区分 no-data / no-match + 清空搜索 | P1 |
| 16 | 详情 `权限` 字段仍是 `join(", ")` | R1 | `role/index.tsx:487` | 分组只读列表 | P0 |
| 17 | 详情用非 `wide` Sheet，22 项时被裁切 | R1 | `detail-sheet.tsx:44` 仅 wide 有 overflow | 详情统一 `wide` + 内部滚动 | P0 |
| 18 | 详情里「系统角色: 系统角色」这种同义字段 | R5 文案 | `role/index.tsx:482-484` | 换成「类型: Tag」 | P2 |
| 19 | 角色状态借用 `namespace.status` 文案（列表 344 行、详情 480 行） | 文案一致性 | `role/index.tsx:344,480` | 新增 `role.status` | P2 |
| 20 | 保存后焦点丢失（菜单项已卸载） | R3/R6 键盘 | `row-actions.tsx:58-85` 菜单触发入口 | 关闭后焦点回到行操作 | P1 |
| 21 | 保存失败清空选择 / 关闭 Sheet | R3 | — | 保留选择 + 页脚错误 | P0 |
| 22 | 空目录静默变成空白列表 | R3 | `role/index.tsx:86-88` | 空态 + 重试 | P1 |
| 23 | 未保存直接关闭，改动丢失 | R3 | 现状无守卫 | 四条退出路径确认 | P1 |
| 24 | 权限顺序随数据库返回变化 | R2 | `role_repo.go:258-292` 未排序 | `ACTION_ORDER` 排序 | P1 |
| 25 | Tooltip 用全局 `delayDuration=0`，扫列表时疯狂弹 | R6 体验 | `tooltip.tsx:7` | 局部 300ms | P2 |
| 26 | `<sm` 之外没有触屏兜底（平板死角） | R6 | C8 | `hover: none` 分支 | P1 |
| 27 | 清空后「撤销」在再次编辑后仍恢复旧快照，覆盖新改动 | R3 | §8.2 未定义 | 编辑即失效撤销 | P2 |
| 28 | Tag 上再挂一个说明 tooltip，每行多一个 Tab 停靠点 | R5 a11y | §6 `withHint` | 说明放详情/只读页 | P2 |
| 29 | 角色名 64 字在分配弹窗里被截断，或把 `?` 挤出可见区 | R6 | `member/index.tsx:487` 无 wrap 约束 | 名称 2 行换行 + `?` 首行对齐 | P1 |
| 30 | `k` 计数用 i18n 复数硬拼 `item(s)` | R2 文案 | §10.1 | 用 i18next 复数或改无复数措辞 | P2 |

---

## 7. 验收清单（可勾选）

### 契约与数据
- [ ] C1–C12 全部落地，且设计文档同步更新（不再出现 `tone="primary"`、`sticky footer`、`w-[24rem]`、`sm 才兜底`）。
- [ ] `RoleItem.remark` 出现在 list/create/update/select 响应里；`PUT /v1/role/{uid}/remark` 对系统角色可用。
- [ ] 迁移后老租户 4 个系统角色备注已补；空备注为 `''`。
- [ ] `permission-catalog.ts` 双语文案与 `perm.go` 的 9 resource / 6 action 一一对应；未知 code 有兜底。
- [ ] 角色名 `maxLength=64`、备注 `maxLength=200`（按 code point），前后端一致。

### 列表（R5）
- [ ] 名称列不再出现 `(系统角色)`；类型列独立、`7rem`、nowrap。
- [ ] 4 系统 + N 自定义同屏可区分；中英切换文案正确。
- [ ] 64 字角色名换行，不横滚撑破列表卡片。
- [ ] 备注列 `≥lg` 显示、单行截断、空值 `—`、全文可达；`<lg` 名称下第二行。
- [ ] 空态区分「无角色」与「搜索无结果」，后者有「清空搜索」。
- [ ] 1440/1280/1024 宽下表格不出现页面级横滚；390 下表格在自身视口横滚且表头吸顶。

### 详情（R1/R2）
- [ ] 权限分组渲染，标题 `权限（N 项）`；组头 `成员 · Members` + `N 项`。
- [ ] owner 22 项（含 delete/transfer）、admin 20、viewer 9、自定义角色实际项数。
- [ ] 未知 code 落「未分类」，显示原始编码。
- [ ] 系统角色：只读说明，无「配置权限」；无 `role:write` 用户同样。
- [ ] 900px 高下内容可滚到底，头部固定，无裁切。
- [ ] 中英切换后两种语言同时可见，顺序随语言翻转。

### 授权器（R3）
- [ ] 搜索：中文 / English / `member:read` 三种都能命中；筛选不影响已选。
- [ ] 组三态：`mixed` 正确、Indicator 显示 `MinusIcon`、`aria-checked="mixed"`。
- [ ] 组折叠记忆；搜索/仅看已选时强制展开。
- [ ] 全选只作用于可见集；搜索时按钮文案带数量。
- [ ] 清空 + 5 秒撤销；撤销后再编辑不可恢复。
- [ ] `已选 a / b` 与 `新增 n · 移除 m` 实时正确（b=20）。
- [ ] `⌘/Ctrl+S` 保存并阻止浏览器保存；0 选中时禁用并给出原因。
- [ ] 脏值退出四条路径（Esc/遮罩/✕/取消）都弹确认；确认后不保存。
- [ ] 保存成功 toast + 列表刷新 + 焦点回到行操作；失败保留选择可重试。
- [ ] 只读态（系统角色/无写权限）：无勾选框、无批量按钮、只有关闭。
- [ ] 390 宽整宽单列、工具条换行；100+ 项时搜索/滚动不卡。
- [ ] 目录加载失败显示重试，不显示空白分类。

### 备注（R4）
- [ ] 新建/编辑写入备注，列表/详情/授权头部/成员 tooltip 五处一致。
- [ ] 200 字上限、emoji 计数正确、超限禁用保存并给字段错误。
- [ ] 系统角色备注可改、名称只读且有说明。
- [ ] 空备注 `—` / 「暂无备注」，详情不留空白。
- [ ] 换行在详情与 tooltip 中保留（`pre-wrap`）。

### 成员分配（R6）
- [ ] 角色来源 `listRoles`，含类型 Tag 与 `?`。
- [ ] 备注 tooltip：hover 300ms / focus / `Esc` / `aria-describedby` 全部通过。
- [ ] 点击 `?` 不改变勾选；`?` 热区 ≥24px 且有 focus ring。
- [ ] 空备注、200 字、换行、最后一行翻转四种情况无死角。
- [ ] 390 触屏 + 768 触屏（`hover: none`）都能读到备注。
- [ ] 停用角色按 C4 策略处理，已分配项不静默消失。
- [ ] >100 角色的限制写在代码注释与本文件。

### 无障碍与主题
- [ ] 纯键盘走完：列表 → 详情 → 授权 → 保存 → 回列表，焦点不丢、不出 Sheet。
- [ ] 所有新控件有可见 focus ring（`ring-2 ring-ring ring-offset-2`）。
- [ ] 浅色/深色下：Tag ≥4.5:1、tooltip ≥15:1、选中行左侧描边 ≥3:1。
- [ ] `prefers-reduced-motion` 下折叠/弹层无位移动效。
- [ ] 200% 缩放不重叠、不裁切。

### 构建与回归
- [ ] `cd web && pnpm typecheck && pnpm lint && pnpm build` 全绿。
- [ ] 真正在浏览器里跑过 1366×768、390×844、768×1024(touch)、深浅色、中英双语；截图留档。
- [ ] 本次未改动的共享组件（`Table`/`Dialog`/`Sheet`/`RowActions`/`StatusBadge`）无回归。

---

## 8. 判定规则

- **PASS**：§7 全部勾选，且 §6 中 P0/P1 全部关闭，浏览器证据齐全（含窄屏、触屏、深浅色、双语）。
- **HOLD**：任一 P0 未关闭，或任一「必须验证」的状态没有证据。不接受「构建通过 = 验收通过」，也不接受只看首屏截图。
- **不允许的通过理由**：「看起来更现代」「和别的后台一致」「组件库默认就是这样」「截图里没报错」。
- **可以保留的现状**：`StatusBadge` 统一色彩体系、`RowActions` 二段式（详情 + 更多）、`ListFrame` + `Table` 内部滚动、`FormField` 的 `?` hint 模式、zh/en 镜像 i18n、`DetailSheet` 复用。

---

## 9. 验证限制（如实声明）

- 本文件基于代码、设计 token 与实测对比度数值产出；**未**启动开发服务器、**未**在浏览器中实际点击验证（本次会话无登录态与浏览器验证条件）。§7 里所有视口/触屏/键盘项必须由实现在真实浏览器中执行，并把截图附在验收记录里。
- 对比度用 sRGB 相对亮度计算，底色取 `index.css` token 与 `bg-*/10` 叠加后的近似值（卡片 `rgba(255,255,255,.85)` 叠在 `#f8fafc` 上 ≈ `#fefeff`）。最终以浏览器取色为准，临界值（`neutral` 4.50）必须复核。
- 组件行为（Radix Tooltip 的 focus 展开、Sheet 内 Tooltip 的 portal、AlertDialog 与 Sheet 的叠层顺序）以实测为准；如与描述不符，按「必须验证」项失败处理。


---

## 变更记录

- 2026-10-25：备注上限由 200 字调整为 **5000 字**（ent `MaxLen(5000)`、proto `max_len=5000`、前端计数 `/5000`）。列表「备注」列改为**最多两行 `line-clamp-2`、超出省略、不挂 Tooltip**，全文只在角色详情里查看；成员分配弹窗的 `?` Tooltip 仍保留，但内容区限高可滚动。本文档中所有 200 字的旧验收项以此为准。
- 2026-10-25：权限目录细化为 **30 项 / 可分配 28 项**。8 个业务模块新增 `*:delete` 且删除操作改为校验 `*:delete`；`namespace:delete` / `namespace:transfer` 仍为 owner 专属，编辑器里展示但禁用。本文档中「22 项 / 20 项 / admin 20」等旧数字与「删除随 write」的假设全部作废：owner=30、admin=28、viewer=9、member=11。
