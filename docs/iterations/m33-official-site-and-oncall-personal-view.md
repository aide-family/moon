# M33 官网公开站、登录入口与值班个人视图

M33 给 Alioth 补上对外可见的另一半：一个免登录的官网（公开站）承载官方公布信息与产品说明，官网内容、页脚备案信息与登录方式开关收进后台系统配置；登录后可以留在官网，也可以选一个租户进后台。同时收尾两处已有缺口：实时告警「认领人」列保存必然 400，以及 OnCall 值班时间视图没有「我的」视角。

本文件先定口径再动手。第 2 节的决策未拍板之前 R1–R5 不开工；R6 已经修完；R7 必须先定「我」怎么映射、按什么轴画，否则做出来的个人视图会把通知组接收时段误当成个人值班。

## 1. 目标与依赖

操作者可以：

1. 不登录访问官网首页与产品说明页，看到的是后台已发布的内容；未发布内容对外 404，不泄露草稿。
2. 超级管理员在后台编辑官网页面与站点设置，保存草稿、预览、发布、下架；公开接口只读发布快照。
3. 在官网点「登录」完成登录后回到原页面（不跳后台）；点「进入后台」时按租户数量直接进入或弹出选择。
4. 页脚备案号、公安备案号、版权与联系邮箱来自后端接口并可在后台修改；留空则整条不渲染。
5. 第三方（OAuth2）登录入口能在后台按平台开关；关闭时后端不下发、前端不渲染，直接调登录接口也不签发凭证。
6. 实时告警「自定义列」勾选或保留「认领人」后保存成功，不再弹「请求不符合规则，请检查后重试。」。
7. OnCall 值班时间视图提供「全部 / 我的」切换，只看当前登录用户自己的班次。

依赖：M1/M9（身份与租户）、M13/M29（OnCall 与值班轮换）、M28（认领与事件时间线）、M30（公开状态页的原生 `fetch` no-auth 先例）、M22（菜单化权限与 `system:*` 权限族）。

## 2. 待决策（阻塞）

| 编号 | 决策 | 推荐 | 影响面 |
| --- | --- | --- | --- |
| D1 | 官网挂在哪：`/` 让给官网、后台着陆迁到 `/app`；还是官网放 `/site/*`、`/` 保持后台着陆 | 官网用 `/`，后台着陆改 `/app` | 前者语义正确、符合「登录后可留在官网」；代价是要同步改 `App.tsx`、`lib/redirect.ts` 的默认 `next`、`lib/last-page.ts` 的 `knownPaths`、`app-shell.tsx` 的 `tenantFreePaths` 与侧栏入口。选 `/site/*` 则零路由风险，但官网被降级成站点里的一个子路径 |
| D2 | 官网内容归属：平台级还是租户级 | 平台级（`system:site:*`，超级管理员维护） | 决定接口前缀（`/v1/system/site/*`）、权限矩阵与多租户隔离用例 |
| D3 | 首页是固定区块（hero / 能力卡片 / 产品说明摘要 / 尾 CTA，字段在设置里编）还是整页 Markdown | 固定区块 + 每个子页 Markdown 正文 | 固定区块能做 SEO 与品牌一致性；纯 Markdown 首页排版不可控 |
| D4 | 双语是否本期做 | 本期做，`(slug, locale)` 一条记录，可只发布中文 | 只做中文则 `LocaleMenu` 与语言回落在官网不生效，二期补要改唯一键 |
| D5 | 「选择租户」放在官网弹窗还是只保留后台的租户切换器 | 官网弹窗 | 需求明确要求登录后可留在官网再选租户 |
| D6 | 个人视图的「我」怎么确定、按什么轴画 | 后端在排班项上给 `isMe`，并提供自服务接口；轴用「未来 7/14 天日期 × 小时」 | 前端按姓名/邮箱猜身份会因重名、无邮箱、代填邮箱误判；沿用现有「星期 × 小时」周网格画不出按绝对日期轮转的班次 |
| D7 | 备案主体与真实号码、官方邮箱、logo 是否改色、是否公开定价 | 号码与邮箱由运营提供；logo 保持蓝色星徽（蓝色只作商标资产色） | 备案号缺失时官网页脚不显示该行，上线前必须补齐；公开定价需要单独的定价页规范 |

## 3. 数据模型

新增 schema 落在 `app/alioth/internal/data/ent/schema/`，迁移只由 `alioth` 执行。

### `site_settings`（单例，平台级）

| 字段 | 说明 |
| --- | --- |
| `id` | 固定 1，单例 |
| `site_name` / `logo_url` / `favicon_url` | 站点名与标识 |
| `contact_email` / `repository_url` | 页脚联系信息，可空 |
| `copyright_text` | 空则前端回落 `© {{year}} Alioth` |
| `filing_icp_number` / `filing_icp_url` | 空则官网页脚不渲染该行；`url` 空则默认工信部备案查询地址 |
| `filing_police_number` / `filing_police_url` / `filing_extra_text` | 公安备案与补充说明，均可空 |
| `hero_image_url` | 首页主视觉，空则用品牌淡底占位 |
| `updated_at` / `updated_by` | 审计 |

### `site_settings_locale`（每语言一行，`locale` 唯一）

`hero_title`（必填）、`hero_subtitle`、`cta_primary_label` / `cta_primary_href`、`cta_secondary_label` / `cta_secondary_href`、`feature_section_title` / `feature_section_subtitle`、`closing_title` / `closing_description` / `closing_button_label`、`footer_tagline`、`seo_title` / `seo_description` / `og_image_url`。

### `site_features` / `site_footer_links`

能力卡片与页脚链接：`uid`、`locale`、`sort_order`（同语言内唯一）、`icon_key`（白名单枚举）、`title`、`description` / `section`（`products` / `resources` / `legal`）、`label`、`href`。

### `site_pages`

| 字段 | 说明 |
| --- | --- |
| `uid` / `slug` / `locale` | `slug` 只允许小写字母、数字与连字符；未删除记录上 `(slug, locale)` 唯一 |
| `state` | `DRAFT` / `PUBLISHED` / `ARCHIVED` |
| `show_in_nav` / `sort_order` | 导航位置 |
| `draft_*` | 草稿：标题、摘要、正文 Markdown、封面、SEO |
| `published_*` | 发布快照，公开接口只读这一套 |
| `published_at` / `published_by` | 最近发布 |

约束：`slug=home` 的记录禁止下架与删除；`hasUnpublishedChanges` 由服务端比较两套字段得出，不由前端 diff。

### `auth_login_settings`（单例，平台级）

`email_enabled`（默认 `true`）、`oauth2_enabled`（默认 `true`）、`updated_at` / `updated_by`。与 `biz/auth` 的 env 配置取交集：开关关闭时 `OAuthProviders` 返回空列表，`StartOAuth` 直接拒绝，只挡一半等于没挡。

## 4. 接口

新增 proto：`api/alioth/v1/site.proto`（`SiteService`）、登录方式开关并入 `api/alioth/v1/auth.proto`。

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/v1/site/config` | 公开 | 站点设置 + 首页区块 + 导航 + `loginMethods`，按 `locale`（缺省回落 `zh`） |
| `GET` | `/v1/site/pages/{slug}?locale=` | 公开 | 只读发布快照；不存在或未发布一律 404，不区分「没写过」与「已下线」 |
| `GET` | `/v1/system/site` | `system:site:read` | 站点设置（双语文案、页脚链接、能力卡片） |
| `PUT` | `/v1/system/site` | `system:site:write` | 全量保存，长度与 URL scheme 在 biz 层校验 |
| `GET` / `POST` | `/v1/system/site/pages` | `system:site:read` / `write` | 列表（关键字、状态、语言、分页）与新建 |
| `GET` / `PUT` | `/v1/system/site/pages/{uid}` | `system:site:read` / `write` | 详情与保存草稿；请求带 `updatedAt` 做乐观锁，冲突 409 |
| `POST` | `/v1/system/site/pages/{uid}/publish` | `system:site:write` | 草稿 → 发布快照，可重复执行 |
| `POST` | `/v1/system/site/pages/{uid}/unpublish` | `system:site:write` | 下架保留快照，`home` 拒绝 |
| `DELETE` | `/v1/system/site/pages/{uid}` | `system:site:write` | 软删，`home` 拒绝 |
| `GET` / `PUT` | `/v1/system/auth/login-methods` | `system:login:read` / `write` | 两种登录方式的开关 |
| `GET` | `/v1/auth/oauth2/providers`（已有） | 公开 | 开关关闭时返回空列表 |

权限常量加进 `app/alioth/pkg/perm/perm.go` 的 `SystemCatalog()`：`system:site:read`、`system:site:write`、`system:login:read`、`system:login:write`。`SUPER_ADMIN` 自动全量，`SYSTEM_ADMIN` 只加读，`SYSTEM_READONLY` 因 `:read` 后缀自动获得。

R7 的接口：`GET /v1/oncall-rotations/my-shifts?days=14`，按当前登录账号解析成员身份，只返回本人班次（含 `current`、轮换名、时区、起止），不要求 `notification_rotation:read`；排班接口现有返回里补 `isMe`，供管理端复用。

## 5. Web

### 路由

| 路径 | 页面 | 登录 | 说明 |
| --- | --- | --- | --- |
| `/`（D1 推荐） | `SiteHomePage` | 否 | 官网首页，独立 `SiteShell`，不挂 `SessionProvider` / `AppShell` |
| `/p/:slug` | `SiteDocumentPage` | 否 | 产品说明等子页 |
| `/login`、`/register` | 现有页面 | 否 | 页脚备案改为读站点配置，删除 `i18n` 里的备案占位 |
| `/app` | `LandingRedirect` | 是 | 从 `/` 迁来，`ignoredPaths` 与 `tenantFreePaths` 同步 |
| `/system/site` | `SiteAdminPage` | 是 | 平台级，`tenantFreePaths` 与侧栏 `nav.site` 入口 |
| `/notification-oncall` | 现有页面 | 是 | 时间视图加「全部 / 我的」切换，视角写进 URL（`?view=mine`） |

官网取数走原生 `fetch`（仿 `web/src/api/status-page/public.ts`），不得走 `@/api/http`：后者带 401/403 拦截，会把免登录访客送回登录页。备案、版权、登录方式共用一份站点配置（`useSiteConfig`），一个页面只请求一次。

### 新增组件

`SiteShell`、`SiteHeader`、`SiteFooter`、`SiteFilingBar`（官网页脚 / 登录页脚 / 后台预览三处共用）、`SiteAccountMenu`、`SiteTenantPickerDialog`、`SitePageBody`（Markdown 渲染，公开页与后台预览共用）、`AccountAvatar`（从 `app-shell.tsx` 抽出）。后台侧：`SitePageList`、`SitePageEditor`、`SitePublishDialog`、`SiteSettingsForm`、`SiteFeatureEditor`、`SiteFilingFields`。

### 官网首页区块

Hero（`text-display-1` 标题、`text-lead` 副标、≤2 个 CTA、主视觉 4:3）→ 能力网格（1/2/3 列，空则整块不渲染）→ 产品说明摘要 → 尾 CTA。容器 `max-w-6xl`，区块纵向 40/56/80px，正文阅读宽 68ch。

### 状态

四态必须区分：加载（骨架复刻版式）、空（尚未发布 / 无能力卡片）、失败（可重试，`role="alert"`）、无权限（后台直达 `/forbidden`）。未发布的 `/p/:slug` 统一 404 文案，不泄露草稿存在性。官网不弹全局错误 toast，也不因取数失败跳登录页。

### 后台编辑器

左列页面列表（标题、状态徽标、slug、语言、有未发布改动），右列编辑器（基本信息、Markdown 正文、导航、SEO 折叠区），操作条粘性包含预览 / 保存草稿 / 发布。离开前有 `dirty` 保护；`409` 给「内容已在别处修改」并保留用户输入。

## 6. 品牌与视觉

官网是只做加法的第四层：L0 品牌层锁 `data-brand="sunset"` 不暴露切换器，L1 语义层原样复用，新增 `web/src/styles/site/site.tokens.css` 只放 `--site-*`（容器宽、段落节奏、阴影、色带、展示字号），`index.css` 的 `@theme inline` 追加映射。**官网页面里不允许出现任何颜色字面量或调色板类。**

| 项 | 规定 |
| --- | --- |
| 主色 | 暖阳橙 `#DB7A0E`（填充）/ `#8C4D0D`（文字图标）；CTA 用 `#171717` 压主色（5.81:1），**永远不要白字压主色**（3.09:1） |
| 蓝色 | logo 自带星蓝 `#1677FF→#0A43B5` 只作商标资产色，不升为第二主色；蓝色不得作正文或链接色（4.10:1 不达 AA） |
| 字体 | 现有 `--font-heading` 不含中文字形，官网显式声明 CJK 回退（PingFang SC / HarmonyOS Sans SC / Noto Sans SC / Microsoft YaHei），中文只用 400 与 600 |
| 展示字号 | display-1 72/44、display-2 56/36、display-3 40/28、section 32/24（桌面/移动） |
| 圆角 | 按钮与输入 10/8px、卡片 18px、截图框 22px；**不用胶囊按钮** |
| 阴影 | 只给浮层与截图框，普通卡片用 1px 边框 |
| 页脚 | 4 栏（品牌 + 产品 / 资源 / 条款），法律条 12px 不低于、备案号每个页面都可见、链接 `target="_blank" rel="noreferrer"` |
| 深浅色 | 默认浅色、尊重系统偏好、三态开关；深色下品牌填充换 `#F0891F`（7.44:1），截图必须进 `--card` 容器 |
| 文案 | 工程师写给工程师，能力带边界；禁用「赋能 / 闭环 / 一站式 / 毫秒级 / 海量」等词 |
| 红线 | 未证实数字、未授权客户 logo、P 图或自制 mock、把未上线能力当现售、自动轮播与弹窗订阅、品牌色承载告警状态 |

## 7. 验收

每条都写清角色、入口、断言与请求证据；缺一项视为未验证。**浏览器验证目前未执行，任何一条在跑完前都不能判 PASS。**

- [ ] R1 官网：匿名打开首页停在官网（不跳 `/login`），`GET /v1/site/config` 无 `Authorization` 且全程 0 次 `/v1/self/*`；草稿不出现；`/p/:slug` 未发布 404；接口 500 时显示可重试错误态且不跳登录；正文里的 `<script>`、`onerror`、`javascript:` 不生效。
- [ ] R2 后台：超级管理员可编辑、保存草稿、预览、发布、下架；发布后匿名侧即时可见，保存草稿不影响线上；`home` 不能下架删除；并发保存 409 且不静默覆盖；租户 owner 与匿名调接口分别 403 / 401。
- [ ] R3 登录：官网点登录 → `/login?next=…` → 登录后回原页；1 个租户直进，多个弹选择，0 个走建租户引导；停用租户不可选；`next=https://evil.com`、`//evil.com` 不跳出站外；刷新不跳后台。
- [ ] R4 备案：页脚文本等于接口值（不再是 `京ICP备00000000号-1` 占位），链接与 `rel` 正确；后台改为空后该行整条消失且布局不塌；`javascript:` 之类的 scheme 不渲染成可点链接；登录页、官网页、公开状态页三处一致，单页只请求一次配置。
- [ ] R5 OAuth：全关时无第三方按钮、无分隔线、邮箱表单直接展开且可登录；部分开时只出现已开平台且顺序与配置一致；关掉后刷新不出现；开关关闭时直接调 `/v1/auth/oauth2/login` 不签发 token；个人资料里已绑定平台仍可见。
- [ ] R6 认领人列（已修）：保存 `PUT /v1/alert-pages/my-columns` 返回 200，表头出现「认领人」，刷新后仍在，页面无错误文案；三个 scope（我的 / 告警页面 Tab / 历史明细）一致；未认领单元格显示文字占位而非留空。
- [ ] R7 个人视图：出现独立于「范围」下拉的「全部 / 我的」切换，`aria-pressed` 与可见态一致、URL 可回放；只出现本人班次且与排班接口逐段一致；跨午夜拆段正确；无值班时给空态与「去值班轮换」入口；排班失败显示失败 + 重试，不伪装成「没有值班」；无 `notification_rotation:read` 的成员也能看到自己的班次。

通用拦截项（任一命中即 HOLD）：四态不可区分或失败态无重试；出现未映射英文原文或只给兜底文案；错误不指向具体对象、失败后草稿丢失；切换状态只靠颜色、键盘不可达；390×844 横向滚动或弹窗超 80vw/80vh；同一视图两个纵向滚动条；Tabs 的 `value`、`aria-selected`、可见面板与 URL 不一致；写入口未按权限显隐；新增字面色值或调色板类。

## 8. 验收用例索引

完整用例矩阵共 75 条（P0 23 / P1 39 / P2 13），按需求分组，P0 是合并门禁：

| 组 | 需求 | 条数 | P0 用例 |
| --- | --- | --- | --- |
| `SITE` | R1 | 9 | SITE-01 匿名可读、SITE-02 草稿不泄露 |
| `ADM` | R2 | 9 | ADM-01 回填、ADM-03 草稿不影响线上、ADM-04 发布即生效 |
| `LOGIN` | R3 | 12 | LOGIN-01 登录后留在官网、LOGIN-02 零租户、LOGIN-03 单租户、LOGIN-04 多租户选择 |
| `ICP` | R4 | 9 | ICP-01 文本等于接口值、ICP-02 登录页同步、ICP-03 空值不渲染 |
| `OAUTH` | R5 | 11 | OAUTH-01 全关、OAUTH-02 单平台、OAUTH-03 顺序、OAUTH-04 关闭后消失 |
| `COL` | R6 | 12 | COL-01 默认列可保存、COL-02 勾选后保存 200、COL-03 持久化 |
| `ONCALL` | R7 | 13 | ONCALL-01 无 read 权限也可用、ONCALL-02 只出现本人、ONCALL-03 当前班次、ONCALL-04 下一班 |

选择器优先级：`getByRole` → `getByLabel` → 现有结构属性（`[data-provider]`、`[data-sort-key]`）→ `data-testid`。`data-testid` 全仓目前为 0，限量新增（官网登录/进入后台入口、备案链接、个人视图班次行等语义拿不到的锚点）。时间与语言钉死（`locale: zh-CN`、`timezoneId: Asia/Shanghai`），不写 `waitForTimeout`，写操作等响应、读断言用 web-first 自动重试。

自动化前置（阻塞，需拍板）：图形验证码测试通道、JWT 直签注入会话、`megrez` 的 `ReportEval` 造事件；三者都不批就只能靠 OCR 与真实评测链路，稳定性不可接受。站点信息与登录开关是平台级单例，用例必须放在串行 project 并在前后做快照恢复。

## 9. 不做什么

- 封面图与 logo 只收 URL，不做文件上传。
- 不做页面版本历史、定时发布、多站点、草稿公开预览链接、官网站内搜索、sitemap 生成。
- 不把官网做成租户级营销页；不做自动轮播、弹窗订阅、视差动效。
- 不在本期公开定价（D7 未定）。
- 不改认领接口语义、不改时间视图的编码方式（位置 / 颜色 / 纹理 / 文字 / 空档底纹 / `sr-only` 摘要），个人视图复用这套表达，不另起饼图或卡片流。

## 10. 风险

| 风险 | 说明 | 处理 |
| --- | --- | --- |
| 路由迁移（D1） | `/` 改官网会牵动 `redirect.ts` 默认 `next`、`last-page.ts` 白名单、`tenantFreePaths`、所有深链 | 选 `/site/*` 则零风险；选 `/` 就按第 5 节清单逐条改并把 404 回退跑一遍 |
| 身份映射（D6） | 平台用户 UID 与租户成员不是同一个标识，按姓名猜会误判 | 后端给 `isMe` 与自服务接口，前端不猜 |
| 时间轴（D6） | 周网格画不出按绝对日期轮转的班次 | 日期 × 24 小时轴，或退化为「当前班次 / 下一班」摘要 |
| 全局单例配置 | 站点信息与登录开关被并行用例互相覆盖 | 串行 project + 快照恢复 |
| 实时告警轮询 | 10s 刷新会覆盖本地乐观更新 | 断言服务端最终态再断言 UI |
| 未验证 | 本轮只做代码与契约交叉核对，浏览器、登录、跳转都没跑 | 逐条按第 7 节补验，未跑就写「未验证」 |

## 11. 现状记录

- R6 的根因不是认领按钮：前端默认列与可选字段都含 `acked`，后端 `allowedFieldColumns` 缺 `acked`，任何一次保存都 400 `invalid request`，弹窗落兜底文案。已在 `fix/alert-page-acked-column` 修复（`e8c9a46e`），并加回归测试钉住「实时告警默认列能直接保存」。
- R4 的现状是硬编码占位：`web/src/i18n/index.ts` 的 `login.icp = 京ICP备00000000号-1`，登录页直接渲染。
- R5 的现状是只看 env：`OAuthProviders` 依据 `client_id` / `client_secret` 是否为空决定是否下发，没有开关；`StartOAuth` 也没有第二道拦截。
- R7 的现状是入口不存在：时间视图只有「周/日」「范围」「包含停用」，排班只出现在轮换详情抽屉里，且需要 `notification_rotation:read`。
- 公开面现状：`App.tsx` 的公开路由只有 `/login`、`/register`、`/auth/callback`、`/invite/:token`、`/status/:token`，其余都在 `AppShell` 内，无 token 直接跳登录。
