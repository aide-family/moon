GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
APPS ?= $(shell ls apps)
path := $(shell pwd)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif


.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: errors
# generate errors
errors:
	protoc --proto_path=./api \
			 --proto_path=./third_party \
			 --go_out=paths=source_relative:./api \
			 --go-errors_out=paths=source_relative:./api \
			 $(API_PROTO_FILES)

.PHONY: validate
# generate validate proto
validate:
	protoc --proto_path=./api \
		   --proto_path=./third_party \
		   --go_out=paths=source_relative:./api \
		   --validate_out=paths=source_relative,lang=go:./api \
		   $(API_PROTO_FILES)

.PHONY: api
# generate api proto
api: errors validate
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)
.PHONY: data
# generate service proto
data:
	@kratos proto data -t apps/node/internal/data api/strategy/v1/pull/pull.proto
	@kratos proto data -t apps/node/internal/data api/strategy/v1/push/push.proto
	@kratos proto data -t apps/node/internal/data api/strategy/v1/load/load.proto
	@kratos proto data -t apps/node/internal/data api/ping.proto
	@kratos proto data -t apps/master/internal/data api/ping.proto
	@kratos proto data -t apps/master/internal/data api/strategy/v1/crud.proto
	@kratos proto data -t apps/master/internal/data api/prom/v1/dir.proto
	@kratos proto data -t apps/master/internal/data api/prom/v1/file.proto
	@kratos proto data -t apps/master/internal/data api/prom/v1/group.proto
	@kratos proto data -t apps/master/internal/data api/prom/v1/rule.proto
	@kratos proto data -t apps/master/internal/data api/prom/v1/node.proto


.PHONY: biz
# generate service proto
biz:
	@kratos proto biz -t apps/node/internal/biz api/strategy/v1/pull/pull.proto
	@kratos proto biz -t apps/node/internal/biz api/strategy/v1/push/push.proto
	@kratos proto biz -t apps/node/internal/biz api/strategy/v1/load/load.proto
	@kratos proto biz -t apps/node/internal/biz api/ping.proto
	@kratos proto biz -t apps/master/internal/biz api/ping.proto
	@kratos proto biz -t apps/master/internal/biz api/strategy/v1/crud.proto
	@kratos proto biz -t apps/master/internal/biz/prom/v1 api/prom/v1/dir.proto
	@kratos proto biz -t apps/master/internal/biz/prom/v1 api/prom/v1/file.proto
	@kratos proto biz -t apps/master/internal/biz/prom/v1 api/prom/v1/group.proto
	@kratos proto biz -t apps/master/internal/biz/prom/v1 api/prom/v1/rule.proto
	@kratos proto biz -t apps/master/internal/biz/prom/v1 api/prom/v1/node.proto

.PHONY: service
# generate service proto
service:
	@kratos proto server -t apps/node/internal/service api/strategy/v1/pull/pull.proto
	@kratos proto server -t apps/node/internal/service api/strategy/v1/push/push.proto
	@kratos proto server -t apps/node/internal/service api/strategy/v1/load/load.proto
	@kratos proto server -t apps/node/internal/service api/ping.proto
	@kratos proto server -t apps/master/internal/service api/ping.proto
	@kratos proto server -t apps/master/internal/service api/strategy/v1/crud.proto
	@kratos proto server -t apps/master/internal/service/prom/v1 api/prom/v1/dir.proto
	@kratos proto server -t apps/master/internal/service/prom/v1 api/prom/v1/file.proto
	@kratos proto server -t apps/master/internal/service/prom/v1 api/prom/v1/group.proto
	@kratos proto server -t apps/master/internal/service/prom/v1 api/prom/v1/rule.proto
	@kratos proto server -t apps/master/internal/service/prom/v1 api/prom/v1/node.proto


.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: test
# test
test:
	go test -v ./... -cover

.PHONY: generate
# generate
generate:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
	go generate ./...

.PHONY: config
# generate internal config
config:
	@for app in $(APPS); do \
		echo "generate internal config for $$app"; \
		cd $(path)/apps/$$app && make config; \
	done

.PHONY: all
# generate all
all:
	make api;
	make generate;

.PHONY: model
# generate model
model:
	@cd gen && go run .
	@git add .

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
