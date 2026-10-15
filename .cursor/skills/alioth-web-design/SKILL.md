---
name: alioth-web-design
description: >-
  Applies Alioth admin UI layout, spacing, and shadcn dashboard patterns. Use
  when editing web/ pages, the shell, login, dialogs, tables, spacing, or
  visual style, or when the user mentions 样式, 间距, 设计, UI, or dashboard.
---

# Alioth 前端设计

改 `web/` 界面前先读本技能，再读 [revisions.md](revisions.md)。修订记录里的条目优先于下文。组件安装仍遵循 [alioth-web](../alioth-web/SKILL.md) 和 shadcn/ui 技能。

视觉体系只用仓库已有的一套：`radix-nova`、`neutral`、Geist、lucide，token 在 [web/src/index.css](../../../web/src/index.css)。不引入 `antd`，不另写配色、圆角或字号。

## 壳

管理页用 shadcn `Sidebar` + `SidebarInset`，不要再用裸 `aside` 加顶栏。尚未安装时用 shadcn 技能安装 `sidebar`，不要手写一套侧栏。

- 侧栏可折叠。品牌在顶，导航在中。当前项用 sidebar accent，不用自行拼 `bg-muted`。
- 菜单仍按当前用户权限生成，文案走 `t()`。系统管理放在可折叠子菜单里。个人资料只从顶栏进入。
- 顶栏放 Namespace 切换，以及语言、主题、个人的图标按钮。多项用下拉菜单。主按钮留在页头，不塞进顶栏。
- 顶栏：`flex h-14 items-center gap-2 border-b px-4`。
- 主内容：`flex flex-1 flex-col gap-6 p-4 md:p-6`。内容不贴边。

## 页

每一页先是页头，再是内容。

- 页头：`flex items-start justify-between gap-4`。左侧标题 `text-2xl font-semibold tracking-tight`，说明 `text-sm text-muted-foreground`。右侧是该页主操作。
- 列表页顺序：页头、筛选条、表格卡片、分页。筛选条件放进 URL。
- 表格放在 `Card` 里，表头与行用 shadcn `Table`。行 key 用 `uid`。
- 新建和编辑用 `Dialog`。字段少的详情用右侧 `Sheet`。需要独立地址或分成多块的详情用当前页的子路由。主从页用左侧列表加右侧明细，间距 `gap-6`。
- 每个输入框和选择框都有 placeholder。标签说不清约束或后果时，在 label 旁放 `CircleHelp` tooltip。
- 一行操作超过两项时只留「详情」和「更多」。更多是下拉菜单。删除、封禁、停用、移交负责人先确认，确认按钮用 destructive。
- 区块之间 `gap-6`。同一行按钮 `gap-2`。

## 间距

用这些类，不按页面临时加减像素。

| 位置 | 类 |
| --- | --- |
| 页边距 | `p-4 md:p-6` |
| 区块 | `gap-6` |
| 卡片内边距 | 使用 `Card` 默认内边距 |
| 表单项 | `grid gap-2` |
| 对话框内容 | `grid gap-4` |
| 按钮组 | `flex gap-2` |

控件高度、圆角、边框用组件默认值。

## 状态

功能要闭环，不能只画成功路径。

- 请求中显示 loading，卸载后不再 setState。失败保留空状态，并给出可重试的失败提示。
- 无数据用空状态，说明下一步，而不是空白表格。
- 无权限不渲染写操作。服务端拒绝不能靠隐藏按钮代替，但界面上不要露出无效按钮。
- 关闭对话框时清空草稿。提交中禁用确认按钮。
- 文案、空状态、表头、按钮都走 `t()`。

## 认证页

登录和注册在视口中居中：外层 `flex min-h-svh items-center justify-center bg-muted/40 p-4`，卡片 `w-full max-w-sm`。主提交按钮拉满卡片宽度。错误信息放在表单下方，使用 `text-sm text-destructive`。输入框同样要有 placeholder；密码、验证码这类约束用 tooltip 说明。

## 改样式之后

用户指出具体样式问题，并且改法已经确认时：

1. 先把可复用的结论写入 [revisions.md](revisions.md)：日期、适用页面或全局、必须遵守的布局或类。
2. 只影响当次页面、不能推广的改动不写入。
3. 再改代码。下次做界面时以修订记录为准。
