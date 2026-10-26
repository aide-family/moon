# 个人资料页「OAuth2 账号关联」区块设计

目标页面：`web/src/pages/profile/index.tsx`（路由 `/profile`，`AppShell` 内、租户无关页）。
现状：页面只有一张 `glass-card max-w-2xl` 的「基本资料」卡片（身份 + 角色/状态 + 手机号/备注表单）。
`GetSelfInfoResponse` 只有 `user/system_role/system_permissions`，`AuthService` 只提供登录与绑定邮箱，
`SelfService` 没有任何自侧 OAuth 关联能力，`user_oauth(uid, user_uid, provider, open_id, email)` 表已存在但只能由登录链路写入。

---

## 1. 信息架构与布局

### 1.1 与「基本资料」卡片的关系

- **同页第二个平级卡片**，不做 Tab。理由：资料页只有两个关注点（我是谁 / 我能用什么方式登录），
  当前信息量一屏可容纳；Tab 会引入 URL 状态、`aria-selected` 与键盘验证成本（见 `web/AGENTS.md`
  的 tabs 回归要求），收益为零。
- 卡片顺序：**基本资料 → 第三方账号**。账号关联是低频操作，放在可编辑资料之后，避免首屏被绑定按钮占据。
- 两张卡片视觉同级：都用 `glass-card`、`max-w-2xl`、`--card-spacing` 默认密度（16px），卡片间距 `gap-6`。

### 1.2 位置与栅格

页面现在的滚动所有者是 `div.min-h-0.flex-1.overflow-y-auto`（profile/index.tsx:117），保留不变，
只在内部套一层两卡片栅格。`AppShell` 是 `h-svh overflow-hidden`，因此**不允许**新增第二层滚动容器。

```tsx
<div className="min-h-0 flex-1 overflow-y-auto">
  <div className="grid max-w-2xl gap-6 pb-2">
    <Card className="glass-card">{/* 基本资料，现有内容不动，去掉自身的 max-w-2xl */}</Card>
    <OAuthAccountsCard />
  </div>
</div>
```

- 单列（`grid`，不写 `md:grid-cols-2`）：两卡片内容性质不同，并排会让表单列宽掉到 ~320px。
- 宽度继承 `max-w-2xl`（42rem），与现有卡片严格对齐；窄屏（<672px）自然占满，`p-4` 由 shell 提供。
- 密度：卡片内 `CardHeader` + `CardContent`；行高由 36px provider 图标 + `py-3` 决定 ≈ 56px，
  四个 provider 合计约 224px，加表头与摘要 ≈ 320px，与基本资料卡片（≈420px）合计仍在一屏 768px 内可滚动读完。

### 1.3 单行结构（横向）

```
┌──────────────────────────────────────────────────────────────┐
│ [图标 36px]  GitHub                          [已绑定] [解绑] │
│              已绑定的第三方邮箱 / open_id 尾号                │
└──────────────────────────────────────────────────────────────┘
```

- `≥sm`：`flex flex-nowrap items-center gap-x-3`，文本块 `min-w-0 flex-1`，操作区 `shrink-0`。
- `<sm`：行内 `flex-wrap gap-y-2`，操作区落在第二行、右对齐（`ml-auto`），文本 `break-all`。
- 行间用 `divide-y divide-border/60`（**不**用嵌套 Card，避免卡中卡）。
- 标题行右侧可放一个弱化摘要 `StatusBadge`：`已绑定 2 / 4`（`neutral` tone），不参与操作。

### 1.4 卡片头部文案

- 标题：`第三方账号`
- 描述：`绑定第三方账号后可以用它直接登录 Alioth；解绑后该账号不能再用于登录。`

---

## 2. 组件树与关键 props / 数据形状

```
ProfilePage                                   （现有，只加一行挂载）
└── OAuthAccountsCard                          web/src/components/oauth-accounts-card.tsx
    ├── Card / CardHeader / CardDescription / CardContent
    │   ├── OAuthAccountsSkeleton              （loading）
    │   ├── InlineFailure                      （failed：role=alert + 重试 Button）
    │   ├── InlineForbidden                    （403：说明 + 重试）
    │   ├── InlineEmpty                        （无任何已配置 provider）
    │   └── ul.divide-y
    │       └── OAuthAccountRow × n            （同文件内的纯展示子组件）
    │           ├── OAuthProviderMark          web/src/components/oauth-provider-mark.tsx
    │           └── Button(outline) 绑定 / Button(destructive) 解绑
    └── ConfirmAction（单个实例，由 confirmTarget 驱动）  components/confirm-action.tsx（复用）
```

### 2.1 数据形状（`web/src/api/self/types.ts` 追加）

```ts
import type { OAuth2Provider } from "@/api/auth/types"

/** 当前账号已绑定的一条第三方身份。 */
export type SelfOAuthLink = {
  uid: string                  // user_oauth.uid，解绑按它对位
  provider: OAuth2Provider
  openId: string               // 只展示尾 4 位，前端掩码
  email: string                // 绑定时从第三方拿到的邮箱，可能为空
  boundAt?: string             // RFC3339，来自 user_oauth.created_at
}

export type SelfOAuthLinks = { items: SelfOAuthLink[] }

export type BindSelfOAuthParams = {
  provider: OAuth2Provider
  code: string
  redirectUri: string
}

export type SelfOAuthBindingResult = { item: SelfOAuthLink }
```

### 2.2 行模型（卡片内）

```ts
type ProviderRow = {
  provider: OAuth2Provider
  link?: SelfOAuthLink     // 有值 = 已绑定
  configured: boolean      // 平台已配置该来源的 client_id/secret
}
```

行集合 = **已配置来源 ∪ 已绑定来源**，顺序按 catalog（`GITHUB, GITEE, FEISHU, DINGTALK`），未知来源追加在末尾。
理由：只列已配置来源能避免死按钮；把「已绑定但来源已下线」保留成行，用户仍有解绑出口。

### 2.3 组件 props

```ts
// oauth-provider-mark.tsx（从 oauth-provider-button.tsx 抽出，复用同一份图标映射）
export function oauthProviderIcon(provider: OAuth2Provider): ComponentType<{ className?: string }>
export function OAuthProviderMark({ provider, className }: { provider: OAuth2Provider; className?: string })

// oauth-accounts-card.tsx
export function OAuthAccountsCard({ className }: { className?: string })   // 自持数据，无外部 props

// 同文件内
function OAuthAccountRow(props: {
  row: ProviderRow
  pending: boolean                  // 本行绑定跳转中
  disabled: boolean                 // 其他行在跑 / 全局 pending
  onBind: (provider: OAuth2Provider) => void
  onUnbind: (link: SelfOAuthLink) => void
})
```

### 2.4 卡片内部状态

```ts
const [links, setLinks] = useState<SelfOAuthLink[] | null>(null)   // null = 未加载
const [providers, setProviders] = useState<OAuth2Provider[]>([])   // 已配置来源
const [loading, setLoading] = useState(true)
const [failed, setFailed] = useState(false)
const [forbidden, setForbidden] = useState(false)
const [loadError, setLoadError] = useState("")
const [attempt, setAttempt] = useState(0)
const [binding, setBinding] = useState<OAuth2Provider | "">("")
const [confirmTarget, setConfirmTarget] = useState<SelfOAuthLink | null>(null)
const [unbinding, setUnbinding] = useState(false)
const [unbindError, setUnbindError] = useState("")
```

---

## 3. 状态清单与交互流程

### 3.1 状态清单（渲染分支，互斥）

| # | 状态 | 触发 | 渲染 |
|---|------|------|------|
| S1 | 加载中 | 首次挂载 / 重试 | 3 行 `OAuthAccountsSkeleton`（36px 图标块 + 两行文字 + 32px 按钮位），`CardContent` 带 `aria-busy` |
| S2 | 加载失败 | list 非 404/403 错误 | `role="alert"` 错误文案（`loadError \|\| t("state.failed")`）+ `t("action.retry")`（与基本资料卡片错误块完全同构） |
| S3 | 权限不足 | list 返回 403 | `InlineForbidden`：`t("profile.oauth.forbiddenHint")` + 重试 |
| S4 | 空态（无来源） | 已配置来源 = 0 且已绑定 = 0 | 图标 + `emptyTitle` + `emptyHint`（引导联系管理员） |
| S5 | 全部未绑定 | rows 全部无 link | 每行 `未绑定` 徽标 + `绑定` 按钮 |
| S6 | 单行已绑定 | 该行有 link | `已绑定` 徽标（success）+ 第三方邮箱/尾号 + `解绑`（destructive） |
| S7 | 绑定进行中 | 已点绑定、正在取授权地址或跳转 | 该行按钮 `aria-busy` + `Loader2 animate-spin` + `t("profile.oauth.binding")`；其余行 `disabled`；按钮保持原位不塌陷 |
| S8 | 解绑二次确认 | 点解绑 | `ConfirmAction` 打开，`confirmLabel = 解绑`，说明含 provider 名 + 已绑定身份 + 「已有邮箱仍可登录」 |
| S9 | 解绑进行中 | 对话框确认后 | `ConfirmAction` 按钮 disabled（新增可选 `pendingLabel`，默认沿用 confirmLabel） |
| S10 | 解绑失败 | 409/404/5xx | 对话框内 `error`（409 = 最后一种登录方式；其他 = 通用失败）；404 视为已解绑 → 关窗 + 静默重拉 |
| S11 | 绑定回落提示 | 回到 `/profile?oauth=linked\|conflict\|failed&provider=X` | 一次性 toast（success/error），随后 `navigate(replace)` 清 query 并重拉列表 |

补充说明：

- **U1 后端未部署（404）**：list 走 `{ silent: true }`，404 直接 `setFailed(false)` 并**整块不渲染**
  （返回 null），让页面在后端 RPC 上线前保持原样，不出现「加载失败」。
- **U2 部分来源取不到授权地址**：`bind` 时找不到该 provider 的 authorizeUrl → `toast.error(authorizeUnavailable)`，复位 pending。
- **U3 多标签页并发**：解绑 404 按成功处理；列表在回到页面（`window.focus`）不轮询，只在 11 的一次性重拉里对齐。

### 3.2 绑定流程（含 state 参数）

```
[Profile] 点击「绑定」
  → setBinding(provider)
  → state = newOAuthState()                       // crypto.randomUUID()，一次性
  → providers = await listOAuth2Providers({ redirectUri: oauthRedirectUri(), state })   // 复用现有公开接口
  → 找不到 provider.authorizeUrl → toast.error(authorizeUnavailable) → 复位
  → storeOAuthRedirect({ state, provider, redirectUri: oauthRedirectUri(), next: "/profile", intent: "bind" })
  → window.location.assign(authorizeUrl)          // 整页跳转（与登录流程一致）
[第三方] 授权 → 302 /auth/callback?code=...&state=...
[Callback] readOAuthRedirect(state)
  ├─ 无 redirect / state 不匹配 → 现有 stateMismatch 文案
  ├─ redirect.intent === "bind"
  │   ├─ 有 error 参数 → 显示 providerError + 详情，出口改为「返回个人资料」（不碰会话）
  │   ├─ 无 code → missingCode
  │   ├─ localStorage token 为空 → t("profile.oauth.sessionExpired") + 「返回登录」
  │   └─ bindSelfOAuth({ provider, code, redirectUri })
  │        ├─ 成功 → clearOAuthRedirect(state) → navigate("/profile?oauth=linked&provider=GITHUB", { replace: true })
  │        ├─ 409 "oauth account is already bound" → navigate("/profile?oauth=conflict&provider=GITHUB", { replace: true })
  │        └─ 其他 → 现有失败块（detail 保留），出口「返回个人资料」
  └─ intent === "login"（默认）→ 现有登录/补邮箱逻辑，一行不改
[Profile] useSearchParams 读到 oauth → toast 一次 → navigate("/profile", { replace: true }) → 重拉列表
```

**state 参数约定**

- `state` 仍然只存 `sessionStorage["alioth.oauth.<state>"]`，回调凭 URL 的 `state` 取回并校验（现有机制不变）。
- `OAuthRedirect` 增加可选字段 `intent?: "login" | "bind"`；`readOAuthRedirect` 缺省补 `"login"`，
  老会话里的现场不会失效。`clearOAuthRedirect` 在成功、provider 报错、code 缺失、冲突四种出口都调用，保证一次性。
- **bind 路径永不调用 `oauth2Login`**：若第三方身份已绑定到另一个账号，`OAuth2Login` 会直接给那个人签发会话，
  在已登录页面里等同于账号劫持。绑定必须走带鉴权的 `BindSelfOAuth`，由它返回 409。
- 每次点击生成新 state（不是页面级复用），避免上一次残留 state 被复用后被回调页判为 mismatch。

### 3.3 解绑流程

```
[Profile] 点击「解绑」→ setConfirmTarget(link)，清空 unbindError
[Dialog] 确认 → setUnbinding(true) → unbindSelfOAuth(link.uid, { silent: true })
  ├─ 成功 → setLinks(prev => prev.filter(l => l.uid !== link.uid))
  │          toast.success(t("profile.oauth.unlinkSuccess", { provider }))
  │          关闭对话框；焦点回到该行（此时按钮已变成「绑定」）
  ├─ 404  → 视为已解绑：本地剔除 + 关窗（不报错）
  ├─ 409  → setUnbindError(t("profile.oauth.unbindLastCredential"))，对话框保持打开
  └─ 其他 → setUnbindError(t("profile.oauth.unlinkFailed"))
```

### 3.4 错误提示与 toast 文案矩阵

| 场景 | 位置 | i18n key |
|------|------|----------|
| 列表加载失败 | 卡片内 `role="alert"` + 重试 | `state.failed` / `action.retry` |
| 权限不足（403） | 卡片内文案 + 重试 | `profile.oauth.forbiddenHint` |
| 无来源可绑 | 卡片内空态 | `profile.oauth.emptyTitle` + `emptyHint` |
| 绑定成功 | toast.success | `profile.oauth.linkSuccess`（`{{provider}}`） |
| 绑定冲突（已绑到他人） | toast.error + 该行说明 | `profile.oauth.alreadyBound` |
| 绑定失败（其他） | 回调页失败块 / toast.error | `profile.oauth.linkFailed` |
| 绑定时会话失效 | 回调页 | `profile.oauth.sessionExpired` |
| 授权地址取不到 | toast.error | `profile.oauth.authorizeUnavailable` |
| 解绑成功 | toast.success | `profile.oauth.unlinkSuccess` |
| 解绑最后一个凭证 | 对话框内 | `profile.oauth.unbindLastCredential` |
| 解绑失败 | 对话框内 | `profile.oauth.unlinkFailed` |
| 后端原文兜底 | `api/http.ts` `messageKeys` | 见 §5.3 |

toast 形态沿用现有 `toast.success(title, { description })` 习惯：标题为结果，描述为 provider 名或下一步提示；
不使用长句堆在标题里。

### 3.5 无障碍与暗色

- `CardContent` 在加载中带 `aria-busy="true"`；每行是 `<li>`，`ul` 无 `role` 覆盖。
- 徽标复用 `StatusBadge`（`value="BOUND"` + `label` + `tone`），不新增颜色体系；
  图标区的品牌色沿用 `.oauth-provider[data-provider]` 的 token（含 `.dark` 覆盖），不改硬编码色值。
- 按钮一律用 `Button` 默认尺寸（`h-8`）；**不**为了对齐 36px 图标块而改写 `h-9/h-10`（`web/AGENTS.md` 硬规则），
  用 `items-center` 居中即可。
- 解绑按钮为 `variant="destructive"` 文本按钮，命中区域 32px；不改高度换 44px。
- 不引入自定义动画，仅沿用 `animate-spin`；`prefers-reduced-motion` 由 tailwind 默认行为兜底。

---

## 4. i18n 文案键清单

全部落在 `web/src/i18n/index.ts` 的 `profile.oauth`（zh 与 en 各一份，key 完全对齐），
以及 `oauthCallback.*` 的 3 个新增键。provider 展示名统一用 `t("enum." + provider)`
（`enum.GITHUB/GITEE/FEISHU/DINGTALK` 已在两语言存在），不使用后端返回的中文 `name`。

| key | zh | en |
|-----|----|----|
| `profile.oauth.title` | 第三方账号 | Linked accounts |
| `profile.oauth.description` | 绑定第三方账号后可以用它直接登录 Alioth；解绑后该账号不能再用于登录。 | Link a GitHub, Gitee, Feishu, or DingTalk account to sign in directly. Unlinking removes that sign-in method. |
| `profile.oauth.summary` | 已绑定 {{bound}} / {{total}} | {{bound}} of {{total}} linked |
| `profile.oauth.bound` | 已绑定 | Linked |
| `profile.oauth.notBound` | 未绑定 | Not linked |
| `profile.oauth.bind` | 绑定 | Link |
| `profile.oauth.binding` | 正在跳转… | Redirecting… |
| `profile.oauth.unbind` | 解绑 | Unlink |
| `profile.oauth.unbinding` | 解绑中… | Unlinking… |
| `profile.oauth.accountLabel` | 第三方账号 | Third-party account |
| `profile.oauth.accountUnknown` | 平台未返回账号信息 | The provider returned no account details |
| `profile.oauth.boundAt` | 绑定于 {{time}} | Linked {{time}} |
| `profile.oauth.emptyTitle` | 没有可绑定的第三方登录 | No third-party sign-in available |
| `profile.oauth.emptyHint` | 平台还没有配置任何第三方登录来源。请联系管理员在服务端配置后再绑定。 | No provider is configured on the server yet. Ask an administrator to configure one before linking. |
| `profile.oauth.forbiddenHint` | 当前账号不能管理第三方账号关联。 | Your account cannot manage linked accounts. |
| `profile.oauth.linkSuccess` | 已绑定 {{provider}} | {{provider}} linked |
| `profile.oauth.unlinkSuccess` | 已解绑 {{provider}} | {{provider}} unlinked |
| `profile.oauth.linkFailed` | 绑定失败，请重试。 | Linking failed. Please try again. |
| `profile.oauth.unlinkFailed` | 解绑失败，请重试。 | Unlinking failed. Please try again. |
| `profile.oauth.alreadyBound` | 该 {{provider}} 账号已绑定到其他账号。 | That {{provider}} account is already linked to another account. |
| `profile.oauth.unbindLastCredential` | 这可能是你最后的登录方式，不能解绑。请先确认邮箱可以正常登录。 | This may be your only sign-in method, so it cannot be unlinked. Make sure email sign-in works first. |
| `profile.oauth.authorizeUnavailable` | 暂时无法获取 {{provider}} 的授权地址，请稍后重试。 | Cannot get the {{provider}} authorization URL right now. Please try again later. |
| `profile.oauth.sessionExpired` | 登录状态已失效，请重新登录后再绑定。 | Your session expired. Sign in again before linking. |
| `profile.oauth.confirmTitle` | 解绑 {{provider}}？ | Unlink {{provider}}? |
| `profile.oauth.confirmDescription` | 解绑后该第三方账号不能再直接登录 Alioth。已有邮箱仍可正常登录。 | After unlinking, this account can no longer sign in to Alioth. Email sign-in still works. |
| `oauthCallback.linking` | 正在绑定 {{provider}}… | Linking {{provider}}… |
| `oauthCallback.linkedTitle` | 正在绑定第三方账号 | Linking a third-party account |
| `oauthCallback.backToProfile` | 返回个人资料 | Back to profile |

复用既有键（不重复定义）：`action.retry`、`action.cancel`、`action.save`、`state.failed`、`state.loading`、
`form.required`、`enum.GITHUB/GITEE/FEISHU/DINGTALK`、`error.oauthFailed`。

---

## 5. 依赖的后端能力缺口与建议的 RPC / 字段

### 5.1 现状缺口

| 缺口 | 现状 | 影响 |
|------|------|------|
| 列出当前用户绑定 | 无。`OAuthAccount(provider, open_id)` 只能按第三方身份反查账号 | 前端无法渲染任何已绑定行 |
| 自侧绑定 | 无。只有 `OAuth2Login`（换会话）与 `BindOAuthEmail`（补邮箱） | 已登录状态下无法把新来源绑到当前账号 |
| 自侧解绑 | 无。repo 只有 `LinkOAuth/MoveOAuth`，没有删除 | 无法解绑 |
| 关联时间 | `user_oauth` 有 `created_at`（Audit mixin），但 proto 无字段 | 「绑定于」无法展示 |
| Feishu | 后端 catalog 已包含 FEISHU 且 `ListOAuth2Providers` 会返回它 | 前端若硬编码四家会与登录页不一致 |

### 5.2 建议 RPC（`api/alioth/v1/identity.proto`，挂在 `SelfService`）

```proto
// 列出当前账号已绑定的第三方身份。
rpc ListSelfOAuthLinks(ListSelfOAuthLinksRequest) returns (ListSelfOAuthLinksResponse) {
  option (google.api.http) = {get: "/v1/self/oauth/links"};
}
// 用第三方授权码把身份绑定到当前账号。不签发会话，不改变登录身份。
rpc BindSelfOAuth(BindSelfOAuthRequest) returns (BindSelfOAuthResponse) {
  option (google.api.http) = {post: "/v1/self/oauth/bind" body: "*"};
}
// 解绑当前账号的一条第三方身份。
rpc UnbindSelfOAuth(UnbindSelfOAuthRequest) returns (UnbindSelfOAuthResponse) {
  option (google.api.http) = {delete: "/v1/self/oauth/links/{uid}"};
}

message OAuthLink {
  string uid = 1;                 // user_oauth.uid
  string provider = 2;            // GITHUB/GITEE/FEISHU/DINGTALK
  string open_id = 3 [json_name = "openId"];
  string email = 4;               // 绑定时第三方返回的邮箱，可能为空
  string bound_at = 5 [json_name = "boundAt"];   // RFC3339，来自 created_at
}

message ListSelfOAuthLinksRequest {}
message ListSelfOAuthLinksResponse { repeated OAuthLink items = 1; }

message BindSelfOAuthRequest {
  string provider = 1 [(buf.validate.field).string = {in: ["GITHUB","GITEE","FEISHU","DINGTALK"]}];
  string code = 2 [(buf.validate.field).string.min_len = 1];
  string redirect_uri = 3 [json_name = "redirectUri"];   // 与授权时一致
}
message BindSelfOAuthResponse { OAuthLink item = 1; }

message UnbindSelfOAuthRequest {
  string uid = 1 [(buf.validate.field).string.min_len = 1];
}
message UnbindSelfOAuthResponse {}
```

实现约定（重要，关系到前端状态可达性）：

1. 三个 RPC 都只要求**登录态**，不加权限码（与 `GetSelfInfo`/`ChangePhone` 一致）。因此正常路径不会产生 403。
2. 业务拒绝一律用 `merr.Client/Conflict`（400/409），**禁止用 `merr.Forbidden`（403）**：
   前端 `api/http.ts` 对 403 会整页跳 `/forbidden`，会让「权限不足」内联态失效（见 §6-R1）。
3. `BindSelfOAuth` 逻辑：
   - 用现有 `exchangeOAuth` 换 profile；`(provider, open_id)` 已绑定**本账号** → 幂等返回该 `OAuthLink`；
   - 已绑定**其他账号** → `merr.Conflict("oauth account is already bound")`；**不调用 `MoveOAuth`**；
   - 未绑定 → `LinkOAuth(self, provider, openID, profile.email)`，返回新 `OAuthLink`。
4. `UnbindSelfOAuth` 逻辑：`DELETE WHERE uid = ? AND user_uid = self`（必须带 `user_uid`，防 IDOR）；
   0 行 → `merr.NotFound("oauth link not found")`；若删除后该账号**没有任何可用登录方式**
   （既无可用邮箱也无其他 OAuth 绑定，含历史 `@users.alioth.local` 占位邮箱）→
   `merr.Conflict("cannot unlink the last login method")`。
5. repo 追加（`app/alioth/internal/data/repository/oauth.go` + `data/oauth_repo.go`）：
   `ListOAuthByUser(ctx, userUID) ([]bo.OAuthLink, error)`、`UnlinkOAuth(ctx, userUID, uid) error`、
   `OAuthByUserProvider(ctx, userUID, provider) (*bo.OAuthLink, error)`。
   `bo` 侧新增 `OAuthLink{UID, Provider, OpenID, Email, BoundAt}`。
6. 代码生成：`make api`（buf generate）+ `cd app/alioth/internal/data/ent && go generate ./...`（仅当改动 schema 时需要，本次不需要）。

### 5.3 前端 `api/http.ts` 的兜底映射（可选但建议）

```ts
"oauth account is already bound": "profile.oauth.alreadyBound",
"oauth link not found": "profile.oauth.linkMissing",        // 可复用 error.notFound
"cannot unlink the last login method": "profile.oauth.unbindLastCredential",
```

同时前端分支判断用 `ApiError.raw`（现有 `oauth-email-binding.tsx` 的既有做法），不比对本地化文本。

### 5.4 可选的后续增强（不在本次范围）

- `GetSelfOAuthBindURL(provider)`：服务端签发带 kind 的 state，替代复用公开的
  `POST /auth/oauth2/providers`，可把「绑定意图」从客户端 sessionStorage 搬到服务端，安全性更好。
- 列表响应增加 `can_manage` / `denied_reason`，让权限不足态可主动渲染而不是等 403。
- 审计：`UserOAuth` 目前是硬删，解绑不留痕；如需审计应改软删并加 `revoked_at`。

---

## 6. 按文件排列的可执行实现步骤

> 顺序 = 依赖顺序。后端先行，前端 feature-detect（404 隐藏卡片）可以并行开发。

**后端**

1. `api/alioth/v1/identity.proto`
   - `SelfService` 增加 `ListSelfOAuthLinks` / `BindSelfOAuth` / `UnbindSelfOAuth`（§5.2 注释与 http 注解）；
   - 新增 `OAuthLink` 与三个 request/response message。
2. 运行 `make api`；确认 `api/alioth/v1/*.pb.go` 与 `*_http.pb.go` 更新。
3. `app/alioth/pkg/bo/bo.go`：新增 `OAuthLink` 结构体。
4. `app/alioth/internal/data/repository/oauth.go`：接口追加 3 个方法（含注释）。
5. `app/alioth/internal/data/oauth_repo.go`：实现 `ListOAuthByUser` / `OAuthByUserProvider` / `UnlinkOAuth`
   （`UnlinkOAuth` 必须 `useroauth.UIDEQ(uid), useroauth.UserUIDEQ(userUID)`；0 行返回 `merr.NotFound`）。
6. `app/alioth/internal/biz/auth/oauth.go`：新增 `SelfOAuthLinks(ctx)`、`BindSelfOAuth(ctx, provider, code, redirectURI)`、
   `UnbindSelfOAuth(ctx, uid)`；复用 `exchangeOAuth`，`actor.UserID(ctx)` 取当前账号；
   约束 409/400 文案与 §5.2 一致。
7. `app/alioth/internal/service/self.go`：三个 handler，做 bo ↔ proto 映射（新增 `oauthLinkItem` 辅助函数，风格对齐 `userItem`）。
8. 测试：`app/alioth/internal/biz/auth/oauth_self_test.go`（新增）
   - 绑定已在他人名下的 open_id → 409；
   - 重复绑定到本账号 → 幂等；
   - 解绑他人 uid → 404（不泄露存在性）；
   - 解绑最后一个登录方式 → 409。
9. 校验：`go test ./app/alioth/internal/biz/auth/... ./app/alioth/internal/data/...`、`make lint`（gofmt/命名/buf lint）。

**前端**

10. `web/src/i18n/index.ts`：写入 §4 全部 `profile.oauth.*` 与 `oauthCallback.linking/linkedTitle/backToProfile`（zh + en）。
11. `web/src/api/http.ts`：`del` 增加可选 `options?: RequestOptions`（向后兼容）；
     `messageKeys` 补 §5.3 三条。
12. `web/src/api/self/types.ts`：加 `SelfOAuthLink`、`SelfOAuthLinks`、`BindSelfOAuthParams`、`SelfOAuthBindingResult`。
13. `web/src/api/self/index.ts`：
    ```ts
    /** GET /v1/self/oauth/links */
    export function listSelfOAuthLinks(options?: RequestOptions): Promise<SelfOAuthLinks>   // 用 get(..., { silent: true })
    /** POST /v1/self/oauth/bind */
    export function bindSelfOAuth(body: BindSelfOAuthParams): Promise<SelfOAuthBindingResult>
    /** DELETE /v1/self/oauth/links/{uid} */
    export function unbindSelfOAuth(uid: string, options?: RequestOptions): Promise<Record<string, never>>
    ```
14. `web/src/auth/oauth-redirect.ts`：`OAuthRedirect` 加 `intent?: "login" | "bind"`；
    `readOAuthRedirect` 归一化时补 `intent: parsed.intent === "bind" ? "bind" : "login"`；
    `storeOAuthRedirect` 原样保存。补一行注释说明「bind 现场只由个人资料页写入」。
15. `web/src/components/oauth-provider-mark.tsx`（新增）：导出 `oauthProviderIcon` 与 `OAuthProviderMark`
    （`<span className="oauth-brand grid size-9 place-items-center rounded-lg border border-border/60 bg-muted/40" data-provider=...>`
    + `<span className="oauth-provider-icon"><Icon className="size-5" /></span>`，未知来源回落 `KeyRound`）。
16. `web/src/components/oauth-provider-button.tsx`：改为从新文件 import `oauthProviderIcon`，删掉本文件的
    `providerIcons` 常量（纯搬运，行为不变）。
17. `web/src/index.css`：把 `.oauth-provider[data-provider="X"]` 与 `.dark .oauth-provider[data-provider="X"]`
    的选择器改成 `.oauth-brand[data-provider="X"]`（8 条），并给登录按钮加上 `oauth-brand` class（步骤 16 顺手改），
    让「品牌色」成为可复用 token；`.oauth-provider` 只保留描边/底色。
18. `web/src/pages/oauth-callback.tsx`：按 §3.2 增加 `intent === "bind"` 分支
    （bind → `bindSelfOAuth`；标题用 `oauthCallback.linkedTitle`；失败块出口改 `backToProfile` 链到 `/profile`）。
    登录分支原样保留。
19. `web/src/components/oauth-accounts-card.tsx`（新增）：实现 §2 组件树、§3 全部状态与流程；
    挂载时 `Promise.allSettled([listSelfOAuthLinks({silent:true}), listOAuth2Providers({redirectUri, state})])`。
20. `web/src/pages/profile/index.tsx`：按 §1.2 包一层 `grid max-w-2xl gap-6 pb-2`，挂 `<OAuthAccountsCard />`；
    读取 `oauth` query 做一次性 toast 并 `navigate("/profile", {replace:true})`。
21. 校验：`cd web && pnpm typecheck && pnpm lint && pnpm build`。
22. 浏览器回归（`web/AGENTS.md` 要求，需登录态本地会话）：
    - 1366×768：两张卡片都能滚到、OAuth 卡片底部「解绑」按钮可点；
    - 390×844：行内按钮换行、无横向滚动、文本 `break-all`；
    - 绑定：点绑定 → 跳第三方 → 回调 → 回 `/profile` 有 toast 且该行为「已绑定」，query 已清；
    - 解绑：确认框 Esc 关闭、焦点回归触发按钮、成功后按钮变「绑定」；
    - 键盘：Tab 顺序 卡片标题 → 各行按钮；`aria-busy` 只在实际请求期间为真；
    - 暗色：四种图标品牌色在 `.dark` 下可见（GITHUB 变浅）。
    - 若无法登录或没有浏览器，明确记录「浏览器验证未执行」，不要声称通过。

---

## 7. 与现有约定的冲突 / 风险点

**R1（阻断级）403 会整页跳转，内联「权限不足」态无法实现。**
`web/src/api/http.ts` 的 `request()` 对任何 403 执行 `window.location.assign("/forbidden")`（只放行 `/invitation/`）。
若新 RPC 返回 403，个人资料页会被整页踢走，S3 态永远不可达。二选一：
(a) 后端三个 self RPC 只要求登录态、业务拒绝用 400/409（推荐，本文档按此设计）；
(b) 同时在 `request()` 的例外列表加入 `/self/oauth/` 前缀，允许页面自己渲染 403。
**必须二选一，不能两不做**，否则 S3 只是摆设。

**R2（高）bind 复用公开的 `POST /v1/auth/oauth2/providers` 是权宜之计。**
它属于登录链路（`publicOperation`），返回的 authorizeUrl 内嵌一次性 state，语义上没有「绑定模式」。
当前可行，因为 intent 存在客户端；但若后端将来按登录策略裁剪该接口（例如禁用来宾注册期隐藏来源），
绑定入口会跟着消失。建议后续加 `GetSelfOAuthBindURL`，并把 intent 放进服务端签名的 state。

**R3（高）绝不能在 bind 路径复用 `OAuth2Login`。**
若第三方身份已绑定他人，`OAuthLogin` 会直接给那个人签发会话，在已登录页面里就是账号劫持 / 静默换号。
本次设计里 bind 只走带鉴权的 `BindSelfOAuth`，且 409 文案与 `error.oauthAccountBound` 区分开
（后者是「请返回登录」，用于登录页，不能复用）。

**R4（中）`OAuthRedirect` 结构变更影响登录链路。**
新增 `intent` 必须可选、读取时缺省 `"login"`，否则 `sessionStorage` 里的历史现场与登录页 `storeOAuthRedirect`
的调用形状会不匹配。回归点：登录页第三方登录全流程、`/auth/callback` 的补邮箱分支。

**R5（中）provider 展示名不一致。**
后端 `OAuth2Provider.name` 是固定中文（"GitHub"、"飞书"、"钉钉"），登录页直接展示它，英文界面会出现中文。
新卡片统一用 `t("enum." + provider)`，于是同一平台在登录页与资料页可能显示不同文案。修登录页属于另外的改动，
本文档只在 §4 标注；不要在两处各写一套映射。

**R6（中）未硬编码四家来源。**
`web/src/auth/oauth-redirect.ts` 的注释称「飞书后端支持但管理台不展示」，但后端 `oauthProviderCatalog` 已包含 FEISHU，
`login.tsx` 也是照后端返回渲染。若新卡片按固定四家渲染，会出现「登录页有飞书、资料页没有飞书」的裂缝；
若飞书已绑定而来源被下线，固定列表还会让这个绑定变成无法解绑的僵尸行。故行集合 = 已配置 ∪ 已绑定。

**R7（中）解绑的授权边界。**
`UnlinkOAuth` 必须按 `user_uid = self` 过滤；只按 `uid` 删除即构成 IDOR（`user_oauth.uid` 是全局唯一字符串）。
同时 `(provider, open_id)` 唯一索引 + 硬删除意味着「解绑后重新绑定」会新建一行（新 uid、新 created_at），
不做软删就没有审计痕迹——如果合规需要留痕，要在后端改软删，前端展示逻辑不用变。

**R8（中）「最后一个登录方式」规则需要后端拍板。**
前端只能呈现 409 文案；判定必须在后端，且要覆盖历史占位邮箱（`@users.alioth.local`，见 `biz/auth/oauth.go`）
这类「有邮箱但收不到信」的账号，否则会出现「解绑后永远登不进来」。

**R9（低）`ConfirmAction` 缺 pending 文案位。**
它只有 `pending: boolean`（禁用按钮）。S9 需要「解绑中…」。建议给它加**可选** `pendingLabel`，
默认回落到 `confirmLabel`——纯增量、不影响现有调用方；若不想动共享组件，只能在解绑时把 confirmLabel 换成
「解绑中…」，语义上会让确认按钮文案抖动，不推荐。

**R10（低）404 会被 http.ts 自动 toast。**
`send()` 对非 silent 的 404 会弹 `error.notFound`（"没有找到请求的资源"）。解绑的 404 是「已经解绑」的正常分支，
必须 `{ silent: true }` 调 `unbindSelfOAuth`，否则会同时出现错误 toast 和「解绑成功」。这就是步骤 11 要改 `del` 的原因。

**R11（低）后端未上线时的降级。**
list 的 404 必须被当作「功能未部署」→ 卡片整体不渲染（U1），否则上线顺序颠倒会在资料页常驻一个「加载失败」块。
这要求 `listSelfOAuthLinks` 固定 `silent: true`。

**R12（低）返回时的 query 清理。**
`?oauth=linked&provider=` 必须 `navigate(..., { replace: true })` 清掉，否则刷新/回退会重复弹 toast；
`rememberLastPath` 只记 `pathname`，不会污染着陆页记录。

**R13（低）布局规则约束。**
不许在这张卡片里引入第二层 `overflow-y-auto`（页面已有一个滚动所有者），不许把按钮改成 `h-9/h-10`
规避 36px 图标对齐，行内 `min-w-0`/`break-all` 必须写全，否则窄屏 390px 会出现横向溢出（`web/AGENTS.md`）。

---

## 8. 实现状态（已落地）

后端与前端均按 §5.2 / §6 实现，工作区当前 diff 即交付内容。

**后端**（`api/alioth/v1/identity.proto` 及生成物、`bo`、`repository/oauth.go`、`data/oauth_repo.go`、
`biz/auth/oauth_self.go`、`service/self.go`、`biz/auth/auth.go`）

- 三个 RPC 与 §5.2 的路径/字段/校验一致；`auth.go` 额外把三个 operation 放进 `namespaceOptional`，
  否则未选租户时会被 `Authenticate` 以 400 `X-Namespace is required` 拒绝，与「只要求登录态」冲突。
- `BindSelfOAuth` 不调用 `MoveOAuth`，他人已绑 → 409；`UnbindSelfOAuth` 删除条件同时带 `uid` 与 `user_uid`，
  且删除前判定「最后一个登录方式」（占位邮箱不算可用），拒绝时不删除。
- `bo.OAuthLink.BoundAt` 用 `time.Time`（仓库既有约定），RFC3339 在 service 层格式化。

**前端**（i18n、`api/http.ts`、`api/self/*`、`auth/oauth-redirect.ts`、`confirm-action.tsx`、
`oauth-provider-button.tsx`、`index.css`、`oauth-callback.tsx`、`profile/index.tsx`、新增卡片与图标块）

- 与 §1–§4 一致；错误出口：bind 409 → `?oauth=conflict`，其它失败留在回调页并给「返回个人资料」。
- **偏差：行集合按「每一条绑定一行」渲染，而不是「每个来源一行」。**
  后端允许同一来源绑多个第三方身份（`user_oauth` 只对 `(provider, open_id)` 唯一），
  按来源聚合会让第二条绑定从界面消失、`summary` 出现 `2 / 1`。已绑行用绑定 `uid` 做 key，
  「已配置但未绑定」的来源才补一条未绑定行。
- **偏差：`messageKeys` 里 `"oauth account is already bound"` 保持映射 `error.oauthAccountBound`。**
  `localizeMessage` 不做插值，映射到带 `{{provider}}` 的 `profile.oauth.alreadyBound` 会在通用错误位
  渲染出字面占位符；绑定冲突由资料页/回调页按 status 用 `t()` 自行插值。

**已知限制与未执行项**

- 浏览器回归未执行（无已登录本地会话）：1366×768 / 390×844 滚动与换行、绑定整页跳转、解绑确认框
  Esc/焦点、键盘 Tab 顺序、暗色品牌色。
- `./internal/data` 的 Postgres 集成测试因缺 `ALIOTH_TEST_DSN` 全部 SKIP，未对真实数据库验证。
- `make lint` 在两个**改动前就存在**的问题上失败，与本次无关：`pkg/esagg/esagg.go` 未 gofmt、
  `web/src/hooks/use-query-store.ts` 与导出的 hook 不同名；本次改动文件 gofmt / eslint / buf lint 均通过。
