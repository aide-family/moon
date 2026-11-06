# Alioth 工作约束

## 分支与提交

改仓库内容前先读 [docs/git-workflow.md](docs/git-workflow.md)（强制）。要点：

1. 先 `git fetch` 把本地 `main` 同步到 `origin/main`（不一致按该文的表处理，不擅自 push），再从 `main` 切 `<类型>/<kebab-case-描述>` 分支开发。
2. 完成后只提交到该分支：不推送、不合并、不打 tag、不建 PR。
3. 推送还是本地合并进 `main` 由用户决定；合并只用 rebase 后 `--ff-only`，不留合并提交。
4. 未得到用户明确指示前，推送、合并、rebase、打 tag、建 PR 一条都不执行。
5. 不动别人的未提交改动；不用 `git add -A`、`git add .`、`git commit -a`。

## 工具用法（影响耗时和准确率，别当风格建议）

- 读文件用 `read`/`ops_read_files`，找文件用 `glob`，搜内容用 `grep`。**不要用 bash 的 `cat/sed/head/tail/ls/grep -r`**。
- 同一路径一次会话只读一次；大文件用 `offset`/`limit` 续读。
- 互不依赖的调用放在同一轮并发发出；长命令丢后台，用 `job_output` 取结果，禁止 `sleep` 轮询。
- 单任务预期 ≤80 step；到 120 step 还没收尾，先停下来汇报剩余工作，不要继续硬推。
- 多智能体只在有 ≥2 个互不依赖的写入范围时使用；一个 agent 能做完就直接做。
- 验证分级执行，别全仓重跑：见 [docs/quality-gate.md §3.1](docs/quality-gate.md#31-验证矩阵)，
  完整工作方式见 [docs/agent-workflow.md](docs/agent-workflow.md)。

## 文档地图（按需读，不要整篇加载）

| 需要什么 | 读哪份 |
| --- | --- |
| 分支、提交、交付 | [docs/git-workflow.md](docs/git-workflow.md)（细则在 detail） |
| 改完怎么验收、DoD | [docs/quality-gate.md](docs/quality-gate.md) |
| 后端分层、命名、端口、接口 | [docs/iterations/conventions.md](docs/iterations/conventions.md) |
| 前端界面约束 | [web/AGENTS.md](web/AGENTS.md) |
| 当前里程碑与进度 | [docs/roadmap.md](docs/roadmap.md)（只读当前里程碑一节） |
| 系统设计背景 | [docs/design.md](docs/design.md)（按章节 `grep` 定位再读） |
