# 控制面多副本部署

命名空间 `alioth`。清单里统一使用 GHCR 上发布的镜像
`ghcr.io/aide-family/moon/<service>:sha-<12>`（`.github/workflows/release.yml` 产出，
多架构）。这里固定用 `sha-` 标签而不是 `main`：`main` 会随主分支漂移，`sha-` 才是不可变引用。
升级版本时用 `kubectl -n alioth set image deploy/<service> <service>=ghcr.io/aide-family/moon/<service>:sha-<新提交>`
逐个换，或者直接改清单里的标签再 apply。

仓库里也有本地构建路径（`make images VERSION=v1.2.3`），产物是 `alioth/<service>:<version>`，
适合离线环境或改动还没进主干时自测；此时把清单里的镜像地址换成本地标签即可。

**镜像来源**：GHCR 包必须公开，否则节点拉不到。私有包需要先建 imagePullSecret 并挂到每个
Deployment 和迁移 Job 上，本仓库目前不提供这部分清单。

**集群依赖的镜像站**：节点的 docker.io 直连不可用，所以 `postgres.yaml` / `redis.yaml`
用的是 `docker.m.daocloud.io` 镜像站地址；配置了 registry mirror 的集群可以换回官方地址。

## 部署顺序

```bash
kubectl apply -f namespace.yaml
kubectl apply -f storage.yaml          # postgres 的 hostPath PV（无动态存储类时必须先建）
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f alioth.yaml           # 只含 Deployment/Service，Secret 单独创建
kubectl apply -f migrate-job.yaml
kubectl wait --for=condition=complete --timeout=180s job/alioth-migrate
kubectl apply -f megrez.yaml
kubectl apply -f merak.yaml
kubectl apply -f web.yaml
kubectl apply -f ingress.yaml          # Kong 网关按域名转发到 web
kubectl apply -f pdb.yaml
kubectl apply -f networkpolicy.yaml
```


迁移必须先于首次滚动完成。`alioth` 启动不会自动建表，`alioth-migrate` Job 调
`alioth migrate`（`ent Schema.Create` + seed + 周分区，可重复执行）。Job 失败时 `backoffLimit: 0`
会直接停下，先看 `kubectl -n alioth logs job/alioth-migrate` 再重试，不要跳过。

## 副本

| 组件 | replicas | 对外 | 探针 |
| --- | --- | --- | --- |
| web | 2 | Kong Ingress `alioth.moon.com` → web:8080 | `/health` |
| alioth | 2 | 仅集群内 8000/9000 | `/livez` + `/readyz`（PG/Redis） |
| megrez | 2 | 仅集群内 8001 | `/livez` + `/readyz`（Redis） |
| merak | 2 | 仅集群内 8002 | `/livez` + `/readyz`（Redis） |

PostgreSQL 使用 `postgres-data` PVC，Pod 重建不丢业务数据。Redis 只保存租约锁，丢失后由 TTL
与下一轮调度恢复，因此不挂持久卷、关闭 RDB/AOF。

`readiness` 查依赖、`liveness` 只查进程：数据库或 Redis 抖动时 Pod 会被摘流量，但不会被重启。
每个 Deployment 都配了 `resources`、`securityContext`（非 root、只读根文件系统、drop ALL
capabilities）、`terminationGracePeriodSeconds` 和 `maxUnavailable: 0` 的滚动策略；`pdb.yaml`
保证节点 drain 时至少留一个副本。

## 排空（preStop）

四个应用 Deployment 都带 `lifecycle.preStop.sleep.seconds: 5`。这不是保险丝，是必需的：
删除 Pod 时 kubelet 立刻发 SIGTERM，而 kube-proxy/iPtables 摘除该 endpoint 还有几百毫秒到
两秒的延迟；`alioth` 收到 SIGTERM 后 kratos 会马上关闭监听，落在这段窗口里的新连接会被拒绝，
`web` 的 nginx 就回 502。

实测数据（700 次 `GET /v1/captcha` 经 Kong，期间滚动重启 alioth + web）：

| 配置 | 结果 |
| --- | --- |
| 无 preStop | 398×200 + 2×502（502 精确出现在两个旧 Pod 退出的时刻） |
| preStop.sleep 5s | 700×200，0 个非 2xx |

用 `sleep` 动作而不是 `exec: ["sleep","5"]`，因为 `alioth` 镜像是 distroless，里面没有
`sleep` 二进制；`preStop.sleep` 是 k8s 1.30+ 的原生能力，不依赖容器内任何命令。

`networkpolicy.yaml` 只收口 PostgreSQL/Redis 的入向（分别只允许 alioth 与 alioth/megrez/merak），
应用 HTTP/gRPC 端口不限制，避免不同 CNI 对 kubelet 探针流量的处理差异把 Pod 摘掉。

## 密钥

`secrets.example.yaml` 只是模板，**不要 apply**，它里面的值都是占位值。清单里已经不含内联
Secret，所以 `kubectl apply -f alioth.yaml` 不会覆盖线上真实密钥。三个 Secret 这样生成：

```bash
PG_PASSWORD="$(openssl rand -hex 24)"   # 只用十六进制字符，免去 URL 转义
kubectl -n alioth create secret generic alioth-postgres \
  --from-literal=POSTGRES_USER=alioth \
  --from-literal=POSTGRES_PASSWORD="${PG_PASSWORD}" \
  --from-literal=POSTGRES_DB=alioth \
  --from-literal=DATABASE_SOURCE="postgres://alioth:${PG_PASSWORD}@postgres:5432/alioth?sslmode=disable" \
  --dry-run=client -o yaml | kubectl apply -f -
```

`POSTGRES_PASSWORD` 和 `DATABASE_SOURCE` 里的密码必须是同一份，所以先落一个变量再复用；
写死了两份随机值的命令会让 postgres 初始化出一套密码、alioth 却拿另一套去连。

```bash
kubectl -n alioth create secret generic alioth-auth \
  --from-literal=JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=ENCRYPT_KEY="$(openssl rand -hex 32)" \
  --from-literal=MEGREZ_SERVICE_KEY="sk-$(openssl rand -hex 24)" \
  --from-literal=MERAK_SERVICE_KEY="sk-$(openssl rand -hex 24)" \
  --from-literal=WEB_URL="http://alioth.moon.com" \
  --dry-run=client -o yaml | kubectl apply -f -
```

`ENCRYPT_KEY` 是数据源密码/Webhook/SSH 凭据的 AES-GCM 主密钥（内部先做 sha256），
**一旦换了就无法解密历史密文**，务必留着备份。

## 私有 CA（Elasticsearch）

集群里的 ECK Elasticsearch 走 TLS，证书由集群内 CA 签发，而数据源客户端用 Go 默认信任库、
没有跳过校验的开关。要接这类自签 HTTPS 数据源，把「系统 CA + 该 CA」合成为一个 bundle，
建 Secret `alioth-ca-bundle`，`alioth.yaml` 会把它挂到 `/etc/ssl/certs/ca-certificates.crt`：

```bash
kubectl -n alioth exec deploy/webhook-sink -- cat /etc/ssl/certs/ca-certificates.crt > sys-ca.crt
kubectl -n es get secret es-es-http-certs-public -o jsonpath='{.data.ca\.crt}' | base64 -d > es-ca.crt
cat sys-ca.crt es-ca.crt > ca-bundle.crt
kubectl -n alioth create secret generic alioth-ca-bundle --from-file=ca-certificates.crt=ca-bundle.crt
kubectl -n alioth rollout restart deploy/alioth
```

Secret 不存在时这个卷是 `optional` 的，镜像自带信任库不受影响。另外 ECK 证书的 SAN 是
`es-es-http.es.svc`，**不含** `.cluster.local`，端点要写成 `https://es-es-http.es.svc:9200`。


## SMTP

登录和注册只有邮箱验证码，没有密码登录，所以 SMTP 不通等于完全无法登录：

```bash
kubectl -n alioth create secret generic alioth-smtp \
  --from-literal=ALIOTH_SMTP_HOST=smtp.example.com \
  --from-literal=ALIOTH_SMTP_PORT=587 \
  --from-literal=ALIOTH_SMTP_USERNAME=... \
  --from-literal=ALIOTH_SMTP_PASSWORD=... \
  --from-literal=ALIOTH_SMTP_FROM=alioth@example.com \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n alioth rollout restart deploy/alioth
```

实现用的是 `net/smtp`：只支持 **STARTTLS（587/25）**，不支持 465 隐式 SSL；填了用户名就会走
`smtp.PlainAuth`，端口协商不出 STARTTLS 时会被拒发。改完 Secret 必须重启 `alioth`，
配置是进程启动时读的。


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

## 历史告警分库分表

恢复的告警事件写入「租户库 + 周表」：库名为 `ALIOTH_HISTORY_DATABASE_PREFIX` + 租户标识
（`Namespace` UID），表名为 `ALIOTH_HISTORY_TABLE_PREFIX` + 周一日期 `YYYYMMDD`。默认前缀为
`alioth_alert_` 与 `alert_events_`。

- 数据库账号需要 `CREATEDB` 权限，alioth 会在首次归档某个租户时创建该租户库。
- 历史库与主库在同一个 PostgreSQL 实例；归档失败时事件保留在主库，历史列表会兜底展示，恢复权限后由每小时回填迁移。
- 连接按租户缓存，连接数会计入 PostgreSQL 的 `max_connections`，租户很多时需要相应上调。

## nginx

`web` 只把 `/v1` 转到 `alioth:8000`。配置只有一份来源：[web/nginx.conf](../../web/nginx.conf)，
构建进镜像；k8s 清单不再挂 ConfigMap，避免两份配置漂移导致 WebSocket 握手失败。

## 验收要点

- 就绪副本均 ≥ 2，`kubectl -n alioth get endpoints` 不含未就绪 Pod。
- 停一个 alioth / megrez / merak 副本后，经 `http://alioth.moon.com` 的登录、评测、通知仍可用。
- `kubectl -n alioth delete pod -l app=alioth --wait=false` 滚动期间，重复请求不出现 5xx。
- 停掉 PostgreSQL 后 `deploy/alioth` 的 ready 副本变为 1，恢复后自动回到 2；`restartCount` 不变。
- 集群外只能通过 Kong 的 80/443 访问 `alioth.moon.com`；8000、8001、8002、8022、9000、9022 不可达。
- 集群内可访问 `megrez:8001/readyz` 与 `merak:8002/readyz`。
- 经过 `http://alioth.moon.com` 打开机器页交互式 shell，devtools 显示 `101 Switching Protocols`，空闲 5 分钟不断线。
- 摇光安装见 [../alkaid/README.md](../alkaid/README.md)。
