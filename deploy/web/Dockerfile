# syntax=docker/dockerfile:1
# web 镜像同时提供静态资源与反向代理；nginx 配置只有 web/nginx.conf 一个来源。
ARG NODE_VERSION=22
ARG NGINX_VERSION=1.27

# 产物只有 dist/ 里的静态文件，与 CPU 架构无关：构建阶段固定在 buildx 的原生平台（BUILDPLATFORM）
# 上跑，多架构发布时 amd64 与 arm64 共用同一份构建层，arm64 不再在 QEMU 里重复等 npm ci。
FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-alpine AS build
WORKDIR /app
# 锁文件里的 tarball 走官方源（GitHub Runner 直连最快）；国内构建可传
# --build-arg NPM_REGISTRY=https://registry.npmmirror.com 覆盖。
ARG NPM_REGISTRY=https://registry.npmjs.org
COPY web/package.json web/package-lock.json ./
# npm 缓存挂到 BuildKit cache mount，重试/重构建不用重下 628 个 tarball；
# fetch-retries 兜住网络抖动，别让一次 ETIMEDOUT 直接废掉整个发布。
RUN --mount=type=cache,target=/root/.npm \
    npm ci --registry="${NPM_REGISTRY}" --no-audit --no-fund \
      --fetch-retries=5 --fetch-retry-mintimeout=10000 --fetch-retry-maxtimeout=60000
COPY web/ ./
RUN npm run build

FROM nginx:${NGINX_VERSION}-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY web/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 8080
