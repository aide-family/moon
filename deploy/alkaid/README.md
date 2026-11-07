# 摇光（alkaid）安装

alkaid 只出站连接 `alioth:9000`。控制面不回拨 Agent。8022 / 9022 仅供本机或集群内健康检查，不要发布到集群外，也不要写进 Web nginx。

## 控制台自动安装

机器页的「自动安装」入口会由 `alioth` 用录入时保存的 SSH 凭据登录目标机器，完成以下步骤并把每一步和命令输出显示在页面上：

1. 建立 SSH 连接，读取 `uname` 和 `id -u`。
2. 按目标架构（amd64 / arm64）准备摇光安装包。
3. 上传并以安装用户解压到 `/tmp`。
4. 以 root（SSH 用户本身是 root 时不再套 `sudo`）执行包内 `install.sh --no-start`。
5. 重签 Agent 令牌，写入 `/etc/alioth/alkaid.env`（控制面 gRPC 地址与令牌）。
6. `systemctl daemon-reload && systemctl enable --now alkaid`。
7. 等待 Agent 向控制面注册，成功后机器页显示在线。

前置条件：目标是 Linux，装有 `systemd`、`tar`、`sudo`（非 root 用户时）；`alioth`
能主动访问目标的 22 端口（或录入的 SSH 端口），目标能访问 `deploy.alioth_grpc_addr`
指向的控制面 gRPC 地址。

安装包按下面的顺序取，前一个拿不到才看下一个：

1. **镜像内置包**：官方 `alioth` 镜像在构建时把 amd64 / arm64 两个
   `alkaid_<version>_linux_<arch>.tar.gz` 打进 `/app/alkaid/`，默认配置直接命中，
   不依赖 GitHub Release，内网也能用。默认路径是
   `/app/alkaid/alkaid_<version>_linux_<arch>.tar.gz`，占位符按控制面版本和目标架构展开。
2. **显式本地包**：`deploy.alkaid_package_path` 指向一个固定文件（离线部署最直接）。
   显式固定路径不存在时直接报错，不会偷偷换成远程下载。
3. **远程下载**：内置包和本地包都不可用时，按 `deploy.alkaid_download_url` 模板下载；
   默认地址是 GitHub Release 的同版本资产，只有打过 `v*` tag、发过 Release 才存在。

```yaml
deploy:
  # 默认走镜像内置包；离线或特殊版本时改成固定本地文件。
  alkaid_package_path: /app/alkaid/alkaid_<version>_linux_<arch>.tar.gz
  # 可选：没有内置包时从制品库/Release 下载。
  alkaid_version: v0.1.0
  alkaid_download_url: https://github.com/aide-family/moon/releases/download/<version>/alkaid_<version>_linux_<arch>.tar.gz
  alioth_grpc_addr: 10.0.0.10:9000
  timeout: 30s
  online_timeout: 60s
```

直接用裸二进制运行 `alioth`（不在镜像里）时没有内置包，需要配置
`alkaid_package_path`（本机 `make package-alkaid` 可生成）或指向可访问的下载地址。

控制面默认不校验目标机的主机密钥（与 `alkaid` 的 `ALKAID_SSH_KNOWN_HOSTS` 默认行为一致），
安装动作只接受持有 `agent:write` 的租户成员。安装会话保存在 `alioth` 进程内存里，
多副本部署时要让轮询请求落到发起请求的同一副本（使用会话保持或单副本入口）。

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

