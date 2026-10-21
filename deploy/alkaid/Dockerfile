# syntax=docker/dockerfile:1
# 构建上下文是仓库根目录：镜像内直接编译 alkaid，并附带 sshd（SFTP 需要）。
ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
ARG VERSION=dev
COPY api/ api/
COPY pkg/ pkg/
COPY app/alkaid/ app/alkaid/
RUN go build -C app/alkaid -trimpath \
	-ldflags "-s -w -X main.version=${VERSION}" \
	-o /out/alkaid ./cmd/alkaid

FROM public.ecr.aws/docker/library/ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
	&& apt-get install -y --no-install-recommends openssh-server ca-certificates tzdata curl \
	&& rm -rf /var/lib/apt/lists/* \
	&& mkdir -p /run/sshd /etc/alioth

COPY --from=build /out/alkaid /usr/local/bin/alkaid
COPY deploy/alkaid/entrypoint.sh /entrypoint.sh
COPY deploy/alkaid/alkaid.yaml /etc/alioth/alkaid.yaml
RUN chmod 755 /entrypoint.sh

EXPOSE 22 8022 9022

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
	CMD curl -fsS http://127.0.0.1:8022/health || exit 1

ENTRYPOINT ["/entrypoint.sh"]
