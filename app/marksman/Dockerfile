# 多阶段构建 - 构建阶段
FROM golang:1.25.3-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache \
    protobuf-dev \
    protoc \
    git \
    make

WORKDIR /moon

# 复制构建文件
COPY Makefile Makefile

# 初始化环境
RUN make init

# 复制源代码
COPY . .

# RUN git clone https://github.com/aide-family/magicbox.git ../magicbox
# RUN git clone https://github.com/aide-family/kratos.git ../kratos

# 构建应用
RUN make build

# 最终运行阶段 - 使用 Alpine
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

WORKDIR /moon

# 复制二进制文件
COPY --from=builder /moon/bin/marksman /usr/local/bin/marksman

# 设置可执行权限
RUN chmod +x /usr/local/bin/marksman

# 暴露端口（根据实际配置调整）
EXPOSE 10070 18080 19090 

# 运行应用
ENTRYPOINT ["/usr/local/bin/marksman"]
CMD ["run", "all"]