# 分支与提交规范

强制约束，不是建议。每个会改动仓库内容的任务都按本文执行。
完整细则（同步的四种情况、分支命名正反例、commit 正文要求、rebase 冲突处理）见
[git-workflow-detail.md](git-workflow-detail.md)。

只读任务不受约束。用户明确要求"直接在 main 上改"时按用户指示执行，并在交付说明里写明走了例外。

## 开工（一条命令 + 一次切分支）

```sh
# 1. 基线：记进交付说明
git branch --show-current; git rev-parse --short HEAD; git status --porcelain

# 2. 同步
git fetch origin; git rev-parse --short main origin/main
```

- 两个哈希一致 → 继续。
- `main` 落后 → 快进到 `origin/main`（`git fetch origin main:main`，或在干净 checkout 里 `git merge --ff-only origin/main`）。
- `main` 领先或分叉 → **停下问用户**，不擅自 push / merge / rebase。

```sh
# 3. 切分支：基线是 main 本身
git switch -c <类型>/<kebab-case-描述> main
git merge-base --is-ancestor main HEAD && echo "起点是 main"
```

分支名正则：`^(feat|fix|hotfix|refactor|docs|test|chore|perf|build|ci|revert)/[a-z0-9]+(-[a-z0-9]+)*$`，
总长 ≤50，小写 ASCII，2–5 个英文词，不用 `wip/tmp/test/updates/changes/new/fix` 这类空词，不加个人前缀。
新分支不设 upstream。

checkout 被别的会话占用、或带着他人未提交改动时，不要在里面 `git switch`，改用
`git worktree add .agents/worktrees/<名> -b <分支> main`，做完 `git worktree remove`。

## 开发

- 一个任务一个分支，中途不为别的事切分支。
- 生成代码（`make conf`、ent、Wire、`buf generate`）在分支上重跑，生成物与源文件同一次提交。
- 不顺手格式化、不重排、不做"顺便优化"。
- 不 `stash`、不 `checkout --`、不 `restore`、不 `clean` 别人的未提交改动；不动不属于本任务的条目。

## 提交

- 跑本任务相关验证，分级见 [quality-gate.md §3.1](quality-gate.md#31-验证矩阵)；
  没跑就写"未验证"，不写"应该没问题"。
- 只 `git add <path>` 逐路径暂存。**禁止 `git add -A`、`git add .`、`git commit -a`**。
- 提交后 `git status --porcelain` 复核：不属于本任务的条目必须仍未提交。
- 信息用 Conventional Commits：`<type>(<scope>): <英文祈使句小写开头不加句号>`，正文说"为什么"和原失败现象。

## 完成

默认停在这里：**不推送、不合并、不 rebase 任务分支、不打 tag、不建 PR**。

交付说明必须给：分支名与起点、提交短哈希与标题、变更文件清单、验证结论（跑了什么、结果、什么没跑）、
工作区里本任务之外的未提交改动是否仍原样保留（没有写"无"）。

后续动作等用户明确指示：

| 用户选择 | 执行 |
| --- | --- |
| 推送分支自开 PR | `git push -u origin <分支>`，不加 `--force` |
| 本地合并进 main | 先 `git rebase main`，再 `git switch main && git merge --ff-only <分支>` |
| 保留分支 | 不执行任何命令 |
| 删除分支 | `git branch -d <分支>`，用户明确要求时才删，不用 `-D` |

合并只走 rebase + `--ff-only`，`main` 保持单线；冲突不自动解，停下报给用户。

## 禁止

- 在 `main` 上提交本任务改动（用户明确指示除外）；在 `main` 与 `origin/main` 不一致时切分支。
- 从非 `main` 的分支或提交切本任务分支（用户明确要求接着某个分支做时除外）。
- `git push`、`git push --force`、`git commit --amend` 已推送提交、`git reset --hard`。
- 对他人的分支 rebase；对本任务分支 rebase 后强推（除非用户明确同意）。
- 用 `git merge`（含 `--no-ff`）把任务分支合进 `main`。
- 自动合并、自动解冲突、自动打 tag、自动建 PR、自动删分支。
