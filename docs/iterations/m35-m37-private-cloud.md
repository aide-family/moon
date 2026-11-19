# M35–M37 私有云：镜像仓库与应用商店

目标：在「私有云」菜单下补齐镜像仓库与应用商店，二者共享镜像、发布包与部署链路；后端沿用 Alioth 的 `server → service → biz → data` 分层，Web 继续使用 Vite + shadcn/ui。

## M35 镜像仓库

### 已交付

- 仓库连接：Harbor / Docker V2，保存前探测；密码加密存储；删除连接时清理密文。
- 镜像浏览：项目、仓库、Tag、多架构平台与 digest；镜像详情。
- 镜像同步：已接入仓库或外部镜像源 → 已接入目标仓库；全部/指定平台、覆盖开关、状态机、取消、重试、租约执行。
- 外部源临时凭据加密保存，终态 24 小时后清理；预览不落库。
- 权限与菜单：`registry_access`、`registry_image`、`registry_sync`；菜单「私有云 → 镜像仓库」。

### 主要接口

- `/v1/registries`、`/v1/registries/{uid}`、`/v1/registries/{uid}/projects|repositories|tags`
- `/v1/image-syncs`、`/v1/image-syncs/{uid}/retry|cancel`

### 已知边界

- Harbor 专属统计（pull count 等）未接入，按标准 OCI 目录读取。
- 同步的 blob 进度暂不逐 blob 统计；自动重试上限未实现，失败后由用户重试。

## M36 应用商店

### 已交付

- `Application → ApplicationVersion → ApplicationVersionImage`：应用、多版本、镜像快照。
- 发布包版本支持任意 Kubernetes 资源 YAML：
  - 解析、必填字段、重复资源、大小限制；
  - 禁止 `Namespace`、`Node`、`Event`、`APIService`、`ComponentStatus`、`TokenReview`、`SubjectAccessReview`；
  - 静态 kind→scope 映射，未知类型标记「部署时判定」；
  - Secret 值与 `stringData` 抽取后加密，清单里只留 key；
  - 扫描 `image` 字段，匹配已接入仓库时解析 digest/平台，否则标记外部。
- 草稿校验、发布不可变、下架；发布时服务端重新校验，不依赖前端。
- 部署模型：`ApplicationDeployment`（组）→ 每集群 `ApplicationRelease` → 每实例 `ApplicationReleaseOperation`。
- 执行器：数据库 `SKIP LOCKED` 认领、租约、Server-Side Apply、归属 labels、CRD 先行、工作负载就绪等待、按上次成功操作裁剪、卸载。
- 权限与菜单：`appstore_package`、`appstore_deploy`、`appstore_cluster_scope`；菜单「私有云 → 应用商店」。

### 主要接口

- `/v1/application`、`/v1/applications`、`/v1/application/{uid}/versions`
- `/v1/application/{uid}/version/{versionUid}/validate|publish|deprecate`
- `/v1/application-deployments`、`/{uid}/upgrade|rollback|redeploy|uninstall|operations`
- `/v1/application-releases/{uid}`、`/v1/application-release-operations/{uid}`

### 2026-11-19 增补：版本内服务模型

- 模型从 `Application → Version → Resource` 扩为 `Application → Version → Component（服务）→ Resource → Image`：一个版本由多个服务组成，每个服务包含若干 Kubernetes 资源，界面以服务卡片展示。
- 新增 `application_version_components`、`application_version_resources` 两张表；`application_versions` 增加 `revision`、`composition_version`、`status_changed_at`、`deprecated_at`、`copied_from_version_uid`；`application_version_images` 增加 `component/resource_uid/container_path/project/source/resolved_image`。
- 兼容：迁移时按旧 `spec` 回填一个 `default` 服务与资源行并标记 `composition_version=1`；新写入继续双写 `spec` 与版本级 Secret 密文，读取优先归一化表，回填失败回退 `spec`。
- 版本状态机：`DRAFT → PUBLISHED → DEPRECATED`（可重新上架）。只有草稿可增删改资源；上架要求每个服务至少一个资源并把镜像固定到 digest；已上架版本只读，改为「复制为新版本」。
- 新接口：资源级 `Create/Get/Save/Delete/Validate ApplicationVersionResource`（带 `expectedRevision` 乐观锁）、`DuplicateApplicationVersion`、`ListApplicationReleases`（按应用/集群查发布记录与集群当前版本）。
- 发布规则：首次发布与升级只接受 `PUBLISHED`；回滚接受 `PUBLISHED` 或曾部署过该 release 的 `DEPRECATED`；重新部署要求当前版本非草稿。部署组详情的集群矩阵同时返回期望版本、集群当前版本与 `inSync`。
- Web：版本页默认「服务」Tab；资源详情 Sheet、表单/YAML 双 Tab 编辑、镜像仓库级联选择；YAML 编辑器接入 monaco-yaml 与内置 K8s schema；发布记录、集群版本列与复制版本弹窗。

## M37 多集群与集群级资源

### 已交付

- 部署对话框多选 1–20 个集群；一次请求创建组、N 个实例与 N 条操作。
- 升级/回滚/重新部署/卸载支持只作用于选中的集群；部署组详情按集群矩阵展示，并提供单集群重试。
- 预检 all-or-nothing：命名空间、集群可达、RESTMapper scope、禁止类型、集群级权限、镜像 digest；同包 CRD 的 CR 跳过 dry-run，待 CRD Established 后执行。
- 集群级资源：
  - 需要 `release:cluster`；
  - 卸载默认保留，显式删除需权限并二次确认；
  - `keepPersistentVolumeClaims` 默认 true。
- 引用保护：
  - 删除 Tag：被 `PUBLISHED`/`DEPRECATED` 版本引用时 409 并列出应用/版本；草稿引用只记录告警。
  - 删除仓库连接：有活动同步任务或任何版本镜像引用时 409。
  - 删除应用：有非草稿版本或未卸载部署组时 409。
  - 删除版本：仅草稿；仍被未卸载实例引用时 409。
- 副本数覆盖：按资源 `item` 覆盖 Deployment/StatefulSet 的 `spec.replicas`，范围 0–100。
- 调度器启动两个执行 worker，仍由数据库部分唯一索引与 `SKIP LOCKED` 保证同一实例不并发。
- 执行租约 10 分钟；超时/失败写回消息，不自动重放，由用户按集群重新部署或回滚。

### 关键安全规则

- `release:write` 是高权限：可在目标命名空间创建任意命名空间级资源；`release:cluster` 再扩展集群级资源。
- 冲突检测：目标资源已存在且 `alioth.io/release-uid` 不是当前实例时返回 409，不 force 抢占。
- Secret 值只写不读；列表/详情只返回 key。
- 禁止类型永不由应用商店创建或删除；目标命名空间必须预先存在。

## 数据表

- `registries`、`image_syncs`
- `applications`、`application_versions`、`application_version_images`
- `application_deployments`、`application_releases`、`application_release_operations`

迁移包归档：`applications`、`application_versions`、`application_version_images`、`registries`；运行态表不进入归档。

## Web 路由

- `/registries`、`/registries/syncs`、`/registries/:uid`、`/registries/:uid/repository`
- `/appstore`、`/appstore/apps/:uid`、`/appstore/apps/:uid/versions/:versionUid`
- `/appstore/deployments/:uid`

## 验证

- `go build ./...`、相关 `go test`、`gofmt`、`check-naming.sh`、`buf lint`
- `npm --prefix web run check`、`npm --prefix web run build`
- 浏览器烟测：应用列表、应用详情、版本编辑器（添加资源、保存、校验）、多集群部署对话框、部署详情集群选择与卸载选项。

## 尚未覆盖

- 镜像 Tag、Ingress Host、Secret 值覆盖只完成存储/接口，执行器目前只应用副本数覆盖。
- 部署详情未实时读取集群资源状态，只展示操作记录与集群矩阵。
- 未在开发环境对真实业务集群执行实际 apply；执行链路通过编译、预检代码路径与 UI 验证。
- 引用保护中的草稿 Tag 引用只写日志，不返回警告结构。
