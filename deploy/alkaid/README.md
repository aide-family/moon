# 摇光（alkaid）安装

alkaid 只出站连接 `alioth:9000`。控制面不回拨 Agent。8022 / 9022 仅供本机或集群内健康检查，不要发布到集群外，也不要写进 Web nginx。

## systemd

1. 把 `alkaid` 二进制放到 `/usr/local/bin/alkaid`。
2. 安装配置：
   - `/etc/alioth/alkaid.yaml`（可用本目录 `alkaid.yaml`）
   - `/etc/alioth/alkaid.env`（从 `alkaid.env.example` 复制，填入 `ALKAID_AGENT_TOKEN`）
3. 权限：`chmod 600 /etc/alioth/alkaid.env`，属主 root。
4. 安装单元：`cp alkaid.service /etc/systemd/system/`，然后 `systemctl daemon-reload && systemctl enable --now alkaid`。
5. 在管理台机器页确认在线。

## Kubernetes

1. 在控制台创建机器并复制令牌，写入 `k8s-deployment.yaml` 的 Secret（不要提交真实令牌）。
2. `kubectl apply -f k8s-deployment.yaml -n alioth`。
3. 需要访问节点本地文件时，把 Deployment 改成 DaemonSet，并在说明里保留 ServiceAccount。
4. 集群内访问 API Server 时使用挂载的 ServiceAccount；也可用环境变量 `ALKAID_K8S_API` / `ALKAID_K8S_TOKEN` 覆盖。
5. TLS 默认校验 API Server 证书，CA 顺序为 `ALKAID_K8S_CA`、ServiceAccount 的 `ca.crt`。只有本地调试才设置 `ALKAID_K8S_INSECURE=true`。
6. SFTP 默认沿用 SSH 账户权限。需要收口时设置 `ALKAID_SFTP_ROOT`，所有列目录、上传、下载、建目录和删除都会被限制在该目录内。单次传输上限 4 MiB，超过时明确失败，不会返回截断内容。
7. SSH 主机密钥默认不校验（兼容旧部署）。生产建议配置 `ALKAID_SSH_KNOWN_HOSTS`；配置后无法通过校验的主机会直接失败。
8. Kubernetes API 默认使用集群内 ServiceAccount。设置 `ALKAID_KUBECONFIG` 后改为按 kubeconfig 的 current-context 连接，支持内嵌 CA、客户端证书、token、tokenFile 和 Basic Auth；不执行 `exec`、`auth-provider` 等客户端插件。
9. 与 `alioth` 的 gRPC 连接默认明文。设置 `ALKAID_ALIOTH_TLS_CA` 启用 TLS；同时设置 `ALKAID_ALIOTH_TLS_CERT/KEY` 时启用 mTLS。控制面用 `ALIOTH_GRPC_TLS_CERT/KEY` 开启 TLS，配置 `ALIOTH_GRPC_TLS_CLIENT_CA` 后要求客户端证书。

## 网络

| 端口 | 用途 | 集群外 |
| --- | --- | --- |
| 8022 | HTTP 健康检查 | 不可达 |
| 9022 | gRPC 健康检查 | 不可达 |
| alioth:9000 | 出站双向流 | 仅集群内 |

## 多副本控制面

`alioth` 多副本时，Agent 会对每个副本各建一条出站流，任意副本都能处理该 Agent 的请求。

- Kubernetes：把 `ALKAID_ALIOTH_GRPC_ADDR` 指向无头 Service（`alioth-headless:9000`）。Agent 每 30 秒重新解析一次，副本扩缩容后自动建立或释放对应的流。
- systemd / 集群外：在 `ALKAID_ALIOTH_GRPC_ADDR` 里用逗号分隔所有控制面地址，例如 `10.0.0.11:9000,10.0.0.12:9000`。只写一个地址时，只有该副本能处理这个 Agent 的请求。

