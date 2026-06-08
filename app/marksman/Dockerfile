FROM golang:1.25.3 AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV CGO_ENABLED=0

WORKDIR /moon

COPY magicbox/ magicbox/
COPY proto/ proto/
COPY app/goddess/ app/goddess/
COPY app/rabbit/ app/rabbit/
COPY app/marksman/ app/marksman/

WORKDIR /moon/app/marksman

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

COPY --from=builder /moon/app/marksman/bin/marksman /usr/local/bin/marksman

RUN chmod +x /usr/local/bin/marksman

EXPOSE 7003 8003 9003

ENTRYPOINT ["/usr/local/bin/marksman"]
CMD ["run", "all"]
