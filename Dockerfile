#FROM golang:1.21.0 AS builder
FROM aidemoonio/build:latest AS builder

COPY . /src
WORKDIR /src

RUN make build

#FROM debian:stable-slim
FROM aidemoonio/deploy:latest

WORKDIR /app

ARG CMD_PARAMS="palace"
ENV CMD_PARAMS_ENV=${CMD_PARAMS}
ENV CONFIG_TYPE="yaml"
# 复制脚本到容器中
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=builder /src/bin/${CMD_PARAMS} /app/${CMD_PARAMS}
COPY --from=builder /src/third_party/swagger_ui /app/third_party/swagger_ui/

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf
# 设置 ENTRYPOINT
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["sh"]
