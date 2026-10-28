GO ?= go
ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BIN := $(ROOT)bin
GOBIN := $(shell $(GO) env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)
PROTO_INC ?= /opt/homebrew/include

# 根目录 .env 与 docker compose 共用同一份变量：include 让本地开发值覆盖下面的默认值，
# 再按 .env 里出现的键导出给各命令（命令行参数仍然优先，例如 make images VERSION=v1）。
ifneq (,$(wildcard $(ROOT).env))
include $(ROOT).env
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' $(ROOT).env)
endif

# 镜像版本戳：注入 main.version 并作为镜像 tag，避免所有部署都用 :latest。
VERSION ?= $(shell git -C $(ROOT) describe --tags --always --dirty 2>/dev/null || echo dev)
IMAGE_PREFIX ?= alioth
# 摇光部署包的目标架构；本地验证时可只跑一个，例如 ARCHES=arm64。
ARCHES ?= amd64 arm64
# 摇光 RPM 开关：auto 表示装了 rpmbuild 就打，1 强制打（缺 rpmbuild 直接失败），0 只打 tar.gz。
RPM ?= auto

.PHONY: tools api conf vet lint build build-alioth build-megrez build-merak build-alkaid build-alkaid-linux build-web \
	image-alioth image-megrez image-merak image-alkaid image-web images package-alkaid \
	run-alioth run-megrez run-merak run-alkaid run-web \
	test test-pkg test-race migrate migrate-in-docker compose-images compose-up compose-down compose-alkaid dev-up dev-down loki-demo loki-demo-stop vlogs-demo vlogs-demo-stop tidy wire

# 版本戳注入：$(call ldflags) 用于本地构建，Dockerfile 内用 --build-arg VERSION 注入。
version_ldflags = -X main.version=$(VERSION)

# Kratos v3 CLI and HTTP/error generators. Pin a known-good pseudo-version so PATH
# does not keep a v2 protoc-gen-go-http. CLI install uses module cache + go build
# because go install @latest fails when the module has no tagged release yet.
KRATOS_CMD_VER ?= v3.0.0-20260626125723-668db92c2c00

tools:
	$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v3@$(KRATOS_CMD_VER)
	$(GO) install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v3@$(KRATOS_CMD_VER)
	$(GO) mod download github.com/go-kratos/kratos/cmd/kratos/v3@$(KRATOS_CMD_VER)
	@mod="$$($(GO) env GOMODCACHE)/github.com/go-kratos/kratos/cmd/kratos/v3@$(KRATOS_CMD_VER)"; \
	tmpdir=$$(mktemp -d); \
	cp -R "$$mod" "$$tmpdir/kratos"; \
	chmod -R u+w "$$tmpdir/kratos"; \
	$(GO) build -C "$$tmpdir/kratos" -o $(GOBIN)/kratos .; \
	rm -rf "$$tmpdir"

api: tools
	cd $(ROOT) && buf dep update && buf generate

conf:
	$(MAKE) conf-app APP=alioth
	$(MAKE) conf-app APP=megrez
	$(MAKE) conf-app APP=merak
	$(MAKE) conf-app APP=alkaid

conf-app:
	protoc --proto_path=$(ROOT)app/$(APP) \
		--proto_path=$(PROTO_INC) \
		--go_out=$(ROOT)app/$(APP) --go_opt=module=github.com/aide-family/moon/app/$(APP) \
		$(ROOT)app/$(APP)/internal/conf/conf.proto

vet:
	$(GO) vet -C $(ROOT)pkg ./...
	$(GO) vet -C $(ROOT)api ./...
	$(GO) vet -C $(ROOT)app/alioth ./...
	$(GO) vet -C $(ROOT)app/megrez ./...
	$(GO) vet -C $(ROOT)app/merak ./...
	$(GO) vet -C $(ROOT)app/alkaid ./...

# gofmt、命名检查与 buf lint 是 CI 的硬门禁；本地先跑一遍避免 PR 变红。
lint:
	@out=$$(gofmt -l $(ROOT)pkg $(ROOT)api $(ROOT)app); \
	if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	$(ROOT)scripts/check-naming.sh
	cd $(ROOT) && buf lint

build: build-alioth build-megrez build-merak build-alkaid build-web

build-alioth:
	$(GO) build -C $(ROOT)app/alioth -ldflags "$(version_ldflags)" -o $(BIN)/alioth ./cmd/alioth

build-megrez:
	$(GO) build -C $(ROOT)app/megrez -ldflags "$(version_ldflags)" -o $(BIN)/megrez ./cmd/megrez

build-merak:
	$(GO) build -C $(ROOT)app/merak -ldflags "$(version_ldflags)" -o $(BIN)/merak ./cmd/merak

build-alkaid:
	$(GO) build -C $(ROOT)app/alkaid -ldflags "$(version_ldflags)" -o $(BIN)/alkaid ./cmd/alkaid

build-alkaid-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -C $(ROOT)app/alkaid -ldflags "$(version_ldflags)" -o $(BIN)/alkaid-linux ./cmd/alkaid

# 摇光 systemd 一键部署包与 RPM（二进制 + 单元 + install.sh + rpm）。release.yml 调用的也是这一条。
package-alkaid:
	OUT=$(ROOT)dist VERSION=$(VERSION) ARCHES="$(ARCHES)" RPM="$(RPM)" $(ROOT)scripts/package-alkaid.sh

build-web:
	npm --prefix $(ROOT)web ci
	npm --prefix $(ROOT)web run build

# 构建上下文固定为仓库根目录，四个 Go 服务都依赖 api/ 与 pkg/。
image-alioth:
	docker build -f $(ROOT)deploy/alioth/Dockerfile --build-arg VERSION=$(VERSION) \
		-t $(IMAGE_PREFIX)/alioth:$(VERSION) $(ROOT)

image-megrez:
	docker build -f $(ROOT)deploy/megrez/Dockerfile --build-arg VERSION=$(VERSION) \
		-t $(IMAGE_PREFIX)/megrez:$(VERSION) $(ROOT)

image-merak:
	docker build -f $(ROOT)deploy/merak/Dockerfile --build-arg VERSION=$(VERSION) \
		-t $(IMAGE_PREFIX)/merak:$(VERSION) $(ROOT)

image-alkaid:
	docker build -f $(ROOT)deploy/alkaid/Dockerfile --build-arg VERSION=$(VERSION) \
		-t $(IMAGE_PREFIX)/alkaid:$(VERSION) $(ROOT)

image-web:
	docker build -f $(ROOT)deploy/web/Dockerfile \
		-t $(IMAGE_PREFIX)/web:$(VERSION) $(ROOT)

images: image-alioth image-megrez image-merak image-alkaid image-web

run-alioth:
	$(GO) run -C $(ROOT)app/alioth ./cmd/alioth serve --conf configs/config.yaml

run-megrez:
	$(GO) run -C $(ROOT)app/megrez ./cmd/megrez serve --conf configs/config.yaml

run-merak:
	$(GO) run -C $(ROOT)app/merak ./cmd/merak serve --conf configs/config.yaml

run-alkaid:
	$(GO) run -C $(ROOT)app/alkaid ./cmd/alkaid serve --conf configs/config.yaml

run-web:
	npm --prefix $(ROOT)web run dev

test: test-pkg
	$(GO) test -C $(ROOT)api ./...
	$(GO) test -C $(ROOT)app/alioth ./...
	$(GO) test -C $(ROOT)app/megrez ./...
	$(GO) test -C $(ROOT)app/merak ./...
	$(GO) test -C $(ROOT)app/alkaid ./...

test-pkg:
	$(GO) test -C $(ROOT)pkg ./...

# 本地多副本/并发问题靠竞态检测暴露；CI 的 backend job 也跑这一档。
test-race:
	$(GO) test -C $(ROOT)pkg -race ./...
	$(GO) test -C $(ROOT)api -race ./...
	$(GO) test -C $(ROOT)app/alioth -race ./...
	$(GO) test -C $(ROOT)app/megrez -race ./...
	$(GO) test -C $(ROOT)app/merak -race ./...
	$(GO) test -C $(ROOT)app/alkaid -race ./...

migrate:
	$(GO) run -C $(ROOT)app/alioth ./cmd/alioth migrate --conf configs/config.yaml

# compose 全栈里迁移由一次性容器执行，避免宿主机没装 Go 时卡住。
migrate-in-docker:
	docker compose -f $(ROOT)docker-compose.yaml run --rm migrate

compose-images:
	$(ROOT)scripts/pull-compose-images.sh

compose-up: compose-images
	docker compose -f $(ROOT)docker-compose.yaml up -d --wait

compose-down:
	docker compose -f $(ROOT)docker-compose.yaml down

loki-demo: compose-images
	docker compose -f $(ROOT)docker-compose.yaml --profile loki-demo up -d loki-demo-logs

loki-demo-stop:
	docker compose -f $(ROOT)docker-compose.yaml --profile loki-demo stop loki-demo-logs

vlogs-demo: compose-images
	docker compose -f $(ROOT)docker-compose.yaml --profile vlogs-demo up -d vlogs-demo-logs

vlogs-demo-stop:
	docker compose -f $(ROOT)docker-compose.yaml --profile vlogs-demo stop vlogs-demo-logs

# 一条命令起依赖 + 迁移 + 四个服务 + web；数据落在 pgdata 卷里。
dev-up: compose-images
	docker compose -f $(ROOT)docker-compose.yaml up -d --build --wait

dev-down:
	docker compose -f $(ROOT)docker-compose.yaml down

compose-alkaid: build-alkaid-linux
	docker compose -f $(ROOT)docker-compose.yaml --profile alkaid down alkaid
	docker compose -f $(ROOT)docker-compose.yaml --profile alkaid up -d --build alkaid

tidy:
	$(GO) mod tidy -C $(ROOT)pkg
	$(GO) mod tidy -C $(ROOT)api
	$(GO) mod tidy -C $(ROOT)app/alioth
	$(GO) mod tidy -C $(ROOT)app/megrez
	$(GO) mod tidy -C $(ROOT)app/merak
	$(GO) mod tidy -C $(ROOT)app/alkaid

wire:
	cd $(ROOT)app/alioth/cmd/alioth && wire
	cd $(ROOT)app/megrez/cmd/megrez && wire
	cd $(ROOT)app/merak/cmd/merak && wire
	cd $(ROOT)app/alkaid/cmd/alkaid && wire
