FROM golang:1.25.3 AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /moon

COPY magicbox/ magicbox/
COPY proto/ proto/
COPY app/jade_tree/ app/jade_tree/

WORKDIR /moon/app/jade_tree

ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG REPO=https://github.com/aide-family/moon
RUN make build-image VERSION="$VERSION" BUILD_TIME="$BUILD_TIME" REPO="$REPO"

FROM debian:stable-slim

RUN apt-get update -y && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    netbase \
    && apt-get update -y \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get autoremove -y && apt-get autoclean -y

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /moon

COPY --from=builder /moon/app/jade_tree/bin/jade_tree /usr/local/bin/jade_tree

RUN chmod +x /usr/local/bin/jade_tree

EXPOSE 8004 9004

ENTRYPOINT ["/usr/local/bin/jade_tree"]
CMD ["run", "all"]
