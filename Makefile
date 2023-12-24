GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
APPS ?= $(shell ls app)
path := $(shell pwd)

# 获取输入的参数
APP_NAME ?= $(app)

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

.PHONY:
dev:
	@echo "make dev"
	@echo "APP_NAME: $(APP_NAME)"
	@echo "VERSION: $(VERSION)"
	@cd $(APP_NAME) && make dev

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
	mkdir -p bin/ &&  CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build  -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

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
	@cd apps/master/web && yarn start

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



TAG ?= latest
REPO ?= docker.hub# TODO: set your repository address

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY:
docker-build: # test ## Build docker image with the manager.
	@echo "Building docker image with the manager..."
	docker build -t "${REPO}/prometheus-manager:${TAG}" .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${REPO}/prometheus-manager:${TAG}


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