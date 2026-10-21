# syntax=docker/dockerfile:1
# 构建上下文是仓库根目录：megrez 依赖同仓库的 api 与 pkg 模块（go.mod replace）。
ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
ARG VERSION=dev
COPY api/ api/
COPY pkg/ pkg/
COPY app/megrez/ app/megrez/
RUN go build -C app/megrez -trimpath \
	-ldflags "-s -w -X main.version=${VERSION}" \
	-o /out/megrez ./cmd/megrez

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/megrez /app/megrez
COPY app/megrez/configs/ /app/configs/
USER nonroot:nonroot
EXPOSE 8001
ENTRYPOINT ["/app/megrez"]
CMD ["serve", "--conf", "/app/configs/config.yaml"]
