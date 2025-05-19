GOHOSTOS:=$(shell go env GOHOSTOS)
VERSION=$(shell git describe --tags --always)
APP_NAME ?= $(app)

# Define path separator based on OS
ifeq ($(GOHOSTOS), windows)
	PATHSEP=\\
	PSEP=;
else
	PATHSEP=/
	PSEP=:
endif

# Normalize paths for different OS
NORMALIZE_PATH=$(subst /,$(PATHSEP),$1)
DENORMALIZE_PATH=$(subst $(PATHSEP),/,$1)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find cmd -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find proto/api/$(APP_NAME) -name *.proto")
	# Use mkdir -p equivalent for Windows
	MKDIR=mkdir
	RM=del /f /q
else
	INTERNAL_PROTO_FILES=$(shell find cmd -name *.proto)
	API_PROTO_FILES=$(shell find proto/api/$(APP_NAME) -name *.proto)
	MKDIR=mkdir -p
	RM=rm -f
endif

.PHONY: init
init:
	@echo "Initializing moon environment"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/moon-monitor/stringer@latest
	go install github.com/protoc-gen/i18n-gen@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

.PHONY: all
all:
	@echo "Initialization of moon project"
	@if [ -z "$(APP_NAME)" ]; then echo "app name is required"; echo "usage: make all app=<app_name>"; exit 1; fi
	make api app=palace
	make api app=houyi
	make api app=rabbit
	make api app=laurel
	make errors
	make conf
	make stringer-$(APP_NAME)
	make conf-$(APP_NAME)
	make wire-$(APP_NAME)
	make gen-$(APP_NAME)

.PHONY: api
# generate api proto
api:
	@if [ -z "$(APP_NAME)" ]; then echo "app name is required"; echo "usage: make api app=<app_name>"; exit 1; fi
	@echo "Generating api proto"
	@echo "APP_NAME: $(APP_NAME)"
	@if [ "$(GOHOSTOS)" = "windows" ]; then \
		$(Git_Bash) -c "rm -rf ./pkg/api/$(APP_NAME)"; \
		if [ ! -d "./pkg/api" ]; then $(MKDIR) ./pkg/api; fi \
	else \
		rm -rf ./pkg/api/$(APP_NAME); \
		if [ ! -d "./pkg/api" ]; then $(MKDIR) ./pkg/api; fi \
	fi
	protoc --proto_path=./proto/api \
	       --proto_path=./proto/api \
	       --proto_path=./proto/config \
	       --proto_path=./proto/third_party \
 	       --go_out=paths=source_relative:./pkg/api \
 	       --go-http_out=paths=source_relative:./pkg/api \
 	       --go-grpc_out=paths=source_relative:./pkg/api \
	       --openapi_out=fq_schema_naming=true,default_response=false:./cmd/$(APP_NAME)/internal/server/swagger \
	       --experimental_allow_proto3_optional \
	       $(API_PROTO_FILES) ./proto/api/common/*.proto

.PHONY: errors
# generate errors
errors:
	mkdir -p ./pkg/merr
	protoc --proto_path=./proto/merr \
           --proto_path=./proto/third_party \
           --go_out=paths=source_relative:./pkg/merr \
           --go-errors_out=paths=source_relative:./pkg/merr \
           ./proto/merr/*.proto
	make i18n

.PHONY: conf
# generate config
conf:
	mkdir -p ./pkg/config
	protoc --proto_path=./proto/config \
           --proto_path=./proto/third_party \
           --go_out=paths=source_relative:./pkg/config \
           --experimental_allow_proto3_optional \
           ./proto/config/*.proto

.PHONY: i18n
# i18n
i18n:
	i18n-gen -O ./i18n/ -P ./proto/merr/err.proto -L en,ja,zh

.PHONY: gen-palace
# generate gorm gen
gen-palace:
	rm -rf ./cmd/palace/internal/data/query
	go run cmd/palace/migrate/gen/gen.go

.PHONY: conf-palace
# generate palace-config
conf-palace:
	make conf
	protoc --proto_path=./proto/config \
           --proto_path=./proto/third_party \
           --proto_path=./cmd/palace/internal/conf \
           --go_out=paths=source_relative:./cmd/palace/internal/conf \
           --experimental_allow_proto3_optional \
           ./cmd/palace/internal/conf/*.proto

.PHONY: conf-rabbit
conf-rabbit:
	make conf
	make api app=rabbit
	protoc --proto_path=./proto/config \
           --proto_path=./proto/api \
           --proto_path=./proto/third_party \
           --proto_path=./cmd/rabbit/internal/conf \
           --go_out=paths=source_relative:./cmd/rabbit/internal/conf \
           --experimental_allow_proto3_optional \
           ./cmd/rabbit/internal/conf/*.proto

.PHONY: conf-houyi
conf-houyi:
	make conf
	make api app=houyi
	protoc --proto_path=./proto/config \
           --proto_path=./proto/api \
           --proto_path=./proto/third_party \
           --proto_path=./cmd/houyi/internal/conf \
           --go_out=paths=source_relative:./cmd/houyi/internal/conf \
           --experimental_allow_proto3_optional \
           ./cmd/houyi/internal/conf/*.proto

.PHONY: conf-laurel
conf-laurel:
	make conf
	make api app=laurel
	protoc --proto_path=./proto/config \
           --proto_path=./proto/api \
           --proto_path=./proto/third_party \
           --proto_path=./cmd/laurel/internal/conf \
           --go_out=paths=source_relative:./cmd/laurel/internal/conf \
           --experimental_allow_proto3_optional \
           ./cmd/laurel/internal/conf/*.proto

.PHONY: wire-palace
wire-palace:
	cd ./cmd/palace && wire

.PHONY: wire-rabbit
wire-rabbit:
	cd ./cmd/rabbit && wire

.PHONY: wire-houyi
wire-houyi:
	cd ./cmd/houyi && wire

.PHONY: wire-laurel
wire-laurel:
	cd ./cmd/laurel && wire

.PHONY: stringer-palace
stringer-palace:
	cd ./cmd/palace/internal/biz/vobj && go generate

.PHONY: stringer-rabbit
stringer-rabbit:
	@echo "Generating rabbit stringer"
	cd ./cmd/rabbit/internal/biz/vobj && go generate

.PHONY: stringer-houyi
stringer-houyi:
	@echo "Generating houyi stringer"
	cd ./cmd/houyi/internal/biz/vobj && go generate

.PHONY: stringer-laurel
stringer-laurel:
	@echo "Generating laurel stringer"
	cd ./cmd/laurel/internal/biz/vobj && go generate
	
.PHONY: gen-rabbit
gen-rabbit:
	@echo "Generating rabbit db"

.PHONY: gen-houyi
gen-houyi:
	@echo "Generating houyi db"

.PHONY: gen-laurel
gen-laurel:
	@echo "Generating laurel db"

.PHONY: build
build:
	@if [ -z "$(APP_NAME)" ]; then echo "app name is required"; echo "usage: make build app=<app_name>"; exit 1; fi
	@echo "Building moon app=$(APP_NAME)"
	make all
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./cmd/$(APP_NAME)

.PHONY: run
run:
	@if [ -z "$(APP_NAME)" ]; then echo "app name is required"; echo "usage: make run app=<app_name>"; exit 1; fi
	@echo "Running moon app=$(APP_NAME)"
	make all
	golangci-lint run -v --fix
	go run ./cmd/$(APP_NAME) -c ./cmd/$(APP_NAME)/config

.PHONY: migrate-table
migrate-table:
	@if [ -z "$(APP_NAME)" ]; then echo "app name is required"; echo "usage: make migrate-table app=<app_name>"; exit 1; fi
	@echo "Migrating moon app=$(APP_NAME)"
	go run ./cmd/$(APP_NAME)/migrate/auto_migrate/migrate.go

.PHONY: docker-build
docker-build:
	@if [ -z "$(APP_NAME)" ]; then echo "app name is required"; echo "usage: make docker-build app=<app_name>"; exit 1; fi
	@echo "Building moon app=$(APP_NAME)"
	docker build -t ghcr.io/moon-monitor/$(APP_NAME):$(VERSION) \
      --build-arg APP_NAME=$(APP_NAME) \
      -f deploy/server/Dockerfile .

.PHONY: builder-image
builder-image:
	@echo "Building moon builder image"
	docker build -f deploy/base/DockerfileBuilder -t ghcr.io/moon-monitor/moon:builder .

.PHONY: deploy-image
deploy-image:
	@echo "Building moon deploy image"
	docker build -f deploy/base/DockerfileDeploy -t ghcr.io/moon-monitor/moon:deploy .