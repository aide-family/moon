GO ?= go
ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BIN := $(ROOT)bin
GOBIN := $(shell $(GO) env GOPATH)/bin
export PATH := $(GOBIN):$(PATH)
PROTO_INC ?= /opt/homebrew/include

.PHONY: api conf build build-alioth build-megrez build-merak build-alkaid build-web \
	run-alioth run-megrez run-merak run-alkaid run-web \
	test test-pkg migrate compose-up compose-down tidy wire

api:
	cd $(ROOT) && buf dep update && buf generate

conf:
	$(MAKE) conf-app APP=alioth
	$(MAKE) conf-app APP=megrez
	$(MAKE) conf-app APP=merak
	$(MAKE) conf-app APP=alkaid

conf-app:
	protoc --proto_path=$(ROOT)app/$(APP) \
		--proto_path=$(PROTO_INC) \
		--go_out=$(ROOT)app/$(APP) --go_opt=module=github.com/aide-family/Alioth/app/$(APP) \
		$(ROOT)app/$(APP)/internal/conf/conf.proto

build: build-alioth build-megrez build-merak build-alkaid build-web

build-alioth:
	$(GO) build -C $(ROOT)app/alioth -o $(BIN)/alioth ./cmd/alioth

build-megrez:
	$(GO) build -C $(ROOT)app/megrez -o $(BIN)/megrez ./cmd/megrez

build-merak:
	$(GO) build -C $(ROOT)app/merak -o $(BIN)/merak ./cmd/merak

build-alkaid:
	$(GO) build -C $(ROOT)app/alkaid -o $(BIN)/alkaid ./cmd/alkaid

build-web:
	npm --prefix $(ROOT)web ci
	npm --prefix $(ROOT)web run build

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

migrate:
	$(GO) run -C $(ROOT)app/alioth ./cmd/alioth migrate --conf configs/config.yaml

compose-up:
	docker compose -f $(ROOT)docker-compose.yaml up -d --wait

compose-down:
	docker compose -f $(ROOT)docker-compose.yaml down

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
