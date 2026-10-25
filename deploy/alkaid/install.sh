#!/bin/sh
# 摇光（alkaid）一键安装：二进制 + 配置 + systemd 单元。
# 在解压后的部署包目录里用 root 执行：
#   sudo ./install.sh
#
# 可选参数：
#   --no-start   只安装并 enable，不立即启动服务
#   -h, --help   显示帮助
# 可选环境变量（打包与调试用）：
#   PREFIX=/usr/local  CONF_DIR=/etc/alioth  UNIT_PATH=/etc/systemd/system/alkaid.service
#   DESTDIR=/tmp/rootfs  只准备文件，不调用 systemctl
set -eu

SRC_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
PREFIX="${PREFIX:-/usr/local}"
CONF_DIR="${CONF_DIR:-/etc/alioth}"
UNIT_PATH="${UNIT_PATH:-/etc/systemd/system/alkaid.service}"
DESTDIR="${DESTDIR:-}"
START=1

usage() {
  cat <<'EOF'
用法：sudo ./install.sh [--no-start]

把 alkaid 安装到 PREFIX/bin，配置写到 CONF_DIR，systemd 单元写到 UNIT_PATH，
然后 daemon-reload 并 enable。ALKAID_AGENT_TOKEN 还是占位值时不会启动服务。

  --no-start   安装并 enable，但不立即启动
  -h, --help   显示本帮助

装好后编辑 CONF_DIR/alkaid.env，填入控制台签发的 ALKAID_AGENT_TOKEN，
再执行 systemctl restart alkaid。
EOF
}

for arg in "$@"; do
  case "${arg}" in
    --no-start) START=0 ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数：${arg}" >&2
      usage >&2
      exit 2
      ;;
  esac
done

die() {
  echo "错误：$*" >&2
  exit 1
}

[ -x "${SRC_DIR}/alkaid" ] || die "找不到 ${SRC_DIR}/alkaid，请在解压后的部署包目录里执行"

if [ -z "${DESTDIR}" ] && [ "$(id -u)" -ne 0 ]; then
  die "需要 root 权限：sudo ./install.sh"
fi

BIN_PATH="${DESTDIR}${PREFIX}/bin/alkaid"
mkdir -p "${DESTDIR}${PREFIX}/bin" "${DESTDIR}${CONF_DIR}" "$(dirname -- "${DESTDIR}${UNIT_PATH}")"

version="$(cat "${SRC_DIR}/VERSION" 2>/dev/null || true)"
[ -n "${version}" ] || version="dev"
echo "安装 alkaid ${version} → ${PREFIX}/bin/alkaid"
install -m 0755 "${SRC_DIR}/alkaid" "${BIN_PATH}"

# 配置和令牌只在缺失时写入：升级包不能覆盖现场修改。
if [ -f "${DESTDIR}${CONF_DIR}/alkaid.yaml" ]; then
  echo "保留已有配置 ${CONF_DIR}/alkaid.yaml"
else
  install -m 0644 "${SRC_DIR}/alkaid.yaml" "${DESTDIR}${CONF_DIR}/alkaid.yaml"
  echo "写入配置 ${CONF_DIR}/alkaid.yaml"
fi

ENV_FILE="${DESTDIR}${CONF_DIR}/alkaid.env"
if [ -f "${ENV_FILE}" ]; then
  echo "保留已有 ${CONF_DIR}/alkaid.env"
else
  install -m 0600 "${SRC_DIR}/alkaid.env.example" "${ENV_FILE}"
  echo "写入 ${CONF_DIR}/alkaid.env（0600，令牌是占位值）"
fi

install -m 0644 "${SRC_DIR}/alkaid.service" "${DESTDIR}${UNIT_PATH}"
echo "写入 systemd 单元 ${UNIT_PATH}"

if [ -n "${DESTDIR}" ]; then
  echo "DESTDIR=${DESTDIR}：跳过 systemctl，文件已就绪"
  exit 0
fi

systemctl daemon-reload
systemctl enable alkaid

# 令牌为空或还是占位值时不要启动：agent 会一直被控制面拒绝，日志刷屏。
token="$(sed -n 's/^ALKAID_AGENT_TOKEN=//p' "${ENV_FILE}" | tail -n 1)"
case "${token}" in
  ""|*replace-me*)
    echo
    echo "已 enable，但 ${CONF_DIR}/alkaid.env 里的 ALKAID_AGENT_TOKEN 还没填。"
    echo "填好后执行：systemctl restart alkaid"
    exit 0
    ;;
esac

if [ "${START}" -eq 1 ]; then
  systemctl restart alkaid
  echo "已启动。状态：systemctl status alkaid"
else
  echo "已 enable。按需执行：systemctl start alkaid"
fi

echo "健康检查：curl -fsS http://127.0.0.1:8022/health"
