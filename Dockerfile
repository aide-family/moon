ARG APP_NAME

FROM golang:1.21.0 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build
RUN cat /src/bin/${APP_NAME} > /src/bin/server
RUN chmod +x /src/bin/server

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
