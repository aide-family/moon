# 角色管理（RBAC）重构设计规格

面向 Alioth 中后台「角色管理」模块的重构设计。目标是让权限可读、可查、可批量操作，并把「角色备注」变成贯穿列表、详情、编辑、成员分配的一等信息。

相关代码：`web/src/pages/role/`、`web/src/pages/member/`、`web/src/api/role/`、`app/alioth/pkg/perm/perm.go`、`app/alioth/internal/biz/role/`、`app/alioth/internal/data/role_repo.go`、`app/alioth/internal/data/ent/schema/role.go`。

---

## 1. 现状盘点

| 位置 | 现状 | 问题 |
| --- | --- | --- |
| 角色列表 | 名称列内联 `角色名 (系统角色)`；列：名称 / 状态 / 操作 | 系统与自定义混在名称里，不可扫读、不可按类型识别 |
| 角色列表 | 无备注列 | 角色用途只能靠名字猜 |
| 角色详情 `DetailSheet` | 权限字段是 `permissions.join(", ")` 的一整行 | 长串权限不可读，看不出模块分布，也没有中英文标注 |
| 授权弹窗 | `Dialog` 内一个平铺的 `Checkbox` 列表，标签是 `member:read` 原始编码 | 22 项无分组、无搜索、无全选、无计数；看不懂编码的人无法授权 |
| 授权弹窗 | 关闭即丢弃选择，无脏值提示 | 误关丢改动 |
| 系统角色 | 行内不显示「权限」操作（`roleActions` 对 `isSystem` 返回空） | 想确认系统角色到底有什么权限，只能看文档 |
| 成员分配角色 | `Dialog` 内平铺 Checkbox，只用 `name` | 同名/近义角色无法区分，不知道这个角色能做什么 |
| 数据来源 | 成员页用 `selectRoles`（`SelectItem{uid,name}`） | 拿不到 `isSystem` / 备注，无法做类型标识与 tooltip |
| 后端 | `RoleItem` 无备注字段 | 备注无处存、无处传 |

---

## 2. 信息架构

角色模块由 5 个界面组成，全部留在 `/roles` 单页，不新增路由：

```
角色管理 /roles
├─ 角色列表（主视图，Table + ListFrame）
│   └─ 行操作：详情 / 权限 / 编辑 / 停用·启用 / 删除
├─ 角色详情（Sheet，只读）
│   └─ 字段：名称 · 类型 · 状态 · 备注 · 权限（分组只读） · UID
├─ 新建角色（Dialog）
│   └─ 名称 + 备注
├─ 编辑角色（Dialog）
│   └─ 名称（系统角色只读）+ 备注
└─ 权限配置器（Sheet，右侧宽面板；系统角色为只读态）

成员管理 /members
├─ 分配角色（Dialog）
│   └─ 搜索 + 角色勾选清单（角色名 + 类型 Tag + 备注 ? 提示）
└─ 邀请成员（Dialog，沿用 Select，不做 tooltip）
```

权限编辑器与详情面板都用右侧 Sheet：这是「在同一上下文里看/改一个角色」的同一类任务，用一致的容器和宽度（`sm:max-w-3xl`），避免用户在两套弹层语法之间切换。

---

## 3. 页面布局（文字线框）

### 3.1 角色列表

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 角色                                                     [ 新建角色 ]          │
│ 管理角色及其权限。                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│ [🔍 按名称搜索角色            ] [搜索]        （P2：类型 Tabs 全部/系统/自定义）│
├──────────────────────────────────────────────────────────────────────────────┤
│ 共 12 条                                                        [上一页][1/1][下一页]│
│ ┌──────────────┬──────────┬──────────────────────────┬────────┬─────────────┐│
│ │ 名称          │ 类型      │ 备注                      │ 状态    │ 操作         ││
│ ├──────────────┼──────────┼──────────────────────────┼────────┼─────────────┤│
│ │ owner        │ 系统角色  │ 租户负责人，拥有全部权限…   │ 启用    │ 详情  ⌄     ││
│ │ admin        │ 系统角色  │ 可管理成员与配置，不能删除… │ 启用    │ 详情  ⌄     ││
│ │ 值班运维      │ 自定义    │ 处理告警与 SSH 执行        │ 启用    │ 详情  ⌄     ││
│ │ 只读审计      │ 自定义    │ —                        │ 停用    │ 详情  ⌄     ││
│ └──────────────┴──────────┴──────────────────────────┴────────┴─────────────┘│
└──────────────────────────────────────────────────────────────────────────────┘
```

列规范（复用 `TABLE_COLUMN_WIDTH`）：

| 列 | 宽度 | 内容 | 说明 |
| --- | --- | --- | --- |
| 名称 | 弹性 `minmax(0,1fr)` | 角色名，`text-sm font-medium` | 主列；不再拼接「(系统角色)」 |
| 类型 | `w-[7rem]`（`type`）+ nowrap | `RoleTypeTag` | 独立列，见 §6 |
| 备注 | `w-[24rem]`（新增 `remark`）+ nowrap | 单行 `truncate`，空值 `—` | 全文放 Tooltip；≥`lg` 才显示，窄屏降级为名称下方第二行 |
| 状态 | `w-[6.5rem]`（`status`） | `StatusBadge` | 沿用 |
| 操作 | `w-[7rem]`（`actions`） | `RowActions` | 沿用 |

表格最小宽度 `TABLE_MIN_WIDTH.default`（64rem）。

行操作变化：

- 自定义角色：详情 / 权限 / 编辑 / 停用·启用 / 删除（沿用 `RowActions` 的「详情 + 更多」二段式）。
- 系统角色：详情 / 权限（**新增，只读查看**）。其余写操作仍不出现（后端本来就拒绝）。

「类型」列只做**展示**。按类型筛选（Tabs / Select）需要 `ListRolesRequest` 增加 `is_system` 可选过滤并在 SQL 层生效，属于 P2，不在本次范围；不要用前端过滤假装实现——列表是服务端分页的。

### 3.2 角色详情（Sheet，`wide`）

```
┌──────────────────────── 角色详情 ──────────────────────────── ✕ ┐
│ 值班运维   [自定义]                                              │
│ 名称      值班运维                                               │
│ 类型      自定义角色                                             │
│ 状态      [启用]                                                 │
│ 备注      处理告警与 SSH 执行，默认分配给夜班同学。                │
│ ──────────────────────────────────────────────────────────────── │
│ 权限（12 项）                                       [配置权限]    │
│  成员 · Members            2/2                                   │
│    [读取 / Read] [写入 / Write]                                    │
│  告警 · Alerts             2/3                                   │
│    [读取 / Read] [干预 / Intervene]                                │
│  …                                                               │
│ UID      0f3a…                                                   │
└──────────────────────────────────────────────────────────────────┘
```

- 权限渲染为 `PermissionSummary`（只读分组）：分组头 = `中文模块 · 英文模块` + 右侧 `已选/总数`；每条是 Badge（`读取 / Read`），Badge 的 Tooltip 给完整形态 `成员 / 读取` + mono 编码 `member:read`。
- 权限项多时（> 1 屏）在分组容器内滚动，不把整个 Sheet 拉长；`PermissionSummary` 自带 `max-h-[24rem] overflow-y-auto`。
- 系统角色在此处右上角显示 `role.grant.readonly` 说明，不出现「配置权限」按钮。

### 3.3 新建 / 编辑角色（Dialog）

```
新建角色
┌──────────────────────────────────────────┐
│ 名称 *                                   │
│ [ 例如：值班运维                      ]   │
│ 备注 ⓘ                                   │
│ [ 说明这个角色能做什么、分配给谁…       ]  │
│ [                                    ]   │
│                                         │
│                              0/5000      │
│                          [ 取消 ] [创建 ]│
└──────────────────────────────────────────┘
```

- 备注用 `Textarea`（3 行，`resize-none`），右下角 `0/5000` 计数，超限按钮禁用 + 错误文案。
- 编辑态：自定义角色名称可改；系统角色名称 `disabled` 并附 `role.systemNameLocked` 说明（后端仍会拒绝改名，前端提前拦住）。备注对两者都可改。

### 3.4 权限配置器（Sheet，核心重构）

```
┌──────────────── 配置权限 ─ 值班运维 [自定义] ──────────────────── ✕ ┐
│ 处理告警与 SSH 执行，默认分配给夜班同学。                            │
│ 已选 12 / 20        新增 2 · 移除 1                                  │
├──────────────────────────────────────────────────────────────────────┤
│ [🔍 搜索权限：成员 / member / read    ]  [仅看已选 12] [展开全部] [全选] [清空]│
├──────────────────────────────────────────────────────────────────────┤
│ ▾ ☑ 成员 · Members                                        2/2       │
│    ☑ 读取 / Read        member:read                                  │
│    ☑ 写入 / Write       member:write                                 │
│ ▾ ◪ 告警 · Alerts                                          2/3       │
│    ☑ 读取 / Read        alert:read                                   │
│    ☑ 干预 / Intervene   alert:intervene                              │
│    ☐ 写入 / Write       alert:write                                  │
│ ▸ ☐ 通知 · Notifications                                   0/2       │
│ ▸ ☐ 机器 · Agents                                          0/2       │
│ ▸ ☐ SSH 命令 · SSH commands                                3/3       │
├──────────────────────────────────────────────────────────────────────┤
│ ⓘ 删除租户 / 移交负责人 只属于负责人角色，不可授予自定义角色。          │
│                                            [ 取消 ]  [ 保存修改 ]     │
└──────────────────────────────────────────────────────────────────────┘
```

布局：`SheetContent`（右侧，`data-[side=right]:sm:!max-w-3xl`）+ 三段式

1. **头部**：`SheetHeader` — 标题「配置权限」；第一行副标题 `角色名 + RoleTypeTag`，第二行角色备注（`text-xs text-muted-foreground line-clamp-2`，空备注不渲染该行，避免空行）；右侧统计行 `已选 a / b`，脏值存在时追加 `新增 n · 移除 m`。
2. **工具条**（`sticky top-0 bg-background/95 backdrop-blur`，`border-b`，高 44px）：搜索 Input（左带 `Search` 图标）、「仅看已选」Toggle（带数量）、「展开全部 / 收起全部」文字按钮、「全选」「清空」。
3. **主体**（`overflow-y-auto`）：分组卡片列表；每组一个 `Collapsible`。
4. **页脚**（`sticky bottom-0`，`border-t`）：左说明（负责人专属权限），右 `取消` + `保存修改`。

分组卡片规范：

- 组头高 44px，整行可点（展开/收起），右侧：`已选/总数` + 三态 `Checkbox`（选中显示 `CheckIcon`，部分选中显示 `MinusIcon`，`aria-checked="mixed"`）。组头左侧有 12px 的 `ChevronRight`，展开时 `rotate-90`，`transition-transform duration-150`。
- 组内行高 44px，`sm:grid-cols-2`、`grid-cols-1`，行本身是 `<label>`，点文字等于点勾选框（消除 16px 勾选框的点击精度问题）。
- 行内容三段：`动作中文`（`text-sm`）+ `动作英文`（`text-xs text-muted-foreground`）+ `模块:动作` mono 编码（`text-xs font-mono text-muted-foreground/70`，`ml-auto`，`hidden sm:inline`）。中文在前、英文其后，两者同权——这是需求 (2) 的「中英文同时标注」，编码只作为辅助信息。
- 已选行视觉：行底色 `bg-primary/5`，行左侧 2px `border-primary` 描边，保证在长列表里一眼看出已选区段；`dark` 下用 `bg-primary/10`。

### 3.5 成员分配角色（Dialog）

```
分配角色
成员 zhang@example.com
┌────────────────────────────────────────────────────────┐
│ [🔍 搜索角色                    ]                       │
├────────────────────────────────────────────────────────┤
│ ☑ 值班运维  [自定义]  ⓘ⑤                               │
│ ☐ 只读审计  [自定义]  ⓘ                                 │
│ ☑ admin     [系统]   ⓘ                                 │
│ ☐ member    [系统]   ⓘ                                 │
└────────────────────────────────────────────────────────┘
  已选 2 个角色                        [ 取消 ] [ 保存 ]
```

- 列表 `max-h-[50vh] overflow-y-auto`，行高 44px，`:hover` 背景 `bg-muted/50`。
- `ⓘ` 是 `CircleHelp` 图标（`size-4 text-muted-foreground`），触发器是 `<button type="button" aria-label="{{role}} 的备注">`，与 `FormField` 的 hint 图标完全一致。
- 窄屏（`< sm`）不依赖 hover：行内第二行直接渲染 `text-xs text-muted-foreground line-clamp-2` 的备注，图标隐藏。保证触屏可用。

---

## 4. 组件层级与 shadcn 选型

```
RolePage
├─ PageHeader                     (既有)
├─ 搜索 form: Input + Button       (ui/input, ui/button)
├─ ListFrame                      (既有)
│  └─ Table / TableHeader / TableBody            (ui/table)
│     ├─ TableHead ×3（名称/类型/备注）+ TableStatusHead + TableActionsHead
│     └─ TableRow
│        ├─ TableCell: name
│        ├─ TableCell: RoleTypeTag                ← 新增 components/role-type-tag.tsx
│        ├─ TableCell: RoleRemarkCell             ← 新增（truncate + ui/tooltip）
│        ├─ TableCell: StatusBadge                (既有)
│        └─ TableCell: RowActions                 (既有)
├─ RoleDetailSheet                                 ← 复用 DetailSheet（新增字段 span）
│  └─ PermissionSummary                            ← 新增（只读分组）
├─ RoleFormDialog (create / edit 共用)             (ui/dialog + ui/textarea + FormField)
├─ PermissionEditor                                ← 新增 components/permission-editor.tsx
│  └─ Sheet (ui/sheet)
│     ├─ 工具条: Input / Button / Collapsible 控制
│     ├─ PermissionGroup (ui/collapsible + ui/checkbox + ui/label)
│     └─ 页脚 Button
└─ ConfirmAction (删除/停用/未保存离开)            (既有)

MemberPage
├─ AssignRolesDialog                               ← 从 member/index.tsx 抽出
│  └─ ui/dialog + ui/input + ui/checkbox + ui/label + ui/tooltip
└─ RoleRemarkHint                                  ← 新增 components/role-remark-hint.tsx
```

选型要点：

| 需求 | 组件 | 理由 |
| --- | --- | --- |
| 分组折叠 | `ui/collapsible` | Radix 原生维护 `aria-expanded` / `data-state`，不自己写开合状态的话动画和键盘都免费 |
| 三态组勾选 | `ui/checkbox` + 小幅增强 | Radix 支持 `checked="indeterminate"`、自动输出 `aria-checked="mixed"`；但仓库里的 Indicator 只渲染 `CheckIcon`，需要补一个 `MinusIcon` 分支（见 §11 批次 3） |
| 备注全文 | `ui/tooltip` | 与 `FormField` 的 hint 同源，视觉一致；`max-w-xs` 已内置 |
| 权限编辑器容器 | `ui/sheet` | 右侧宽面板，不遮挡列表上下文；比 Dialog 更适合长表单 |
| 新建/编辑 | `ui/dialog` | 短表单，聚焦强，沿用现有交互 |
| 计数/新增移除 | `ui/badge`（`variant="secondary"`） | 与 `StatusBadge` 同一套圆角与字号 |

不引入新依赖：全部使用 `web/src/components/ui/` 已有组件（`alert-dialog`、`badge`、`button`、`checkbox`、`collapsible`、`dialog`、`input`、`label`、`sheet`、`skeleton`、`textarea`、`tooltip`）。

---

## 5. 权限分组与中英映射表结构

### 5.1 数据形态

后端 `/v1/permissions` 返回 `PermissionItem{uid,name,resource,action}`，分组在前端完成（不改后端目录结构）：

```ts
type PermissionGroup = {
  resource: string            // 分组键
  items: PermissionItem[]     // 已按动作顺序排序
  selectedCount: number
}
```

纯函数放在 `web/src/lib/permission-catalog.ts`，便于单测：

- `ACTION_ORDER = ["read","write","intervene","exec","delete","transfer"]` — 组内顺序：读写在前，强动作在后。
- `groupPermissions(items)` — 按首次出现顺序排列模块（与 `perm.Catalog()` 一致），组内按 `ACTION_ORDER`，未识别动作排到最后。
- `permissionLabel(t, item)` — 返回 `{ module, moduleEn, action, actionEn, code }`，供上面两种形态拼装，避免各处自己拼 i18n key。
- `matchesPermission(t, item, query)` — 命中中文、英文、`resource`、`action`、`name` 任一即算命中（大小写不敏感）。
- `OWNER_ONLY = new Set(["namespace:delete","namespace:transfer"])` — 编辑器过滤掉；页面已有过滤逻辑，收敛到这里。

### 5.2 模块（resource）映射

| code | 中文 | English | 可分配动作 | 编辑器项数 |
| --- | --- | --- | --- | --- |
| `member` | 成员 | Members | read, write, delete | 3 |
| `role` | 角色 | Roles | read, write, delete | 3 |
| `namespace` | 租户 | Namespaces | read, write（delete/transfer 仅 owner，展示但禁用） | 2 |
| `datasource` | 数据源 | Datasources | read, write, delete | 3 |
| `alert_rule` | 告警规则 | Alert rules | read, write, delete | 3 |
| `alert` | 告警 | Alerts | read, write, delete, intervene | 4 |
| `notification` | 通知 | Notifications | read, write, delete | 3 |
| `agent` | 机器 | Agents | read, write, delete | 3 |
| `ssh` | SSH 命令 | SSH commands | read, write, delete, exec | 4 |
| — | 未分类 | Uncategorized | 兜底 | — |

合计：目录 30 项，编辑器可见 28 项（只把 `namespace:delete`、`namespace:transfer` 留成「仅负责人」禁用项）。
删除普通业务数据（成员、角色、数据源、告警规则/字典、告警页面、通知渠道/模板/组/订阅、机器、SSH 命令）是**独立可授权**的权限，不再和 `*:write` 绑在一起；删除租户与移交负责人仍然只属于 owner。

### 5.3 动作（action）映射

| code | 中文 | English |
| --- | --- | --- |
| `read` | 读取 | Read |
| `write` | 写入 | Write |
| `intervene` | 干预 | Intervene |
| `exec` | 执行 | Execute |
| `delete` | 删除 | Delete |
| `transfer` | 移交 | Transfer |

### 5.4 展示规则

权限标签有两种形态，取决于有没有「分组头」提供模块上下文：

| 形态 | 文案 | 用在哪 |
| --- | --- | --- |
| 完整 | `{{模块中文}} / {{动作中文}}`，en 下 `{{module}} / {{action}}` —— 即需求里的 `member:read → 成员 / 读取` | 无分组上下文的地方：Badge 的 Tooltip、跨页面的权限 chip、日志/审计文本、以及本文件的映射表 |
| 分组内 | `{{动作中文}} / {{动作英文}}`（如 `读取 / Read`） | 编辑器行、详情分组内——模块已由分组头 `成员 · Members` 给出，行内再重复会淹没动作信息 |

- 编码（`member:read`）始终以 mono 小字保留：编辑器行内右侧、详情 Badge 的 Tooltip 里。它是排障和对接后端的锚点，但不能是唯一标签。
- 分组头：`{{中文模块}} · {{英文模块}}`，例如 `成员 · Members`。
- 兜底：任一 `resource` / `action` 在字典里缺失时，直接显示原始 code（`perm.Catalog()` 新增模块而前端未同步时不至于空白），并在分组头追加 `role.unknownModule` 提示。
- 需求 (1) 的「权限列表」在详情页是 Badge 组，在编辑器里是勾选行——同一份 `groupPermissions` 输出，两处渲染不同壳子。
- 代码里出现的 `member:read → 成员 / 读取` 这一形式，落到字典是 `permission.resource.member` + `permission.action.read`，不写死成一条字符串（否则中英切换会失效）。

---

## 6. 系统角色 / 自定义角色 Tag 规范

```
RoleTypeTag
├─ isSystem = true  → Badge variant="outline" tone "primary"  文案：系统角色 / System role
└─ isSystem = false → Badge variant="outline" tone "neutral"  文案：自定义角色 / Custom role
```

| 维度 | 系统角色 | 自定义角色 |
| --- | --- | --- |
| 底色 | `border-primary/30 bg-primary/10 text-primary` | `border-border bg-muted/60 text-muted-foreground` |
| 文案（zh） | 系统角色 | 自定义角色 |
| 文案（en） | System role | Custom role |
| 尺寸 | `h-5 px-2 text-xs`，与 `StatusBadge` 完全一致 | 同左 |
| 变体 | `full` = 系统角色 / 自定义角色（列表、详情、授权面板）；`compact` = 系统 / 自定义（成员分配行，横向空间紧张） | 同左 |
| 位置 | 列表「类型」独立列；详情「类型」字段；授权面板副标题 | 同左 |
| 附加说明 | Tooltip：预置角色，权限由平台维护，不能改名或删除 | 无 |
| 权限可编辑 | 否（只读查看） | 是 |
| 可停用/删除 | 否 | 是 |

实现：`components/role-type-tag.tsx`，内部复用 `StatusBadge` 的思路（同一套 `toneClasses`），通过 props 传入 `isSystem` + 可选 `withHint`。**不要**用 `StatusBadge` 直接传 `isSystem`，因为它按枚举值映射颜色，布尔值不适合塞进枚举表。

列表里旧行为（`{item.name}{item.isSystem ? " (系统角色)" : ""}`）必须删除，否则会和独立列重复。

---

## 7. 备注字段（remark）

### 7.1 存储与接口

| 层 | 变更 |
| --- | --- |
| ent schema | `role.go` 增加 `field.String("remark").MaxLen(5000).Default("")`，中文注释说明「给成员分配角色时的说明」 |
| 迁移 | 启动时 `client.Schema.Create`（`internal/data/partition.go`）自动加列；`NOT NULL DEFAULT ''` 是纯增量变更，老行回填空串，无需手工 SQL，也无需 `WithDropColumn` |
| `pkg/bo` | `bo.Role.Remark string` |
| repository | `CreateRole(ctx, ns, name, remark)`；新增 `UpdateRoleRemark(ctx, ns, uid, remark)`；`roleViews` 填充 `Remark` |
| biz | `CreateRole` 透传 remark（trim）；新增 `UpdateRoleRemark`：校验 `role:write`，**允许系统角色**（备注是运维说明，不改变预置权限边界，而列表和分配弹窗需要它），`MaxLen` 由 proto 校验兜底 |
| proto | `RoleItem.remark = 7`；`CreateRoleRequest.remark = 2`（`max_len = 200`）；新增 `rpc UpdateRoleRemark` → `PUT /v1/role/{uid}/remark`，`body: "*"`，`UpdateRoleRemarkRequest{uid, remark}` / `UpdateRoleRemarkResponse{role}` |
| 默认值 | 4 个系统角色的预设备注放在 `perm` 包（`DefaultRoleRemark(role)`，与 `RolePermissions` 相邻），两处写入：① `data/namespace_repo.CreateNamespace` 新建租户建 4 个系统角色时直接带上；② `data/seed.syncTenantRoles` 对**已有**系统角色在 `remark = ''` 时补齐（不覆盖管理员改过的值）。文案见下表 |

系统角色默认备注：

| 角色 | 写入数据的中文备注（≤5000 字，前端不做翻译，按用户数据展示） |
| --- | --- |
| `owner` | 租户负责人，拥有全部权限，包括删除租户与移交负责人。 |
| `admin` | 可管理成员、角色与配置，不能删除租户或移交负责人。 |
| `member` | 日常运维：查看全部资源，处理告警与 SSH 执行。 |
| `viewer` | 只读：可以查看但不能修改任何配置。 |

注意现有 `role_repo.UpdateRole` 与 `biz.UpdateRole` **各有一道** `IsSystem` 拒绝。新的 `UpdateRoleRemark` 在这两层都不做这个判断，两处都写一行注释指向本节原因，避免后来者「照着邻居补一个判断」把系统角色备注锁死。

为什么不把 remark 塞进现有 `UpdateRole`：`UpdateRole` 的语义是「自定义角色改名」，biz 里对系统角色直接拒绝。备注需要能被系统角色修改，混在一起会把「改名」和「写备注」两种权限判断耦合成一个分支，容易漏。独立接口后：

- 自定义角色保存：`updateRole` + `updateRoleRemark` 并行发出；
- 系统角色保存：只发 `updateRoleRemark`。

### 7.2 读取位置

| 界面 | 位置 | 形态 |
| --- | --- | --- |
| 角色列表 | 「备注」独立列（`w-[18rem]`） | 最多两行 `line-clamp-2`，超出省略；**不挂 Tooltip**（悬停不展示全文），全文去角色详情看；空值 `—` |
| 角色详情 | 备注字段 | 普通文本，`whitespace-pre-wrap` 保留换行，`break-all` 沿用 `DetailSheet` |
| 编辑角色 | 备注 `Textarea` | 3 行 + `0/200` 计数 |
| 权限配置器 | 头部副标题下方 | 只读一行，`line-clamp-2`，帮助在授权时确认角色用途 |
| 成员分配角色 | 角色名后的 `?` 图标 Tooltip | 见 §9 |

### 7.3 校验

- 前端：`trim` 后 ≤5000 字符（按 `Array.from(value).length` 计，避免 emoji 算两个），超限实时红框 + `role.remarkTooLong`。
- 后端：proto `string.max_len = 5000`，biz 不做二次规则，`merr.Client` 由统一错误映射回字段（沿用 `readFormError`）。
- 空备注合法，写 `''` 而不是 `null`。

---

## 8. 授权交互重设计

### 8.1 目标

从「平铺 22 个裸编码勾选框」变成「可分组的、能搜的、能批量的、有计数和脏值保护的权限清单」。核心动作只有三个：**找**（搜索）、**勾**（单点/整组/全部）、**存**（脏值保护）。

### 8.2 交互定义

| 行为 | 触发 | 结果 |
| --- | --- | --- |
| 单点勾选 | 点行任意位置 / 点勾选框 / 空格键 | 切换该权限；组头三态随之更新；页脚统计即时刷新 |
| 整组全选 | 点组头勾选框 | 按当前状态推进：部分选中 → 全选；全选 → 全不选；全不选 → 全选。不记忆曾经的部分选择 |
| 整组展开/收起 | 点组头左侧 / 整行（勾选框除外） | 折叠状态记在 `collapsed: Set<resource>`；搜索或「仅看已选」时强制展开且组头置灰不可折叠 |
| 搜索 | 输入 `成员` / `Read` / `member` / `alert_rule:write` | 实时过滤（`useDeferredValue` 防抖 150ms）；命中组自动展开；组头 `已选/命中` 双计数；无命中显示空态 |
| 仅看已选 | Toggle | 只渲染 `selected` 非空的组与行；组头「全选」变为「取消本组」；与搜索可叠加（先过滤再取交集） |
| 全选 | 工具条按钮 | 对**当前可见集合**全选；搜索激活时按钮文案变 `role.grant.selectMatches`（避免误选 20 项） |
| 清空 | 工具条按钮 | 清空全部（不看过滤）；点击后按钮变「撤销」，5 秒内可撤销，之后落定 |
| 保存 | 页脚按钮 / `Ctrl(⌘)+S` | 仅在 `dirty` 且非 `saving` 时可用；成功后关闭 + toast + `reload()` |
| 取消 / 关闭 | 页脚 / ✕ / `Esc` | `dirty` 时弹 `ConfirmAction`（放弃修改？），否则直接关 |
| 系统角色只读 | 从列表点「权限」 | 同容器，`readonly` 模式：无勾选框（改显示 `Check` 图标），无工具条批量按钮，页脚仅「关闭」，顶部一行 `role.grant.readonly` 说明 |

### 8.3 状态模型

```ts
type EditorState = {
  baseline: Set<string>     // 打开时的角色权限
  selected: Set<string>     // 当前选择
  query: string
  onlySelected: boolean
  collapsed: Set<string>
  status: "idle" | "saving" | "failed"
  error: string
}
```

派生值（全部纯计算，不存 state）：

- `dirty = symDiff(baseline, selected).size > 0`
- `added` / `removed` 计数 → 页脚 `新增 n · 移除 m`
- `groups` → `groupPermissions(catalog)` 后套用搜索/仅看已选过滤
- `visibleAllSelected` → 工具条「全选」的禁用条件

### 8.4 状态矩阵

| 状态 | 头部 | 主体 | 页脚 |
| --- | --- | --- | --- |
| 加载目录中 | 骨架行（`ui/skeleton`） | 3 个分组骨架 | 保存禁用 |
| 目录加载失败 | 标题正常 | 空态 + `action.retry`（复用 `listPermissions`） | 保存禁用 |
| 无权限选中（自定义） | `已选 0 / 20` | 正常列表 + 顶部提示 `role.grant.emptyHint`（未选任何权限的角色只能看到自己） | 保存禁用（沿用后端 `min_items = 1`） |
| 搜索无命中 | `已选 n / 20` 不变 | 空态：`role.grant.noMatch` + 「清空搜索」按钮 | 正常 |
| 仅看已选且为空 | — | 空态：`role.grant.noSelected` | 正常 |
| dirty | `新增 2 · 移除 1` badge | 正常 | 保存可用 |
| saving | 保存按钮 `Loader2` + 禁用全部勾选 | 只读 | 取消禁用 |
| failed | 顶部 `Alert`（`text-destructive`） | 保持选择不变 | 保存可重试 |

### 8.5 无障碍与精度

- 组头勾选框：`aria-checked="mixed"`；`aria-controls` 指向组内容；`aria-label="全选 成员"`。
- 每一行是 `<label htmlFor>` 结构，键盘 `Tab` 落在勾选框、`Space` 切换；焦点环用 `focus-visible:ring-2 ring-ring ring-offset-2`（与 `Button` 一致）。
- 行高 44px，勾选框 16px 但有整行热区。
- 折叠动画 `duration-150`，`prefers-reduced-motion` 下用 `motion-reduce:transition-none`。
- 中文动作是主标签，英文是副标签，**不用** `title` 属性承载编码；编码是可见的 mono 文本，屏幕阅读器会连读，用 `aria-hidden` 标记编码 span 避免噪音。

### 8.6 不做的

- 不做权限模板 / 从其他角色复制（超出本次范围，容易引入「模板与角色双写」的一致性负担）。
- 不做权限搜索的服务端分页（目录只有 22 项）。
- 不做拖拽排序或权限分组自定义（分组由 `perm.Catalog()` 决定，属于产品语义）。

---

## 9. 成员分配角色时的备注 Tooltip

### 9.1 数据来源改造

成员页当前用 `selectRoles()`（`SelectItem{uid,name}`），拿不到 `isSystem` 和 `remark`。改为 `listRoles({ page: 1, pageSize: 100 })`：

- 一次拿到 `RoleItem{uid,name,isSystem,remark,status}`，类型 Tag 与 tooltip 同源；
- 保留现有过滤 `name !== "owner"`，并追加 `status === "ACTIVE"`（停用角色不该再分配）；
- 已知限制：`pageSize` 上限 100，角色超过 100 个时列表不完整。当前产品形态（单租户 4 系统角色 + 少量自定义）可接受，超出后改为带搜索的分页选择器（`RemoteSelect`），在文件里留 `TODO` 注释。

### 9.2 行结构

```
[☐] 值班运维 [自定义] ⓘ        ← 桌面：ⓘ hover/focus 打开 Tooltip
[☐] admin    [系统]   ⓘ
```

- 图标：`CircleHelp`，`size-4`，`text-muted-foreground`，`hover:text-foreground`，`shrink-0`。
- 触发器：`<button type="button" aria-label={t("member.roleRemarkLabel", { name })}>`，`onClick` 阻止冒泡（避免误触发行勾选）。
- Tooltip 内容：`<div className="grid gap-1"><span className="font-medium">{t("member.roleRemark")}</span><span>{remark}</span></div>`；`side="right"`，`align="start"`，局部 `TooltipProvider delayDuration={300}`（全局 provider 是 0，长文本瞬间弹出会打扰扫读）。
- 备注为空：图标保留但 `disabled` + `aria-disabled`，Tooltip 文案 `member.roleRemarkMissing`（「暂无备注」）。**保留图标位**，避免有/无备注的角色名左右错位。
- 备注超长（200 字）：`TooltipContent` 已是 `max-w-xs`，内部 `whitespace-pre-wrap break-words`，最长约 8 行，超出该长度不再截断（200 字是上限，可接受）。
- 窄屏（`< sm`）：不渲染图标，改为角色名下方 `text-xs text-muted-foreground line-clamp-2` 行内展示备注，空备注不渲染该行。这是触屏下 hover 不可用的兜底。
- 键盘：`Tab` 能到达图标按钮，focus 即展开 tooltip（Radix 默认行为）。

### 9.3 分配弹窗其他调整

- 顶部搜索框（客户端过滤 `name` + `remark`）。
- 页脚显示 `已选 {{count}} 个角色`。
- 「至少选 1 个角色」校验保持不变；保存后 `reload()`。
- 无匹配：`member.roleSearchEmpty`。

---

## 10. i18n key 命名

命名沿用现有约定：`<模块>.<语义>`、camelCase、中英两份镜像。新增两个顶层段：`permission`（权限字典）、以及 `role` / `member` 的补充键。

### 10.1 `permission.*`（新增段）

| key | zh | en |
| --- | --- | --- |
| `permission.resource.member` | 成员 | Members |
| `permission.resource.role` | 角色 | Roles |
| `permission.resource.namespace` | 租户 | Namespaces |
| `permission.resource.datasource` | 数据源 | Datasources |
| `permission.resource.alert_rule` | 告警规则 | Alert rules |
| `permission.resource.alert` | 告警 | Alerts |
| `permission.resource.notification` | 通知 | Notifications |
| `permission.resource.agent` | 机器 | Agents |
| `permission.resource.ssh` | SSH 命令 | SSH commands |
| `permission.resource.unknown` | 未分类 | Uncategorized |
| `permission.action.read` | 读取 | Read |
| `permission.action.write` | 写入 | Write |
| `permission.action.intervene` | 干预 | Intervene |
| `permission.action.exec` | 执行 | Execute |
| `permission.action.delete` | 删除 | Delete |
| `permission.action.transfer` | 移交 | Transfer |
| `permission.action.unknown` | 未知动作 | Unknown action |
| `permission.count`（可选） | {{count}} 项 | {{count}} item(s) |
| `permission.groupTitle`（可选） | {{module}} · {{moduleEn}} | {{module}} · {{moduleEn}} |

### 10.2 `role.*` 补充

| key | zh | en |
| --- | --- | --- |
| `role.type` | 类型 | Type |
| `role.typeSystem` | 系统角色 | System role |
| `role.typeCustom` | 自定义角色 | Custom role |
| `role.typeSystemHint` | 预置角色，权限由平台维护，不能改名或删除。 | Built-in role. Permissions are managed by the platform; it cannot be renamed or deleted. |
| `role.remark` | 备注 | Remark |
| `role.remarkPlaceholder` | 说明这个角色能做什么、分配给谁 | Describe what this role does and who should get it |
| `role.remarkHint` | 最多 5000 字。列表只显示两行，全文在角色详情里看。 | Up to 5000 characters. The list shows two lines; open the role details for the full text. |
| `role.remarkEmpty` | 暂无备注 | No remark |
| `role.remarkTooLong` | 备注最多 5000 字 | Remark must be 5000 characters or fewer |
| `role.remarkCount` | {{count}}/5000 | {{count}}/5000 |
| `role.systemNameLocked` | 系统角色不能改名 | System roles cannot be renamed |
| `role.permissionCount` | {{count}} 项 | {{count}} item(s) |
| `role.unknownModule` | 未在内置字典中的模块 | Module not in the built-in dictionary |
| `role.grant.title` | 配置权限 | Configure permissions |
| `role.grant.viewTitle` | 查看权限 | View permissions |
| `role.grant.readonly` | 系统角色的权限由平台预置，只能查看。 | System role permissions are preset and read-only. |
| `role.grant.selected` | 已选 {{selected}} / {{total}} | {{selected}} of {{total}} selected |
| `role.grant.diff` | 新增 {{added}} · 移除 {{removed}} | +{{added}} · −{{removed}} |
| `role.grant.search` | 搜索权限：中文、English 或 member:read | Search by name, English, or member:read |
| `role.grant.onlySelected` | 仅看已选 | Selected only |
| `role.grant.expandAll` | 展开全部 | Expand all |
| `role.grant.collapseAll` | 收起全部 | Collapse all |
| `role.grant.selectAll` | 全选 | Select all |
| `role.grant.selectMatches` | 全选搜索结果（{{count}}） | Select all matches ({{count}}) |
| `role.grant.clear` | 清空 | Clear |
| `role.grant.undo` | 撤销 | Undo |
| `role.grant.selectGroup` | 全选 {{module}} | Select all {{module}} |
| `role.grant.noMatch` | 没有匹配的权限 | No permissions match |
| `role.grant.noMatchHint` | 换个关键词，或清空搜索框。 | Try another keyword or clear the search. |
| `role.grant.clearSearch` | 清空搜索 | Clear search |
| `role.grant.noSelected` | 还没有选择任何权限 | No permissions selected yet |
| `role.grant.emptyHint` | 至少选择一个权限，否则该角色看不到任何资源。 | Select at least one permission; otherwise the role can see nothing. |
| `role.grant.loadFailed` | 权限目录加载失败 | Failed to load the permission catalog |
| `role.grant.ownerOnly` | 删除租户、移交负责人只属于负责人角色，不可授予自定义角色。 | Deleting a namespace and transferring ownership belong to the owner role only. |
| `role.grant.unsavedTitle` | 放弃未保存的修改？ | Discard unsaved changes? |
| `role.grant.unsavedDescription` | 关闭后本次勾选的权限不会保存。 | The permissions you changed will not be saved. |
| `role.grant.saved` | 权限已更新 | Permissions updated |

### 10.3 `member.*` 补充

| key | zh | en |
| --- | --- | --- |
| `member.roleRemark` | 备注 | Remark |
| `member.roleRemarkLabel` | {{name}} 的备注 | Remark for {{name}} |
| `member.roleRemarkMissing` | 暂无备注 | No remark |
| `member.roleSearch` | 搜索角色 | Search roles |
| `member.roleSearchEmpty` | 没有匹配的角色 | No roles match |
| `member.roleSelected` | 已选 {{count}} 个角色 | {{count}} role(s) selected |
| `member.roleDisabled` | 已停用 | Disabled |

（`role.remarkEmpty` 与 `member.roleRemarkMissing` 分开，是为了后续文案可独立调整；若要减 key，可合并为一个 `role.remarkEmpty`。）

---

## 11. 文件清单与实施顺序

### 批次 1：数据层（后端可独立验证）

1. `app/alioth/internal/data/ent/schema/role.go` — 加 `remark` 字段。
2. `app/alioth/internal/data/ent/` — 在 `ent/` 目录执行 `go generate ./...`（对应 `ent/generate.go` 的 `go:generate ent generate ./schema`），重新生成 `role.go`、`role_create.go`、`role_update.go`、`migrate/schema.go` 等。
3. `app/alioth/pkg/bo/bo.go` — `bo.Role.Remark`。
4. `app/alioth/internal/data/repository/role.go` — 修改 `CreateRole`、新增 `UpdateRoleRemark`。
5. `app/alioth/internal/data/role_repo.go` — 实现 `CreateRole` 新签名、新增 `UpdateRoleRemark`（**不加 isSystem 判断**，注释说明原因）；`roleViews` 里补 `Remark` 一处即可覆盖 列表/详情/创建/改状态/授权的全部返回路径。
6. `app/alioth/pkg/perm/perm.go` — 增加 `DefaultRoleRemark(role string) string`，与 `RolePermissions` 放在一起，作为系统角色备注的唯一来源。
7. `app/alioth/internal/data/namespace_repo.go` — 新建租户时 `Create()` 带上默认备注。
8. `app/alioth/internal/data/seed.go` — `syncTenantRoles` 里对已有系统角色补空备注（不覆盖非空值）。

验证：`make migrate` 起库后 `\d roles` 能看到 `remark`；`go build ./...`。

### 批次 2：接口契约

9. `api/alioth/v1/identity.proto` — `RoleItem.remark`、`CreateRoleRequest.remark`、`UpdateRoleRemark` RPC + 两个 message。
10. `make api`（`buf dep update && buf generate`）→ `api/alioth/v1/*.pb.go`、`*_http.pb.go`。
11. `app/alioth/internal/service/role.go` — 映射 `Remark`、实现 `UpdateRoleRemark`。
12. `app/alioth/internal/biz/role/role.go` — `CreateRole(ctx, name, remark)`、新增 `UpdateRoleRemark`（允许系统角色，注释写明原因）。
13. Wire：本次只改接口方法与既有构造函数签名，ProviderSet 不变，通常**不需要**重生成；只有新增构造器时才跑 `make wire`（生成物在 `app/alioth/cmd/alioth/wire_gen.go`）。

验证：`curl` 三个端点（create 带 remark、list 返回 remark、`PUT /v1/role/{uid}/remark` 改系统角色备注）。

### 批次 3：前端契约与基础件

14. `web/src/api/role/types.ts` — `RoleItem.remark`、`CreateRoleParams.remark?`、`UpdateRoleRemarkParams`。
15. `web/src/api/role/index.ts` — `updateRoleRemark()`。
16. `web/src/lib/permission-catalog.ts`（新）— 分组、排序、搜索、owner-only 过滤。
17. `web/src/components/role-type-tag.tsx`（新）— 类型 Tag。
18. `web/src/components/role-remark-hint.tsx`（新）— `?` 图标 + Tooltip（供成员页与角色列表复用）。
19. `web/src/components/ui/checkbox.tsx` — Indicator 增加三态：`data-state="indeterminate"` 时显示 `MinusIcon`（`group-data-[state=indeterminate]:hidden` / `block` 成对切换），保持竖线/对勾尺寸 14px。这是本地组件，可直接改。
20. `web/src/components/detail-sheet.tsx` — `DetailField` 增加可选 `span?: 1 | 2`，权限分组占满整行（向后兼容）。
21. `web/src/i18n/index.ts` — 中英两份新增键（§10）。

### 批次 4：权限配置器（最大单点）

22. `web/src/components/permission-editor.tsx`（新）— Sheet + 工具条 + 分组列表 + 页脚 + 脏值保护；`mode: "edit" | "readonly"`。
23. `web/src/components/permission-summary.tsx`（新）— 只读分组渲染（详情页与授权面板头部共用分组排版）。

验证：单独挂一个临时路由或 Story 式预览，手工过一遍 §8.4 状态矩阵。

### 批次 5：页面接线

24. `web/src/pages/role/index.tsx` —
    - 列表：删掉名称里的类型文案，加「类型」「备注」两列；
    - 行操作：系统角色加只读「权限」入口；
    - 新建/编辑 Dialog：加备注 `Textarea` + 计数；系统角色名称只读；保存拆成 `updateRole` + `updateRoleRemark`；
    - 详情：换成 `wide` + `PermissionSummary` + 备注；
    - 授权弹窗：`Dialog` → `PermissionEditor`；
    - `catalog` 过滤逻辑移到 `permission-catalog.ts`，并保留加载失败的重试。
25. `web/src/pages/member/index.tsx` — 角色数据源换 `listRoles`；分配弹窗抽出为 `AssignRolesDialog`（可留在同文件下方，若超过 600 行再拆文件），加搜索、类型 Tag、`RoleRemarkHint`、已选计数、窄屏行内备注。

验证：`cd web && pnpm typecheck && pnpm lint`，再手工走一遍角色创建 → 授权 → 成员分配闭环。

### 批次 6：收尾

26. `docs/rbac-role-redesign.md`（本文件）随实现更新「已知限制」。
27. 若有截图/验收记录，追加到 `docs/roadmap.md` 对应里程碑。

---

## 12. 验收清单

| # | 需求 | 验收点 |
| --- | --- | --- |
| 1 | 详情展示权限列表 | 打开任意角色详情，权限按组分块，显示 `已选/总数`，长列表不溢出（`wide` sheet + 内部滚动） |
| 2 | 模块 + 动作分组、中英标注 | 分组头为 `成员 · Members`；每项为 `读取 / Read` + mono `member:read`；切到 en 全部变英文；字典缺键时回落 code 不空白 |
| 3 | 授权交互重设计 | 搜索（中文/英文/code）、组内全选、三态组头、仅看已选、全选/清空+撤销、脏值关闭确认、`⌘S` 保存、`已选 a/b` 与 `新增/移除` 计数、保存失败保留选择 |
| 4 | 备注字段 | 列表列、详情字段、新建+编辑可写、系统角色可写备注但不可改名、200 字上限前后端一致、空值显示 `—` |
| 5 | 系统/自定义 Tag | 独立「类型」列，两套底色，中英文案随语言切换；名称列不再出现 `(系统角色)` |
| 6 | 分配角色 tooltip | 角色旁 `?`，hover 300ms 显示备注；空备注显示「暂无备注」且图标位不塌陷；键盘 focus 可见；`< sm` 改为行内两行文本 |

附加检查：

- 权限不足（无 `role:read` / `role:write`）时按钮、行操作、只读态正确隐藏；系统角色永不出现编辑/删除。
- 空态：无角色、搜索无结果、权限目录加载失败（带重试）。
- 深浅色下 Tag 与 `bg-primary/5` 选中行对比度 ≥ 3:1，文本对比度 ≥ 4.5:1（`muted-foreground` 已满足）。
- 键盘全流程：列表 → 详情 → 授权面板 → 保存，`Esc` 与焦点返回正确（Radix 默认，勿覆盖 `onOpenAutoFocus`）。

---

## 13. 风险与取舍

| 风险 | 处理 |
| --- | --- |
| ent 启动自动加列在生产库上执行 DDL | 与现有迁移策略一致（`Schema.Create` 已在启动路径），列为 `NOT NULL DEFAULT ''` 的可空安全变更；大表（角色通常 <1000 行）无锁风险 |
| 编辑弹窗发两个请求，部分失败 | 先 `updateRoleRemark` 再 `updateRole`；失败时保留面板、提示具体字段错误，重试幂等（都是全量覆盖写） |
| 成员分配改用 `listRoles` 后角色 >100 不可见 | 明确记为已知限制，落在代码注释 + 本文件；后续换分页搜索选择器 |
| 权限字典与 `perm.Catalog()` 漂移 | 前端兜底显示 code；`permission-catalog` 增加一条注释指向 `perm.go`；后续可考虑由后端返回 `resourceLabel`/`actionLabel`（本次不做，i18n 留前端） |
| 成员页从「角色名」扩展成「角色名 + Tag + 图标」后行宽拥挤 | Tag 用 `sm:` 以上才显示，窄屏只留名称 + 行内备注 |
| 编码 span 影响屏幕阅读器 | `aria-hidden`，标签用中文动作 + 英文动作 |

---

## 14. 实现记录（本次落地）

已按 §11 的批次 1–5 实现，实际落地与设计稿的差异如下（其余按稿）：

**数据层 / 接口**
- `role.remark`：ent schema（`MaxLen(200) Default("")`）→ `bo.Role.Remark` → `repository.RoleRepo`（`CreateRole` 带 remark、新增 `UpdateRoleRemark`）→ `role_repo.roleViews` 统一回填。
- 预置备注：`perm.DefaultRoleRemark` 是唯一来源，建租户（`namespace_repo`）与 `seed.syncTenantRoles`（仅补空值）两处写入。
- proto：`RoleItem.remark = 7`、`CreateRoleRequest.remark = 2`、新增 `UpdateRoleRemark` → `PUT /v1/role/{uid}/remark`。service/biz 都**不**对系统角色设限，只校验 `role:write`。
- 服务启动时的 `client.Schema.Create` 自动加列（`NOT NULL DEFAULT ''`），无需手工 SQL。

**前端**
- 新增 `lib/permission-catalog.ts`：中英字典用 `{zh, en}` 字面量而不是 `t()`，保证中英同时显示；含分组顺序、动作顺序、`OWNER_ONLY`、搜索与兜底。
- `RoleTypeTag`：系统角色 sky、自定义角色中性灰（文字提到 foreground 级别，避免 4.5:1 临界）；列表独立「类型」列，名称列不再拼「(系统角色)」。
- 详情 Sheet：`wide` + `PermissionSummary`，模块头 `成员 · Members` + `已选/总数`，动作用 `读取 / Read` Badge，Tooltip 给 `成员 / 读取 · Members / Read` 与 code；**展示目录 22 项**（owner 的 `namespace:delete/transfer` 不再被过滤），编辑器才用 20 项的可分配目录。
- `PermissionEditor`：右侧 Sheet（`sm:max-w-3xl`，窄屏整宽）、搜索（中文/English/code）、组三态（`indeterminate` + `MinusIcon`，本地 checkbox 组件新增该分支）、仅看已选、展开/收起、全选（按可见集，搜索时文案变「全选搜索结果（n）」）、清空 5s 撤销、`新增 n · 移除 m`、`⌘/Ctrl+S`、脏值关闭确认、保存失败保留选择。
- 备注：列表列（空值 `—`）、详情字段（保留换行）、新建/编辑 Textarea + `0/200` 计数（按字符计）；系统角色名称禁用并给 `role.systemNameLocked` 提示，备注可改。
- 成员分配：数据源从 `selectRoles` 换成 `listRoles`，行内是「勾选框 + 名称 + compact 类型 Tag + ? 备注 Tooltip（300ms、空备注显示「暂无备注」、图标位不塌陷）」；窄屏与 `hover:none` 时改为名称下方两行备注。

**与设计稿的有意取舍**
1. **不单独做只读权限面板**：系统角色的「权限」入口直接打开详情 Sheet（本就只读且含 22 项分组），避免同一份分组渲染两套壳子。系统角色因此多了一个「编辑备注」入口（详情 + 更多：权限 / 编辑）。
2. **不做类型筛选 Tabs**：`ListRolesRequest` 没有 `is_system`，服务端分页下前端过滤是假实现，维持 P2。
3. **备注列常显**：设计稿写 `≥lg` 才显示；实现里常显并允许换行，表格自身横向滚动（`TABLE_MIN_WIDTH.default`），避免同一列在断点间跳动。
4. **成员分配保留停用角色**（灰显 + 禁勾，已选中的保持可取消），不让已有分配被静默丢弃；邀请下拉只列 ACTIVE。

**验证情况（如实）**
- `pnpm typecheck` / `pnpm lint` / `pnpm build` 通过；`go build ./...`、`go vet ./...`、`gofmt -l` 通过。
- 浏览器实测（headless Chrome + 真实登录会话，dev server：Vite 8080 / 旧版 alioth 8000）：1366×768 下列表列头为「名称/类型/备注/状态/操作」、行内类型 Tag 正确、owner 详情显示 22 项权限且分组中英并标、权限配置器 20 项 / 9 组 / 已选计数与 diff、组头勾选、搜索 `ssh` 只剩 SSH 组且按钮变「全选搜索结果（3）」、脏值 Esc 出现放弃确认、成员分配弹窗的 `?` hover 显示「备注 / 暂无备注」、系统角色编辑弹窗名称禁用；390×844 下详情/配置器整宽、保存按钮在视口内、表格横向滚动不撑破页面（`documentElement.scrollWidth = 390`）。
- **未验证**：dev 环境的 alioth 仍是改动前的二进制，因此 remark 的真实写入/回读、启动自动加列、`PUT /v1/role/{uid}/remark` 均未做端到端实测；上述截图中备注列与 Tooltip 展示的是空备注态。重启后端（重新构建 `bin/alioth`）后需补一次「新建角色带备注 → 列表回显 → 系统角色改备注 → 成员分配 hover 看到备注」的闭环。

**2026-10-25 调整（备注上限与展示）**
- 备注上限从 200 字改为 **5000 字**：ent `MaxLen(5000)`、proto `max_len = 5000`、前端计数与校验同步；`Array.from` 计数不变。
- 列表「备注」列固定 **最多两行**（`line-clamp-2` + 省略号），**不加 Tooltip**——悬停不再展示全文，避免长备注把行高和阅读节奏打乱；全文只在角色详情里看。
- 详情里的备注改为自己的滚动区（`max-h-64 overflow-y-auto whitespace-pre-wrap`），5000 字不会把权限分组挤出 Sheet。
- 成员分配弹窗的 `?` Tooltip 保留（需求 6 明确要求），但内容限高 `max-h-60` 可滚动；窄屏行内备注仍为两行截断。
- 实测：`line-clamp: 2`、span 高 40px（两行 × 20px）、`scrollHeight 880` 被裁切、悬停备注列无 Tooltip 元素；typecheck / lint / build / go build 通过。

**2026-10-25 补全：权限目录里的「删除」**
- 权限配置器不再过滤 `namespace:delete` / `namespace:transfer`：22 项目录全部可见，这两项显示为禁用行 + 「仅负责人」标记（`role.grant.ownerOnlyBadge`），勾选框不可点。
- 「可分配」仍是 20 项：`已选 a / 20`、组头计数、组全选、工具条全选都只作用于可分配项，所以全选结果是 20 而不是 22。
- 后端规则未变：`biz.UpdateRolePermissions` 与 `role_repo` 仍拒绝把这两项授给自定义角色；如果以后允许自定义角色持有「删除租户 / 移交负责人」，要先改这条业务规则（见 `perm.go` 与 `biz/role/role.go` 的注释）。
- 实测：租户组渲染 4 行（读取/写入/删除/移交），删除与移交的 checkbox 为 disabled 且带「仅负责人」；已选 0/20；点「全选」后 20/20。

**2026-10-25 权限细化：删除普通业务数据可单独授权**
- `perm.Catalog()` 为 8 个业务模块补上 `*:delete`（member / role / datasource / alert_rule / alert / notification / agent / ssh），目录从 22 项变为 **30 项**，可分配 **28 项**；`namespace:delete` / `namespace:transfer` 仍是 owner 专属（展示但禁用）。
- 删除操作的鉴权从 `*:write` 改为 `*:delete`：成员、角色、数据源、告警规则/规则组/告警字典、告警页面、通知渠道/模板/组/订阅、机器、SSH 命令。只读角色不受影响；已有自定义角色如果只被授予 write，升级后将不能再删除（这是本次细化的预期结果）。
- 预置角色：owner 30 项（含删除租户/移交）；admin 28 项（业务删除全有，不含删除租户/移交）；member / viewer 不变（11 / 9）。
- 前端无需改代码：权限目录、分组、中英标签都由后端目录驱动，`OWNER_ONLY` 仍只含 namespace 两项。
- 实测（新二进制跑在 127.0.0.1:8001 + 独立 Vite 8081，连同一套 dev 库）：`/v1/permissions` 返回 30 项、9 个 delete；编辑器 `已选 0 / 28`，成员/角色/数据源等组出现可勾选的「删除 / Delete」；勾选 `datasource:delete` 保存后接口回读 `['datasource:delete']`（验证完已把该测试授权清回空）；owner 详情 30/30、admin 28/28。
- 迁移：`role.permissions` 的新行由启动时的 `seed()` 写入，`syncTenantRoles` 给预置角色补授权；本次验证时已对本机 dev 库执行过一次（`permissions` 表 22 → 30）。本机 8000 端口仍是旧二进制，需重新构建并重启才会看到新目录。
