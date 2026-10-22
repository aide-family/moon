---
name: alioth-web
description: >-
  Builds the Alioth admin UI as one Vite and shadcn/ui app. Use when editing
  web/, adding pages, menus, or API clients, or when the user mentions shadcn,
  X-Namespace, or the management console.
---

# Alioth Web

先读本技能，再读 [alioth-web-design](../alioth-web-design/SKILL.md) 和本机的 shadcn/ui 技能：[shadcn-ui/SKILL.md](file:///Users/hubiao/.agents/skills/shadcn-ui/SKILL.md)。布局和间距以设计技能为准，组件安装以 shadcn/ui 技能为准。不要引入 `antd` 或 ahooks。

验收页面时再读：

- [web-design-guidelines/SKILL.md](file:///Users/hubiao/.agents/skills/web-design-guidelines/SKILL.md)
- [playwright-cli/SKILL.md](file:///Users/hubiao/.agents/skills/playwright-cli/SKILL.md)

浏览器只打开 `http://<web>:8080`。确认登录、提交和跳转，不要只看静态渲染。

## 应用形态

一个 Vite + React + TypeScript + Tailwind 应用。不拆微前端。页面在 `web/src/pages`，请求函数在 `web/src/api`。组件不直接使用 axios。

每个请求带 `Authorization: Bearer <token>` 和 `X-Namespace`。`/v1` 只代理到 `alioth:8000`，不请求 8001、8002、8022、9022。

接口错误按状态处理，不要只吞掉：

- 400：参数错误，显示到表单字段；对不上字段时显示在表单错误区。
- 401：清除令牌并回到登录页。登录、注册、验证码接口的 401 留在当前表单并显示消息。
- 403：进入无权限页。
- 404：全局 message。
- 5xx：全局 notify。
- 409：写操作留在表单错误区；列表加载显示服务端消息。

菜单按当前用户在当前 Namespace 的权限生成。文案使用 i18n，不在页面里写死业务文案。

## 命名

完整规则见 [conventions Web/命名](../../../docs/iterations/conventions.md)。要点：

- 文件 kebab-case，一个文件一个主导出；跨页面组件文件名等于导出名的 kebab（`app-shell.tsx` → `AppShell`）。
- 页面目录用路由名，入口 `index.tsx`，导出 `<业务实体>Page`（`workers/index.tsx` → `WorkerStatusPage`）；领域名以 Page 结尾时写 `<领域>ManagePage`。
- 页面私有组件放该页面目录，文件用 `<部件>.tsx`，导出加领域前缀（`column-dialog.tsx` → `AlertColumnDialog`）。
- 壳用 `<用途>Shell`（`AppShell`、`AuthShell`）；专有名词按官方写法（`GitHub`、`PromQL`、`LogQL`、`LogsQL`、`SSH`）。
- hook 以 `use` 开头，文件名 `use-<hook 名>.ts`；跨页面 hook 放 `src/hooks`，页面私有 hook 留页面目录。
- `src/lib` 只放纯函数（`runAction`、`readFormError`、`formErrorMessage`、`failureMessage`），不用 `utils`、`helpers`、`common`、`misc`、`shared` 文件名。
- API 函数名等于 RPC 名 lowerCamel；类型 `XxxItem`、`ListXxxParams`、`ListXxxResult`、`Create/UpdateXxxParams`，单对象响应用业务对象名，共用写载荷 `Save<资源>Params`，嵌套写载荷 `<资源><子对象>Input`。
- shadcn 生成目录 `src/components/ui` 不参与重命名。

做具体页面时遵循 [alioth-web-page](../alioth-web-page/SKILL.md)。后端 OpenAPI 变更时遵循 [alioth-update-api](../alioth-update-api/SKILL.md)。
