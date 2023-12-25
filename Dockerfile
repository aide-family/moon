FROM golang:1.21.0 AS builder

COPY . /src
WORKDIR /src

# protocol buffer的头文件还有动态库都会放在/usr/local下
RUN cd protobuf-3.11.2 && ./configure -prefix=/usr/local/ && make && make install
RUN protoc --version

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

CMD ["./prom_server", "-conf", "/data/conf"]
