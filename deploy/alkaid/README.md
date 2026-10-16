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

## 网络

| 端口 | 用途 | 集群外 |
| --- | --- | --- |
| 8022 | HTTP 健康检查 | 不可达 |
| 9022 | gRPC 健康检查 | 不可达 |
| alioth:9000 | 出站双向流 | 仅集群内 |

