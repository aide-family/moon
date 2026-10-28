# 分支与提交规范

本文是强制约束，不是建议。每个新会话任务都按本文执行：凡是会改动 git 仓库内容的任务，先基于当前分支位置切出工作分支，在分支上完成开发，结束时只提交到分支，不推送、不合并，把推送还是本地合并进 `main`（只用 rebase 快进，不留合并提交）的决定权交回用户。

只读任务（只读文件、只跑查询命令、不改工作区）不受约束。用户明确要求"直接在 main 上改"时按用户指示执行，并在交付说明里写明这次走了例外。

## 适用范围

- 会改动仓库内容的任务：新增、修改、删除文件，重跑生成代码，改文档，改配置。
- 判断方法：任务开始时记录 `git status --porcelain`，任务结束时非空即视为发生了变更，必须走分支流程。

## 开工前

记录基线，并写进最后的交付说明：

```sh
git branch --show-current   # 当前分支
git rev-parse --short HEAD  # 起点提交
git status --porcelain      # 任务开始前已有的未提交改动
```

从当前分支的当前位置切出工作分支。不先 `pull`，不先 `rebase`，不先 `reset`：

```sh
git switch -c <类型>/<kebab-case-描述>
```

"当前分支位置"指任务开始时 HEAD 指向的提交。默认不基于 `origin/main`，避免把远端新提交悄悄带进本任务。

工作区已经有上一任务或他人的未提交改动时，原样保留：不 `stash`、不 `checkout --`、不 `restore`、不 `clean`，也不把这些改动提交进本任务。只改本任务自己的文件。

一个任务一个分支。分支不合并其他分支；任务中途不为别的事再切分支——那是另一个任务、另一个分支。

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

## 开发

- 全部改动在当前工作分支完成。不要切回 `main` 改文件再切回来。
- 生成代码（`make conf`、ent、Wire、`buf generate`）在分支上重跑，生成物与源文件进同一次提交。
- 不顺手格式化、重排与本任务无关的文件，不做"顺手优化"。

## 提交

提交前：

- 跑本任务相关的验证。后端 `make vet`、`make lint`、`make test`；`web/` 跑 typecheck、lint、build。交付说明里如实写结果，没跑就写"未验证"，不写"应该没问题"。
- 只 stage 本任务文件，逐路径 `git add <path>`。禁止 `git add -A`、`git add .`、`git commit -a`——工作区里可能有别人的改动。
- 提交后用 `git status --porcelain` 复核：不属于本任务的条目必须仍处于未提交状态，一个都不能被带进提交。

提交信息用 Conventional Commits，与仓库现有提交一致：

```
<type>(<scope>): <英文祈使句，小写开头，不加句号>
```

- `type` 与分支类型同义：`feat`、`fix`、`refactor`、`docs`、`test`、`chore`、`perf`、`build`、`ci`、`revert`。
- `scope` 用改动所在模块：`web`、`api`、`data`、`biz`、`service`、`deploy`、`docs`、`alert-rule` 等。
- 标题说明改了什么，例如 `fix(web): localize the backend error copy`。
- 正文说明"为什么"和原来的失败现象——哪个请求报什么错、哪个页面看不到什么、为什么要这样取舍，不逐条复述 diff。跨层的同一个逻辑变更放一个提交，一次提交只做一个逻辑变更。
- 多行消息用 heredoc 传入，不要用多个 `-m` 拼转义。

## 完成后

默认停在这里：不推送，不合并，不对任务分支 rebase（rebase 只在你明确要求合并时执行），不打 tag，不建 PR。

交付说明必须给出：

- 分支名与起点，例如 `fix/web-dialog-scroll`，基于 `main` @ `742226b`。
- 提交短哈希与标题，按顺序列出。
- 变更文件清单。
- 验证结论（跑了什么，结果如何，什么没跑）。
- 工作区里本任务之外的未提交改动是否仍原样保留（没有就写"无"）。

后续动作由用户决定，得到明确指示前一条都不执行：

| 用户的选择 | 执行 |
| --- | --- |
| 推送分支，自己去开 PR | `git push -u origin <分支>`，不加 `--force` |
| 本地合并进 `main` | 见下方「进 main 只用 rebase」 |
| 保留分支暂不处理 | 不执行任何命令 |
| 删除分支 | `git branch -d <分支>`，用户明确要求时才删，不用 `-D` |

### 进 main 只用 rebase

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

任何一步出现冲突都不自动解决：停下来把冲突文件报给用户，不绕过、不硬推。

推送和合并为什么交给用户：`.github/workflows/ci.yml` 对任意分支的 push 跑检查，`.github/workflows/release.yml` 在 push 到 `main` 和 `v*` tag 时构建并发布 GHCR 镜像。推送就会产生对外可见的构建结果，时机由用户定。

## 禁止

- 在 `main` 上直接提交本任务的改动（用户明确指示除外）。
- `git push`、`git push --force`、`git commit --amend` 已推送的提交、`git reset --hard`。
- 对 `main` 或他人的分支执行 `git rebase`；对本任务分支 rebase 后强推，除非用户明确同意。
- 用 `git merge` 把任务分支合进 `main`，包括 `--no-ff`：`main` 只接受 rebase 后的 `--ff-only` 快进。
- 用 `git stash`、`git checkout --`、`git restore`、`git clean` 动不属于本任务的改动。
- 自动合并、自动解冲突、自动打 tag、自动建 PR、自动删分支。
- 任务中途换分支处理另一件事。
