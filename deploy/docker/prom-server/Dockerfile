FROM golang:1.21.0 AS serverBuilder
ARG VERSION=v0.0.0

WORKDIR /src
COPY ../../../go.mod ../../../go.sum .
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY ../../../.. /src
RUN mkdir -p bin/ && go build -ldflags "-X main.Version=$VERSION" -o ./bin/ ./app/prom_server/...

# 使用官方Node.js镜像作为基础镜像
FROM node:latest as webBuilder

# 设置工作目录
WORKDIR /usr/src/app

# 将前端应用程序代码复制到容器中
COPY ../../../web .

RUN rm -rf node_modules

# 设置npm源
RUN npm config set registry https://registry.npmmirror.com/

# 安装依赖
RUN npm install

# 编译前端应用程序
RUN npm run build-docker

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        nginx \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

# /usr/share/nginx/html
COPY --from=webBuilder /usr/src/app/dist /usr/share/nginx/html
COPY --from=webBuilder /usr/src/app/default.conf /etc/nginx/conf.d/default.conf
COPY --from=webBuilder /usr/src/app/nginx.conf /etc/nginx/nginx.conf
COPY --from=serverBuilder /src/bin /app
COPY --from=serverBuilder /src/app/prom_server/configs /data/conf

WORKDIR /app

EXPOSE 8000
EXPOSE 8888
EXPOSE 9000
EXPOSE 80
VOLUME /data/conf
VOLUME /app/log
VOLUME /app/cache
VOLUME /app/db
VOLUME /etc/nginx/conf.d

# TZ=Asia/Shanghai
ENV TZ=Asia/Shanghai

#CMD ["nginx", "", "./prom_server", "-conf", "/data/conf"]
# 同时执行nginx 和 prom_server
CMD ["bash", "-c", "nginx && ./prom_server -conf /data/conf"]
