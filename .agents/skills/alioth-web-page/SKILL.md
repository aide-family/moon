---
name: alioth-web-page
description: >-
  Builds an Alioth feature page with shadcn/ui and the web API layer. Use when
  the user asks to complete a page, module, or feature (完成页面、完成模块、完成功能)
  in web/.
---

# 功能页

先读 [alioth-web](../alioth-web/SKILL.md)、[alioth-web-design](../alioth-web-design/SKILL.md)，再读 [shadcn-ui/SKILL.md](file:///Users/hubiao/.agents/skills/shadcn-ui/SKILL.md)。布局和间距以设计技能为准。UI 只用 shadcn/ui 和 Tailwind。禁止从 `antd` 或 `@ant-design/icons` 引入组件。

## 请求

页面和组件只引用 `web/src/api` 已导出的函数和类型。开发页面时不修改 `web/src/api`。

接口与后端不一致时，先按 [alioth-update-api](../alioth-update-api/SKILL.md) 同步，再改页面。不用 `any`、类型断言或页面内自定义类型绕过 API 类型。

## 结构

- 页面放在 `web/src/pages/<领域>/`。跨页组件放在 `web/src/components/`。
- 页面组件写 `<业务实体>Page`（`MemberPage`、`WorkerStatusPage`），领域名本身以 Page 结尾时写 `<领域>ManagePage`；页面私有组件放该页目录，文件用 `<部件>.tsx`，导出加领域前缀（`AlertColumnDialog`）。
- 单页过长时，把只供本页使用的子组件放到该页目录下的 `components/`。
- Props 用 `interface`。类型用 `import type`。
- 列表、筛选和分页的字段与 API 类型一致，并把查询条件放进 URL。
- 新建和编辑用对话框，高度不超过 `80vh`。字段多就分步。字段少的详情用右侧抽屉。需要独立地址或分成多块的详情用当前页的子路由。主从页面用左侧列表加右侧明细。
- 列表总数在左上，翻页在右上。没有数据不显示翻页。表头固定，表体在剩余高度里滚动。
- 一行操作超过两项时只留「详情」和「更多」。删除、封禁、停用、移交负责人先确认。
- 文案使用 `t()`。按钮、表头、空状态和提示都不写死中文或英文。
- 请求处理 loading。卸载后不再 setState。失败时保留可用的空状态。
- **表单（必须）**：每个 `Input` / `Textarea` / `Select` 都有 i18n placeholder。提交前按后端 proto/biz 必填做前端校验（`required` + trim；必填 Select 禁止空/`NONE`；`StepForm` 下一步先校验当前步），空值不发无效请求。可用 `web/src/lib/form.ts` 的 `requireTrimmed` / `requireSelected`。不引入 zod/RHF。
- 登录与注册保持分路由，底部保留互跳链接。

## 组件

需要表格、表单、对话框或下拉菜单时，用 shadcn/ui 技能里的组件名安装后再使用。表格行使用稳定的 `uid` 作为 key。关闭对话框时清空本地编辑状态。
