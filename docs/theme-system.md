# Alioth 多品牌主题系统（15 品牌 × 浅色/深色）

本文是可直接落地的技术方案：token 分层与命名、浅深两套变量文件、主题切换组件、文件清单与关键代码。落地顺序见第 8 节。

## 0. 落地状态（本轮）

**已实现并验证**（`web/` 下可直接运行）：

| 层 | 文件 | 说明 |
| --- | --- | --- |
| L0 品牌层 | `src/styles/theme/tokens.brand.css`（生成） | 15 套 `:root[data-brand]`，每套 11 级色阶 + 浅/深两套角色色（fill / fill-fg / hover / ink / ring） |
| L1 语义层 | `src/styles/theme/semantic.light.css`、`semantic.dark.css` | 品牌角色 → `--primary*` / `--ring`；中性色、状态色、图表色、材质（overlay/scrollbar/selection/gradient） |
| L2 组件层 | `src/index.css` | 三层 import + `@theme inline` 映射（新增 `--color-brand-*`、`--color-primary-ink` 等） |
| 状态 | `src/lib/appearance.ts` + `components/theme-provider.tsx` | `data-brand` + `light/dark` 类 + `color-scheme` + `theme-color`；`useLayoutEffect`；跨标签同步 |
| 防闪烁 | `index.html` | `<head>` 内联引导脚本（早于样式表与入口 module） |
| 切换 UI | `components/appearance-menu.tsx`、`brand-swatch.tsx` | 模式分段控件 + 15 张双模式色卡（3 列网格，超出视口高度时面板内部滚动）；radiogroup 语义、方向键/Home/End、`aria-live` |
| 语言入口 | `components/locale-menu.tsx`（原 `locale-theme-menu.tsx`） | 只保留语言，主题不再有两个入口 |
| 守卫 | `scripts/gen-brand-tokens.mjs --check`、`scripts/check-theme-boot.mjs` | 色值产物是否过期 + 引导脚本品牌列表是否与注册表一致 |

**验证结果**（详见第 8.3 节）：

- `npm run tokens:check` / `typecheck` / `lint` / `build` 全绿。
- 浏览器实测 15 品牌 × 2 模式 × 12 组语义对比度 **360 组全部达标**（最低 4.55:1）。
- 交互实测：切换即生效、Popover 保持打开、`data-brand` 落盘、`theme-color` 跟随、方向键移动并选中、Esc 关闭。

**本轮未做（明确的技术债）**：

1. **25 个文件里 169 处 Tailwind 调色板类**（`emerald-700`、`amber-500/10` 等）。它们都带 `dark:` 变体，换品牌不会串色，但绕开了语义层；这是第 6.2 节清单里的批量替换工作。
2. **14 张 unDraw 插图里 181 处 `#1677ff`**。`<img>` 引用的 SVG 读不到 CSS 变量，换品牌后插图仍是蓝色（登录页/403/404/502/503 最明显）。需要单独把 SVG 改成内联或构建期替换，属于独立工程。
3. **跨设备偏好同步**：`/v1/self/*` 没有偏好接口，目前仅 localStorage（与改动前一致）。
4. `scripts/check-colors.mjs` 硬编码守卫未落地（第 8.5 节）；现有调色板类会让它直接红，需先完成第 1 项。

> 另：仓库曾并行存在第二套方案（`docs/brand-tokens.md` 对应 10 级 / Python 实现，会定义同名 `--primary` 系列）。已合并为一套：运行时只保留本节这套，`web/src/styles/brand-tokens.css` 与 `scripts/brand-tokens/` 已移除；`docs/brand-tokens.md` 降级为设计依据（品牌×状态冲突矩阵、S1–S6、面积预算），仍值得在设计新页面时读。

配套产物：

- `web/scripts/gen-brand-tokens.mjs`：品牌色阶与角色色的生成器（含 WCAG 断言，断言失败退出码非 0）
- `web/src/styles/theme/tokens.brand.css`：15 套品牌色阶 + 角色色（生成产物，勿手改）
- `web/src/lib/brand-palette.generated.ts`：品牌注册表 + 切换器色卡用的预览色（生成产物，勿手改）
- `web/package.json`：新增 `tokens:gen`、`tokens:check`
- `docs/theme-preview.html`：15 品牌 × 双模式的静态预览页（直接用浏览器打开，不改动应用代码即可评审配色与切换器观感）

---

## 1. 现状核对（读代码得到的事实）

| 项 | 事实 | 证据 |
| --- | --- | --- |
| 技术栈 | React 19 + Vite 8 + TypeScript 6 + Tailwind CSS v4 + shadcn/ui（`radix-nova`，`radix-ui` 原生包）+ lucide | `web/package.json`、`web/components.json` |
| 主题实现 | **自研** `ThemeProvider`（`next-themes` 在依赖里但未被引用）；`localStorage.theme` 存 `light/dark/system`，`useEffect` 里改 `documentElement.classList` | `web/src/components/theme-provider.tsx` |
| token 现状 | 单一 `:root` / `.dark` 两段，`--primary` 恒为 `#1677ff`；`@theme inline` 把 40 余个 CSS 变量映射成 Tailwind 工具类 | `web/src/index.css:11-159` |
| 深浅变体 | `@custom-variant dark (&:is(.dark *))`，即 `.dark` 元素自身不匹配 `dark:`，只有后代匹配 | `web/src/index.css:6` |
| 切换入口 | 头部 `LocaleThemeMenu` 下拉（语言 + 三选一模式），位于顶栏右侧 | `web/src/components/locale-theme-menu.tsx`、`web/src/components/app-shell.tsx:545` |
| 持久化 | 仅 localStorage；`/v1/self/*` 没有偏好设置接口（`self` 只有 info/namespaces/permissions/change-*） | `web/src/api/self/index.ts` |
| 首帧防闪烁 | **缺失**：`index.html` 没有任何引导脚本，主题类名要等 React 挂载后才写入 | `web/index.html`、`main.tsx` |
| 硬编码色 | 169 处 Tailwind 调色板类（`emerald/amber/red/sky/violet/slate/…`），分布 25 个文件；`bg-black/10` 3 处遮罩；`text-white` 2 处 | 见 6.2 清单 |
| 十六进制 | 代码里 57 处：`code-editor.tsx` 40（Monaco 双主题，按模式二选一）、`chart-colors.ts` 8、`oncall-time-view.tsx` 8、`alert-dictionary` 3、`event-detail` 1（`i18n/index.ts` 4 处是文案里的示例色值，不是渲染色） | `grep -rn '#[0-9a-fA-F]\{6\}'` |
| 渐变 | `--app-gradient` 写死 `rgba(22,119,255,…)`，不跟随主题 | `web/src/index.css:100,148` |
| 图表色 | `CHART_COLORS` 8 个固定 hex，SVG 上以 `stroke={chartColor(index)}` 直接给出 | `web/src/lib/chart-colors.ts`、`components/metric-graph.tsx:302` |
| 语义色 | `--success/--warning/--danger/--info` 存在但只有 4 个文件在用（`detail-values`、`cluster/workloads-tab`、`cluster/pods-tab`、`message/index`）；状态标签实际用的是调色板类 + `dark:` 变体 | `web/src/components/status-badge.tsx:15-25` |
| 对比度现状 | 现主色 `#1677ff`：白字压底 **4.10:1**、作为文字压页面底 **3.92:1**，两者都不到 AA 的 4.5:1；`--ring` 同为 `#1677ff`，而 shadcn 用 `ring-ring/50` 又把它减半 | 实算，见 3.4 的新取值 |
| 结构化结论 | 现在的 token 结构里"品牌色"没有位置：`--primary` 既是填充又是文字又是焦点环，三者对对比度的要求并不相同 | 3.2 的 `-foreground` / `-ink` / `-ring` 拆分 |
| 布局约束 | `web/AGENTS.md` 的页签/滚动/表格/表单规范不受本次改动影响，但改 `ui/*` 组件颜色时必须遵守 | `web/AGENTS.md` |

结论：这是"单品牌 + 双模式"的实现，缺的是**品牌维度**（`data-brand`）、**品牌色阶**、**语义层与组件层的隔离**，以及**切换时的防闪烁**。

---

## 2. 目标与硬约束的对应关系

| 需求 | 方案落点 |
| --- | --- |
| 15 套品牌 × 浅/深 | `:root[data-brand="…"]` 提供 15 套原始色阶（品牌维度）；`semantic.light.css` / `semantic.dark.css` 提供 2 套角色映射（模式维度）。**15 + 2 个文件块，而不是 30 套硬编码** |
| 品牌色派生色阶 | OKLCH 生成器：固定色相、固定明度刻度、彩度锚定品牌色并按 sRGB 色域钳制；品牌基色原样落在最近色阶上 |
| 不得只用品牌色做唯一状态标识 | 状态语义色与品牌解耦（success/warning/danger/info/violet 固定色相）；状态组件一律"图标/文字 + 颜色"；色卡选中态用对勾 + 名称 + `aria-checked` 表达 |
| 切换持久化 | `localStorage.theme`（沿用现有键，不迁移）+ `localStorage.brand`；跨标签页 `storage` 同步；首帧由 `index.html` 内联脚本读取 |
| 所有页面组件 token 化 | 第 6 节清单 + `scripts/check-colors.mjs` 守卫（CI 拦截新增 hex/调色板类） |
| 防闪烁 | 内联引导脚本（CSS 前执行）+ `useLayoutEffect` + 切换期禁用过渡 + `color-scheme`/`theme-color` 同步 |
| 即时反馈 | 点击即写 DOM 属性（同帧生效）+ 选中色卡的确认动画（`animation`，不受"禁用过渡"影响）+ `aria-live` 播报 |

---

## 3. Token 分层与命名

### 3.1 三层

```
L0  品牌原始层    --brand-50 … --brand-950        每个品牌 11 阶，15 套
     └ 角色别名   --brand-fill / --brand-ink / …   每个品牌 10 个（浅/深各 5）
L1  语义层        --primary / --card / --border …  每个模式 1 套（2 套文件）
     └ 感官层     --success / --warning / --danger / --info / --violet
L2  组件层        @theme inline → Tailwind 工具类（bg-primary、text-primary-ink、border-border…）
     组件只允许用 L2，禁止出现 hex / rgb / 调色板类
```

组件永不直接引用 L0；L1 也不互相引用除 L0 之外的层。这样加品牌只改 L0（跑一次生成器），换模式只改 L1。

### 3.2 命名规范

| 前缀 | 含义 | 举例 | 允许被谁用 |
| --- | --- | --- | --- |
| `--brand-{50…950}` | 品牌色阶（原始） | `--brand-600` | 语义层、图表、装饰性强调 |
| `--brand-{fill,fill-fg,fill-hover,ink,ring}` | 品牌角色色（浅色模式，生成器按对比度自动选阶） | `--brand-fill` | 语义层 |
| `--brand-*-dark` | 同上，深色模式 | `--brand-fill-dark` | 语义层 |
| `--{primary,background,card,border,ring,…}` | 语义角色 | `--primary-ink` | `@theme inline` → 组件 |
| `--{success,warning,danger,info,violet}` | 语义状态色（可读色调，可作填充也可作文字） | `--danger` | 组件 |
| `--{success,warning,danger,info,violet}-{soft,border,foreground}` | 状态色配套的三件套 | `--warning-soft` | 组件 |
| `--chart-1…8` | 分类图表色 | `--chart-3` | 图表 |
| `--{sidebar,scrollbar,overlay,code,selection}-*` | 分区/材质 | `--sidebar-ring` | 组件 |

命名后缀约定（与现有代码保持一致）：

- `-foreground`：**该色块上的文字色**（`bg-primary` 配 `text-primary-foreground`）
- `-ink`：**在页面底色上作为文字/图标使用的品牌色**（`text-primary-ink`、`text-primary-soft-foreground`）
- `-soft`：低饱和底（12% 叠色）
- `-border`：边框/描边专用
- `-hover`：固态填充的 hover 值

> **为什么把 `-foreground` 和 `-ink` 分开**：品牌色当填充时要挑能和它达到 4.5:1 的文字（柠檬黄只能配近黑字），而当文字时要挑能在页面底色上达到 4.5:1 的色阶（柠檬黄本身做不到）。同一枚 `--primary` 无法同时满足这两件事，硬合并的结果必然有一侧不达 AA。现有代码里 36 处 `text-primary` 就是被合并的那个——迁移时按"文字/图标 → `text-primary-ink`，填充 → `bg-primary`"逐条替换。

### 3.3 品牌色阶派生算法

`web/scripts/gen-brand-tokens.mjs` 做的事：

1. 品牌 hex → OKLab → OKLCH，取 `(L0, C0, H)`，色相 `H` 全程固定（不跑偏到邻近色）。
2. 明度按固定刻度取 11 档：`50=.965 100=.935 200=.878 300=.808 400=.726 500=.648 600=.566 700=.486 800=.408 900=.330 950=.248`。
3. 彩度 `C(L) = min(C0 × amp(L), cap(L) × maxChroma(L, H))`：
   - `amp(L)`：向浅端放大最多 1.55 倍并衰减，向深端收敛到 0.28 倍——避免深色品牌的浅色阶灰掉、浅色品牌的深色阶糊掉；
   - `cap(L)`：50/100 档压制彩度（0.60 / 0.72），保证浅底 tint 不刺眼；
   - `maxChroma`：二分求该明度/色相下 sRGB 色域上限，保证输出一定在色域内。
4. 品牌基色**原样**覆盖到最接近的色阶（如 `#DB7A0E → --brand-500`、`#722169 → --brand-800`），保证设计稿与主题页可逐值核对。
5. 角色色按 WCAG 从色阶里挑，不是手填：
   - 浅色填充：从品牌所在阶向深/浅逐阶试，取第一个"填充 vs 自动前景 ≥ 4.5"的阶；前景在白/近黑中取对比更高者。
   - 浅色文字（`--brand-ink`）：从浅到深取第一个"vs `#ffffff` ≥ 4.5"的阶。
   - 深色填充：找 `L≈clamp(L0, 0.70, 0.79)` 的最近阶（深底上要能分离出来，且前景放得下）。
   - 深色文字（`--brand-ink-dark`）：取 300 → 400 → 200 中第一个"vs `#0b1120` ≥ 4.5"的阶。
   - 焦点环：直接复用 ink/ink-dark（≥ 3:1）。

生成器每次运行都断言 8 项对比度，失败即非 0 退出：

```
sunset      暖阳橙   浅色 fill#500  ink    5.81:1  ink#700   6.60:1  ring 6.31:1 | 深色 fill#400  ink 7.09:1  ink#300  10.05:1  ring 10.05:1
cornflower  矢车菊蓝  浅色 fill#500  ink    5.70:1  ink#700   6.43:1  ring 6.15:1 | 深色 fill#400  ink 7.34:1  ink#300  10.33:1  ring 10.33:1
jade        翠微    浅色 fill#600  white  4.55:1  ink#700   6.18:1  ring 5.91:1 | 深色 fill#400  ink 7.66:1  ink#300  10.69:1  ring 10.69:1
lagoon      碧青    浅色 fill#400  ink    7.35:1  ink#700   6.15:1  ring 5.88:1 | 深色 fill#400  ink 7.35:1  ink#300  10.69:1  ring 10.69:1
lemon       柠檬黄   浅色 fill#200  ink   11.57:1  ink#700   6.38:1  ring 6.10:1 | 深色 fill#300  ink 9.82:1  ink#300  10.32:1  ring 10.32:1
merlot      梅洛粉   浅色 fill#400  ink    6.12:1  ink#700   6.80:1  ring 6.50:1 | 深色 fill#400  ink 6.12:1  ink#300   9.90:1  ring 9.90:1
royal       皇家紫   浅色 fill#800  white  9.83:1  ink#800   9.83:1  ring 9.39:1 | 深色 fill#400  ink 6.87:1  ink#300   9.72:1  ring 9.72:1
dawn        拂晓蓝   浅色 fill#700  white  6.58:1  ink#700   6.58:1  ring 6.29:1 | 深色 fill#400  ink 7.36:1  ink#300  10.35:1  ring 10.35:1
cyan        明青    浅色 fill#400  ink    8.13:1  ink#700   6.12:1  ring 5.85:1 | 深色 fill#400  ink 8.13:1  ink#300  10.82:1  ring 10.82:1
purple      酱紫    浅色 fill#700  white  6.94:1  ink#700   6.94:1  ring 6.63:1 | 深色 fill#400  ink 7.06:1  ink#300  10.09:1  ring 10.09:1
geekblue    极客蓝   浅色 fill#700  white  5.85:1  ink#700   5.85:1  ring 5.59:1 | 深色 fill#400  ink 7.30:1  ink#300  10.29:1  ring 10.29:1
magenta     法式洋红  浅色 fill#500  ink    4.59:1  ink#700   7.10:1  ring 6.79:1 | 深色 fill#400  ink 6.72:1  ink#300   9.84:1  ring 9.84:1
marigold    金盏花   浅色 fill#300  ink    9.44:1  ink#700   6.44:1  ring 6.16:1 | 深色 fill#300  ink 9.44:1  ink#300   9.91:1  ring 9.91:1
lime        青柠    浅色 fill#300  ink   10.61:1  ink#700   6.18:1  ring 5.91:1 | 深色 fill#300  ink 10.61:1  ink#300  11.14:1  ring 11.14:1
dusk        日暮    浅色 fill#400  ink    7.54:1  ink#700   6.60:1  ring 6.31:1 | 深色 fill#400  ink 7.54:1  ink#300  10.05:1  ring 10.05:1
```

### 3.4 15 个品牌的角色色（生成结果）

具体色值以生成产物为准（`web/src/styles/theme/tokens.brand.css`），不要在这里抄一份会过期的表。
随时用 `cd web && node scripts/gen-brand-tokens.mjs` 打印当前结果（下面是一次真实输出）：

```
sunset      暖阳橙   浅色 fill#500  ink    5.81:1  ink#700   6.60:1  ring 6.31:1 | 深色 fill#400  ink 7.09:1  ink#300  10.05:1  ring 10.05:1
cornflower  矢车菊蓝  浅色 fill#500  ink    5.70:1  ink#700   6.43:1  ring 6.15:1 | 深色 fill#400  ink 7.34:1  ink#300  10.33:1  ring 10.33:1
jade        翠微    浅色 fill#600  white  4.55:1  ink#700   6.18:1  ring 5.91:1 | 深色 fill#400  ink 7.66:1  ink#300  10.69:1  ring 10.69:1
lagoon      碧青    浅色 fill#400  ink    7.35:1  ink#700   6.15:1  ring 5.88:1 | 深色 fill#400  ink 7.35:1  ink#300  10.69:1  ring 10.69:1
lemon       柠檬黄   浅色 fill#200  ink   11.57:1  ink#700   6.38:1  ring 6.10:1 | 深色 fill#300  ink 9.82:1  ink#300  10.32:1  ring 10.32:1
merlot      梅洛粉   浅色 fill#400  ink    6.12:1  ink#700   6.80:1  ring 6.50:1 | 深色 fill#400  ink 6.12:1  ink#300   9.90:1  ring 9.90:1
royal       皇家紫   浅色 fill#800  white  9.83:1  ink#800   9.83:1  ring 9.39:1 | 深色 fill#400  ink 6.87:1  ink#300   9.72:1  ring 9.72:1
```

读法：`fill#N` 是色阶档位、紧跟的 `ink/white` 是**填充上的前景字色**及其对比度；`ink#N` 是品牌当文字用的色阶及对底色的对比度；`ring` 是焦点环对底色的对比度。

三条已实现的取舍：

1. **浅色下 10 个品牌是"深色字压品牌底"**（暖阳橙、矢车菊蓝、碧青、柠檬黄、梅洛粉、明青、法式洋红、金盏花、青柠、日暮），只有翠微、皇家紫、拂晓蓝、酱紫、极客蓝配白字。这是对比度自动判定的结果——品牌色不改亮度、只换字色，比"把品牌压暗再配白字"更保品牌色。
2. **浅色 ink 从品牌档位往深走**取第一个 ≥5.2:1 的色阶（不是从 50 号往深扫）。所以皇家紫的链接色就是品牌原色 `#722169`（9.4:1），而不是更艳的 600 号玫红。
3. **深色 fill 提到 0.70–0.79 明度带，并把彩度钳回基色 ×1.15**。只提明度不钳彩度，皇家紫会变成荧光玫红 `#f366e1`；钳完是 `#e27cd2`，色相不变、不发荧光。

新增品牌只在 `BRANDS` 里加一行（如 `{ id: "dawn", zh: "拂晓蓝", en: "Dawn Blue", hex: "#1677ff" }`），跑 `npm run tokens:gen` 即可；色阶、角色色、注册表、`index.html` 引导脚本的品牌列表、预览页都会跟着更新，任何一条对比度断言不过就退出码非 0。

### 3.5 状态语义色（与品牌解耦）

状态色**不参与品牌轮换**，否则"失败=橙/黄/紫"会随主题变化，破坏语义。深浅各一套，全部通过"文字 vs 底色 ≥ 4.5、文字 vs soft 底 ≥ 4.5、填充 vs 前景 ≥ 4.5"校验：

| token | 浅色 | 深色 | 用途 |
| --- | --- | --- | --- |
| `--success` | `#047857` | `#34d399` | 成功、在线、已恢复 |
| `--warning` | `#a34a06` | `#fbbf24` | 待处理、被抑制 |
| `--danger` | `#b42318` | `#fca5a5` | 失败、触发中 |
| `--info` | `#0369a1` | `#38bdf8` | 进行中、中性提示 |
| `--violet` | `#6d28d9` | `#c4b5fd` | 类型区分（日志源等） |

深色模式的 `--danger` 取 `#fca5a5` 而不是 `#f87171`：`#f87171` 在 12% soft 底上只有 4.01:1，不达 AA。soft 底默认 12%，深色模式**不要**调到 16% 以上（会掉到 4.0 以下）。

浅色 `--warning` 从 `#b45309` 再压到 `#a34a06`：只要降 12% 的 soft 底，`#b45309` 就掉到 4.27:1（浏览器实测），压一档后回到 4.8 以上。**状态色当文字用时，永远要连 soft 底一起算**，不能只算白底。

实心红底（未读 badge 之类）不能直接用 `--destructive`/`--danger`——它们在深色下是浅红墨色，白字会掉到 2:1 以下。语义层为此单独提供了 `--danger-strong`（浅 `#b42318` / 深 `#dc2626`）配 `--danger-strong-foreground`（白），两模式都是 4.8:1 以上。

---

## 4. 变量文件结构（7 + 2）

### 4.1 目录

```
web/src/styles/theme/
├── tokens.brand.css        # L0：15 套品牌色阶 + 角色色（生成产物）
├── semantic.light.css      # L1：浅色角色映射（:root）
└── semantic.dark.css       # L1：深色角色映射（:root.dark）
```

`index.css` 只做入口与 L2 映射，不再持有颜色字面量：

```css
@import "tailwindcss";
@import "tw-animate-css";
@import "shadcn/tailwind.css";
@import "@fontsource-variable/geist";

/* L0 品牌层：15 套，按 html[data-brand] 命中 */
@import "./styles/theme/tokens.brand.css";
/* L1 语义层：浅/深两套 */
@import "./styles/theme/semantic.light.css";
@import "./styles/theme/semantic.dark.css";

@custom-variant dark (&:is(.dark *));

@theme inline {
  /* …保留现有映射，新增项见 4.5… */
}
```

> Tailwind v4 要求 `@import` 位于其它规则之前；三行主题 import 必须排在 `@import "tailwindcss"` 之后、`@custom-variant` 之前。品牌层与语义层的先后无所谓——它们只声明变量，解析发生在计算值阶段。

### 4.2 品牌层长什么样（生成，节选）

```css
/* 暖阳橙 Sunset Orange — #DB7A0E */
:root[data-brand="sunset"] {
  --brand-50: #faf2eb;
  --brand-100: #f9e5d7;
  --brand-200: #f6ceb0;
  --brand-300: #f8ad72;
  --brand-400: #f0891f;
  --brand-500: #db7a0e;
  --brand-600: #ac6012;
  --brand-700: #8c4d0d;
  --brand-800: #6e3b08;
  --brand-900: #502b06;
  --brand-950: #331a04;

  --brand-fill: #db7a0e;
  --brand-fill-fg: #171717;
  --brand-fill-hover: #ac6012;
  --brand-ink: #ac6012;
  --brand-ring: #ac6012;

  --brand-fill-dark: #f0891f;
  --brand-fill-fg-dark: #171717;
  --brand-fill-hover-dark: #f8ad72;
  --brand-ink-dark: #f8ad72;
  --brand-ring-dark: #f8ad72;
}
```

选择器用 `:root[data-brand="…"]`（特异性 0,2,0），不用 `html[data-brand]` 之后再去覆盖——品牌块之间互不重叠，无需 `:not()` 兜底。兜底品牌由引导脚本写入 `data-brand`，因此**不存在**"没有任何品牌块命中"的中间态。

### 4.3 浅色语义层（`semantic.light.css` 全文）

```css
/* 浅色模式角色映射：只引用 L0 变量，不写字面量除中性灰/材质外 */
:root {
  color-scheme: light;

  --background: #f8fafc;
  --foreground: #0f172a;
  --card: rgb(255 255 255 / 0.85);
  --card-foreground: #0f172a;
  --popover: #ffffff;
  --popover-foreground: #0f172a;

  /* 品牌角色 */
  --primary: var(--brand-fill);
  --primary-foreground: var(--brand-fill-fg);
  --primary-hover: var(--brand-fill-hover);
  --primary-ink: var(--brand-ink);
  --primary-soft: color-mix(in oklab, var(--primary) 12%, var(--background));
  --primary-soft-foreground: var(--primary-ink);
  --ring: var(--brand-ring);

  --secondary: #f1f5f9;
  --secondary-foreground: #0f172a;
  --muted: #f1f5f9;
  --muted-foreground: #64748b;   /* 4.55:1 @ page, 4.76:1 @ card */
  --accent: #e2e8f0;
  --accent-foreground: var(--primary-ink);

  /* 状态语义 */
  --success: #047857;
  --success-foreground: #ffffff;
  --success-soft: color-mix(in oklab, var(--success) 12%, var(--background));
  --success-border: color-mix(in oklab, var(--success) 32%, transparent);

  --warning: #b45309;
  --warning-foreground: #ffffff;
  --warning-soft: color-mix(in oklab, var(--warning) 12%, var(--background));
  --warning-border: color-mix(in oklab, var(--warning) 32%, transparent);

  --danger: #b42318;
  --danger-foreground: #ffffff;
  --danger-soft: color-mix(in oklab, var(--danger) 12%, var(--background));
  --danger-border: color-mix(in oklab, var(--danger) 32%, transparent);
  --destructive: var(--danger);
  --destructive-foreground: var(--danger-foreground);
  --error: var(--danger);

  --info: #0369a1;
  --info-foreground: #ffffff;
  --info-soft: color-mix(in oklab, var(--info) 12%, var(--background));
  --info-border: color-mix(in oklab, var(--info) 32%, transparent);

  --violet: #6d28d9;
  --violet-foreground: #ffffff;
  --violet-soft: color-mix(in oklab, var(--violet) 12%, var(--background));
  --violet-border: color-mix(in oklab, var(--violet) 32%, transparent);

  /* 结构与材质 */
  --border: rgb(15 23 42 / 0.1);
  --input: rgb(15 23 42 / 0.12);
  --overlay: rgb(15 23 42 / 0.32);
  --selection-bg: color-mix(in oklab, var(--primary) 22%, transparent);
  --scrollbar-thumb: rgb(15 23 42 / 0.2);
  --scrollbar-thumb-hover: rgb(15 23 42 / 0.32);

  --sidebar: #ffffff;
  --sidebar-foreground: #0f172a;
  --sidebar-primary: var(--primary);
  --sidebar-primary-foreground: var(--primary-foreground);
  --sidebar-accent: #f1f5f9;
  --sidebar-accent-foreground: var(--primary-ink);
  --sidebar-border: rgb(15 23 42 / 0.08);
  --sidebar-ring: var(--ring);

  /* 分类图表色：与背景 ≥ 3:1，且色相两两拉开（红绿色弱下仍可区分顺序） */
  --chart-1: #2563eb;
  --chart-2: #0e9f6e;
  --chart-3: #b45309;
  --chart-4: #be185d;
  --chart-5: #0e7490;
  --chart-6: #6d28d9;
  --chart-7: #c2410c;
  --chart-8: #4338ca;

  --app-gradient:
    radial-gradient(ellipse 80% 50% at 50% -20%, color-mix(in oklab, var(--primary) 12%, transparent), transparent),
    radial-gradient(ellipse 60% 40% at 100% 0%, color-mix(in oklab, var(--brand-400) 10%, transparent), transparent);

  --code-bg: #ffffff;
  --code-fg: #24292e;
}
```

### 4.4 深色语义层（`semantic.dark.css` 全文）

```css
/* 深色模式角色映射 */
:root.dark {
  color-scheme: dark;

  --background: #0b1120;
  --foreground: #e2e8f0;
  --card: rgb(17 24 39 / 0.75);
  --card-foreground: #e2e8f0;
  --popover: #111827;
  --popover-foreground: #e2e8f0;

  --primary: var(--brand-fill-dark);
  --primary-foreground: var(--brand-fill-fg-dark);
  --primary-hover: var(--brand-fill-hover-dark);
  --primary-ink: var(--brand-ink-dark);
  --primary-soft: color-mix(in oklab, var(--primary) 16%, var(--background));
  --primary-soft-foreground: var(--primary-ink);
  --ring: var(--brand-ring-dark);

  --secondary: #1e293b;
  --secondary-foreground: #e2e8f0;
  --muted: #1e293b;
  --muted-foreground: #94a3b8;   /* 7.34:1 @ page, 6.92:1 @ card */
  --accent: #1e293b;
  --accent-foreground: var(--primary-ink);

  --success: #34d399;
  --success-foreground: #04120c;
  --success-soft: color-mix(in oklab, var(--success) 12%, var(--background));
  --success-border: color-mix(in oklab, var(--success) 32%, transparent);

  --warning: #fbbf24;
  --warning-foreground: #1a1206;
  --warning-soft: color-mix(in oklab, var(--warning) 12%, var(--background));
  --warning-border: color-mix(in oklab, var(--warning) 32%, transparent);

  --danger: #fca5a5;             /* 不是 #f87171：后者在 soft 底上只有 4.01:1 */
  --danger-foreground: #2a0d0d;
  --danger-soft: color-mix(in oklab, var(--danger) 12%, var(--background));
  --danger-border: color-mix(in oklab, var(--danger) 32%, transparent);
  --destructive: var(--danger);
  --destructive-foreground: var(--danger-foreground);
  --error: var(--danger);

  --info: #38bdf8;
  --info-foreground: #04121c;
  --info-soft: color-mix(in oklab, var(--info) 12%, var(--background));
  --info-border: color-mix(in oklab, var(--info) 32%, transparent);

  --violet: #c4b5fd;
  --violet-foreground: #12081f;
  --violet-soft: color-mix(in oklab, var(--violet) 12%, var(--background));
  --violet-border: color-mix(in oklab, var(--violet) 32%, transparent);

  --border: rgb(148 163 184 / 0.15);
  --input: rgb(148 163 184 / 0.2);
  --overlay: rgb(2 6 23 / 0.62);
  --selection-bg: color-mix(in oklab, var(--primary) 28%, transparent);
  --scrollbar-thumb: rgb(226 232 240 / 0.2);
  --scrollbar-thumb-hover: rgb(226 232 240 / 0.32);

  --sidebar: #0f172a;
  --sidebar-foreground: #e2e8f0;
  --sidebar-primary: var(--primary);
  --sidebar-primary-foreground: var(--primary-foreground);
  --sidebar-accent: #1e293b;
  --sidebar-accent-foreground: var(--primary-ink);
  --sidebar-border: rgb(148 163 184 / 0.12);
  --sidebar-ring: var(--ring);

  --chart-1: #60a5fa;
  --chart-2: #34d399;
  --chart-3: #fbbf24;
  --chart-4: #f472b6;
  --chart-5: #22d3ee;
  --chart-6: #a78bfa;
  --chart-7: #fb923c;
  --chart-8: #818cf8;

  --app-gradient:
    radial-gradient(ellipse 80% 50% at 50% -20%, color-mix(in oklab, var(--primary) 10%, transparent), transparent),
    radial-gradient(ellipse 60% 40% at 100% 0%, color-mix(in oklab, var(--brand-400) 8%, transparent), transparent);

  --code-bg: #24292e;
  --code-fg: #e1e4e8;
}
```

**浅色块不能省 `:root`、深色块必须是 `:root.dark`**：特异性 0,2,0 > 0,1,0，模式文件先后顺序不再影响结果；也不会出现"切到深色后某些变量仍旧值"。

### 4.5 `@theme inline` 的增量

```css
@theme inline {
  /* 品牌色阶 → bg-brand-500 / text-brand-700 / border-brand-200 … */
  --color-brand-50: var(--brand-50);
  --color-brand-100: var(--brand-100);
  --color-brand-200: var(--brand-200);
  --color-brand-300: var(--brand-300);
  --color-brand-400: var(--brand-400);
  --color-brand-500: var(--brand-500);
  --color-brand-600: var(--brand-600);
  --color-brand-700: var(--brand-700);
  --color-brand-800: var(--brand-800);
  --color-brand-900: var(--brand-900);
  --color-brand-950: var(--brand-950);

  /* 品牌文字/底/环 */
  --color-primary-ink: var(--primary-ink);
  --color-primary-hover: var(--primary-hover);
  --color-primary-soft: var(--primary-soft);
  --color-primary-soft-foreground: var(--primary-soft-foreground);

  /* 状态三件套：text-success / bg-success-soft / border-success-border */
  --color-success-soft: var(--success-soft);
  --color-success-border: var(--success-border);
  --color-warning-soft: var(--warning-soft);
  --color-warning-border: var(--warning-border);
  --color-danger-soft: var(--danger-soft);
  --color-danger-border: var(--danger-border);
  --color-info-soft: var(--info-soft);
  --color-info-border: var(--info-border);
  --color-violet: var(--violet);
  --color-violet-foreground: var(--violet-foreground);
  --color-violet-soft: var(--violet-soft);
  --color-violet-border: var(--violet-border);

  /* 材质 */
  --color-overlay: var(--overlay);
  --color-chart-6: var(--chart-6);
  --color-chart-7: var(--chart-7);
  --color-chart-8: var(--chart-8);
}
```

现有 `--color-primary`、`--color-card`、`--color-sidebar-*`、`--color-chart-1..5` 等映射保留不动（`chart-1..5` 的**值**由语义文件重新定义，语义文件里已扩到 8 个）。`--color-error` / `--color-info` 保留。

---

## 5. 主题切换组件

### 5.1 结构与交互

```
顶栏按钮（icon）
 ├─ 模式图标（Sun / Moon / SunMoon）
 └─ 右下角品牌色圆点（padding-box ring，不靠颜色单独表意：图标本身已区分模式）
Popover（w-[19rem]）
 ├─ 「模式」分段控件   浅色 / 深色 / 跟随系统   role=radiogroup
 ├─ 分隔线
 ├─ 「品牌主题」15 张色卡 3 列网格   role=radiogroup
 │   └─ 每张：上浅下深两段预览 + 名称 + 选中对勾（aria-checked）
 └─ 隐藏的 aria-live 区域：「已切换到 翠微 · 深色」
```

要点：

- **色卡预览必须用品牌自己的字面值**（来自 `brand-palette.generated.ts`），不能引用当前 token，否则 15 张卡片会跟着当前主题一起变。
- 选中态 = 2px ring + 对勾图标 + 名称加粗 + `aria-checked`，**不只靠颜色**。
- 点击即应用（不点"确定"），Popover 保持打开便于对比；Esc/外部点击关闭。
- 键盘：分段控件 ←/→ 切换并即时选中；色卡 3 列网格 ←/→/↑/↓ 移动焦点（±1 / ±3），Enter/Space 选中，Home/End 到首尾。焦点环用 `focus-visible:ring-2 ring-ring`。
- 确认动画用 `animation`（keyframes）而不是 `transition`：切换瞬间会全局禁用 transition，`animation` 不受影响。

### 5.2 品牌注册表与状态（`web/src/lib/appearance.ts`）

```ts
import { BRAND_IDS, type BrandId } from "@/lib/brand-palette.generated"

export type ThemeMode = "light" | "dark" | "system"
export type ResolvedMode = "light" | "dark"

/** 沿用既有键名，老用户的 theme 偏好不会丢；brand 是新键。 */
export const STORAGE_KEYS = { mode: "theme", brand: "brand" } as const
export const DEFAULT_BRAND: BrandId = "sunset"
export const DEFAULT_MODE: ThemeMode = "system"

export function isThemeMode(value: unknown): value is ThemeMode {
  return value === "light" || value === "dark" || value === "system"
}

export function isBrandId(value: unknown): value is BrandId {
  return typeof value === "string" && (BRAND_IDS as readonly string[]).includes(value)
}

export function systemMode(): ResolvedMode {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light"
}

/** 首选品牌与模式，供 index.html 的引导脚本读取（键名必须与这里一致）。 */
export function readStoredAppearance(): { mode: ThemeMode; brand: BrandId } {
  const mode = localStorage.getItem(STORAGE_KEYS.mode)
  const brand = localStorage.getItem(STORAGE_KEYS.brand)
  return {
    mode: isThemeMode(mode) ? mode : DEFAULT_MODE,
    brand: isBrandId(brand) ? brand : DEFAULT_BRAND,
  }
}
```

### 5.3 Provider 改动（`web/src/components/theme-provider.tsx`）

在现有实现上加品牌维度，其余语义（`theme/setTheme/disableTransitionOnChange`）保持不变，避免牵连 45 处调用点：

```tsx
type ThemeProviderState = {
  theme: ThemeMode
  resolvedTheme: ResolvedMode
  setTheme: (theme: ThemeMode) => void
  brand: BrandId
  setBrand: (brand: BrandId) => void
}

const THEME_COLOR_META = 'meta[name="theme-color"]'

/** 把状态写到 <html>：类名、data-brand、color-scheme、theme-color。 */
function syncDocument(theme: ThemeMode, brand: BrandId) {
  const root = document.documentElement
  const resolved = theme === "system" ? systemMode() : theme

  root.classList.toggle("dark", resolved === "dark")
  root.classList.toggle("light", resolved === "light")
  root.dataset.brand = brand
  root.style.colorScheme = resolved
  // 引导脚本为了消白闪写过内联底色，挂载后交还给 token。
  root.style.removeProperty("background-color")

  requestAnimationFrame(() => {
    const background = getComputedStyle(root).getPropertyValue("--background").trim()
    const meta = document.querySelector<HTMLMetaElement>(THEME_COLOR_META)
    if (meta && background) meta.content = background
  })
}

export function ThemeProvider({ children, disableTransitionOnChange = true }: ThemeProviderProps) {
  const [theme, setThemeState] = React.useState<ThemeMode>(() => readStoredAppearance().mode)
  const [brand, setBrandState] = React.useState<BrandId>(() => readStoredAppearance().brand)

  const apply = React.useCallback(
    (nextTheme: ThemeMode, nextBrand: BrandId) => {
      const restore = disableTransitionOnChange ? disableTransitionsTemporarily() : null
      syncDocument(nextTheme, nextBrand)
      restore?.()
    },
    [disableTransitionOnChange]
  )

  // useLayoutEffect：在浏览器绘制前落类名，避免二次闪
  React.useLayoutEffect(() => apply(theme, brand), [apply, theme, brand])

  const setTheme = React.useCallback((next: ThemeMode) => {
    localStorage.setItem(STORAGE_KEYS.mode, next)
    setThemeState(next)
  }, [])

  const setBrand = React.useCallback((next: BrandId) => {
    localStorage.setItem(STORAGE_KEYS.brand, next)
    setBrandState(next)
  }, [])

  // 系统模式变化
  React.useEffect(() => {
    if (theme !== "system") return
    const query = window.matchMedia("(prefers-color-scheme: dark)")
    const onChange = () => apply("system", brand)
    query.addEventListener("change", onChange)
    return () => query.removeEventListener("change", onChange)
  }, [apply, brand, theme])

  // 跨标签页同步：两个键都要听
  React.useEffect(() => {
    const onStorage = (event: StorageEvent) => {
      if (event.storageArea !== localStorage) return
      if (event.key === STORAGE_KEYS.mode) {
        setThemeState(isThemeMode(event.newValue) ? event.newValue : DEFAULT_MODE)
      }
      if (event.key === STORAGE_KEYS.brand) {
        setBrandState(isBrandId(event.newValue) ? event.newValue : DEFAULT_BRAND)
      }
    }
    window.addEventListener("storage", onStorage)
    return () => window.removeEventListener("storage", onStorage)
  }, [])

  const value = React.useMemo(
    () => ({ theme, resolvedTheme: theme === "system" ? systemMode() : theme, setTheme, brand, setBrand }),
    [brand, setBrand, setTheme, theme]
  )

  return <ThemeProviderContext.Provider value={value}>{children}</ThemeProviderContext.Provider>
}
```

注意 `resolvedTheme` 在渲染函数里读 `matchMedia` 会有 SSR 隐患，本项目纯 CSR，可接受；若要更严谨，用 `useSyncExternalStore` 订阅媒体查询（第 9 节列为后续优化）。

### 5.4 防闪烁时序

```
HTML 解析
  └─ <head> 内联脚本（阻塞，晚于 localStorage 可读、早于 CSS/JS 请求）
       ├─ 读 theme/brand → 决定 dark
       ├─ html.classList.add('dark'|'light')
       ├─ html.dataset.brand = 规范化后的品牌 id
       ├─ html.style.colorScheme = dark?'dark':'light'   ← 原生滚动条/表单控件立即正确
       └─ html.style.backgroundColor = dark?'#0b1120':'#f8fafc'  ← 消白闪
  └─ CSS 解析（token 就位）
  └─ React 挂载 → useLayoutEffect 再同步一次（校正 + 清内联底色 + 更新 theme-color）
```

`web/index.html`（`<head>` 内，放在 `<link rel="icon">` 之后即可）：

```html
<meta name="theme-color" content="#f8fafc" />
<script>
  // 首帧防闪烁：必须在样式表与 React 之前同步执行，只能内联。
  ;(() => {
    try {
      const root = document.documentElement
      const mode = localStorage.getItem("theme")
      const stored = localStorage.getItem("brand")
      const brands = ["sunset", "cornflower", "jade", "lagoon", "lemon", "merlot", "royal"]
      const brand = brands.includes(stored) ? stored : "sunset"
      const dark =
        mode === "dark" ||
        ((mode === null || mode === "system") &&
          window.matchMedia("(prefers-color-scheme: dark)").matches)

      root.classList.add(dark ? "dark" : "light")
      root.dataset.brand = brand
      root.style.colorScheme = dark ? "dark" : "light"
      // 占位底色，挂载后由 ThemeProvider 清掉
      root.style.backgroundColor = dark ? "#0b1120" : "#f8fafc"
    } catch {
      // 隐私模式下 localStorage 会抛错，忽略即可（回落默认品牌/跟随系统）
    }
  })()
</script>
```

> 品牌 id 列表在引导脚本里硬编码一次是刻意的：它不能 import 模块（会变成额外请求，失去"首帧"意义）。`tokens:check` 之外再加一条 grep 守卫，保证脚本里的数组与 `BRAND_IDS` 一致（见 8.5）。

另外把这几个当前写死的颜色一起收进 token：

```css
/* index.css base 层 */
::selection { background: var(--selection-bg); }
body { background-image: var(--app-gradient); }
* { scrollbar-color: var(--scrollbar-thumb) transparent; }
*::-webkit-scrollbar-thumb { background-color: var(--scrollbar-thumb); }
```

### 5.5 切换器组件（`web/src/components/appearance-menu.tsx`）

```tsx
import { Check, Moon, Sun, SunMoon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { BrandSwatch } from "@/components/brand-swatch"
import { useTheme } from "@/components/theme-provider"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import { BRAND_IDS, BRAND_PALETTES, type BrandId } from "@/lib/brand-palette.generated"
import { cn } from "@/lib/class-names"

const MODES = [
  { value: "light", icon: Sun, labelKey: "appearance.light" },
  { value: "dark", icon: Moon, labelKey: "appearance.dark" },
  { value: "system", icon: SunMoon, labelKey: "appearance.system" },
] as const

export function AppearanceMenu() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme, brand, setBrand } = useTheme()
  const [announcement, setAnnouncement] = useState("")
  const zh = i18n.language.startsWith("zh")
  const labelOf = (id: BrandId) => (zh ? BRAND_PALETTES[id].labelZh : BRAND_PALETTES[id].labelEn)
  const TriggerIcon = MODES.find((mode) => mode.value === theme)?.icon ?? SunMoon

  function pickBrand(next: BrandId) {
    if (next === brand) return
    setBrand(next)
    setAnnouncement(t("appearance.switched", { brand: labelOf(next) }))
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative shrink-0" aria-label={t("appearance.label")}>
          <TriggerIcon />
          {/* 右下角品牌点：颜色是附加信息，模式由图标表达 */}
          <span
            aria-hidden
            className="absolute right-1 bottom-1 size-1.5 rounded-full ring-2 ring-background"
            style={{ backgroundColor: "var(--brand-500)" }}
          />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[19rem] p-3">
        <p className="px-1 pb-2 text-xs font-medium text-muted-foreground">{t("appearance.mode")}</p>
        <div role="radiogroup" aria-label={t("appearance.mode")} className="grid grid-cols-3 gap-1 rounded-lg bg-muted p-1">
          {MODES.map((mode) => {
            const active = theme === mode.value
            return (
              <button
                key={mode.value}
                type="button"
                role="radio"
                aria-checked={active}
                onClick={() => setTheme(mode.value)}
                className={cn(
                  "flex h-8 items-center justify-center gap-1.5 rounded-md text-xs font-medium transition-colors",
                  "focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                  active ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
                )}
              >
                <mode.icon className="size-3.5" aria-hidden />
                {t(mode.labelKey)}
              </button>
            )
          })}
        </div>

        <Separator className="my-3" />

        <p className="px-1 pb-2 text-xs font-medium text-muted-foreground">{t("appearance.brand")}</p>
        <div role="radiogroup" aria-label={t("appearance.brand")} className="grid grid-cols-3 gap-2">
          {BRAND_IDS.map((id) => (
            <button
              key={id}
              type="button"
              role="radio"
              aria-checked={brand === id}
              onClick={() => pickBrand(id)}
              className={cn(
                "group relative rounded-lg p-1 text-left",
                "focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                brand === id && "ring-2 ring-primary ring-offset-2 ring-offset-background"
              )}
            >
              <BrandSwatch palette={BRAND_PALETTES[id]} />
              <span className="mt-1 flex items-center gap-1 px-0.5 text-[11px] leading-4">
                {labelOf(id)}
                {brand === id ? <Check className="size-3 text-primary-ink" aria-hidden /> : null}
              </span>
              {announcement && brand === id ? (
                <span aria-hidden className="pointer-events-none absolute inset-0 rounded-lg animate-[brand-confirm_420ms_ease-out]" />
              ) : null}
            </button>
          ))}
        </div>

        <p aria-live="polite" className="sr-only">{announcement}</p>
      </PopoverContent>
    </Popover>
  )
}
```

`brand-swatch.tsx`（预览用品牌自身色值，浅深两段并排展示"这套主题在两边的样子"）：

```tsx
import type { BrandPalette } from "@/lib/brand-palette.generated"

/** 预览色卡：刻意使用品牌字面值，避免跟随当前主题变化。 */
export function BrandSwatch({ palette }: { palette: BrandPalette }) {
  return (
    <span className="block overflow-hidden rounded-md border border-border/60">
      {[palette.previewLight, palette.previewDark].map((tone, index) => (
        <span key={index} className="flex h-4 items-center gap-1 px-1.5" style={{ backgroundColor: tone.bg }}>
          <span className="h-2 flex-1 rounded-[2px]" style={{ backgroundColor: tone.card }} />
          <span className="size-2 rounded-[2px]" style={{ backgroundColor: tone.fill }} />
          <span className="h-2 w-3 rounded-[2px]" style={{ backgroundColor: tone.tint }} />
        </span>
      ))}
    </span>
  )
}
```

`index.css` 里的确认动画：

```css
@keyframes brand-confirm {
  0%   { box-shadow: 0 0 0 0 color-mix(in oklab, var(--primary) 55%, transparent); }
  100% { box-shadow: 0 0 0 10px color-mix(in oklab, var(--primary) 0%, transparent); }
}
@media (prefers-reduced-motion: reduce) {
  .animate-\[brand-confirm_420ms_ease-out\] { animation: none; }
}
```

顶栏接入（`app-shell.tsx:545`）：`<LocaleThemeMenu />` → `<AppearanceMenu />`；`locale-theme-menu.tsx` 只保留语言下拉（或改名 `locale-menu.tsx`），避免两个地方都能改主题。

---

## 6. 需要改动的文件清单

### 6.1 新增

| 文件 | 内容 | 状态 |
| --- | --- | --- |
| `web/scripts/gen-brand-tokens.mjs` | 色阶/角色色生成 + 对比度断言 + `--check` | ✅ 已加 |
| `web/src/styles/theme/tokens.brand.css` | 15 套品牌色阶 + 角色色 | ✅ 已生成 |
| `web/src/lib/brand-palette.generated.ts` | `BRAND_IDS` / `BRAND_PALETTES`（含预览色） | ✅ 已生成 |
| `web/src/styles/theme/semantic.light.css` | 浅色语义层（4.3 全文） | ✅ 已建 |
| `web/src/styles/theme/semantic.dark.css` | 深色语义层（4.4 全文） | ✅ 已建 |
| `web/src/lib/appearance.ts` | 常量、类型守卫、存储读取（5.2） | ✅ 已建 |
| `web/src/components/appearance-menu.tsx` | 切换器主体（5.5） | ✅ 已建 |
| `web/src/components/brand-swatch.tsx` | 预览色卡（5.5） | ✅ 已建 |
| `web/scripts/check-theme-boot.mjs` | 引导脚本 × 品牌注册表一致性守卫 | ✅ 已建（并入 `tokens:check`） |
| `web/scripts/check-colors.mjs` | 硬编码颜色守卫（8.5） | 未建（见第 0 节技术债 4） |

### 6.2 修改

> 本轮完成情况见第 0 节。下表中"隔离层 + 核心组件"（index.css、index.html、Provider、切换入口、图表色、状态标签/圆点、遮罩、i18n）已改；页面级调色板类与插图仍是待办，不要按"已完成"阅读。

| 文件 | 动作 | 备注 |
| --- | --- | --- |
| `web/index.html` | 加内联引导脚本；`theme-color` 改 `#f8fafc` | 5.4 |
| `web/src/index.css` | 拆出三层 import；`@theme inline` 增量；`::selection`/`scrollbar`/`body` 走 token；`:root`/`.dark` 里的颜色字面量整体搬进语义层；`--app-gradient` 改由 `--primary`+`--brand-400` 派生 | 4.1 / 4.5 / 5.4 |
| `web/src/index.css`（oauth 段） | 第三方平台色（GitHub `#24292f`、Gitee `#c71d23`、飞书 `#3370ff`、钉钉 `#1677ff`）是**品牌方规定的颜色**，不参与我们的品牌轮换：挪到 `styles/theme/vendor.css` 并加豁免注释 | 保持原观感，只是换个位置 |
| `web/src/components/code-editor.tsx` | Monaco 双主题（`github-light` / `github-dark`）保留字面量：Monaco 只吃具体 hex，读不到 CSS 变量。只需保证取色与 `--code-bg/--code-fg` 同源（把两套 hex 提到该文件顶部的 `EDITOR_THEMES` 常量，并在切主题时按 `resolvedTheme` 重设） | 40 处 hex，Lint 白名单 |
| `web/src/components/theme-provider.tsx` | 加 `brand` 维度、`useLayoutEffect`、`data-brand`、`theme-color`、跨标签同步 | 5.3 |
| `web/src/components/app-shell.tsx` | 顶栏换成 `AppearanceMenu`；`bg-destructive … text-white` 两处改 `text-danger-foreground` | :545、:538、:629 |
| `web/src/components/locale-theme-menu.tsx` | 移除主题下拉，只留语言（或拆分重命名） | 避免双入口 |
| `web/src/lib/chart-colors.ts` | 导出 `var(--chart-N)` 字符串，不再返回 hex | 见 7.1 |
| `web/src/components/metric-graph.tsx` | `stroke={chartColor(i)}` → `style={{ stroke: chartColor(i) }}`；图例 `backgroundColor` 同理 | SVG 属性里 `var()` 兼容性不稳，改走 `style` |
| `web/src/pages/datasource/result-view.tsx` | 同上（`fill` / `backgroundColor`） | :227 / :290 |
| `web/src/components/status-badge.tsx` | `toneClasses` 换成 token 类；`LevelBadge` 的 `color-mix` 改为 `color-mix(in oklab, …)` 并用 token 兜底 | 见 7.2 |
| `web/src/components/status-dot.tsx` | 改 token，并强制配文字：新增 `label` 必填，或改成 `StatusBadge` | 见 7.3 |
| `web/src/components/role-type-tag.tsx` | `sky-*`/`slate-*` → `info`/`muted` token | |
| `web/src/components/tenant-switcher.tsx` | 6 个 `bg-*-500/10` 头像色 → `--chart-1..6`（身份色，用固定分类色而不是品牌色） | :44-49、:122 |
| `web/src/pages/overview/cards.tsx` | 11 处 red/amber/emerald → `danger`/`warning`/`success` tooltip | :33-34、:251、:360-380、:454-502 |
| `web/src/pages/cluster/detail/overview-cards.tsx` | 同上（11 处） | |
| `web/src/pages/datasource/metadata-panel.tsx` | emerald/amber/red → `success`/`warning`/`danger`（6 处） | |
| `web/src/pages/alert-dictionary/index.tsx` | 色板兜底 `#1677ff` → `DEFAULT_LEVEL_COLOR`（`#d92d20`）；新增色默认值同源 | :72 :149 :141 |
| `web/src/pages/alert/event-detail.tsx` | 级别圆点兜底 `#1677ff` → `var(--muted-foreground)` 或字典默认色常量 | :234 |
| `web/src/pages/overview/{trend,insight-trend,level-breakdown}.tsx` | 图表色走 `--chart-*` | |
| `web/src/pages/{cluster/index.tsx,cluster/pods-tab.tsx,cluster/kubeconfig-dialog.tsx}` | 调色板类 → 语义 token | |
| `web/src/pages/ssh-command/{index,results}.tsx` | 同上；`bg-black` 终端底改 `--code-bg`（`agent/shell.tsx:102` 的 xterm 底同理） | |
| `web/src/pages/{alert/list.tsx,alert/history-insights.tsx,message-log/index.tsx,datasource/components/query-statement-row.tsx}` | 调色板类 → 语义 token | |
| `web/src/components/{channel-logo.tsx,copy-button.tsx,detail-values.tsx,date-range-picker.tsx,oncall-time-view.tsx}` | 调色板类 / `rgba(255,255,255,.75)` 斜纹 → `--muted-foreground`、`--background` | oncall 斜纹用 `color-mix` |
| `web/src/components/ui/{dialog,sheet,alert-dialog}.tsx` | `bg-black/10` → `bg-overlay` | 3 处 |
| `web/src/i18n/index.ts` | 新增 `appearance.*`（zh/en），删掉不再使用的 `theme.*` 或保留别名 | :1885 / :4010，文案见 6.3 |
| `web/package.json` | `tokens:gen` / `tokens:check` | ✅ 已加 |
| `web/AGENTS.md` | 增补"颜色只能用语义 token；新增品牌跑 `tokens:gen`"两行 | 落地后补 |

### 6.3 i18n 文案（`web/src/i18n/index.ts`）

品牌名不写进 i18n，直接用生成注册表里的 `labelZh/labelEn`，避免 7×2 条重复文案：

```ts
// zh
appearance: {
  label: "外观",
  mode: "模式",
  brand: "品牌主题",
  light: "浅色",
  dark: "深色",
  system: "跟随系统",
  switched: "已切换到{{brand}}",
  reset: "恢复默认",
},
// en
appearance: {
  label: "Appearance",
  mode: "Mode",
  brand: "Brand theme",
  light: "Light",
  dark: "Dark",
  system: "System",
  switched: "Switched to {{brand}}",
  reset: "Reset",
},
```

旧的 `theme.*` 若删掉，记得 `locale-theme-menu.tsx` 同步移除引用；保留也无害（TypeScript 不会报未使用对象键）。

---

## 7. 其它关键片段

### 7.1 图表色

```ts
// web/src/lib/chart-colors.ts
/** 图表色只给变量名，主题切换时 CSS 变量即时生效，图表无需重渲染。 */
const CHART_TOKENS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
  "var(--chart-6)",
  "var(--chart-7)",
  "var(--chart-8)",
]

export function chartColor(index: number): string {
  return CHART_TOKENS[index % CHART_TOKENS.length]
}
```

调用点保持 `chartColor(i)`；但 SVG **属性**要换成 `style`（Safari 对 presentation attribute 里的 `var()` 支持不一致）：

```tsx
<path d={path} fill="none" style={{ stroke: chartColor(index) }} strokeWidth={1.5} />
<span style={{ backgroundColor: chartColor(index) }} className="size-2 rounded-full" />
```

### 7.2 状态标签

```tsx
// 语义 token 化：浅深差异由变量文件承担，不再写 dark: 变体
const toneClasses: Record<BadgeTone, string> = {
  success: "border-success-border bg-success-soft text-success",
  warning: "border-warning-border bg-warning-soft text-warning",
  danger: "border-danger-border bg-danger-soft text-danger",
  info: "border-info-border bg-info-soft text-info",
  violet: "border-violet-border bg-violet-soft text-violet",
  primary: "border-primary-soft-foreground/30 bg-primary-soft text-primary-ink",
  neutral: "border-border bg-muted/60 text-muted-foreground",
}
```

`LevelBadge` 用后端字典色时保持 `style` 内联（数据色），只把 `color-mix` 的插值空间统一到 `oklab`，兜底用 `--primary-ink`。

### 7.3 状态圆点（满足"颜色不是唯一标识"）

```tsx
// 圆点必须与文字成对出现；单独一枚色点在色弱/高对比模式下不可读
export function StatusDot({ active, label }: { active: boolean; label: string }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span
        aria-hidden
        className={cn("size-2 shrink-0 rounded-full", active ? "bg-success" : "bg-danger")}
      />
      <span className="text-xs text-muted-foreground">{label}</span>
    </span>
  )
}
```

调用点（`tenant-switcher`、`overview/cards`、`cluster` 等）把原本只有色点的位置补上文案或 `aria-label`；表格里若空间紧张，用 `title` + `<span className="sr-only">`。同理，品牌色卡、侧栏选中项都必须带图标/字重/位置指示，不能只靠颜色。

### 7.4 侧栏

侧栏保持中性底（`--sidebar` 不跟品牌走），品牌只通过 `--sidebar-primary`（选中项文字/图标）、`--sidebar-accent-foreground`、`--sidebar-ring` 出现。这样 15 套主题的"骨架"一致，切换只换强调色——减少"像换了个产品"的割裂感。

---

## 8. 落地步骤与验收

### 8.1 分批（每批可独立合并、独立回滚）

1. **基础设施**：`semantic.light/dark.css`、`index.css` 改造、`theme-provider.tsx`、`index.html` 引导脚本、`appearance.ts`。此时还没有切换 UI，品牌默认 `sunset`，界面除主色外不变。
2. **切换 UI**：`appearance-menu.tsx` + `brand-swatch.tsx` + `app-shell.tsx` + i18n 文案 + 删掉旧主题下拉。
3. **状态色批量化**：`status-badge` / `role-type-tag` / `status-dot` / `overview` / `cluster` / `datasource` 等 20 余处（用 `grep -n` 清单逐条改，改完跑一次 `check-colors`）。
4. **图表与材质**：`chart-colors.ts`、`metric-graph`、`result-view`、`overlay`、`scrollbar`、`oncall-time-view`、xterm 底色。
5. **守卫**：`scripts/check-colors.mjs` + `tokens:check` 接入 CI/Makefile，`AGENTS.md` 补两行规范。

### 8.2 命令

```bash
cd web
node scripts/gen-brand-tokens.mjs          # 打印对比度报告（失败即退出 1）
npm run tokens:check                       # 产物是否与脚本一致（CI 用）
npm run typecheck && npm run lint && npm run build
```

### 8.3 浏览器验收（`AGENTS.md` 要求，逐项记录结果）

- 15 品牌 × 2 模式 = 30 种组合，逐个在 8 个代表页面走一遍：`/overview`、`/alerts`（表格 + 筛选 chips）、`/alert-history`（两个页签 + 长面板滚动）、`/datasources`（查询结果图）、`/cluster`、`/members`（角色标签）、`/profile`、`/login`。
- 1366×768 与 390×844 两个视口各看一遍：颜色改动不得引出新的溢出/裁剪（表头、筛选区高度、页签指示器）。
- 交互态：按钮 hover/active/focus-visible、输入框错误态（`data-invalid` 红框仍在）、页签选中、行 hover、骨架屏、Sonner toast、Dialog/Sheet 遮罩、Tooltip。
- 防闪烁：DevTools 4× CPU 限速 + 强制刷新，浅/深各刷 10 次，首帧不得出现白底/错主题；切到"跟随系统"后改系统外观，页面跟随且不闪。
- 持久化与同步：切换后刷新保持；开两个标签页，A 改品牌 B 跟着变；`localStorage` 被清空后回落 `sunset` + 跟随系统。
- 键盘与读屏：Tab 到切换按钮 → Enter 打开 → 方向键在模式/色卡间移动 → Enter 选中；焦点环始终可见；`aria-checked`、`aria-live` 播报正确。
- 对比度抽查：切换每一套后，用 DevTools 取 `--primary` / `--primary-ink` / `--muted-foreground` / `--danger` 与相邻底色的对比（生成器已覆盖主色对，页面上还要抽查实际渲染像素）。

### 8.4 完成定义

- `node scripts/gen-brand-tokens.mjs` 全绿；`tokens:check` 无 diff。
- `grep -rE '#[0-9a-fA-F]{6}|(bg|text|border|ring)-(slate|gray|red|emerald|amber|sky|violet|pink|cyan|rose)-[0-9]{2,3}' web/src` 只命中 `styles/theme/**`、`brand-palette.generated.ts`、`assets/`。
- 7×2 组合在真实浏览器过一遍，截图存档在 PR 描述里。
- `typecheck` / `lint` / `build` 通过。

### 8.5 守卫脚本（`web/scripts/check-colors.mjs`）

```js
import { readdirSync, readFileSync, statSync } from "node:fs"
import { join, relative } from "node:path"

const SRC = new URL("../src", import.meta.url).pathname
const ALLOW_FILE = [
  /^styles\/theme\//,                     // token 定义处（含 vendor.css）
  /^lib\/brand-palette\.generated\.ts$/,   // 生成产物
  /^components\/code-editor\.tsx$/,        // Monaco 只吃 hex，双主题常量集中在文件顶部
  /^i18n\/index\.ts$/,                     // 示例色值是文案，不是渲染色
  /^components\/(?:feishu|dingtalk|gitee|github)-icon\.tsx$/, // 第三方品牌标志
]
const ALLOW_LINE = [
  /token-allow/,                    // 个别数据色（后端字典色、品牌 logo）可显式豁免
]
const RULES = [
  { name: "hex", re: /#[0-9a-f]{3,8}\b/i },
  { name: "rgb", re: /\brgba?\(/ },
  {
    name: "palette-class",
    re: /\b(?:bg|text|border|ring|from|to|via|fill|stroke|divide)-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}\b/,
  },
]

const hits = []
for (const file of walk(SRC)) {
  const rel = relative(SRC, file)
  if (ALLOW_FILE.some((re) => re.test(rel))) continue
  readFileSync(file, "utf8").split("\n").forEach((line, index) => {
    if (ALLOW_LINE.some((re) => re.test(line))) return
    for (const rule of RULES) {
      if (rule.re.test(line)) hits.push(`${rel}:${index + 1} [${rule.name}] ${line.trim()}`)
    }
  })
}

if (hits.length > 0) {
  console.error(`发现 ${hits.length} 处硬编码颜色：\n${hits.join("\n")}`)
  process.exit(1)
}

function* walk(dir) {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry)
    if (statSync(full).isDirectory()) yield* walk(full)
    else if (/\.(tsx?|css)$/.test(entry)) yield full
  }
}
```

再补一条品牌 id 一致性检查（防止 `index.html` 与注册表漂移）：

```bash
node -e '
const ids = require("fs").readFileSync("src/lib/brand-palette.generated.ts","utf8")
  .match(/export const BRAND_IDS = \[([^\]]+)\]/)[1].replace(/["\s]/g,"").split(",");
const html = require("fs").readFileSync("index.html","utf8");
const missing = ids.filter((id) => !html.includes(`"${id}"`));
if (missing.length) { console.error("index.html 引导脚本缺少品牌：", missing); process.exit(1) }
'
```

---

## 9. 风险与取舍

| 风险 | 说明 | 处置 |
| --- | --- | --- |
| 品牌色与"语义正确"冲突 | 暖阳橙/梅洛粉做危险色会被误读 | 状态色完全解耦（3.5），品牌色只用于主操作与强调；危险操作恒用 `--danger` |
| 浅色按钮文字色随品牌变化 | 5 个品牌是深字、2 个是白字，观感不完全统一 | 这是自动对比度判定的结果；若设计上要求统一白字，把生成器里 `pickFill` 的候选顺序改成"只接受白字并逐阶压深"，代价是主色变暗（暖阳橙→`#ac6012`、矢车菊→`#4776bc`） |
| 切换瞬间的过渡 | 全局禁用 transition 会让 hover 也顿一下 | 只在切换那一帧禁用（现有 `disableTransitionsTemporarily` 已是双击 rAF 后恢复）；色卡反馈用 `animation` |
| `resolvedTheme` 读取 `matchMedia` | 渲染期读取，非严格响应式 | CSR 可接受；后续可换 `useSyncExternalStore` |
| `localStorage` 隐私模式抛错 | 引导脚本与 provider 都要 try/catch | 已包含；失败回落默认值 |
| 服务端持久化缺失 | 换设备不跟随 | `/v1/self` 无偏好接口；后续可加 `PUT /v1/self/preferences`，前端只需在 `setBrand` 里追加一次写请求（localStorage 仍是首帧来源） |
| 品牌数未来增长 | 新品牌要维护两处（脚本 + 引导脚本白名单） | `tokens:gen` + 一致性检查脚本；引导脚本的数组是唯一刻意重复的地方 |
| `next-themes` 未使用 | 依赖里挂着但被自研 provider 取代 | 本次不动；若后续替换，接口形状已对齐（`theme/setTheme/system`） |
| 深色 `--danger` 变浅 | 大红变成 `#fca5a5`，观感偏"粉红错误" | 为保证 soft 底上 4.5:1；若坚持正红，可把 soft 底降到 8%，但需重新验证 |

---

## 10. 附录

### 10.1 生成器输出与校验

```bash
cd web
node scripts/gen-brand-tokens.mjs            # 报告
node scripts/gen-brand-tokens.mjs --write    # 写 tokens.brand.css + brand-palette.generated.ts
node scripts/gen-brand-tokens.mjs --check    # CI：产物过期即失败
```

新增品牌只需在脚本 `BRANDS` 数组加一项（`id / zh / en / hex`），跑 `--write`，再补 `index.html` 白名单一项。品牌数量与"7+2"结构无关：品牌层是 N 个块，模式层恒为 2 个文件。

### 10.2 术语

- **角色色**：从色阶里按对比度挑出来的、有明确职位的颜色（填充/前景/文字/环），组件只认角色不认阶。
- **soft / border / foreground**：同一语义色的低饱和底、描边、反差前景。
- **ink**：品牌色作为文字/图标出现时的可读版本。
