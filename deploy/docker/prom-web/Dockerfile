# 使用官方Node.js镜像作为基础镜像
FROM node:latest as builder

# 设置工作目录
WORKDIR /usr/src/app

# 将前端应用程序代码复制到容器中
COPY ../../../web .

RUN rm -rf node_modules

# 设置npm源
RUN npm config set registry https://mirrors.cloud.tencent.com/npm/

# 安装依赖
RUN npm install

# 编译前端应用程序
RUN npm run build-docker

# 部署到nginx最小系统
FROM nginx:stable-perl

# /usr/share/nginx/html
COPY --from=builder /usr/src/app/dist /usr/share/nginx/html

EXPOSE 80