# 控制面多副本部署

命名空间 `alioth`。镜像由 `make images VERSION=v1.2.3` 在仓库根目录构建，清单里统一使用
`alioth/<service>:latest`；发版时用 `kubectl -n alioth set image deploy/<service> <service>=alioth/<service>:<version>`
把四个 Deployment 换成带版本戳的镜像，不要长期停留在 `latest`。

## 部署顺序

```bash
kubectl apply -f namespace.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f alioth.yaml          # Secret 与 Deployment
kubectl apply -f migrate-job.yaml
kubectl wait --for=condition=complete --timeout=180s job/alioth-migrate
kubectl apply -f megrez.yaml
kubectl apply -f merak.yaml
kubectl apply -f web.yaml
kubectl apply -f pdb.yaml
kubectl apply -f networkpolicy.yaml
```

迁移必须先于首次滚动完成。`alioth` 启动不会自动建表，`alioth-migrate` Job 调
`alioth migrate`（`ent Schema.Create` + seed + 周分区，可重复执行）。Job 失败时 `backoffLimit: 0`
会直接停下，先看 `kubectl -n alioth logs job/alioth-migrate` 再重试，不要跳过。

## 副本

| 组件 | replicas | 对外 | 探针 |
| --- | --- | --- | --- |
| web | 2 | LoadBalancer :8080 | `/health` |
| alioth | 2 | 仅集群内 8000/9000 | `/livez` + `/readyz`（PG/Redis） |
| megrez | 2 | 仅集群内 8001 | `/livez` + `/readyz`（Redis） |
| merak | 2 | 仅集群内 8002 | `/livez` + `/readyz`（Redis） |

PostgreSQL 使用 `postgres-data` PVC，Pod 重建不丢业务数据。Redis 只保存租约锁，丢失后由 TTL
与下一轮调度恢复，因此不挂持久卷、关闭 RDB/AOF。

`readiness` 查依赖、`liveness` 只查进程：数据库或 Redis 抖动时 Pod 会被摘流量，但不会被重启。
每个 Deployment 都配了 `resources`、`securityContext`（非 root、只读根文件系统、drop ALL
capabilities）、`terminationGracePeriodSeconds` 和 `maxUnavailable: 0` 的滚动策略；`pdb.yaml`
保证节点 drain 时至少留一个副本。

`networkpolicy.yaml` 只收口 PostgreSQL/Redis 的入向（分别只允许 alioth 与 alioth/megrez/merak），
应用 HTTP/gRPC 端口不限制，避免不同 CNI 对 kubelet 探针流量的处理差异把 Pod 摘掉。

## 密钥

清单里的 `JWT_SECRET`、`ENCRYPT_KEY`、`MEGREZ_SERVICE_KEY`、`MERAK_SERVICE_KEY`、`POSTGRES_PASSWORD`
都是占位值，**部署前必须替换**，不要把真实值提交到仓库。推荐先创建 Secret，再应用清单：

```bash
kubectl -n alioth create secret generic alioth-auth \
  --from-literal=JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=ENCRYPT_KEY="$(openssl rand -hex 32)" \
  --from-literal=MEGREZ_SERVICE_KEY="sk-$(openssl rand -hex 24)" \
  --from-literal=MERAK_SERVICE_KEY="sk-$(openssl rand -hex 24)" \
  --dry-run=client -o yaml | kubectl apply -f -
```

创建后删除 `alioth.yaml` 顶部的 Secret 清单再 `kubectl apply`（或改用 External Secrets /
Sealed Secrets），避免占位值覆盖真实值。

## gRPC TLS

控制面与工作进程支持可选 TLS/mTLS，通过环境变量提供证书路径（挂载 Secret 成文件）：

| 变量 | 使用方 | 说明 |
| --- | --- | --- |
| `ALIOTH_GRPC_TLS_CERT` / `ALIOTH_GRPC_TLS_KEY` | alioth 服务端 | 启用 TLS |
| `ALIOTH_GRPC_TLS_CLIENT_CA` | alioth 服务端 | 启用 mTLS，要求客户端证书 |
| `ALIOTH_GRPC_TLS_CA` | megrez、merak | 校验服务端证书 |
| `ALIOTH_GRPC_TLS_CERT` / `ALIOTH_GRPC_TLS_KEY` | megrez、merak | 提供客户端证书 |
| `ALIOTH_GRPC_TLS_SERVER_NAME` | megrez、merak | 证书里没有服务名时显式指定 |

`alkaid` 用 `ALKAID_ALIOTH_TLS_*`，见 [../alkaid/README.md](../alkaid/README.md)。

## 通知回收

`alioth` 会回收长时间停留在 SENDING 的通知任务，默认 2 分钟。需要调整时给 `alioth` 增加环境变量
`ALIOTH_NOTIFY_STALE_SECONDS`（秒）。

## nginx

`web` 只把 `/v1` 转到 `alioth:8000`。配置只有一份来源：[web/nginx.conf](../../web/nginx.conf)，
构建进镜像；k8s 清单不再挂 ConfigMap，避免两份配置漂移导致 WebSocket 握手失败。

## 验收要点

- 就绪副本均 ≥ 2，`kubectl -n alioth get endpoints` 不含未就绪 Pod。
- 停一个 alioth / megrez / merak 副本后，经 `http://<web>:8080` 的登录、评测、通知仍可用。
- `kubectl -n alioth delete pod -l app=alioth --wait=false` 滚动期间，重复请求不出现 5xx。
- 停掉 PostgreSQL 后 `deploy/alioth` 的 ready 副本变为 1，恢复后自动回到 2；`restartCount` 不变。
- 集群外只能访问 web:8080；8000、8001、8002、8022、9000、9022 不可达。
- 集群内可访问 `megrez:8001/readyz` 与 `merak:8002/readyz`。
- 经过 `http://<web>:8080` 打开机器页交互式 shell，devtools 显示 `101 Switching Protocols`，空闲 5 分钟不断线。
- 摇光安装见 [../alkaid/README.md](../alkaid/README.md)。
