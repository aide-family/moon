FROM golang:1.25.3 AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

RUN apt-get update && apt-get install -y \
    protobuf-compiler

WORKDIR /moon/app/rabbit

COPY app/rabbit/Makefile ./Makefile

RUN make init-build

COPY app/rabbit/go.mod ./go.mod
COPY app/rabbit/go.sum ./go.sum
COPY app/rabbit/ ./

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

COPY --from=builder /moon/app/rabbit/bin/rabbit /usr/local/bin/rabbit

RUN chmod +x /usr/local/bin/rabbit

EXPOSE 7001 8001 9001

ENTRYPOINT ["/usr/local/bin/rabbit"]
CMD ["run", "all"]
