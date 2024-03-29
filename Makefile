# NOTE: set your repository address
REPO ?= docker.hub

# local
LOCAL_OS:=$(shell go env GOOS)
LOCAL_ARCH:=$(shell go env GOARCH)

# golang build arg
# TODO: go_cc
CGO_ENABLED?=0
GOOS:=linux
GOARCH:=amd64

ifdef BUILD_GOOS
	GOOS=${BUILD_GOOS}
endif

ifdef BUILD_GOARCH
	GOARCH=${BUILD_GOARCH}
endif

GO_BUILD_ARG:=CGO_ENABLED=$(CGO_ENABLED)  GOOS=$(GOOS) GOARCH=$(BUILD_GOARCH)

# version
GIT_TAG=$(shell git describe --tags --always)
GIT_MOD=$(shell if ! git diff-index --quiet HEAD;then echo "-dirty";fi)
VERSION:=$(GIT_TAG)$(GIT_MOD)
PROM-SERVER-VERSION:=$(VERSION)
PROM-AGENT-VERSION:=$(VERSION)

# image
PROM-SERVER-IMAGE:=${REPO}/moon-server:${PROM-SERVER-VERSION}
PROM-AGENT-IMAGE:=${REPO}/moon-agent:${PROM-AGENT-VERSION}

APPS ?= $(shell ls app)
path := $(shell pwd)

# 获取输入的参数
APP_NAME ?= $(app)
# 从app/prom_server字符串 to prom_server
APP := $(subst app/,,$(APP_NAME))
ARGS ?= $(args)

ifeq ($(LOCAL_OS), windows)
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

.PHONY:
env:
	@echo "VERSION=$(VERSION)" > ./deploy/docker/.env

.PHONY:
dev:
	@echo "make dev"
	@echo "APP_NAME: $(APP_NAME)"
	@echo "VERSION: $(VERSION)"
	@cd $(APP_NAME) && make dev

# ------------------ DOCKER ------------------
all-docker-build: prom-server-docker-build prom-agent-docker-build

all-docker-push: prom-server-docker-push prom-agent-docker-push

prom-server-docker-build:
	@echo "Building docker image with the prom-server..."
	docker build --build-arg VERSION=$(VERSION) -t $(PROM-SERVER-IMAGE) -f ./deploy/docker/prom-server/Dockerfile .
	@echo "Successfully build docker image with the prom-server."

prom-server-docker-push:
	@echo "start push image [$(PROM-SERVER-IMAGE)]"
	docker push $(PROM-SERVER-IMAGE)
	@echo "Successfully push image [$(PROM-SERVER-IMAGE)] to target repo."

prom-agent-docker-build:
	@echo "Building docker image with the prom-server..."
	docker build --build-arg VERSION=$(VERSION) -t $(PROM-AGENT-IMAGE) -f ./deploy/docker/prom-agent/Dockerfile .
	@echo "Successfully build docker image with the prom-agent."

prom-agent-docker-push: # test ## push docker image with the prom-server.
	@echo "start push image [$(PROM-AGENT-IMAGE)]"
	docker push $(PROM-AGENT-IMAGE)
	@echo "Successfully push image [$(PROM-AGENT-IMAGE)] to target repo."

# ------------------ DOCKER-COMPOSE ------------------
all-docker-compose-up: env
	docker-compose -f ./deploy/docker/docker-compose.yaml --env-file ./deploy/docker/.env up

all-docker-compose: env
	docker-compose -f ./deploy/docker/docker-compose.yaml --env-file ./deploy/docker/.env $(ARGS)

prom-server-compose-up: clean-server-cache prom-server-docker-build
	docker tag $(PROM-SERVER-IMAGE) docker-prometheus_manager_server:latest
	docker-compose -f ./deploy/docker/docker-compose.yaml up prometheus_manager_server -d --no-build

prom-agent-compose-up: prom-agent-docker-build
	docker tag $(PROM-AGENT-IMAGE) docker-prometheus_manager_agent:latest
	docker-compose -f ./deploy/docker/docker-compose.yaml up prometheus_manager_agent -d --no-build

# The prom-server container has a problem that when it is restarted, 
# the cache file cannot be mounted into the container. It needs to clean up the cache.
# Other services rely on the prom-server, so they also need to do this make action.
# WARNNING: If there is important test data in the cache file, please back it up yourself.
clean-server-cache:
	rm -rf ./app/prom_server/cache/

# ------------------ LOCAL ------------------
.PHONY:
local:
	@echo "make local"
	@echo "APP_NAME: $(APP_NAME)"
	@echo "VERSION: $(VERSION)"
	@cd $(APP_NAME) && make local

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
	      --openapi_out=fq_schema_naming=true,default_response=false:./third_party/swagger_ui \
	       $(API_PROTO_FILES)

.PHONY: build
# build
build:
	mkdir -p bin/ &&  CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: test
# test
test:
	go test -v ./... -cover

.PHONY: wire
# generate
wire:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
	go generate ./...

.PHONY: config
# generate internal config
config:
	@for app in $(APPS); do \
		echo "generate internal config for $$app"; \
		cd $(path)/app/$$app && make config; \
	done

.PHONY: all
# generate all
all:
	make api;
	make generate;

.PHONY: web
# start web
web:
	@cd web && yarn dev

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

# ------------------ KUBERNETES ------------------
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.2)

.PHONY: deploy-yaml
# Generate deploy yaml.
deploy-yaml: kustomize ## Generate deploy yaml.
	$(call gen-yamls)


PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

define gen-yamls
{\
set -e ;\
[ -f ${PROJECT_DIR}/_output/yamls/build ] || mkdir -p ${PROJECT_DIR}/_output/yamls/build; \
rm -rf ${PROJECT_DIR}/_output/yamls/build/manager; \
cp -rf ${PROJECT_DIR}/config/* ${PROJECT_DIR}/_output/yamls/build/; \
cd ${PROJECT_DIR}/_output/yamls/build/manager; \
${KUSTOMIZE} edit set image controller=${REPO}/prometheus-manager:${TAG}; \
set +x ;\
echo "==== create prometheus-manager.yaml in ${PROJECT_DIR}/_output/yamls/ ====";\
${KUSTOMIZE} build ${PROJECT_DIR}/_output/yamls/build/default > ${PROJECT_DIR}/_output/yamls/prometheus-manager.yaml;\
}
endef