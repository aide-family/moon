# 分支与提交规范（细则）

本文是 [git-workflow.md](git-workflow.md) 的完整版，收录表格、示例与理由。
**强制规则以 git-workflow.md 为准**；两份冲突时以速查版为准并修本文。

## 适用范围

- 会改动仓库内容的任务：新增、修改、删除文件，重跑生成代码，改文档，改配置。
- 判断方法：任务开始时记录 `git status --porcelain`，任务结束时非空即视为发生了变更，必须走分支流程。
- 只读任务（只读文件、只跑查询命令、不改工作区）不受约束。
- 用户明确要求"直接在 main 上改"时按用户指示执行，并在交付说明里写明这次走了例外。

## 同步 main 的四种情况

`git fetch` 只更新远端引用，不动工作区；两个哈希必须一致才能继续：

```sh
git fetch origin
git rev-parse --short main origin/main
```

| 情况 | 处理 |
| --- | --- |
| 一致 | 继续切分支 |
| `main` 落后 | 快进到 `origin/main`：`main` 没被检出时用 `git fetch origin main:main`（只接受快进，非快进会直接拒绝）；正被检出且工作区干净时，在那份 checkout 里 `git merge --ff-only origin/main` |
| `main` 领先 | 停下问用户：推送这些本地提交，还是先处理掉。不擅自 `push` |
| 两边分叉 | 停下问用户，不自动 `merge`、不自动 `rebase` |

- 本地合并进 `main` 走 rebase 快进，合并完 `main` 会领先远端；下次开工前必须先按上表处理，这条同步检查就是为此存在。
- 新分支不设 upstream，等用户决定推送时再加 `-u`。
- checkout 正被另一个会话使用、或带着他人未提交改动时，不要在里面 `git switch`：在 `.agents/worktrees/`（已被 git 忽略）里用 `git worktree add <路径> -b <分支> main` 开一份工作区，做完 `git worktree remove <路径>`。

## 分支命名

格式固定为 `<类型>/<kebab-case-描述>`。整名小写 ASCII，只含字母、数字、连字符和中间那一个 `/`：

```
^(feat|fix|hotfix|refactor|docs|test|chore|perf|build|ci|revert)/[a-z0-9]+(-[a-z0-9]+)*$
```

| 类型 | 用于 |
| --- | --- |
| `feat` | 新功能，或给现有能力加新行为 |
| `fix` | 修缺陷 |
| `hotfix` | 线上故障的紧急修复，用户明确要求时 |
| `refactor` | 不改行为的重构 |
| `docs` | 文档、注释、设计说明 |
| `test` | 测试、夹具、测试数据 |
| `chore` | 依赖、脚本、配置、生成产物 |
| `perf` | 性能优化 |
| `build` | 构建、打包、镜像 |
| `ci` | CI/CD 工作流 |
| `revert` | 回滚既有提交 |

类型名与提交信息里的 `type` 保持一致，写 `feat` 不写 `feature`。

描述规则：

- 2–5 个英文单词，说明业务对象和动作，kebab-case；分支名总长不超过 50 字符。
- 不用中文、大写、下划线、点、空格、冒号和括号。
- 不用 `wip`、`tmp`、`test`、`updates`、`changes`、`new`、`fix` 这类不说明内容的词。
- 不加人名、工号或 `hubiao/` 这类个人前缀。
- 有关联 issue 时写成 `<类型>/<描述>-<编号>`，例如 `fix/web-dialog-scroll-141`；仓库目前没有 issue 编号体系，默认不加。

| 分支名 | 判断 |
| --- | --- |
| `feat/datasource-availability` | 符合：类型 + 业务描述 |
| `fix/web-dialog-scroll` | 符合：对象 + 动作 |
| `docs/git-workflow` | 符合 |
| `chore/deploy-pin-images` | 符合 |
| `feature/new` | 不符合：类型不是 `feat`，描述不说明内容 |
| `Fix_WebScroll` | 不符合：大写和下划线 |
| `feat/k8s-工作负载` | 不符合：带中文 |
| `wip` | 不符合：没有类型 |
| `hubiao/fix` | 不符合：个人前缀 |
| `feat/web-dialog-scroll-and-alert-variables` | 不符合：语法合法，但两个变更应拆成两个任务、两个分支 |

已有同名分支时不删除、不复用，换更具体的描述或追加 `-2` 后缀。

## 提交信息

```
<type>(<scope>): <英文祈使句，小写开头，不加句号>
```

- `type` 与分支类型同义。
- `scope` 用改动所在模块：`web`、`api`、`data`、`biz`、`service`、`deploy`、`docs`、`alert-rule` 等。
- 标题说明改了什么，例如 `fix(web): localize the backend error copy`。
- 正文说明"为什么"和原来的失败现象——哪个请求报什么错、哪个页面看不到什么、为什么要这样取舍，不逐条复述 diff。
- 跨层的同一个逻辑变更放一个提交，一次提交只做一个逻辑变更。
- 多行消息用 heredoc 传入，不要用多个 `-m` 拼转义。

## 进 main 只用 rebase

`main` 必须保持单线，不留合并提交。合并任务分支一律先 rebase 再快进，不用 `git merge`（包括 `--no-ff`）：

```sh
git switch <分支>
git rebase main            # 把本任务的提交重放到 main 顶端
git switch main
git merge --ff-only <分支> # 没有快进空间就停下：说明 rebase 没做干净，回到 rebase 重来
```

- 冲突在 `git rebase` 阶段解决。解不了就 `git rebase --abort` 回到原状，把冲突文件报给用户，不绕过。
- rebase 会重写提交，哈希随之改变；交付说明里的短哈希以 rebase 后的为准。
- 走 PR 时合并方式同样选 rebase（Rebase and merge），不要 merge commit，否则 `main` 又出现分叉。
- 分支已经推送过时，rebase 改写了远端历史：由用户确认后用 `git push --force-with-lease`，不用 `--force`；不替用户强推。
- 任何一步出现冲突都不自动解决：停下来把冲突文件报给用户，不绕过、不硬推。

## 为什么推送和合并交给用户

`.github/workflows/ci.yml` 对任意分支的 push 跑检查，`.github/workflows/release.yml` 在 push 到 `main` 和 `v*` tag 时构建并发布 GHCR 镜像。推送就会产生对外可见的构建结果，时机由用户定。

## 分支清理

任务分支按用户指示合并进 `main` 并推送后，主动 `git branch -d <分支>`，不再回头问一次。
cherry-pick 过来的分支不是 `main` 的祖先，`-d` 会拒绝，这时先 `git cherry main <分支>`
确认每条提交都是 `-`（补丁已在主线），再 `-D`；只要有一条是 `+`，说明还有没进主线的提交，
停下问用户，不删。

未合并进 `main` 的分支（用户保留、或还在等指示的）一律不动。

## 禁止清单

- 在 `main` 上直接提交本任务的改动（用户明确指示除外）。
- 在 `main` 与 `origin/main` 不一致时从 `main` 切新分支：落后先快进，领先或分叉停下问用户。
- 从非 `main` 的分支或提交切本任务分支（用户明确要求接着某个分支做时除外）。
- `git push`、`git push --force`、`git commit --amend` 已推送的提交、`git reset --hard`。
- 对 `main` 或他人的分支执行 `git rebase`；对本任务分支 rebase 后强推，除非用户明确同意。
- 用 `git merge` 把任务分支合进 `main`，包括 `--no-ff`。
- 用 `git stash`、`git checkout --`、`git restore`、`git clean` 动不属于本任务的改动。
- 自动合并、自动解冲突、自动打 tag、自动建 PR；删除只限已确认进 `main` 的分支（合并推送后主动删），未合并的分支不删。
- 任务中途换分支处理另一件事。
