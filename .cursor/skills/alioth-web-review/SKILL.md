---
name: alioth-web-review
description: >-
  Reviews uncommitted Alioth web changes for API usage, types, and accidental
  Ant Design imports. Use when the user says 帮我审核当前变更, 帮我审核代码, or
  asks to review the web UI.
---

# 审核前端变更

只审核 `web/` 下尚未提交的 diff。先看 `git diff` 和已暂存的文件，再下结论。

## 检查

- 页面是否直接调用 axios，或在页面里新增请求函数。请求必须来自 `web/src/api`。
- 状态和提交参数是否使用 API 模块的类型与枚举。枚举字段不要再 `String()` 或 `Number()` 转一遍。
- 是否出现 `antd`、`@ant-design/icons` 或 ahooks。发现则改为 shadcn/ui。
- 异步请求是否有 loading，失败是否有空状态，卸载后是否还会 setState。
- 业务文案是否绕过 i18n。
- 请求是否带 Bearer 和 `X-Namespace`，是否指向 8001、8002、8022 或 9022。

发现问题就改。界面结构、键盘操作和可访问性再按 [web-design-guidelines/SKILL.md](file:///Users/hubiao/.agents/skills/web-design-guidelines/SKILL.md) 复核改过的文件。

审核结果按文件列出：必须修改的问题，以及已经改掉的项。
