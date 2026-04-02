FROM golang:1.25.3 AS builder

# 安装必要的构建工具
RUN apt-get update && apt-get install -y \
    protobuf-compiler

WORKDIR /moon

# 复制构建文件
COPY Makefile Makefile

# 初始化环境
RUN make init

# 复制源代码
COPY . .

# 构建应用
RUN make build

# 最终运行阶段 - 使用 Alpine
FROM debian:stable-slim

# 安装运行时依赖
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

# 复制二进制文件
COPY --from=builder /moon/bin/marksman /usr/local/bin/marksman

# 设置可执行权限
RUN chmod +x /usr/local/bin/marksman

# 暴露端口（根据实际配置调整）
EXPOSE 7003 8003 9003 

# 运行应用
ENTRYPOINT ["/usr/local/bin/marksman"]
CMD ["run", "all"]