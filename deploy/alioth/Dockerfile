# syntax=docker/dockerfile:1
# 构建上下文是仓库根目录：alioth 依赖同仓库的 api 与 pkg 模块（go.mod replace）。
ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
ARG VERSION=dev
COPY api/ api/
COPY pkg/ pkg/
COPY app/alioth/ app/alioth/
RUN go build -C app/alioth -trimpath \
	-ldflags "-s -w -X main.version=${VERSION}" \
	-o /out/alioth ./cmd/alioth

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/alioth /app/alioth
COPY app/alioth/configs/ /app/configs/
USER nonroot:nonroot
EXPOSE 8000 9000
ENTRYPOINT ["/app/alioth"]
CMD ["serve", "--conf", "/app/configs/config.yaml"]
