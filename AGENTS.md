# Alioth 工作约束

改仓库内容前先读 [docs/git-workflow.md](docs/git-workflow.md)，它是强制约束。要点：

1. 任何会改动仓库内容的新任务，先 `git fetch` 并把本地 `main` 同步到 `origin/main`（两边不一致就按 [docs/git-workflow.md](docs/git-workflow.md) 的表处理，不擅自 push），再从 `main` 切出 `<类型>/<kebab-case-描述>` 分支，在分支上完成开发。
2. 完成后只提交到该分支：不推送、不合并、不打 tag、不建 PR。
3. 推送分支还是本地合并进 `main`，由用户决定；合并只用 rebase 后 `--ff-only` 快进，不留合并提交。
4. 未得到用户明确指示前，推送、合并、rebase、打 tag、建 PR 一条都不执行。
5. 不动工作区里不属于本任务的未提交改动，不用 `git add -A`、`git add .`、`git commit -a`。

后端分层、命名、端口等通用约定见 [docs/iterations/conventions.md](docs/iterations/conventions.md)；`web/` 的界面约束见 [web/AGENTS.md](web/AGENTS.md)；当前里程碑见 [docs/roadmap.md](docs/roadmap.md)。
