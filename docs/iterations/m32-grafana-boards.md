# M32 Grafana 大盘

通用规范见 [conventions.md](conventions.md)。本文新增租户级「大盘」模块：把租户自己 Grafana 上的面板以 iframe 组合进 Alioth，按 12 栏栅格排布，支持标题、描述、尺寸与顺序维护。

Alioth 只保存面板的展示配置，不代理、不缓存、不转存 Grafana 数据，也不保存 Grafana 账号或 Token；浏览器直接访问 Grafana。大盘与 `/overview` 的平台告警/投递概览是两件事，不合并。

## 目标与依赖

操作者可以：

1. 在「大盘」列表新建、编辑、删除、启用/停用大盘，并进入查看页。
2. 给大盘添加任意多个 Grafana 面板，维护标题、描述、分享地址、宽度（12 栏中占几列）、高度和顺序。
3. 在查看页用 12 栏 CSS grid 渲染 iframe，默认追加 `?kiosk`、允许全屏、禁止发送 Referer；可一键隐藏/显示面板标题与描述，也可「打开原图」。
4. 租户按 `board:read` / `board:write` / `board:delete` 和菜单权限 `menu:board` 控制入口与操作；中英文文案齐全。
5. 存量基础套餐租户升级后自动获得 `board` 能力；新租户随基础套餐获得。

依赖：M19 / M21 / M22 / M24（能力目录、套餐快照、菜单化授权与租户菜单）、M23（升级与迁移）、M1（用户与租户）、Web 壳的主题与 i18n。

## 交付

| 层 | 文件 | 变更 |
| --- | --- | --- |
| proto | `api/alioth/v1/board.proto`（新增） | `BoardService`：大盘 CRUD、启停、列表；面板随大盘整体保存 |
| 业务对象 | `app/alioth/pkg/bo/board.go`（新增） | `Board`、`BoardPanel` |
| 仓储接口 | `app/alioth/internal/data/repository/board.go`（新增） | `BoardRepo` |
| 用例 | `app/alioth/internal/biz/board/composer.go`（新增） | 大盘与面板的组合维护、校验、租户内唯一与软删除 |
| 存储 | `app/alioth/internal/data/ent/schema/board.go`、`board_panel.go`（新增）与 `board_repo.go`（新增） | 实体、索引、事务保存与级联软删除 |
| 权限目录 | `app/alioth/pkg/perm/feature.go`、`capability.go`、`menu.go`、`perm.go` | 模块、能力、菜单、`board:*` 权限 |
| 种子 | `app/alioth/internal/data/seed.go` | 基础套餐补能力、存量基础套餐租户补偿 |
| Web | `web/src/api/board/`、`web/src/pages/boards/`、`web/src/lib/grafana-url.ts` | 列表页、查看页、面板编辑与 URL 处理 |
| Web | `web/src/App.tsx`、`components/app-shell.tsx`、`lib/feature-catalog.ts`、`lib/last-page.ts`、`i18n/index.ts` | 路由、一级菜单、能力目录与中英文 |

### 模块、能力与权限

- 目录新增模块 `board`（中文「大盘」）、能力 `board`、权限 `board:read` / `board:write` / `board:delete`、菜单 `board` 与菜单权限 `menu:board`。能力与菜单都包含三个权限，菜单默认读权限为 `board:read`。
- 模块与能力都插在 `role` 之后、`cluster` 之前，使基础套餐保持连续的前八个模块；目录从 10 个模块 / 57 个能力变为 11 / 63（含 M28 维护窗口、M29 升级/轮换与 M30 状态页/SLA 能力），基础套餐从 30 项变为 36 项。
- `perm.BaseFeatures()` 增加 `board`，`BaseCapabilities()` 自动包含 `board`；`basePackageRemark` 同步补「大盘」。套餐勾选页、角色授权页与 `/v1/system/features` 都随目录自动生效。
- 权限语义：`board:read` 看列表与查看页；`board:write` 新建、编辑、启停；`board:delete` 删除。接口由 `access.Allow` 校验能力与菜单权限，页面隐藏按钮不能代替服务端拒绝。

| 权限 | 允许 |
| --- | --- |
| `board:read` | 列表、详情、查看页、面板 URL 读取 |
| `board:write` | 新建/编辑大盘与面板、启用/停用 |
| `board:delete` | 删除大盘（连带软删除面板） |
| `menu:board` | 侧边栏「大盘」入口；没有它时即使有 `board:read` 也不显示菜单 |

### 升级补偿

- 新租户创建时按基础套餐快照获得 `board`；`ensureBasePackage` 通过 `perm.BaseCapabilities()` 给基础套餐补齐 `board` 明细，重复执行只加不删。
- 存量租户的能力在购买时已快照，不会自动拿到新能力。新增 seed 步骤 `compensateBaseTenantBoard`：只处理**当前套餐为基础套餐**的未删除租户，若快照和额外能力里都没有 `board`，写一条 `namespace_feature`（`enabled=true`）作为额外能力。
- 用额外能力而不是改套餐快照，是为了不重写租户的购买记录，并在后续换套餐时保留；只补基础套餐租户，避免把新模块免费扩散到自定义套餐（额外能力不可撤销，见风险）。
- 补能力不会给角色自动加权限：负责人按身份隐式拥有；其他角色需要租户在角色授权页勾选 `menu:board` 与 `board:*`。
- 迁移只加不减、可重复执行；已有权限与角色授权不被改写。

### 数据模型

`Board`（ent schema，`Audit` + `SoftDelete`）：

| 字段 | 类型/约束 | 说明 |
| --- | --- | --- |
| `uid` | string，唯一、不可变 | 标识 |
| `namespace_uid` | string | 所属租户 |
| `name` | string ≤64 | 租户内未删除记录唯一 |
| `description` | string ≤1024，默认空 | 描述 |
| `status` | 枚举 `ACTIVE` / `DISABLED`，默认 `ACTIVE` | 启用/停用（需求中的 `enabled` 语义；沿用项目其他资源的 status 约定） |
| `created_by` | string | 创建者用户 UID，创建后不变 |

索引：`uid` 唯一；`(namespace_uid, name)` 部分唯一（`deleted_at IS NULL`）；`(namespace_uid)`。

`BoardPanel`（同样带审计与软删除）：

| 字段 | 类型/约束 | 说明 |
| --- | --- | --- |
| `uid` | string，唯一、不可变 | 面板标识，整体保存时用于匹配 |
| `namespace_uid` | string | 冗余租户列，隔离与索引用 |
| `board_uid` | string | 所属大盘 |
| `title` | string 1–128 | 标题 |
| `description` | string ≤1024，默认空 | 描述 |
| `grafana_url` | string ≤2048 | Grafana 分享/嵌入地址 |
| `width` | int 1–12 | 12 栏中占几列 |
| `height` | int 120–1600 | iframe 内容区像素高，不含标题/描述 |
| `sort` | int | 服务端按数组下标从 0 起连续重写，升序展示 |

索引：`uid` 唯一；`(namespace_uid, board_uid)`；`(board_uid, sort)`。项目不使用 ent 外键/edge，面板与大盘的归属由 `board_uid` 逻辑关联，保存和删除都在事务内完成，不给数据库加级联约束。

### 字段语义与校验

- `grafana_url`：去首尾空白，必须能解析为绝对 URL；scheme 只接受 `http` / `https`；拒绝 userinfo、`javascript:` / `data:` 等协议、URL fragment；path 必须包含 `/d/`、`/d-solo/` 或 `/public-dashboards/`（Grafana dashboard / panel / 公开面板的分享或嵌入地址，允许 `/grafana/` 子路径）。query 允许 `from`、`to`、`orgId`、`panelId`、`viewPanel`、`var-*`、`refresh`、`kiosk` 等；参数名含 `token`、`secret`、`password`、`api_key`、`apikey` 的地址拒绝（不区分大小写）。
- `width`：1–12；`height`：120–1600；面板数每个大盘 0–50。空面板合法（新建后先建骨架再补面板）。
- `sort` 不由客户端提交：`panels[]` 数组顺序就是展示顺序，服务端按数组下标从 0 起连续重写。上移/下移只交换数组相邻元素后整体保存。
- 启用状态只存 `status`，不另存 `enabled` 布尔列，避免两处状态不一致；启停只走独立接口。
- 名称去空格后 1–64，租户内精确匹配唯一；冲突返回 `CONFLICT`，软删除后名称可复用。
- 面板 `uid` 只在整体保存中用于匹配：空表示新建；非空必须在**当前大盘**的面板集合内，否则 `INVALID`，不能借 UID 修改其他大盘的面板。

### 接口

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/v1/board` | `board:write` | 新建大盘，默认启用，可同时提交面板 |
| `PUT` | `/v1/board/{uid}` | `board:write` | 整体保存名称、描述与面板；不改启停状态 |
| `PUT` | `/v1/board/{uid}/status` | `board:write` | 只切换 `status`（`ACTIVE` / `DISABLED`），避免整体保存覆盖布局 |
| `DELETE` | `/v1/board/{uid}` | `board:delete` | 软删除大盘及其全部面板 |
| `GET` | `/v1/board/{uid}` | `board:read` | 返回大盘与按 `sort` 升序的面板 |
| `GET` | `/v1/boards` | `board:read` | 分页列表，`page`/`pageSize`/`name`/`status`，按更新时间倒序 |

**面板保存方式：选择随大盘整体保存，不提供独立的面板 CRUD 接口。** 新建/更新请求带 `panels[]`，服务端在一个事务内按 `uid` 匹配：更新已有面板、创建 `uid` 为空的面板、软删除未出现的面板，并重写 `sort`。理由是一次写入即可保证「面板集合 + 顺序」原子一致，避免独立增删排序产生中间态；代价与并发覆盖见风险。

请求校验用 `buf.validate`：`name` 1–64、`description` ≤1024、`panels` ≤50、`title` 1–128、`grafana_url` 1–2048、`width` 1–12、`height` 120–1600。响应统一返回 `board` 与 `panels`；列表项带 `panelCount`。删除、跨租户 UID、其他租户数据源式越权都按 `NOT_FOUND` 或 `FORBIDDEN` 处理。

### Web

- 路由：`/boards` 列表页（`BoardPage`）、`/boards/:uid` 查看页（`BoardViewPage`）。侧边栏在「概览」下增加一级入口「大盘」（模块只有一个菜单，不再套二级分组），显示条件为 `board:read` + 已开通 `board` + `menu:board`；`last-page` 的已知路径同步。
- 列表页：保留筛选、分页并把查询条件放进 URL；表格列为名称、描述、面板数、状态、更新时间、操作（查看 / 编辑 / 启停 / 删除）。新建与编辑用 `Dialog`（`board-dialog.tsx`）完成元数据与面板编辑；面板编辑器（`panel-editor.tsx`）支持添加、删除、上移/下移、宽度/高度输入与 URL 提示。删除用确认对话框；启停用行内开关。
- 查看页：页头显示名称、描述、启停状态与更新时间；工具栏提供「显示标题与描述」开关（按大盘存 `localStorage`，不改实体）、刷新 iframe 与返回列表；停用大盘显示醒目提示但仍可查看。
- 栅格：容器 `grid grid-cols-12 gap-3`；每张面板用 CSS 变量传宽度，`md:` 以上按 `span` 展开，窄屏（<768px）统一占满 12 栏单列。面板高度用 `height` 像素；标题/描述在 iframe 上方，隐藏时 iframe 占满卡片。
- iframe 固定属性：`title`（面板标题）、`allowfullscreen`、`referrerpolicy="no-referrer"`、`loading="lazy"`、`sandbox="allow-scripts allow-same-origin allow-popups allow-forms"`。URL 由 `web/src/lib/grafana-url.ts` 的 `buildGrafanaEmbedURL` 处理：已有 `kiosk` 参数则保留原值，没有则追加 `kiosk`，其余 query 原样保留。
- 每张面板提供「打开原图」链接，`target="_blank"` + `rel="noopener noreferrer"`，打开未追加 `kiosk` 的原始地址。
- 空态：列表与查看页都给出配置说明：Grafana 需开启 `allow_embedding = true`；HTTPS 的 Alioth 不能嵌入 HTTP Grafana；Grafana 的 `X-Frame-Options` / CSP `frame-ancestors` 必须允许 Alioth 来源；需要登录的 Grafana 建议使用匿名访问或公共 kiosk 面板。跨域 iframe 的错误无法被父页面读取，页面不伪造「加载成功」。
- 加载排查：排查入口收成图标放在大盘名称右侧（默认只显示感叹号，点击才展开说明），不再占用整行；说明内网/VPN、Grafana 证书需受信任（iframe 不弹证书例外）、`allow_embedding`、混合内容与登录跳转。面板连通性探测失败时，感叹号同样放在该面板标题右侧；隐藏标题时浮在面板右上角。新增面板在列表最底部，避免编辑完最后一个面板还要回到顶部。
- 面板连通性探测：每个面板额外发一次 `no-cors HEAD` 探测，只判断浏览器网络层能否到达（DNS、证书、混合内容、连接拒绝或 CSP `connect-src` 失败会 reject）。失败时该面板右上角显示警告图标，浮层给出结论、排查步骤、「重新探测并加载」和「打开原图」。探测成功只代表能到达，不能证明允许 iframe 嵌入；X-Frame-Options/CSP/登录跳转仍无法从父页面读取，浮层会明确说明这一点。
- 面板展示：面板网格不保留外圈间距，iframe 与卡片完全贴合；隐藏标题与描述时卡片不绘制边框/背景/圆角，纯图表连续排列。显示标题与描述时卡片有清晰边框和阴影，鼠标悬停上浮并加深阴影（`prefers-reduced-motion` 下不做位移动画）；Grafana 自身在 iframe 内的面板边框/留白由 Grafana 绘制，Alioth 无法跨域修改，需要更贴边时改用 Grafana 的单面板 d-solo 链接。
- i18n：`nav.boards`、`board.*`（列表、表单、面板、空态、安全提示、错误）与 `menu.board` 中英文齐全；能力/模块标签加到 `feature-catalog.ts`。

### 安全

- URL 协议与形状在服务端校验（Web 同步一份做即时提示，服务端为准）；HTML 与 URL 都通过 React 文本/属性渲染，不使用 `srcdoc`、`dangerouslySetInnerHTML` 或字符串拼 HTML。
- 不代理 Grafana 请求：没有服务端转发接口，浏览器直连 Grafana；Alioth 不保存 Grafana 地址以外的凭据，响应不回传任何密钥。面板 URL 由用户粘贴，校验只能拦截可识别的密钥参数，不能穷尽所有 Grafana 鉴权方式。
- iframe `sandbox` 不含 `allow-top-navigation`，脚本不能把 Alioth 顶层导航走；`allow-scripts` 与 `allow-same-origin` 同时存在是 Grafana 读写自身 localStorage/cookie 所需，iframe 与父页面跨源，拿不到父页面 DOM。
- `referrerpolicy="no-referrer"` 不把 Alioth 页面地址泄露给 Grafana；`allowfullscreen` 只用于单面板放大。
- 部署侧：若给 Web 配 CSP，需要允许 `frame-src` 指向租户的 Grafana 源（`https:` / 具体域名）；Grafana 侧必须允许被嵌（`allow_embedding=true`、无 `X-Frame-Options: DENY`、`frame-ancestors` 含 Alioth）。HTTPS 混合内容与第三方 Cookie 限制属于部署边界，代码不做绕过。

## 验收

- [ ] 能力目录：`/v1/system/features` 返回模块 `board`、能力 `board` 与三个权限；模块/能力总数为 11 / 63，基础套餐为 36 项；角色授权页、套餐勾选页与中英文标签正常。
- [ ] 新租户：创建后 `board` 在套餐快照中，负责人能看到「大盘」菜单；非负责人授权 `menu:board` + `board:read` 后可见。
- [ ] 存量基础套餐租户：跑 seed 后获得 `board` 额外能力，重复执行不产生重复行；自定义套餐租户不被自动补，`namespace_feature` 不变。
- [ ] 菜单与鉴权：无 `menu:board` 时侧栏不显示，直接访问 `/boards` 回落到无权页面；无 `board:write` 时新建/编辑/启停入口隐藏且接口 403；无 `board:delete` 时删除入口隐藏且接口 403。
- [ ] 大盘 CRUD：创建、列表分页、详情、编辑、删除成功；同名返回 `CONFLICT`；软删除后列表/详情不可见且名称可复用。
- [ ] 面板：整体保存可新增/修改/删除/上移/下移面板，`sort` 从 0 起连续；已有面板 UID 保持不变；从数组移除的面板软删除；事务失败时大盘与面板都不落库。
- [ ] 校验：`ftp://`、`javascript:`、带 userinfo、带 fragment、path 不含 `/d/`、`/d-solo/` 或 `/public-dashboards/`、query 含 `token`/`password` 的 URL 被拒绝；`width` 0/13 与 `height` 119/1601 被拒绝；51 个面板被拒绝；合法面板保存成功。
- [ ] 启停接口只改 `status`，不重写面板；停用后列表状态正确，查看页显示停用提示；列表按状态筛选正确。
- [ ] 查看页：iframe 带 `allowfullscreen`、`referrerpolicy`、`sandbox` 与追加的 `kiosk`；已有 `kiosk=tv` 不被覆盖；标题/描述开关生效；「打开原图」新开原始地址；空态显示 Grafana 配置说明。
- [ ] 浏览器回归：`/boards` 筛选/分页/新建/编辑/删除/启停/进入查看；1366×768 与 390×844 下无横向溢出，窄屏面板单列、页面滚动仅一处；中英文切换。
- [ ] 单测覆盖 URL 校验、宽高与面板数、名称冲突、整体保存匹配、级联软删除、启停与权限；seed 补偿幂等由真实库集成测试覆盖。
- [ ] `make vet`、`make lint`、`make test` 通过；Web `pnpm typecheck`、`pnpm lint`、`pnpm build` 通过。

## 不做什么与后续

- 不做 Grafana API 集成：不自动创建/同步 dashboard、不拉取 Grafana 面板列表、不校验面板是否存在。
- 不做面板拖拽缩放、自由布局或响应式断点；宽度只支持 1–12 栅格，高度只支持像素。
- 不做大盘级权限、公开分享、复制模板、历史版本、回收站与恢复。
- 不做变量选择器、时间范围控件、刷新频率控制、图片快照或导出；这些继续由面板 URL 自带参数和 Grafana 自己负责。
- 不做服务端代理、鉴权注入或 Grafana Token 托管；后续若需要统一代理，应单独立项评估密钥与 SSRF 风险。
- 后续可加：面板复制排序的拖拽、Board 级标签/分组、多人协作的乐观锁版本号、已停止大盘的隐藏筛选。

## 风险与边界

- **嵌入失败不可检测**：跨域 iframe 即使被 `X-Frame-Options` 拒绝也常触发 `onload`，前端无法区分内容与错误页；只能靠空态说明与验收时人工确认，不能承诺「加载失败提示」。
- **Grafana 侧配置依赖**：`allow_embedding`、CSP/`frame-ancestors`、匿名访问、Cookie 与 HTTPS 混合内容都可能让面板空白；这是本方案只保存 URL 的直接边界。
- **整体保存会覆盖并发修改**：两个编辑者先后保存时后写者覆盖前者的面板集合；本期按单人配置场景接受，后续可加版本号做乐观锁。
- **额外能力不可撤销**：升级补偿写 `namespace_feature` 后，按现有「能力只增不减」模型无法收回；因此只补基础套餐租户，自定义套餐需人工评估再开。
- **URL 校验是启发式**：Grafana 版本/子路径差异可能造成误拒或漏拒；黑名单不能代替「不要把 Token 放进分享 URL」的运维要求。
- **性能与可访问性**：面板多时 iframe 同时加载会拖慢页面，已限制 50 个并加 `loading="lazy"`；固定高度由 Grafana 内部滚动，键盘焦点进入 iframe 后可能产生嵌套滚动。
- **删除不可恢复**：大盘与面板均为软删除，本期没有恢复入口；删除操作需要二次确认。
- **创建者展示**：实体只存 `created_by` UID，列表暂不展示创建者昵称，避免为展示再查用户；后续需要时再加解析。

## 文档同步

- 实现后更新 [../design.md](../design.md) 的 Web 章节与能力目录计数（11 模块 / 63 能力 / 基础套餐 36 项），并在数据对象一览补 `Board`、`BoardPanel`。
- 在 [../roadmap.md](../roadmap.md) 增加 M32 条目；[conventions.md](conventions.md) 同步模块数、能力数与基础套餐描述。
- kubernetes-console 文档不涉及大盘；本模块的面板地址规则与安全边界以本文为准。
