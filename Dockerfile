#FROM golang:1.21.0 AS builder
FROM aidemoonio/build:latest AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build

#FROM debian:stable-slim
FROM aidemoonio/deploy:latest

WORKDIR /app

ARG CMD_PARAMS="palace"
ENV CMD_PARAMS_ENV=${CMD_PARAMS}
# 复制脚本到容器中
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=builder /src/bin/${CMD_PARAMS} /app/${CMD_PARAMS}

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf
# 设置 ENTRYPOINT
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["sh"]
