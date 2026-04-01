GOHOSTOS:=$(shell go env GOHOSTOS)
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date '+%Y-%m-%dT%H:%M:%SZ')
AUTHOR=$(shell git log -1 --format='%an')
AUTHOR_EMAIL=$(shell git log -1 --format='%ae')
REPO=$(shell git config remote.origin.url)

ifeq ($(GOHOSTOS), windows)
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find ../../proto/jade_tree/ -name *.proto")
	MKDIR=mkdir
	RM=del /f /q
else
	API_PROTO_FILES=$(shell find ../../proto/jade_tree -name *.proto)
	MKDIR=mkdir -p
	RM=rm -f
endif

.PHONY: init
init:
	@echo "Initializing jade_tree environment"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: conf
conf:
	@echo "Generating conf files"
	protoc --proto_path=./internal/conf 	       --proto_path=../../proto/third_party 	       --proto_path=../../proto 	       --go_out=paths=source_relative:./internal/conf 	       --experimental_allow_proto3_optional 	                  ./internal/conf/*.proto

.PHONY: api
api:
	@echo "Generating api files"
	protoc --proto_path=../../proto \
	       --proto_path=../../proto/third_party \
 	       --go_out=. --go_opt=module=github.com/aide-family/jade_tree \
 	       --go-http_out=. --go-http_opt=module=github.com/aide-family/jade_tree \
 	       --go-grpc_out=. --go-grpc_opt=module=github.com/aide-family/jade_tree \
	       --openapi_out=fq_schema_naming=true,default_response=false:./internal/server/swagger \
	       --experimental_allow_proto3_optional \
	       $(API_PROTO_FILES)

.PHONY: wire
wire:
	@echo "Generating wire files"
	wire ./...

.PHONY: all
all:
	make api conf wire

.PHONY: gen
gen:
	@echo "Generating gen files"
	go test -tags=generate ./internal/data/impl/do -run TestGenerate

.PHONY: build
build: all
	@echo "Building jade_tree"
	go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" -o bin/jade_tree main.go

.PHONY: dev
dev:
	@echo "Running jade_tree in development mode"
	go run -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" . run all --log-level=DEBUG

.PHONY: clean
clean:
	rm -rf ./bin

.DEFAULT_GOAL := build
