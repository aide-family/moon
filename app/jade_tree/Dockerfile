FROM golang:1.25.3 AS builder

# Use domestic Go module proxy during image build to avoid timeouts.
ENV GOPROXY=https://goproxy.cn,direct
# Disable sumdb lookup to make the build more robust in restricted networks.
ENV GOSUMDB=off

# 安装必要的构建工具
RUN apt-get update && apt-get install -y \
    protobuf-compiler

WORKDIR /moon/app/jade_tree

# Copy build files first (to leverage Docker layer cache).
COPY app/jade_tree/Makefile ./Makefile

# 初始化环境
RUN make init

# Copy Go module manifest and sources.
COPY app/jade_tree/go.mod ./go.mod
COPY app/jade_tree/go.sum ./go.sum
COPY app/jade_tree/ ./

# Copy shared local modules and proto files required by make api/conf.
COPY magicbox/ ../../magicbox/
COPY proto/ ../../proto/

# Build application (go build / proto generation).
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

# Copy binary
COPY --from=builder /moon/app/jade_tree/bin/jade_tree /usr/local/bin/jade_tree

# Ensure executable permission
RUN chmod +x /usr/local/bin/jade_tree

# Expose ports (http=8004, grpc=9004)
EXPOSE 8004 9004

# Run application
ENTRYPOINT ["/usr/local/bin/jade_tree"]
CMD ["run", "all"]
