

FROM golang:1.21.0 AS builder

COPY . /src
WORKDIR /src

ARG APP_NAME

RUN echo "APP_NAME=${APP_NAME}"

RUN GOPROXY=https://goproxy.cn make build

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
ENV APP=${APP_NAME}

CMD ["${APP}", "-conf", "/data/conf"]
