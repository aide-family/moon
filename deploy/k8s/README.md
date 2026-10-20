# 控制面多副本部署

命名空间 `alioth`。应用清单顺序：

```bash
kubectl apply -f namespace.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f alioth.yaml
kubectl apply -f megrez.yaml
kubectl apply -f merak.yaml
kubectl apply -f web.yaml
```

## 副本

| 组件 | replicas | 对外 |
| --- | --- | --- |
| web | 2 | LoadBalancer :8080 |
| alioth | 2 | 仅集群内 8000/9000 |
| megrez | 2 | 仅集群内 8001 |
| merak | 2 | 仅集群内 8002 |

PostgreSQL 与 Redis 独立 Deployment。PostgreSQL 使用 `postgres-data` PVC，Pod 重建不丢业务数据。Redis 只保存租约锁，丢失后由 TTL 与下一轮调度恢复，因此不挂持久卷、关闭 RDB/AOF。

## 密钥

清单里的 `JWT_SECRET`、`ENCRYPT_KEY`、`MEGREZ_SERVICE_KEY`、`MERAK_SERVICE_KEY`、`POSTGRES_PASSWORD` 都是占位值，**部署前必须替换**，不要把真实值提交到仓库。推荐先创建 Secret 再应用清单：

```bash
kubectl -n alioth create secret generic alioth-auth \
  --from-literal=JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=ENCRYPT_KEY="$(openssl rand -hex 32)" \
  --from-literal=MEGREZ_SERVICE_KEY="sk-$(openssl rand -hex 24)" \
  --from-literal=MERAK_SERVICE_KEY="sk-$(openssl rand -hex 24)" \
  --dry-run=client -o yaml | kubectl apply -f -
```

创建后把 `alioth.yaml` 里的 Secret 清单删掉再 `kubectl apply`，避免占位值覆盖真实值。

## 迁移

`alioth` 启动不会自动建表。首次部署先执行一次迁移（或在 CI/Job 中执行）：

```bash
kubectl -n alioth exec deploy/alioth -- /app/alioth migrate --conf /app/configs/config.yaml
```

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

`alioth` 会回收长时间停留在 SENDING 的通知任务，默认 2 分钟。需要调整时给 `alioth` 增加环境变量 `ALIOTH_NOTIFY_STALE_SECONDS`（秒）。

## nginx

`web` 只把 `/v1` 转到 `alioth:8000`。配置中不出现 megrez、merak、alkaid。

## 验收要点

- 就绪副本均 ≥ 2。
- 停一个 alioth / megrez / merak 副本后，经 `http://<web>:8080` 的登录、评测、通知仍可用。
- 集群外只能访问 web:8080；8000、8001、8002、8022、9000、9022 不可达。
- 集群内可访问 `megrez:8001/health` 与 `merak:8002/health`。
- 摇光安装见 [../alkaid/README.md](../alkaid/README.md)。
