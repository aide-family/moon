#!/usr/bin/env bash
# 摇光（alkaid）二进制与 systemd 一键部署包。
# 交叉编译 linux 二进制，连同 systemd 单元、配置模板和 install.sh 打成 tar.gz，
# 同时在 OUT 下保留可直接下载的裸二进制，并生成 SHA256SUMS。
#
# 本地与 release.yml 共用同一条命令：
#   make package-alkaid VERSION=v0.1.0
#   ARCHES=arm64 make package-alkaid        # 只打一个架构，加快本地验证
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-dev}"
ARCHES="${ARCHES:-amd64 arm64}"
GO="${GO:-go}"
OUT="${OUT:-${ROOT}/dist}"

mkdir -p "${OUT}"
OUT="$(cd "${OUT}" && pwd)"

files=()
stage_root="${OUT}/.stage"
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
  rm -rf "${stage}"
  files+=("${name}" "${name}.tar.gz")
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
