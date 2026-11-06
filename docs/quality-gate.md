# 交付质量门禁

这份文档是本仓库的**放行标准**，不是建议。它记录了一轮迭代里真实发生的质量事故、根因，
以及从那之后每个改动必须满足的检查项。新增需求同样适用。

## 1. 事故回顾（M33 官网与值班这一轮）

| # | 缺陷 | 根因 | 本可在哪一步拦住 |
| --- | --- | --- | --- |
| 1 | 后台「首页内容」一直加载失败 | 前端接口路径漏了 `/system`，打到了公开接口 | 真实后端的一次页面打开 |
| 2 | 表单被裁切、滚不动 | 用了 `flex-1` 补高度；真因是 `Card` 自带 `overflow-hidden`，作为 flex 子项最小高度为 0 | 浏览器里滚到底 |
| 3 | 版权重复 `© 2026 © 2026` | 把完整版权行当成名字拼进模板 | 看一眼页脚 |
| 4 | 平台级新接口 `400 X-Namespace is required`（两次） | 新接口漏加 `namespaceOptional` | 一次带令牌的 curl |
| 5 | 页面保存永远 409 | 乐观锁令牌截到秒，库里是微秒，永不相等 | 一次 GET→PUT 往返 |
| 6 | `my-shifts` 404 | 静态路径被 `{uid}` 路由抢先匹配 | 一次 curl |
| 7 | 预览页跳过链接常显 | 用位移隐藏，嵌套位置变化后落回视口 | 一次截图 |
| 8 | 点左侧导航后标题消失 | `scrollIntoView` 连外层固定视口一起滚 | 一次点击 |
| 9 | 侧栏两个菜单同时高亮 | 前缀匹配 | 一次点击 |
| 10 | 保存冲突锁死编辑器 | 初版只给「重新加载」，用户改动只能丢 | 设计评审 |
| 11 | 预览只是一段文本 | 实现取舍没和需求对齐 | 设计评审 |
| 12 | Markdown 不支持代码高亮 | 以「不加依赖」为由降级 | 需求对齐 |

## 2. 根因

1. **用构建通过代替用户路径走通**：`typecheck/lint/build` 全绿不等于页面能用，
   上面的 1、4、6、9、10 都是一次真实点击就能暴露的。
2. **把 mock 当浏览器验收**：拦截 `/v1/*` 返回假数据，路径、权限、序列化全没验。
3. **分层堆叠 + 并行协作缺少集成审查**：多层分支叠加，前一层的缺陷一路带下去；
   多人改同一棵树，共享文件（`i18n`、`App.tsx`、页面 `index.tsx`）靠口头约定。
4. **按症状修而不先定根因**：`flex-1`、`end`、`translate` 都是表层修复，第二轮才找到真因。
5. **没有自动化护栏**：没有端到端冒烟、没有 i18n 一致性检查、没有 UI 约定检查、
   没有视觉回归，所以每加一个页面就重复一遍同类错误。

## 3. 每个改动都必须满足的 DoD

### 3.1 验证矩阵

**分级验证**：先按改动范围跑最小集，交付前再跑一次全量。全仓重跑不是"更稳"，是把时间花在没改的代码上。

| 改动位置 | 最小验证（改完就跑） |
| --- | --- |
| 仅 `docs/`、`*.md` | 无；改完核对相对链接可达 |
| `web/src/**` | `npm --prefix web run typecheck`、`npx eslint src`（或 `npm --prefix web run check`） |
| `web/src/styles/theme/**`、`index.css` | 追加 `npm --prefix web run tokens:check` |
| `web/e2e/**`、`web/scripts/**` | `npx eslint e2e scripts` + 跑受影响的 spec |
| `app/alioth/internal/{data,biz,service,access}/**` | `go test -C app/alioth ./internal/<层>/...` |
| `app/alioth/api/**` | `buf lint` + `go vet -C app/alioth ./...` |
| `app/<其它服务>/**` | `go vet -C app/<服务> ./...` + `go test -C app/<服务> ./...` |
| `pkg/**`、`api/**` | 对应模块 `go vet` + `go test` |
| 生成代码（ent / wire / conf / buf） | 重新生成后 `git diff --stat` 复核，再跑受影响包测试 |
| 交付前（一次） | `make vet && make lint && make test`，`npm --prefix web run check && npm --prefix web run build` |

- 同一命令在一次任务里只跑一次；只有你又改了它覆盖的范围，才需要重跑。
- 新接口：带真实令牌的 curl 往返一次；平台级接口要验证「不带 `X-Namespace` 也能调」。
- 乐观锁/状态机/权限分支：补单测，覆盖成功与冲突两条路径。
- 新路由：验证静态路径不会被同级 `{id}` 路由抢先匹配。


### 3.2 前端 UI

- 只使用共享原语（`ui/textarea`、`ui/select`、`ui/dialog`、`StatusBadge`、`FormField`…），
  不允许在页面里写原生表单控件。
- 只使用语义 token；不允许 hex/rgb/调色板类（主题资产文件除外，且需注明）。
- 字段错误就地显示并带 `aria-invalid`；表单错误在操作条附近；不允许只有一条兜底红字。
- 四态齐全：加载、空、失败可重试、无权限。
- 交互边界：未保存改动有保护、切换上下文（语言/标签/页面）前确认、冲突有可恢复路径。
- 可访问性：键盘可达、焦点环可见、状态不只靠颜色、一个 h1。

### 3.3 浏览器验收（真实后端，禁止 mock 业务接口）

- 桌面 1440×900 与窄屏 390×844 各一遍；浅色与深色各一遍。
- 走完主路径 + 空态 + 错误态 + 权限态；截图或录屏保存到 `tmp/<任务>/`。
- 登录态用 JWT 注入（不要走验证码）；只允许 mock **错误注入**，不允许 mock 正常响应。
- 没有浏览器可用时，交付说明必须写「浏览器未验证」，不得写「通过」。

## 4. 工具与命令

```sh
npm --prefix web run check:ui     # 原生控件 / 字面色检查
npm --prefix web run check:i18n   # 中英 key 一致性
node web/scripts/smoke-site.mjs   # 真实后端冒烟（需 dev server 与 JWT）
```

`check:ui` 与 `check:i18n` 是必须在合并前跑的红线；`smoke-site.mjs` 覆盖官网首页、
文档页、后台站点设置、预览页与页面保存的乐观锁往返。

### 4.1 Playwright 端到端

```sh
# 需要先起好：日志服务（默认 :8100）与 web dev server（默认 :8080）
E2E_WEB_URL=http://127.0.0.1:8080 \
E2E_API_URL=http://127.0.0.1:8100 \
E2E_JWT_SECRET=... E2E_USER_UID=... E2E_NAMESPACE=... \
npm --prefix web run e2e
```

- 用本机已安装的 Chrome（`channel: "chrome"`），不下载额外浏览器。
- 登录态用 JWT 注入 `localStorage`，不走验证码；夹具里的默认值只对应本机开发环境。
- 覆盖：公开配置、平台级接口免租户头、乐观锁往返与过期 409、首页、文档页 Markdown
  代码高亮与右侧目录、深色/窄屏无横向滚动、预览页、设置页快捷导航与首页内容预览、
  切换语言确认、编辑器内联预览、保存冲突覆盖、标签与滚动（`tabs-scroll.spec.ts`）。
- 标签/滚动这类逐页检查用环境变量指路由，不再现场手写脚本：
  `E2E_SCROLL_ROUTES=/alert-rule npx --prefix web playwright test tabs-scroll`。
  缺场景就补进 `web/e2e/`，下一次任务直接复用。
- 截图落在 `tmp/acceptance/`，作为浏览器证据随交付一起给出。
- 用例里创建的数据都在 `finally` 里删除，跑完不污染开发库。

## 5. 流程约束

1. 一次只开一个分支；合并前必须过第 3 节 DoD。
2. 共享文件（`web/src/i18n/index.ts`、`web/src/App.tsx`、页面 `index.tsx`）由一个人改，
   并行任务只允许改不相交目录。
3. 每个 UI 批次交付前，由 UI 设计师与 UI 视觉验收两个角色各过一遍（设计一致性 + 放行判定），
   任一不通过不算完成。
4. 交付说明只写事实：改了什么、证据在哪、什么没验证。禁止「应该没问题」。
