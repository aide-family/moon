FROM golang:1.25.3 AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

RUN apt-get update && apt-get install -y \
    protobuf-compiler

WORKDIR /moon/app/goddess

COPY app/goddess/Makefile ./Makefile

RUN make init

COPY app/goddess/go.mod ./go.mod
COPY app/goddess/go.sum ./go.sum
COPY app/goddess/ ./

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

COPY --from=builder /moon/app/goddess/bin/goddess /usr/local/bin/goddess

RUN chmod +x /usr/local/bin/goddess

EXPOSE 7000 8000 9000

ENTRYPOINT ["/usr/local/bin/goddess"]
CMD ["run", "all"]
