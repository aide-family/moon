#!/usr/bin/env bash
# 准备 docker-compose 依赖镜像。
# Docker Hub 不可达时，按顺序尝试镜像代理；拿到镜像后重新 tag 成 compose 使用的官方名字。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# target|repository|tag
images=(
  "postgres:16|library/postgres|16"
  "redis:7|library/redis|7"
  "axllent/mailpit:latest|axllent/mailpit|latest"
  "prom/prometheus:v3.5.0|prom/prometheus|v3.5.0"
  "grafana/loki:3.5.0|grafana/loki|3.5.0"
  "curlimages/curl:8.12.1|curlimages/curl|8.12.1"
)

# 可用 MIRROR 覆盖，例如：MIRROR=dockerproxy.net make compose-images
if [[ -n "${MIRROR:-}" ]]; then
  mirrors=("${MIRROR}")
else
  mirrors=(dockerproxy.net docker.m.daocloud.io)
fi

exists() {
  docker image inspect "$1" >/dev/null 2>&1
}

for entry in "${images[@]}"; do
  IFS='|' read -r target repo tag <<<"${entry}"
  if exists "${target}"; then
    echo "image ${target} already present"
    continue
  fi

  pulled=""
  for mirror in "${mirrors[@]}"; do
    candidate="${mirror}/${repo}:${tag}"
    if exists "${candidate}"; then
      echo "using local mirror image ${candidate}"
      pulled="${candidate}"
      break
    fi
    echo "pulling ${candidate}"
    if docker pull "${candidate}"; then
      pulled="${candidate}"
      break
    fi
  done

  if [[ -z "${pulled}" ]]; then
    echo "failed to obtain ${target} from configured mirrors: ${mirrors[*]}" >&2
    echo "set MIRROR=your.registry to use another mirror" >&2
    exit 1
  fi

  docker tag "${pulled}" "${target}"
  echo "tagged ${pulled} -> ${target}"
done

docker image inspect postgres:16 redis:7 axllent/mailpit:latest prom/prometheus:v3.5.0 grafana/loki:3.5.0 curlimages/curl:8.12.1 \
  --format '{{index .RepoTags 0}} arch={{.Architecture}}/{{.Os}}'
echo "compose images are ready: ${ROOT}/docker-compose.yaml"
