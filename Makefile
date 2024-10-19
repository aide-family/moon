GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
server := "cmd/server"
APPS ?= $(shell ls $(server))
path := $(shell pwd)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find cmd -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
	ERROR_PROTO_FILES=$(shell $(Git_Bash) -c "find pkg -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find cmd -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
	ERROR_PROTO_FILES=$(shell find pkg -name *.proto)
endif

.PHONY: init
# init env
init:
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/aide-cloud/protoc-gen-go-errors@latest

.PHONY: format
format:
	gofmt -s -w .
	golint ./...
	go vet ./...
	go mod tidy
	go mod verify
	goimports -w .

.PHONY: config
# generate internal config
config:
	protoc --proto_path=./third_party \
    		   --proto_path=./pkg \
     	       --go_out=paths=source_relative:./pkg $(ERROR_PROTO_FILES)
	@for app in $(APPS); do \
  		if [ "$$app" = "gen" ] || [ "$$app" = "stringer" ]; then \
			continue; \
		fi; \
		echo "generate internal config for $$app"; \
		cd $(path)/$(server)/$$app && make config; \
	done

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --proto_path=./pkg \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
 	       --go-errors_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:./third_party/swagger_ui \
	       $(API_PROTO_FILES)

.PHONY: error
# generate api proto
error:
	go install github.com/aide-cloud/protoc-gen-go-errors@latest
	protoc --proto_path=./third_party \
		   --proto_path=./pkg \
 	       --go_out=paths=source_relative:./pkg \
		   --go-errors_out=paths=source_relative:./pkg $(ERROR_PROTO_FILES)

.PHONY: error
# generate api proto
errorx:
	protoc --proto_path=./third_party \
		   --proto_path=./pkg \
 	       --go_out=paths=source_relative:./pkg \
		   --go-errors_out=paths=source_relative:./pkg $(ERROR_PROTO_FILES)


.PHONY: build
# build
build: all
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

# CGO_ENABLED=0 GOOS=linux GOARCH=amd64
.PHONY: build-linux
build-linux: all
	mkdir -p bin/linux/ && GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/linux ./...

.PHONY: stringer
stringer:
	go generate $(path)/pkg/vobj

.PHONY: wire
# generate
wire:
	go mod tidy
	go get github.com/google/wire/cmd/wire@latest
	@for app in $(APPS); do \
  		if [ "$$app" = "gen" ] || [ "$$app" = "stringer" ]; then \
			continue; \
		fi; \
		echo "generate internal config for $$app"; \
		cd $(path)/$(server)/$$app && wire; \
	done

.PHONY: model
model:
 	rm -rf pkg/palace/model/query
 	rm -rf pkg/palace/model/bizmodel/bizquery
 	rm -rf pkg/palace/model/alarmmodel/alarmquery
	go run cmd/server/gen/gen/cmd.go -m main
	go run cmd/server/gen/gen/cmd.go -m biz
	go run cmd/server/gen/gen/cmd.go -m alarm

.PHONY: all
# generate all
all: error api config stringer model wire
	go mod tidy

.PHONY: clean
# clean
clean:
	rm -rf ./bin/

.PHONY: houyi
# local run houyi
houyi:
	go run -ldflags "-X main.Version=$(VERSION)" cmd/server/houyi/houyi/cmd.go -c cmd/server/houyi/configs

.PHONY: rabbit
# local run rabbit
rabbit:
	go run -ldflags "-X main.Version=$(VERSION)" cmd/server/rabbit/rabbit/cmd.go -c cmd/server/rabbit/configs

.PHONY: palace
# local run palace
palace:
	go run -ldflags "-X main.Version=$(VERSION)" cmd/server/palace/palace/cmd.go -c cmd/server/palace/configs

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