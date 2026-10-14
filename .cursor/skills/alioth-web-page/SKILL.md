---
name: alioth-web-page
description: >-
  Builds an Alioth feature page with shadcn/ui and the web API layer. Use when
  the user asks to complete a page, module, or feature (完成页面、完成模块、完成功能)
  in web/.
---

# 功能页

先读 [alioth-web](../alioth-web/SKILL.md)，再读 [shadcn-ui/SKILL.md](file:///Users/hubiao/.agents/skills/shadcn-ui/SKILL.md)。UI 只用 shadcn/ui 和 Tailwind。禁止从 `antd` 或 `@ant-design/icons` 引入组件。

## 请求

页面和组件只引用 `web/src/api` 已导出的函数和类型。开发页面时不修改 `web/src/api`。

接口与后端不一致时，先按 [alioth-update-api](../alioth-update-api/SKILL.md) 同步，再改页面。不用 `any`、类型断言或页面内自定义类型绕过 API 类型。

## 结构

- 页面放在 `web/src/pages/<领域>/`。跨页组件放在 `web/src/components/`。
- 单页过长时，把只供本页使用的子组件放到该页目录下的 `components/`。
- Props 用 `interface`。类型用 `import type`。
- 列表、筛选和分页的字段与 API 类型一致，并把查询条件放进 URL。
- 新建和编辑用对话框。主从页面用左侧列表加右侧明细。
- 文案使用 `t()`。按钮、表头、空状态和提示都不写死中文或英文。
- 请求处理 loading。卸载后不再 setState。失败时保留可用的空状态。

## 组件

需要表格、表单、对话框或下拉菜单时，用 shadcn/ui 技能里的组件名安装后再使用。表格行使用稳定的 `uid` 作为 key。关闭对话框时清空本地编辑状态。
