# 摇光（alkaid）RPM 包：二进制由 scripts/package-alkaid.sh 交叉编译后放进 rpmbuild
# 的 SOURCES，这里只负责装文件、装 systemd 单元和保留现场配置。
# 版本号由打包脚本用 --define 传入；直接 rpmbuild 时退回 0.0.0-1，方便本地试打。
#
# 本地打 RPM：
#   RPM=1 make package-alkaid VERSION=v0.1.0 ARCHES=amd64
%{!?alkaid_version:%global alkaid_version 0.0.0}
%{!?alkaid_release:%global alkaid_release 1}

Name:           alkaid
Version:        %{alkaid_version}
Release:        %{alkaid_release}
Summary:        Alioth edge agent (alkaid)
License:        Proprietary
URL:            https://github.com/aide-family/moon
Source0:        alkaid
Source1:        alkaid.service
Source2:        alkaid.yaml
Source3:        alkaid.env.example
Source4:        README.md

# 回连控制面的 TLS 校验要系统 CA；单元文件由 systemd 提供。
Requires:       ca-certificates
Requires:       systemd

# 二进制在打包机上已经 -s -w 编译好并 strip 过。这里不编译，也不让不同版本
# rpmbuild 的 brp 脚本（strip、build-id 链接、debuginfo）再处理一遍：不关掉
# 的话 rpmbuild 会往包里塞 /usr/lib/.build-id 符号链接，本地和 CI 的产物也会不一致。
%global debug_package %{nil}
%global __strip /bin/true
%global __os_install_post %{nil}
%global _build_id_links none

%description
Alioth（瑶光）平台的摇光边缘 Agent：只从本机向 alioth 控制面建立 gRPC 双向流，
在本机执行命令、提供 SSH 会话与健康检查，不要求控制面回拨。

安装后编辑 /etc/alioth/alkaid.env，填入控制台签发的 ALKAID_AGENT_TOKEN，
再执行 systemctl restart alkaid。

%prep
%build

%install
# 不写 install -D：macOS 的 BSD install 没有 -D，本地 rpmbuild 也要能跑。
mkdir -p %{buildroot}/usr/local/bin %{buildroot}/usr/lib/systemd/system \
	%{buildroot}/etc/alioth %{buildroot}/usr/share/doc/alkaid
install -m 0755 %{SOURCE0} %{buildroot}/usr/local/bin/alkaid
install -m 0644 %{SOURCE1} %{buildroot}/usr/lib/systemd/system/alkaid.service
install -m 0644 %{SOURCE2} %{buildroot}/etc/alioth/alkaid.yaml
install -m 0600 %{SOURCE3} %{buildroot}/etc/alioth/alkaid.env
install -m 0644 %{SOURCE4} %{buildroot}/usr/share/doc/alkaid/README.md

%post
# 只 enable 不启动：令牌还是占位值时启动会一直被控制面拒绝、刷错误日志。
if [ -d /run/systemd/system ]; then
	systemctl daemon-reload >/dev/null 2>&1 || :
	systemctl enable alkaid >/dev/null 2>&1 || :
fi

%preun
# 升级（$1 = 1）时不动服务；卸载（$1 = 0）时先停再 disable。
if [ "$1" -eq 0 ] && [ -d /run/systemd/system ]; then
	systemctl disable --now alkaid >/dev/null 2>&1 || :
fi

%postun
if [ "$1" -ge 1 ] && [ -d /run/systemd/system ]; then
	systemctl daemon-reload >/dev/null 2>&1 || :
fi

%files
%dir /etc/alioth
%config(noreplace) /etc/alioth/alkaid.yaml
%attr(0600, root, root) %config(noreplace) /etc/alioth/alkaid.env
/usr/lib/systemd/system/alkaid.service
/usr/local/bin/alkaid
/usr/share/doc/alkaid/README.md

%changelog
* Thu Oct 29 2026 Alioth contributors <noreply@example.com> - 0.0.0-1
- 提供摇光（alkaid）RPM 包。
