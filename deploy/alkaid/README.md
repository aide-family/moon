# 摇光（alkaid）安装

alkaid 只出站连接 `alioth:9000`。控制面不回拨 Agent。8022 / 9022 仅供本机或集群内健康检查，不要发布到集群外，也不要写进 Web nginx。

## 一键部署包（systemd）

每个 `v*` tag 的 GitHub Release 附带 linux/amd64 与 linux/arm64 的裸二进制
（`alkaid_<版本>_linux_<架构>`）和一键部署包，包内是 `alkaid` 二进制、`alkaid.service`、
`alkaid.yaml`、`alkaid.env.example`、`install.sh`、`VERSION` 和本说明：

```bash
sha256sum -c SHA256SUMS          # macOS 用 shasum -a 256 -c SHA256SUMS
tar -xzf alkaid_v0.1.0_linux_amd64.tar.gz
cd alkaid_v0.1.0_linux_amd64
sudo ./install.sh                # 加 --no-start 只安装并 enable
```

`install.sh` 会把二进制装到 `/usr/local/bin/alkaid`、配置写到 `/etc/alioth/`、单元写到
`/etc/systemd/system/alkaid.service`，然后 `daemon-reload` 并 `enable`。`alkaid.env` 已存在时不会被覆盖；
令牌为空或仍是占位值时不会启动服务，填好后 `systemctl restart alkaid`。

本地打同一个包：`make package-alkaid VERSION=v0.1.0`（产物在 `dist/`，`ARCHES=arm64` 可只打一个架构）。

## RPM 包

同一批 Release 资产里还有 `alkaid-<版本>-<release>.x86_64.rpm` 与
`alkaid-<版本>-<release>.aarch64.rpm`（版本号去掉 `v` 前缀，正式版 release 是 `1`，
`v1.2.3-rc1` 这类预发布是 `0.rc1`，`sha-xxxxx` 这类非语义版本落到 `0.0.0-<sha>`）：

```bash
sudo dnf install ./alkaid-0.1.0-1.x86_64.rpm     # RHEL / openEuler；或 rpm -Uvh
sudoedit /etc/alioth/alkaid.env                  # 填 ALKAID_AGENT_TOKEN
sudo systemctl restart alkaid
```

RPM 装完会 `enable` 但不会启动（令牌还是占位值时启动只会刷错误日志），升级时
`/etc/alioth/alkaid.yaml` 和 `alkaid.env` 按 `%config(noreplace)` 保留现场修改。文件位置：
二进制 `/usr/local/bin/alkaid`、单元 `/usr/lib/systemd/system/alkaid.service`、配置
`/etc/alioth/`。本地打 RPM 需要 rpmbuild（Debian/Ubuntu 装 `rpm`，RHEL/openEuler 装 `rpm-build`）：
`RPM=1 make package-alkaid VERSION=v0.1.0`。

## 手工 systemd 安装

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
3. 需要访问节点本地文件时，把 Deployment 改成 DaemonSet。
4. SSH 主机密钥默认不校验（兼容旧部署）。生产建议配置 `ALKAID_SSH_KNOWN_HOSTS`；配置后无法通过校验的主机会直接失败。
5. 与 `alioth` 的 gRPC 连接默认明文。设置 `ALKAID_ALIOTH_TLS_CA` 启用 TLS；同时设置 `ALKAID_ALIOTH_TLS_CERT/KEY` 时启用 mTLS。控制面用 `ALIOTH_GRPC_TLS_CERT/KEY` 开启 TLS，配置 `ALIOTH_GRPC_TLS_CLIENT_CA` 后要求客户端证书。

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

