#!/usr/bin/env bash
# 摇光（alkaid）二进制、systemd 一键部署包与 RPM。
# 交叉编译 linux 二进制，连同 systemd 单元、配置模板和 install.sh 打成 tar.gz；
# 有 rpmbuild 时同时打 x86_64 / aarch64 的 RPM（deploy/alkaid/alkaid.spec），
# 在 OUT 下保留可直接下载的裸二进制，最后生成 SHA256SUMS。
#
# 本地与 release.yml 共用同一条命令：
#   make package-alkaid VERSION=v0.1.0
#   ARCHES=arm64 make package-alkaid        # 只打一个架构，加快本地验证
#   RPM=1 make package-alkaid               # 必须出 RPM，缺 rpmbuild 直接失败
#   RPM=0 make package-alkaid               # 只出 tar.gz 和裸二进制
#
# rpmbuild 来源：Debian/Ubuntu 装 `rpm`，RHEL/openEuler 装 `rpm-build`。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-dev}"
ARCHES="${ARCHES:-amd64 arm64}"
GO="${GO:-go}"
OUT="${OUT:-${ROOT}/dist}"
RPM="${RPM:-auto}"

case "${RPM}" in
  1|0) ;;
  auto)
    if command -v rpmbuild >/dev/null 2>&1; then
      RPM=1
    else
      RPM=0
      echo "未找到 rpmbuild：只打 tar.gz 和裸二进制；需要 RPM 时先装 rpm / rpm-build 再跑 RPM=1" >&2
    fi
    ;;
  *)
    echo "RPM 只能是 1、0 或 auto（当前：${RPM}）" >&2
    exit 1
    ;;
esac
if [ "${RPM}" = 1 ] && ! command -v rpmbuild >/dev/null 2>&1; then
  echo "RPM=1 但找不到 rpmbuild：Debian/Ubuntu 装 rpm 包，RHEL/openEuler 装 rpm-build" >&2
  exit 1
fi

# RPM 的 Version / Release 不允许连字符：v1.2.3 拆成 1.2.3 + 1，预发布
# v1.2.3-rc1 变成 1.2.3 + 0.rc1（排序在正式版之前）；sha-xxxxx、dev 这类
# 非语义版本落到 0.0.0 + 清洗后的标签。tar.gz 和裸二进制仍用原始 VERSION。
rpm_version=""
rpm_release=""
if [ "${RPM}" = 1 ]; then
  raw="${VERSION#v}"
  if [[ "${raw}" =~ ^([0-9]+(\.[0-9]+)*)$ ]]; then
    rpm_version="${raw}"
    rpm_release=1
  elif [[ "${raw}" =~ ^([0-9]+(\.[0-9]+)*)-(.+)$ ]]; then
    rpm_version="${BASH_REMATCH[1]}"
    rpm_release="0.${BASH_REMATCH[3]//-/.}"
  else
    rpm_version=0.0.0
    rpm_release="${raw//-/.}"
  fi
  # 只保留 RPM 认的字符，其它（空格、斜杠、中文等）一律换成点。
  rpm_version="$(printf '%s' "${rpm_version}" | LC_ALL=C tr -c '[:alnum:]._+~' '.')"
  rpm_release="$(printf '%s' "${rpm_release}" | LC_ALL=C tr -c '[:alnum:]._+~' '.')"
fi

temp_dirs=()
cleanup() {
  local dir
  for dir in "${temp_dirs[@]}"; do
    case "${dir}" in
      ""|/) continue ;;
    esac
    rm -rf "${dir}"
  done
}
trap cleanup EXIT

# 打一个架构的 RPM：把已经编译好并装进 stage 的文件塞进 rpmbuild 的 SOURCES，
# 由 spec 负责安装路径、单元和配置。成功时把 rpm 文件名写进全局 rpm_file。
rpm_file=""
build_rpm() {
  local stage="$1" arch="$2" rpm_arch top log meta want
  case "${arch}" in
    amd64) rpm_arch=x86_64 ;;
    arm64) rpm_arch=aarch64 ;;
    *)
      echo "未知架构：${arch}（RPM 只支持 amd64、arm64）" >&2
      exit 1
      ;;
  esac

  rpm_file="alkaid-${rpm_version}-${rpm_release}.${rpm_arch}.rpm"
  top="$(mktemp -d)"
  temp_dirs+=("${top}")
  mkdir -p "${top}/BUILD" "${top}/BUILDROOT" "${top}/RPMS" "${top}/SOURCES" "${top}/SPECS" "${top}/SRPMS"
  install -m 0755 "${stage}/alkaid" "${top}/SOURCES/alkaid"
  install -m 0644 "${ROOT}/deploy/alkaid/alkaid.service" "${top}/SOURCES/alkaid.service"
  install -m 0644 "${ROOT}/deploy/alkaid/alkaid.yaml" "${top}/SOURCES/alkaid.yaml"
  install -m 0600 "${ROOT}/deploy/alkaid/alkaid.env.example" "${top}/SOURCES/alkaid.env.example"
  install -m 0644 "${ROOT}/deploy/alkaid/README.md" "${top}/SOURCES/README.md"
  install -m 0644 "${ROOT}/deploy/alkaid/alkaid.spec" "${top}/SPECS/alkaid.spec"

  echo "打包 ${rpm_file}"
  log="${top}/rpmbuild.log"
  # 目标写全 arch-linux：在 macOS 上 rpmbuild 会把 host 的 os（darwin）带进包，装不上 Linux。
  if ! rpmbuild -bb --target "${rpm_arch}-linux" \
    --define "_topdir ${top}" \
    --define "alkaid_version ${rpm_version}" \
    --define "alkaid_release ${rpm_release}" \
    "${top}/SPECS/alkaid.spec" >"${log}" 2>&1; then
    echo "rpmbuild 失败（${rpm_file}），日志：" >&2
    cat "${log}" >&2
    exit 1
  fi
  install -m 0644 "${top}/RPMS/${rpm_arch}/${rpm_file}" "${OUT}/"

  # 元数据和载荷都要对得上，避免装出一个空包或装错架构。
  meta="$(rpm -qp --qf '%{NAME}-%{VERSION}-%{RELEASE}.%{ARCH}.rpm' "${OUT}/${rpm_file}")"
  if [ "${meta}" != "${rpm_file}" ]; then
    echo "RPM 元数据与文件名不一致：${meta} ≠ ${rpm_file}" >&2
    exit 1
  fi
  for want in /usr/local/bin/alkaid /usr/lib/systemd/system/alkaid.service /etc/alioth/alkaid.yaml /etc/alioth/alkaid.env; do
    if ! rpm -qpl "${OUT}/${rpm_file}" | grep -qx "${want}"; then
      echo "${rpm_file} 里缺少 ${want}" >&2
      exit 1
    fi
  done
}

mkdir -p "${OUT}"
OUT="$(cd "${OUT}" && pwd)"

files=()
stage_root="${OUT}/.stage"
temp_dirs+=("${stage_root}")
for arch in ${ARCHES}; do
  name="alkaid_${VERSION}_linux_${arch}"
  stage="${stage_root}/${name}"
  echo "打包 ${name}"
  rm -rf "${stage}" "${OUT}/${name}.tar.gz" "${OUT}/${name}"
  mkdir -p "${stage}"

  # trimpath + -s -w：二进制里不带构建机路径和调试符号；版本戳走 -ldflags。
  CGO_ENABLED=0 GOOS=linux GOARCH="${arch}" "${GO}" build -C "${ROOT}/app/alkaid" \
    -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${stage}/alkaid" ./cmd/alkaid

  install -m 0644 "${ROOT}/deploy/alkaid/alkaid.service" "${stage}/alkaid.service"
  install -m 0644 "${ROOT}/deploy/alkaid/alkaid.yaml" "${stage}/alkaid.yaml"
  install -m 0644 "${ROOT}/deploy/alkaid/alkaid.env.example" "${stage}/alkaid.env.example"
  install -m 0755 "${ROOT}/deploy/alkaid/install.sh" "${stage}/install.sh"
  install -m 0644 "${ROOT}/deploy/alkaid/README.md" "${stage}/README.md"
  # 版本文件供 install.sh 展示用：跨架构解包时也能说清装的是哪个版本，不必执行二进制。
  printf '%s\n' "${VERSION}" > "${stage}/VERSION"

  # 只想要二进制时可以直接下载裸文件；systemd 单元和 install.sh 在 tar.gz 里。
  install -m 0755 "${stage}/alkaid" "${OUT}/${name}"

  tar -C "${stage_root}" -czf "${OUT}/${name}.tar.gz" "${name}"
  files+=("${name}" "${name}.tar.gz")

  if [ "${RPM}" = 1 ]; then
    build_rpm "${stage}" "${arch}"
    files+=("${rpm_file}")
  fi
  rm -rf "${stage}"
done
rm -rf "${stage_root}"

if [ "${#files[@]}" -eq 0 ]; then
  echo "ARCHES 为空，没有可打包的架构" >&2
  exit 1
fi

# 校验文件放在归档外面，发布页可以直接下载核对。
(
  cd "${OUT}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${files[@]}" > SHA256SUMS
  else
    shasum -a 256 "${files[@]}" > SHA256SUMS
  fi
)
cat "${OUT}/SHA256SUMS"
