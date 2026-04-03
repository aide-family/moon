%{!?jade_tree_version:%global jade_tree_version 0.0.1}
%{!?jade_tree_release:%global jade_tree_release 1}

Name:           jade-tree
Version:        %{jade_tree_version}
Release:        %{jade_tree_release}%{?dist}
Summary:        Moon agent runtime service

License:        Apache-2.0
URL:            https://github.com/aide-family/moon
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang
BuildRequires:  git
BuildRequires:  systemd-rpm-macros
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%global debug_package %{nil}

%description
Jade Tree is the Moon agent runtime service. It provides machine profile
collection/reporting and probe task capabilities.

%prep
%setup -q -n %{name}-%{version}

%build
export CGO_ENABLED=0
export GOPROXY=${GOPROXY:-https://goproxy.cn,direct}
export GOSUMDB=${GOSUMDB:-sum.golang.org}
go build -trimpath -ldflags="-s -w" -o jade_tree ./main.go

%install
install -d -m0755 %{buildroot}/opt/jade-tree/bin
install -d -m0755 %{buildroot}/opt/jade-tree/config
install -m0755 jade_tree %{buildroot}/opt/jade-tree/bin/jade_tree
install -m0644 packaging/rpm/server.yaml.default %{buildroot}/opt/jade-tree/config/server.yaml
install -D -m0644 packaging/rpm/jade-tree.service %{buildroot}%{_unitdir}/jade-tree.service

%post
%systemd_post jade-tree.service

%preun
%systemd_preun jade-tree.service

%postun
%systemd_postun jade-tree.service

%files
/opt/jade-tree/bin/jade_tree
%config(noreplace) /opt/jade-tree/config/server.yaml
%{_unitdir}/jade-tree.service

%changelog
* Fri Apr 03 2026 Moon maintainers <moon@localhost> - 0.0.1-1
- Add initial RPM packaging for jade_tree
