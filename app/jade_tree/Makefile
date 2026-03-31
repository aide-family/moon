GOHOSTOS:=$(shell go env GOHOSTOS)
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date '+%Y-%m-%dT%H:%M:%SZ')
AUTHOR=$(shell git log -1 --format='%an')
AUTHOR_EMAIL=$(shell git log -1 --format='%ae')
REPO=$(shell git config remote.origin.url)

ifeq ($(GOHOSTOS), windows)
	MKDIR=mkdir
	RM=del /f /q
else
	MKDIR=mkdir -p
	RM=rm -f
endif

.PHONY: init
init:
	@echo "Initializing jade_tree environment"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3
	go install github.com/google/wire/cmd/wire@latest

.PHONY: conf
conf:
	@echo "Generating conf files"
	protoc --proto_path=./internal/conf 	       --proto_path=../../proto/third_party 	       --proto_path=../../proto 	       --go_out=paths=source_relative:./internal/conf 	       --experimental_allow_proto3_optional 	       ./internal/conf/*.proto

.PHONY: wire
wire:
	@echo "Generating wire files"
	wire ./...

.PHONY: all
all:
	make conf wire

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
