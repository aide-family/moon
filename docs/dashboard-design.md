# Alioth 概览页（Dashboard）设计规范

> 状态：设计稿，待主会话按本文件开发。**本文件不改代码**，只给结构、视觉与文件级改动建议。
> 依据：`README.md`、`docs/design.md`、`docs/roadmap.md`、`web/AGENTS.md`、`.cursor/skills/alioth-web-design/{SKILL.md,revisions.md}`、
> `web/src/{App.tsx,index.css}`、`app-shell.tsx`、`pages/alert/*`、`api/**`、`i18n/index.ts`、
> 以及 `api/alioth/v1/*.proto` 与 `app/alioth/internal/{service,biz,data}` 的现网实现。
> 设计边界：本页只做**结构与视觉**。统计口径由「统计学家」定义，图表类型与刻度由「数据可视化工程师」定义；
> 本文出现的每个数字都标注了它来自哪个已定稿接口、口径是否已被后端固定。

---

## 0. 结论摘要

| 项 | 结论 |
| --- | --- |
| 页面定位 | **值班态势与处置入口**，不是统计报表，也不是指标探索器 |
| 路由 | 新增 `/overview`；`/` 由 `Navigate to="/profile"` 改为 `Navigate to="/overview"` |
| 首屏 | 4 个 KPI + 「最近触发的告警」清单前 5 行，1366×768 下不滚动即可见 |
| 栅格 | Tailwind 默认断点 640 / 1024 / 1280；KPI 2→2→4 列，主区 xl 起 `1.6fr / 1fr` |
| 视觉方向 | 推荐 **方向 A「机舱仪表盘」**（沿用现有玻璃卡片 + 单一主色），弃用 NOC 大屏式方向 B |
| 新依赖 | **零**。图表用现有 CSS 条形 / 内联 SVG；若 v2 需要时间序列，先复用 `MetricGraph`，再由数据可视化工程师选型 |
| 口径红线 | `recovered` 是**生命周期累计（含归档）**，禁止当作"今日/近24小时恢复数"，也禁止与 firing 放进同一个堆叠比例 |
| 需确认 | 侧栏「概览」作为一级直链是否突破 `revisions.md` 的"一级只放领域分组"约定 |

---

## 1. 业务属性判断与目标用户

### 1.1 项目真实业务属性

Alioth 是 aide-family 的**自建、多租户运维平台**（`README.md`、`docs/design.md` §1）。它不是消费级产品，也不是纯观测查询工具，而是一条闭环：

```
数据源(Prometheus/VictoriaMetrics/Loki/VictoriaLogs/Elasticsearch/Jaeger)
  → 告警规则(级别条件/表达式)  → megrez 周期性评测
  → 告警事件(触发/抑制/恢复)   → 通知订阅/通知组/渠道 → merak 投递 → 消息日志
边缘侧：alkaid Agent → 机器信息 → SSH 命令(提交/审核/执行) → 审计
控制面：alioth 唯一写 PostgreSQL；多租户 + RBAC(resource:action)；Web 是集群外唯一入口
```

由此得出三个**决定界面形态**的事实：

1. **中心是"状态与处置"，不是"探索"**。全站已有页面（实时告警、历史告警、数据源查询、通知投递、机器、SSH 审核）分别承担了深度查看；概览页若再做一遍深度查看就是重复建设。概览的价值是**第一眼判断 + 一跳直达**。
2. **平台自身健康是可观测对象**。`/workers` 已经暴露 megrez/merak 的在线/忙碌（15 秒心跳）、`/datasources` 暴露数据源配置、`/message-logs` 暴露投递结果。自建平台没有厂商 SLA，值班人必须能一眼看出"是没告警，还是评测器/通知器挂了"。
3. **权限差异极大**。权限码按 `resource:action` 细分（`alert:read`、`notification:read`、`agent:read`、`ssh:read`、`system:user:read`…，见 `lib/permission-catalog.ts`）；`/workers` 额外要求 `system:user:read`（`app/alioth/internal/biz/worker/worker.go:72`）。**没有租户时（`session.namespaceUid === ""`）告警与通知类页面整体不可见**（`app-shell.tsx` 的 `alerts/notifies/edges` 过滤）。概览页必须能为"四种可见性组合"都给出合理版面，而不是假设管理员视角。

### 1.2 目标用户与首要诉求

| 角色 | 典型身份 | 进页面第一诉求 | 对应版面 |
| --- | --- | --- | --- |
| 值班 SRE / 运维工程师（主用户） | 租户成员，有 `alert:read`，可能无 `notification:read` | "现在还有什么在响？多久了？我要不要动手？" | KPI + 最近触发的告警 + 态势构成 |
| 租户管理员 | OWNER/ADMIN，`alert:*`、`notification:*`、`member:*` | 同上，加"通知链路是否正常" | 上述 + 通知投递卡 |
| 平台管理员 | `SYSTEM_ADMIN` / `SUPER_ADMIN` | "评测器和通知器在跑吗？卡在哪个任务？" | 上述 + 后台进程卡 |
| 边缘操作者 | `agent:read` / `ssh:read` 或 `ssh:exec` | "机器在线吗？有没有等我审的命令？" | 待办卡（受限，见 §4.3） |
| 新用户 / 空租户 | 已登录，无租户或租户内无任何资源 | "我从哪开始？" | 引导型空态（建模向导入口） |

### 1.3 核心使用场景（也是验收场景）

- **S1 交接班/上班第一眼**：打开即知当前未恢复告警量、其中多少在抑制中、最近触发的是哪几条。
- **S2 一跳处置**：从概览进入 `/alerts`（可带 `?level=<levelUid>` 或 `?name=<关键字>`），再做抑制/恢复/介入。
- **S3 判断"是不是静默故障"**：告警为 0 时，用后台进程卡（在线/忙碌）与数据源数量判断是"真的没事"还是"链路断了"。
- **S4 通知是否送出**：近 24 小时投递成功/失败条数 + 失败样本入口 `/message-logs`。
- **S5 处理个人待办**：未读系统消息、待处理租户邀请。

**明确不做**（防止与历史分析页重复、防止越权造口径）：

- 不在概览页做历史趋势、MTTR、复现指纹排行 —— 这些属于 `/alert-history?historyTab=analysis`（`pages/alert/history-insights.tsx` 已实现，且带 30 天/5000 条截断提示）。
- 不在概览页做任意指标的图表探索 —— 属于 `/datasources` 查询面板。
- 不在概览页做任何写操作（抑制/恢复/审批/删除）—— 概览是发射台，不是操作台；避免误触生产处置。
- v1 不做拖拽自定义看板（需要用户偏好持久化 + 布局引擎，收益低于成本）。

---

## 2. 信息架构与页面布局

### 2.1 路由与导航

- 新路由 `/overview` → `<OverviewPage />`，放在 `AppShell` 内（需要登录与租户上下文）。
- `/` 改为 `Navigate to="/overview"`；`/profile` 仍从顶栏头像进入（`revisions.md` 2026-10-15 约定不变）。
- 侧栏新增「概览」`LayoutDashboard`，位于四个可折叠领域组**之上**。
  ⚠️ **与现有约定冲突，需主会话/用户确认**：`revisions.md` 2026-10-15 写"侧栏第一级只放功能领域，用可折叠子菜单，不放直接进入页面的链接"。概览是工作区首页而非业务领域，建议作为**唯一例外**；若不被接受，退路是只保留 `/` 重定向 + 侧栏品牌区点击进入（不推荐，可发现性差）。折叠成图标时 `tooltip="概览"`。
- 权限：概览本身不设新权限码。可见性由卡片级权限决定（见 §4.3）。

### 2.2 页面分区（信息优先级）

```
┌───────────────────────────────────────────────────────────────────────────────┐
│ R0 PageHeader  概览                          [窗口 24h ▾] [自动刷新] [刷新]    │  固定，不滚动
│    当前租户的告警态势、通知链路与待处理事项 · 数据截至 14:32:08               │
├───────────────────────────────────────────────────────────────────────────────┤
│ R1 KPI 条   ┌──────┬──────┬──────┬──────┐                                     │  首屏
│             │触发中│抑制中│累计恢复│我的告警│  ← 4 格，gap-px 分隔（复用现有写法）  │
│             └──────┴──────┴──────┴──────┘                                     │
│ R2 态势区   ┌──────────────────────────┬───────────────┐                      │
│   (xl 1.6/1)│ 最近触发的告警（5 行）    │ 态势构成       │                      │
│             │ · 级别 · 摘要 · 规则·数据源│ firing/suppressed 分割条             │
│             │ · 持续时长 · 状态         │ + 待办（3 行） │                      │
│             │ [查看全部 N 条 →]         │               │                      │
│             └──────────────────────────┴───────────────┘                      │
├───────────────────────────────────────────────────────────────────────────────┤
│ R3 链路区   ┌──────────────────────┬──────────────────────┐                   │  折叠线以下
│   (xl 1/1)  │ 通知投递（近 24 小时）│ 后台进程（管理员）    │                   │
│             │ 成功 N / 失败 M      │ megrez/merak 在线·忙碌│                   │
│             └──────────────────────┴──────────────────────┘                   │
└───────────────────────────────────────────────────────────────────────────────┘
```

**唯一纵向滚动归属**：`R0` 固定；`R1–R3` 放进 `min-h-0 flex-1 overflow-y-auto overscroll-contain` 的容器。
卡片内部**不再嵌套滚动**（`web/AGENTS.md` 布局规则 3：仪表盘滚主动面板，禁止双滚动）。
告警清单不设内部滚动条，只显示固定 5 行 + "查看全部"链接 —— 这既是规则要求，也让高度可预测、无 CLS。

### 2.3 首屏预算（可验收的算术）

1366×768 下：`48`(顶栏) + `24`(上内边距) + `52`(PageHeader) + `24`(gap-6) = **148px**，底部 `24` → 首屏可视 **596px**。

- KPI 条：`py-3`(24) + 标签 16 + 数值 36 + 辅助 16 + 间距 8 ≈ **104px**
- gap-6 = 24 → 态势区剩 **468px**
- 态势卡：标题区 52 + 内边距 32 + 5 行 × 64 = 320 + 底部链接 40 ≈ **444px**

结论：**首屏 = KPI 4 格 + 告警清单 5 行**，余量 24px。验收按"≥5 行完整可见"执行。

### 2.4 栅格与响应式断点

断点全部沿用 Tailwind 默认值（项目现有页面即 `sm/lg/xl`），**不引入自定义断点**。
内容宽度参考：1366 视口 → `1366 − 256(sidebar 16rem) − 48(p-4 md:p-6)` ≈ **1062px**；1440 → 1136px；1920 → 1616px。
`xl`(1280) 在 1366 已生效，所以 1366 及以上的笔记本就是"宽屏形态"。

| 区域 | base <640 | sm 640–1023 | lg 1024–1279 | xl ≥1280 |
| --- | --- | --- | --- | --- |
| R1 KPI | `grid-cols-2`（2×2） | `grid-cols-2` | `grid-cols-2` | `grid-cols-4` |
| R2 态势区 | 单列堆叠 | 单列堆叠 | `grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)]` | `grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]` |
| R2 右列（构成 + 待办） | 随上 | 随上 | 右列内部 `grid gap-4 content-start`，两张卡纵向堆叠 | 同左 |
| R3 链路区 | 单列堆叠 | 单列堆叠 | `grid-cols-2` | `grid-cols-2` |
| 卡片间间距 | `gap-4` | `gap-4` | `gap-4` | `gap-4` |
| 区块间间距 | `gap-6` | `gap-6` | `gap-6` | `gap-6` |

实现用 **fr 模板**而非 12 列 span，与现有 `xl:grid-cols-[minmax(0,1.7fr)_minmax(17rem,1fr)]`（历史分析）、
`grid-cols-[minmax(0,18rem)_minmax(0,1fr)]`（数据源）保持一致。12 列只是设计意图：R1 `4×3`、R2 `8/4`、R3 `6/6`。

窄屏额外约束（`web/AGENTS.md` 回归项）：
- 390×844 下 `document.scrollingElement.scrollWidth <= clientWidth`（页面不横向滚动），表格类内容若出现在卡片内，横向滚动只能发生在 `Table` 自己的容器里。
- KPI 在 390 下为 2 列，单格约 `(390−32−12)/2 = 173px`，足够容纳 4 位数字 + 标签；数字 `text-2xl`（24/32），`xl` 起 `text-3xl`（30/36）。
- 页头控件（窗口/自动刷新/刷新）用 `flex flex-wrap items-center gap-2`，窄屏换行而非压缩。

### 2.5 URL 与刷新策略

| 状态 | 存放位置 | 理由 |
| --- | --- | --- |
| 时间窗 `window`（`1h`/`6h`/`24h`/`7d`，默认 `24h`） | URL `?window=24h` | 可分享、可回退；仅作用于**窗口型卡片** |
| 自动刷新开关 | `localStorage`（`alioth:overview:auto-refresh`） | 个人偏好，不适合分享 |
| 选中告警级别 | 不在此页保存，跳转时写入 `/alerts?level=` | 避免两处状态打架 |

刷新节奏：**默认 30s**，与 `/alerts` 的 `REALTIME_REFRESH_MS = 30000` 一致。
每个 tick 只发 2 个请求（`alert-event-realtime-stats` + 告警清单），与实时告警页"每 tick 两个请求"的既定做法对齐；
通知/进程卡每 4 个 tick（约 2 分钟）或手动刷新时更新。
`document.visibilityState === "hidden"` 或 `session.switching` 为真时暂停计时，切回立即补一次。

---

## 3. 视觉风格方案

### 3.1 继承基线（必须先尊重，不是可选）

| 项 | 现值 | 出处 |
| --- | --- | --- |
| 主色 / ring / chart-1 | `#1677ff` | `index.css` `:root` |
| 背景 | `#f8fafc` + `--app-gradient`（两处 radial 蓝/青微光，`background-attachment: fixed`） | `index.css` `@layer base` |
| 卡片 | `--card: rgba(255,255,255,.85)`，`.glass-card` = `bg-card/80 backdrop-blur-sm border-border/50`，圆角 `rounded-xl`（`--radius-xl = 1.4 × 10px = 14px`） | `index.css`、`ui/card.tsx` |
| 深色 | `.dark` 类挂在 `<html>`；背景 `#0b1120`，卡片 `rgba(17,24,39,.75)`，边框 `rgba(148,163,184,.15)` | `index.css`、`theme-provider.tsx` |
| 字体 | sans `Inter`→Geist，heading `IBM Plex Sans`→Geist，mono `JetBrains Mono` | `index.css` `@theme inline`、`index.html` |
| 语义色 | `StatusBadge` 的 7 档 tone；`FIRING→danger`、`SUPPRESSED→warning`、`RECOVERED→success`、`OFFLINE→info` | `components/status-badge.tsx` |
| 阴影 | 现有体系**不用卡片投影**，靠 1px 边框 + 玻璃模糊表达层级 | `ui/card.tsx`、`list-frame.tsx` |
| 控件高度 | 表单/按钮 `h-8`(32px)；**禁止**在概览里把 Input/Select 改成 `h-9/h-10` | `web/AGENTS.md` |

概览页是这套语言的**延伸**，不新增圆角档、不新增阴影档、不新增品牌色。

### 3.2 两个候选方向

**方向 A — 机舱仪表盘（Calm Operations Console）｜推荐**

- 描述：冷静、低噪声的仪表盘。信息密度高，颜色克制：主色只用于"可点击/被选中"，语义色只出现在状态徽标与严重度标记上。数字是主角（等宽数字、`tabular-nums`）；卡片是"带 1px 边框的玻璃面"，没有投影、没有内渐变。
- 版式语言：KPI 用 `gap-px + bg-border` 的分隔条（现有历史分析页 `Metric` 行原样复用），内容卡用 `.glass-card`，区块标题 `text-base font-semibold tracking-tight`（与 `history-insights.tsx` 的 h2 一致）。
- 从方向 B 借一个元素：顶部一条**态势分割条**（firing / suppressed 两段 + 图例计数），给出"一眼看懂比例"的 NOC 感，但不做大屏式铺色。
- 风险：偏"朴素"，需要靠间距节奏和数字排版撑住质感。

**方向 B — NOC 大屏（Wallboard）**

- 描述：深色优先、整屏一块大曲线、严重度分色带、粗描边、脉冲呼吸动效。
- 优势：远距离可读，适合挂在值班室电视上。
- 与现网不兼容的点（因此不推荐作为 v1）：
  1. 默认主题是浅色（`ThemeProvider defaultTheme="system"`，浅色用户占多数），深色优先会让 `--app-gradient` 与玻璃卡片语言失效，等于给全站开第二套视觉。
  2. 严重度用整块颜色承载 → 违反"不以颜色为唯一编码"，需要额外的文字/图标补偿，收益打折。
  3. 动效需要 `prefers-reduced-motion` 分支与额外验收；而现有页面只有 skeleton 的 `animate-pulse`。
  4. 整屏大图强依赖图表库与统计口径，正是本阶段最不确定的部分（统计学家尚未定义趋势口径）。

**推荐：方向 A**，理由按权重排序：① 与现有列表页/表单页同一个视觉族，切换页面无"换产品"感；② 零新依赖、零新 token，落地风险最低；③ 深色模式天然成立（现有 token 已适配）；④ 把"图表/趋势"的不确定性留给数据可视化工程师，不被视觉方案倒逼着造口径。
若后续要上值班大屏，用**同一批卡片组件**加一个 `data-density="wallboard"` 变体（放大字号、隐藏辅助文案、强制深色），不重做一套界面。

### 3.3 配色规范（含实测对比度）

对比度用 WCAG 相对亮度公式实测（非估算）：

| 组合 | 比值 | 结论 |
| --- | --- | --- |
| `--foreground #0f172a` / `--background #f8fafc` | 17.06:1 | AA/AAA ✅ |
| `--muted-foreground #64748b` / `#f8fafc` | **4.55:1** | 刚过 AA 正文 ⚠️ 不可再调浅，不可用于 <12px |
| `--muted-foreground` / 白卡 `Card`（0.85 叠底 → `#fefeff`） | 4.72:1 | AA ✅ |
| `--muted-foreground` / `.glass-card`（`bg-card/80` → 实际 `#fdfdfe`） | 4.68:1 | AA ✅ |
| `--muted-foreground` / `bg-muted/20`（空态底） | 4.60:1 | AA ✅（勉强） |
| `--muted-foreground` / `bg-muted/50`（行 hover） | 4.51:1 | AA ✅（临界） |
| `--muted-foreground` / `bg-muted/70`、`bg-muted`（实底） | 4.43 / **4.34**:1 | ❌ 未达 AA |
| 白字 / `--primary #1677ff`（主按钮） | **4.10:1** | 未达 AA 正文 ⚠️ 现存问题 |
| `--primary #1677ff` 作文本 / `#f8fafc` | **3.92:1** | 未达 AA 正文；仅可用于 ≥18.66px 粗体或非文本图形 |
| `.dark` `#e2e8f0` / `#0b1120` | 15.27:1 | AAA ✅ |
| `.dark` `#94a3b8` / `#0b1120` | 7.34:1 | AAA ✅ |
| 语义标签（emerald/amber/red/sky/violet-700 于白底） | 5.48 / 5.02 / 6.47 / 5.93 / 7.10 : 1 | AA ✅ |

由此得到四条**概览页强制规则**：

1. **数值一律 `text-foreground`**；只有标签、口径说明、时间戳用 `text-muted-foreground`。数值绝不用主色。
2. **`text-muted-foreground` 不得放在 `bg-muted/70` 及更实的底上**（4.43:1 起不达 AA）。实底块里的小字改用 `text-foreground`；或把底色收到 `/20`–`/50`（4.60 / 4.51:1）。空态沿用现有 `bg-muted/20` 是对的，别"加深一点更好看"。
3. **主色不做小字号文本色**。链接/动作要么用 `text-foreground` + 下划线/箭头，要么新增一个更深的主色墨色 token。建议（需签字，属全局 token 增量）：
   ```css
   :root      { --primary-ink: #1163d6; }  /* 白底 5.56:1，浅底 5.31:1 */
   .dark      { --primary-ink: #69b1ff; }  /* 深底 8.37:1 */
   ```
   仅用于文本与图标，保留 `--primary` 作为填充/描边/图表色。v1 可只在概览页作用域内使用，稳定后再推广。
4. **状态色只来自 `StatusBadge`/`statusTone`**，不手写绿黄红。级别颜色来自告警字典（`AlertDictionaryItem.color`，`api/alert-dictionary/types.ts`），**事件本身不带颜色**——取不到字典就退化为中性条形 + 名称，不能自己编严重度色。

图表色：优先主题变量 `var(--chart-1..5)`（深浅色各有一套，`index.css` 已定义），超过 5 类再用 `lib/chart-colors.ts` 的 `chartColor(i)`。
⚠️ 技术债提示（不在本页修）：`--chart-3/4/5` 在深浅主题里是不同的色相，而 `MetricGraph` 用的是硬编码 `CHART_COLORS`，深色下两者不一致。概览页选 `--chart-*` 即可，等数据可视化工程师统一。

### 3.4 字体层级

全部落在现有字族与字号档上，不新增字号：

| 元素 | 字族 | 字号/行高 | 字重 | 颜色 |
| --- | --- | --- | --- | --- |
| 页面标题 | heading（IBM Plex Sans） | 24/32 `text-2xl` | 600 | foreground |
| 页面说明 | sans | 14/20 `text-sm` | 400 | muted-foreground |
| 区块标题（卡标题） | heading | 16/22 `text-base` | 500 | foreground |
| 卡片口径副标题 | sans | 12/16 `text-xs` | 400 | muted-foreground |
| KPI 标签 | sans | 12/16 | 500 | muted-foreground |
| KPI 数值 | heading | 30/36 `text-3xl`（<640 用 `text-2xl`） | 600 + `tabular-nums` | foreground |
| KPI 辅助行 | sans | 12/16 | 400 | muted-foreground |
| 清单主行 | sans | 14/20 `text-sm` | 500 | foreground |
| 清单副行 | sans | 12/16 | 400 | muted-foreground |
| 时间 / UID / 指纹 | mono | 12/16 + `tabular-nums` | 400–500 | muted-foreground |
| 徽标 | sans | 12/16 | 400（`StatusBadge` 默认） | tone |

规则：**数字永远等宽对齐**（`tabular-nums`）；**UID/指纹永远 mono**；中文标题不用 mono。

### 3.5 间距节奏

4px 基准，只使用现有档位：`4 / 8 / 12 / 16 / 24 / 32 / 48`。

| 位置 | 值 | 类 |
| --- | --- | --- |
| 页面内边距 | 16 / 24 | 由壳提供 `p-4 md:p-6`，本页不加 |
| 区块之间 | 24 | `gap-6` |
| 卡片之间 | 16 | `gap-4` |
| 卡片内边距 | 16 | 内容卡 `p-4`；`Card` 用默认 `--card-spacing: 16` |
| KPI 格内边距 | 12 / 16 | `px-4 py-3`（对齐历史分析 `Metric`） |
| KPI 格间隙 | 1px 分隔线 | `grid gap-px bg-border` |
| 卡片标题与内容 | 16 | `mb-4` |
| 清单行高 | 64 | `py-3` + 两行文本 |
| 行内元素 | 8 | `gap-2` |

**不按页面临时加减像素**（`SKILL.md`「间距」一节）。

### 3.6 深色主题

- 结构零差异，只靠 token 翻转。卡片 `rgba(17,24,39,.75)` + `backdrop-blur-sm` 在 `--app-gradient` 上呈玻璃质感。
- 分隔线 `border-border/70`（深色为 `rgba(148,163,184,.15 × .7)`）；**不要**用 `border-white/10` 之类手写色。
- 图表网格线用 `border-border`，坐标轴文字 `muted-foreground`（7.34:1，深色下比浅色更宽松）。
- 主色在深色底对比 4.59:1，可做图形与大字；文本仍走 `--primary-ink`。
- 强制要求：两套主题 × 三档断点（390 / 1024 / 1366+）都要真机截图核对，不接受"只验浅色"。

### 3.7 动效

- 只允许：hover 150ms 颜色过渡、skeleton `animate-pulse`、刷新时页面顶部 2px 进度发丝条（复用 `app-shell.tsx` 的 `session.switching` 呈现方式）。
- 禁止：KPI 数字滚动动画、告警行呼吸/闪烁、自动轮播。
- 所有过渡与动画必须能被 `motion-reduce` 关闭（现有代码已用 `motion-reduce:transition-none` 的写法，沿用）。

---

## 4. 核心组件清单与状态

### 4.1 组件树

```
OverviewPage                       页面编排、数据获取、URL 状态
├─ PageHeader (existing)           标题/说明 + 右侧控件
│  ├─ WindowSelect                 Select，h-8，1h/6h/24h/7d
│  ├─ AutoRefreshToggle            Button variant=outline + aria-pressed
│  └─ RefreshButton                Button variant=outline + RefreshCw
├─ OverviewScrollPanel             唯一滚动容器（min-h-0 flex-1 overflow-y-auto）
│  ├─ MetricStrip                  4 格 KPI，gap-px 分隔条
│  │  └─ MetricCard × 4            标签/数值/辅助/状态（loading/error/empty/unavailable）
│  ├─ SituationGrid                态势区
│  │  ├─ TriageCard                最近触发的告警（5 行 + 查看全部）
│  │  │  └─ TriageRow × 5          级别徽标 · 摘要 · 规则/数据源 · 时长 · 状态
│  │  ├─ CompositionCard           态势构成（firing / suppressed 分割条 + 图例）
│  │  └─ TodoCard                  待办（未读消息 / 待处理邀请）
│  ├─ LinkHealthGrid               链路区
│  │  ├─ NotificationCard          近 24h 投递成功/失败
│  │  └─ PlatformCard              后台进程（仅 system:user:read）
│  └─ EmptyWorkspace               无租户/零可见模块时的引导态
└─ CardShell                       卡壳：标题 + 口径副标题 + 右上角"查看全部" + 状态插槽
```

复用现有组件（不要重写）：`PageHeader`、`Card/CardHeader/CardTitle/CardDescription/CardContent`、
`Button`、`Select`、`Skeleton`、`Tooltip`、`Badge`、`StatusBadge`、`LevelBadge`、`Separator`、`FilterChips`（若做级别筛选条）、
`formatSpan`/`formatDuration`（`pages/alert/format.ts`）、`enumLabel`。

### 4.2 各组件规格

**CardShell（新，`pages/overview/card-shell.tsx`）**

| 属性 | 值 |
| --- | --- |
| 容器 | `.glass-card rounded-xl`，`p-4`，`flex flex-col gap-4 min-w-0` |
| 标题 | `text-base font-semibold tracking-tight`（`<h2>`） |
| 副标题 | `text-xs text-muted-foreground`，承载**口径说明**（如"近 24 小时 · 滚动窗口"） |
| 右上动作 | `Button variant="link" size="sm"` + `ArrowRight`，或图标按钮（`size-8` + `aria-label`） |
| 状态插槽 | `children` 之外提供 `state: "ready" | "loading" | "empty" | "error" | "unavailable"` |
| 语义 | `<section aria-labelledby>` 指向自身标题 id |

**MetricCard（新，`pages/overview/metric-strip.tsx`）**

| 项 | 规格 |
| --- | --- |
| 尺寸 | 高度由内容决定（≈104px），`px-4 py-3`，`bg-card`，父级 `gap-px bg-border` 形成 1px 分隔 |
| 结构 | 标签 → 数值（`text-3xl font-semibold tabular-nums`）→ 辅助行（`text-xs`，截断 + `title`） |
| 状态 | loading：骨架三条（16/32/12px 高，宽度递变）；error：`—` + 重试图标按钮；unavailable（无权限）：整格不渲染 |
| 交互 | 可点击时整格为 `<a>`/`<Link>`，`focus-visible:ring-2 ring-ring ring-offset-2`，hover `bg-muted/40` |
| 无障碍 | 数值变化不放在 live region（避免 30s 一次打断屏幕阅读器）；只由页面级"最近更新"承担播报 |

四个 KPI 槽位（**从左到右即优先级**）：

| # | 标题 | 值来源 | 辅助行 | 链接 |
| --- | --- | --- | --- | --- |
| 1 | 正在触发 | `alert-event-realtime-stats.firing` | "未抑制的 FIRING" | `/alerts` |
| 2 | 已抑制 | `.suppressed` | "仍在触发但已抑制" | `/alerts` |
| 3 | 累计已恢复 | `.recovered` | **必须写"累计（含归档）"** | `/alert-history` |
| 4 | 我的告警 | `.tabs[id=mine].count` | "我关注的告警页" | `/alerts` |

> 四个槽位全部依赖 `alert:read`，缺该权限时**整条 KPI 不渲染**，页面直接从态势区开始（不留空栅格）。

**TriageCard（新）**

- 标题「最近触发的告警」，口径副标题「按事件创建时间倒序 · 最多 5 条」。
  ⚠️ **不可写"最严重"**：`ListAlertEvents` 的 REALTIME 排序是 `created_at DESC`（`data/alert_event_repo.go:301`），
  服务端没有严重度优先排序；在**截断集上**再按严重度排会制造错误口径。
- 行结构（`<ul>` + `<li>`，整行 `<Link>`，高 64px，`hover:bg-muted/50`）：
  ```
  [级别徽标]  摘要（单行 truncate + title 给全量）
              规则名 · 数据源名 · 持续 {formatSpan(now - startsAt)}   [状态徽标]
  ```
  该清单是**列表**不是表格：不设固定列宽、不做横向滚动，窄屏只换行不溢出。
- 级别徽标：优先 `LevelBadge name={levelName}`；有字典色才上色（需要 `alert_rule:read` 拉字典），否则中性。
- 状态徽标：`StatusBadge value={event.status}` + 抑制中额外 `SUPPRESSED` 徽标（`event.suppressed`）。
- 底部：`查看全部 {total} 条 →`（`total` 取列表接口的 total，作为"总条数"元信息，允许出现）。
- 空态分两种（关键）：
  - 有规则、当前无事件 → "当前没有未恢复的告警" + 「查看历史告警」链接。
  - 无任何规则 → "还没有告警规则，先创建数据源与规则" + 两个 CTA（`/datasources`、`/alert-rules`），且遵循写权限：无 `alert_rule:write` 时只显示文案不显示"新建"按钮。
    **判定方式**：只在清单为空时才补发一次 `alert-rules?pageSize=1` 取 `total`（常态不多一个请求）。

**CompositionCard（新）**

- 只做 **firing / suppressed 二分割条**。依据后端谓词二者互斥且穷尽 `status = FIRING`：
  - firing：`status=FIRING AND (suppressed=false OR suppressed_until<=now)`
  - suppressed：`status=FIRING AND suppressed=true AND (suppressed_until IS NULL OR suppressed_until>now)`
- 条高 8px，`rounded-full`，两段 `bg-danger/70` 与 `bg-warning/70`（取自 `StatusBadge` tone 色系），段间 2px `bg-background` 间隙。
- 图例：`● 触发中 12 (75%)` `● 抑制中 4 (25%)`，总标题「未恢复告警 16」。
- **`recovered` 绝不进入这条分割条**（生命周期累计 vs 瞬时快照，混算无意义）。
- 数据为 0 时：条退化为 `bg-muted/60` 全宽 + "当前没有未恢复的告警"。
- 无障碍：容器 `role="img"` + `aria-label` 概述；下方附带 `<table class="sr-only">` 两行明细。

**NotificationCard（新）**

- 标题「通知投递」，口径副标题「近 24 小时 · 滚动窗口 · 按投递日志计」。
- 值：`listMessageLogs({status:"SENT"|"FAILED", startTime, endTime, pageSize:1})` 的 `total`。两个并行请求。
- **不显示"成功率"**：分母口径（任务 vs 投递、是否含重试、是否含 SENDING/PENDING）尚未定义 → 留给统计学家，先只展示可数的绝对量。
- 布局：两格 `成功 N` / `失败 M` + 失败为 0 时用 `text-success`，>0 用 `text-danger`；右下 `查看消息日志 →`。
- 无 `notification:read` → 整卡不渲染。

**PlatformCard（新，`system:user:read` 才渲染）**

- 数据：`listWorkerInstances()` → `onlineCount / busyCount / idleCount / observedAt`。
- 展示：Megrez / Merak 各一行"在线 x / 忙碌 y"，附 `observedAt` 本地时间；`onlineCount === 0` 时给出**危险提示**（"评测/通知进程离线，告警链路可能已中断"），这是本页最有价值的"静默故障"信号。
- 口径副标题「15 秒内上报心跳算在线」（与 `/workers` 文案一致，来自 proto 注释）。

**TodoCard（新）**

- 行：未读系统消息（`session.unreadMessages`，或 `getUnreadMessageCount()`）、待处理邀请（`listSelfInvitations()` 中 `status === "PENDING"` 计数）。
- 无待办时："没有待处理事项"（`text-muted-foreground`，不显示 0 行）。
- **不显示"待审 SSH 命令"**：`ListSSHCommands` 没有服务端 `reviewStatus` 过滤，页面是客户端筛的（`pages/ssh-command/index.tsx`），
  想拿准确计数必须翻遍所有分页 → 会把错误口径带上首页。要显示就必须先加服务端过滤或计数接口（见 §6 交付清单"可选后端项"）。

### 4.3 权限矩阵（决定渲染，而不是渲染后禁用）

| 条件 | 表现 |
| --- | --- |
| `session.namespaceUid === ""` | 只显示 `EmptyWorkspace`：「还没有可切换的租户」+ 创建/加入入口（复用 `namespace.emptyTitle` / `emptyHint` 语义） |
| 无 `alert:read` | KPI 1–4、TriageCard、CompositionCard 全部不渲染 |
| 无 `notification:read` | NotificationCard 不渲染 |
| 无 `system:user:read` | PlatformCard 不渲染 |
| 全部卡片不可见（例如仅 `member:read`） | 保留页头 + 一张说明卡："当前账号没有告警或通知的读权限，请联系管理员分配角色" + 「查看我的资料/角色」链接（**不要**跳 `/forbidden`，那不是错误） |
| 服务端 403（权限被中途收回） | 该卡显示 `unavailable` 态 + 文案"没有权限查看该模块"，其余卡片正常 |

**原则**：隐藏而不是禁用（`SKILL.md`「状态」一节：界面上不要露出无效按钮）；但服务端拒绝仍是最终真相，前端隐藏不代替鉴权。

### 4.4 状态矩阵（每个卡片都必须闭环）

| 状态 | 视觉 | 文案（i18n key） | 行为 |
| --- | --- | --- | --- |
| loading（首次） | 与真实内容等高的 `Skeleton`（KPI 3 条 / 清单 5 行 / 卡片 2 行），**不得高度跳变** | `state.loading` | — |
| refreshing（有旧数据） | 内容不变 + 顶部 2px 发丝条 + "更新中" | `overview.refreshing` | 禁用刷新按钮 |
| ready | 正常 | — | — |
| empty（无数据但有权限） | 虚线边框块 `border-dashed` + 主文案 + 下一步说明 | 各卡 `*.empty*` | 提供下一步链接 |
| empty（配置缺失，如无规则/无数据源） | 同上，但带 CTA 与写权限判断 | `*.emptySetup*` | 跳配置页 |
| error（本卡请求失败） | `role="alert"` + `text-destructive` + `Button variant="outline" size="sm"` 重试 | `*.loadFailed` | 仅重试本卡 |
| stale（刷新失败但有旧数据） | 卡片右上角 `Badge variant="secondary"`"数据可能已过期" + 上次成功时间 | `overview.stale` | 保留旧值，不闪空 |
| unavailable（无权限） | 见 §4.3 | `overview.noPermission` | — |
| truncated（仅清单类） | 底部一行合并呈现："仅显示最近 5 条，共 N 条 →"，不额外加提示条 | `overview.truncated` | 链接全量 |

**错误提示不要重复弹 toast**：`api/http.ts` 已对 5xx 统一 `toast.error`、404 走 global message，卡片内只写内联文案。
会话级失败（`session.failed`）已由 `AppShell` 整体替换为 `ServerErrorPage`，概览页不再二次处理。

### 4.5 交互细节

- KPI 与清单行是**链接**，不是按钮；`Cmd/Ctrl+点击` 新标签可用（用 `<Link>`/`<a>`，不用 `onClick` + `navigate`）。
- 深链接只用**已支持**的参数：`/alerts?level=<levelUid>`、`/alerts?name=<关键字>`、`/alert-history?eventLevel=&eventFrom=&eventTo=`。
  「已抑制」KPI 目前**无法**深链到过滤后的列表（实时告警故意不提供抑制/状态筛选，见 `web/AGENTS.md`）：因此该卡链接到 `/alerts`，辅助行写明"在列表内查看"，**不得**发明 `?suppression=` 参数。
- 键盘顺序：页头控件 → KPI 1–4 → 清单 5 行 → 查看全部 → 右列卡片。所有可聚焦元素有 `focus-visible:ring-2 ring-ring`。
- 行高 ≥ 56px，点按目标覆盖整行；图标按钮保持 `size-8`（不违反 `web/AGENTS.md` 的控件高度规则），但其功能均有等价的文本入口。

---

## 5. 技术栈选择

**总原则：能复用就不新增。v1 零新依赖。**

| 关注点 | 选择 | 理由 |
| --- | --- | --- |
| 框架 | React 19 + TypeScript（`~6.0.2`）+ react-router-dom 7 | 现网既有，路由分组已就绪（`App.tsx` 的 `AppShell` 嵌套路由） |
| 构建 | Vite 8 + `@vitejs/plugin-react` | 现网既有；dev 代理已把 `/v1` 转到 alioth（`vite.config.ts`） |
| 样式 | Tailwind v4 + `@tailwindcss/vite`，token 走 `@theme inline` | CSS 变量优先，`.dark` 类切换天然生效，无需再建一套主题系统 |
| 组件 | shadcn（`radix-nova`）+ `radix-ui` + `lucide-react` | 项目硬性约定"不引入 antd、不换 base-nova"；所需原语（Card/Table/Badge/Select/Skeleton/Tooltip）已安装 |
| 变体 | `class-variance-authority`（已在依赖） | 卡壳/徽标的变体用 cva，与 `button.tsx`/`tabs.tsx` 一致 |
| 图表 | **v1 不引入图表库** | 本页只需要"两条分割条 + 计数条"，CSS 足够；`echarts` 未安装且体积大，`recharts` 也未安装。若 v2 需要时间序列：先复用 `components/metric-graph.tsx` 的手写 SVG（零依赖），选型由数据可视化工程师定（ECharts 动态 import / uPlot 二选一） |
| 数据请求 | 现有 `@/api/http` 封装 + 新 hook `useOverviewResource` | 项目是手写 hook 风格（`use-paged-list`、`use-list-query`、`use-query-store`）；引入 react-query/SWR 会造成两套缓存模型并存 |
| 状态管理 | 局部 `useState` + 一个聚合 hook | 页面内状态很浅，不值得引入状态库 |
| URL 状态 | `useSearchParams` | 与 `/alerts`、`/alert-history` 的 URL 约定一致 |
| 国际化 | `react-i18next` + 在 `i18n/index.ts` 的 **zh 与 en 两个块**同时加 `overview` 组 | 现网要求全程 `t()`，且两语言键必须对齐 |
| 时间格式 | 复用 `pages/alert/format.ts` 的 `formatSpan` / `formatDuration`，日期用 `lib/date-time.ts` | 避免第三个时间格式化实现 |
| 无障碍 | 原生语义 + 现有原语 | Radix 已处理 Tabs/Select/Tooltip 的键盘与 ARIA |

**关于要不要新后端聚合接口 `/v1/overview`**

- **A 方案（v1 默认，纯前端）**：组合 5 类既有接口、最多 6 个请求（`alert-event-realtime-stats`、`alert-events`、`worker-instances`、`message-logs`×2、`self/invitations`；无 `notification:read` 时后者不请求）。优点：零后端改动、口径全部是已定稿契约；缺点：一次冷启动最多 6 个请求，且**无法**得到"按级别分布""按严重度 Top N"这类需要全量聚合的数字。
- **B 方案（推荐作为 v2，需后端 + 统计学家签字）**：`GET /v1/overview?window=24h` 单次返回租户级聚合。字段只能来自统计学家定义的口径；前端 `api/overview/` 先建好类型占位，切换时页面不改。
- 明确反对 C 方案：在前端翻页拉全量历史再聚合（`history-insights.tsx` 已因可视化需要这样做，带 5000 条截断告警）。仪表盘每 30 秒这么做会打爆数据库，且必然出现口径漂移。

---

## 6. 验收标准与开发交付清单

### 6.1 验收标准（逐条可在浏览器里验）

**布局与滚动**
1. 1366×768、已登录、有 `alert:read`：首屏完整可见 4 个 KPI 与 **≥5 行**告警清单，无需滚动。
2. 唯一滚动归属：`document.scrollingElement.scrollHeight === clientHeight`（壳内不出现文档级滚动）；概览滚动容器 `scrollHeight > clientHeight` 且能滚到"链路区"底部再滚回；期间页头与 KPI 不消失、不重叠。
3. 390×844：页面 `scrollWidth <= clientWidth`（无横向滚动）；KPI 为 2×2；页头控件换行不重叠；清单行文本折行不溢出。
4. 1024 / 1280 / 1440 三档：态势区在 `lg` 起两列、`xl` 起 1.6/1，右列（构成+待办）不与左列重叠。
5. 折叠侧栏为图标态（`data-collapsible=icon`）时，内容区宽度变化不产生横向滚动。

**数据与状态**
6. 断网/停后端：仅相关卡片显示 error + 重试，其他卡片正常；**不出现**整页白屏（会话级失败除外，走 `ServerErrorPage`）。
7. 单个接口 500（用 devtools 阻断某一路由）：该卡进入 error；其余卡保持 ready；无重复 toast（5xx 已由 `http.ts` 弹一次）。
8. 无数据：无告警时显示"当前没有未恢复的告警"；无规则时显示配置引导；无待办时不显示 0 行。
9. 刷新失败但已有数据：显示"数据可能已过期"徽标与上次成功时间，旧值不清零。
10. 30s 自动刷新：一次 tick 只发 2 个请求（stats + 清单）；切到后台标签页后停表，切回立即刷新一次。
11. 手动刷新按钮在请求中禁用；完成后"最近更新"时间前进。

**权限**
12. 仅 `alert:read` 的账号：通知卡与进程卡都不渲染，版面不留空洞。
13. `SYSTEM_ADMIN`：进程卡出现，`onlineCount=0` 时显示危险提示。
14. 无租户账号：只出现 `EmptyWorkspace`，不闪现告警卡片。
15. 仅有 `member:read` 的账号：出现"无读权限"说明卡，不跳 `/forbidden`。
16. 权限被运营中途收回后刷新：对应卡片进入 unavailable，其余正常。

**视觉与无障碍**
17. 浅色/深色 × 390/1024/1366 六张截图均无缝、无错位；深色下卡片边框、分隔线、骨架可见。
18. 数值为 `--foreground`；页面内**没有**主色小字文本（对照 §3.3 规则 3）；`text-muted-foreground` 不出现在 `bg-muted/70` 及更实的底上（规则 2）。
19. 检查器确认 KPI 数值 `font-variant-numeric: tabular-nums`；时间/UID 用 `JetBrains Mono`。
20. 键盘：Tab 可依次到达所有可点击卡片与行，焦点环可见；Enter 进入目标页；`Cmd+点击` 新标签打开。
21. 屏幕阅读器：KPI 数值变化不被 live region 播报；"最近更新"为 `aria-live="polite"`；构成条有 `role="img"` + 文案 + 隐藏表格。
22. `prefers-reduced-motion: reduce` 下无过渡/脉冲动画。
23. 缩放 200% 不出现内容截断与横向滚动（壳本体除外）。

**i18n 与代码质量**
24. 切换 en：页面无中文残留、无 `overview.*` 原始 key 暴露；zh/en 键数量一致（可写一次性脚本对比）。
25. `pnpm/npm run typecheck`、`lint`、`build` 全绿（`web/package.json` 现有脚本）。
26. 不引入新 npm 依赖（`package.json` 无 diff，除非确认采用 B 方案图表库）。

### 6.2 文件级改动清单

**新增（前端）**

| 文件 | 职责 |
| --- | --- |
| `web/src/pages/overview/index.tsx` | 页面编排、滚动归属、权限分支、刷新节奏 |
| `web/src/pages/overview/use-overview-resource.ts` | 通用资源 hook：`AbortController` + `refreshKey` + 旧值保留 + stale 标记 |
| `web/src/pages/overview/use-overview-data.ts` | 组合各资源、tick 调度、`namespaceUid` 变化重置 |
| `web/src/pages/overview/card-shell.tsx` | 卡壳（标题/口径副标题/动作/状态插槽） |
| `web/src/pages/overview/metric-strip.tsx` | KPI 条 + `MetricCard` |
| `web/src/pages/overview/cards/triage-card.tsx` | 最近触发的告警清单 |
| `web/src/pages/overview/cards/composition-card.tsx` | firing/suppressed 构成条 |
| `web/src/pages/overview/cards/notification-card.tsx` | 通知投递 |
| `web/src/pages/overview/cards/platform-card.tsx` | 后台进程 |
| `web/src/pages/overview/cards/todo-card.tsx` | 个人待办 |
| `web/src/pages/overview/empty-workspace.tsx` | 无租户 / 零可见模块 |
| `web/src/pages/overview/window.ts` | 时间窗 → `startTime/endTime`（滚动窗口；如需自然日，复用 `lib/date-range.ts` 的 `dayStartISO/dayEndISO`，保持 DST 安全） |
| `web/src/api/overview/types.ts`、`index.ts` | 仅在采用 B 方案时创建；A 方案不建（直接复用现有 api 模块） |

**修改（前端）**

| 文件 | 改动 | 风险 |
| --- | --- | --- |
| `web/src/App.tsx` | 加 `/overview` 路由；`/` 重定向改为 `/overview` | 改变所有用户落地页，需确认（见 §8 第 2 条） |
| `web/src/components/app-shell.tsx` | `links` 前置一条「概览」直链（`LayoutDashboard`，无权限码） | 触碰"一级只放领域"约定，需确认 |
| `web/src/i18n/index.ts` | zh + en 各加 `overview` 组（约 40 键） | 遗漏 en 会导致生产环境露出 key |
| `web/src/index.css` | **仅当批准时**加 `--primary-ink`（浅/深各一值） | 全局 token，需单独评审 |
| `web/src/lib/date-time.ts` | 可选：加 `formatRelative`（"3 分钟前"）；也可留在 overview 内部 | 避免第三个时间格式化入口 |

**文档（本次已产出）**

| 文件 | 说明 |
| --- | --- |
| `docs/dashboard-design.md` | 本文件 |
| `.cursor/skills/alioth-web-design/revisions.md` | 开发完成后**必须**追加一条「概览页」约定（`SKILL.md` 末尾"改样式之后"流程） |

**可选后端项（不在本阶段）**

- `GET /v1/overview`：一次性聚合。字段与口径由统计学家定义后再定 proto。
- `GET /v1/ssh-command-audits/count`（或等价的 `ListSSHCommands` 计数）：`reviewStatus` 过滤已随 SSH 命令列表落地，待办卡只差一个只读计数接口。
- 实时告警的服务端"严重度优先 + 时长优先"排序（若要"最该处理的 5 条"而不是"最近 5 条"）。

---

## 7. 与「统计学家」「数据可视化工程师」的接口约定

### 7.1 已确认可用的口径（本页可直接渲染，且必须在卡片上标注）

| 数字 | 定义（来自后端实现） | 标注要求 |
| --- | --- | --- |
| firing | `status=FIRING` 且 `(suppressed=false OR suppressed_until<=now)`；当前租户 | "未抑制的 FIRING" |
| suppressed | `status=FIRING` 且 `suppressed=true` 且 `(suppressed_until IS NULL OR > now)` | "仍在触发但已抑制" |
| recovered | 主库 `status != FIRING` 计数 + 归档索引计数 − 重叠；**生命周期累计** | **必须**写"累计（含归档）" |
| mine（我的告警） | 当前用户"我的告警" tab 下的实时事件数 | "我关注的告警" |
| worker online/busy | 15 秒内上报心跳 | "15 秒内心跳" |
| 投递成功/失败 | 投递日志（非通知任务）按 `status=SENT/FAILED` 的条数 | "按投递日志计 · 近 24 小时" |

### 7.2 必须由统计学家定义后才上页面的（本页先留空槽，不上假数）

1. **通知成功率**的分母：投递日志条数 vs 通知任务数；是否含重试（`retriesLeft`）、是否含 `SENDING`/`PENDING`。
2. **时间窗**：滚动窗口 vs 自然日；时区按浏览器本地还是租户固定。
3. **恢复类指标**的窗口：`recovered` 现在是累计值，"近 24 小时恢复数"需要新查询。
4. **级别分布**：按事件数、按指纹数、还是按"未恢复事件数"；是否含已抑制。
5. **"最该处理的 5 条"**的排序权重（严重度 / 持续时长 / 是否已抑制）。
6. 任何 **MTTR / 抑制覆盖率 / 误报率** 类派生指标 —— 本页 v1 一个都不出现。

### 7.3 与数据可视化工程师的分工

- 我（本文件）负责：容器尺寸与位置、留白、空/错/载态、颜色 token 来源、图例与标注的可读性、无障碍替代文本。
- 他们负责：图表类型、轴域与刻度、分箱、堆叠/分组、交互（hover/tooltip/brush）、序列语义。
- 预留接口：`<ChartSlot height="h-40|h-64" legend emptyText seriesLabel />`，内部实现可整体替换，页面布局不感知。
- 现状约束：本页 v1 无时间序列图；若 v2 引入，深色下请使用 `--chart-*` 而不是 `CHART_COLORS`（见 §3.3 技术债提示）。

---

## 8. 待确认问题（请主会话/用户裁决）

1. **侧栏「概览」是否为一级直链的例外**（与 `revisions.md` 2026-10-15 冲突）。推荐：是，且只此一个。
2. **`/` 落地页从 `/profile` 改为 `/overview`** 是否可接受（影响所有既有用户书签与肌肉记忆）。
3. **是否接受 `--primary-ink` 新 token**（修复主色文本 3.92:1 的 AA 问题）。推荐接受，先在概览页局部使用。
4. **v1 是否必须"级别分布"与"趋势图"**。若必须，则需要后端聚合接口 + 统计口径，工期与风险都会上一个台阶；本文默认 v1 不做。
5. **通知投递卡的时间窗默认值**（本文取 24h 滚动窗口），以及"成功率"是否 v1 就要。
6. **待审 SSH 命令**是否要进待办卡（需要后端加过滤/计数）。
7. 是否需要**值班大屏变体**（`data-density="wallboard"`），若有，可在 v1 之后用同一组件加变体实现。
