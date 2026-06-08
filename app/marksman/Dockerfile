FROM golang:1.25.3 AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

RUN apt-get update && apt-get install -y \
    protobuf-compiler

WORKDIR /moon/app/marksman

COPY app/marksman/Makefile ./Makefile

RUN make init-build

COPY app/marksman/go.mod ./go.mod
COPY app/marksman/go.sum ./go.sum
COPY app/marksman/ ./

COPY magicbox/ ../../magicbox/
COPY proto/ ../../proto/

RUN make build

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
