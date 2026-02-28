GOHOSTOS:=$(shell go env GOHOSTOS)
VERSION=$(shell git describe --tags --always)
BUILD_TIME=$(shell date '+%Y-%m-%dT%H:%M:%SZ')
AUTHOR=$(shell git log -1 --format='%an')
AUTHOR_EMAIL=$(shell git log -1 --format='%ae')
REPO=$(shell git config remote.origin.url)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find ../../proto/rabbit/ -name *.proto")
	# Use mkdir -p equivalent for Windows
	MKDIR=mkdir
	RM=del /f /q
else
	API_PROTO_FILES=$(shell find ../../proto/rabbit -name *.proto)
	MKDIR=mkdir -p
	RM=rm -f
endif

.PHONY: init
# initialize the rabbit environment
init:
	@echo "Initializing rabbit environment"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/aide-family/stringer@v1.1.3
	go install github.com/protoc-gen/i18n-gen@latest
	go install golang.org/x/tools/gopls@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest

.PHONY: conf
# generate the conf files
conf:
	@echo "Generating conf files"
	protoc --proto_path=./internal/conf \
           --proto_path=../../proto/third_party \
		   --proto_path=../../proto \
           --go_out=paths=source_relative:./internal/conf \
           --experimental_allow_proto3_optional \
           ./internal/conf/*.proto

.PHONY: api
# generate the api files
api:
	@echo "Generating api files"
	protoc --proto_path=../../proto \
	       --proto_path=../../proto/third_party \
 	       --go_out=. --go_opt=module=github.com/aide-family/rabbit \
 	       --go-http_out=. --go-http_opt=module=github.com/aide-family/rabbit \
 	       --go-grpc_out=. --go-grpc_opt=module=github.com/aide-family/rabbit \
	       --openapi_out=fq_schema_naming=true,default_response=false:./internal/server/swagger \
	       --experimental_allow_proto3_optional \
	       $(API_PROTO_FILES)

.PHONY: wire
# generate the wire files
wire:
	@echo "Generating wire files"
	wire ./...

.PHONY: all
# generate all files
all: 
	make api conf wire

.PHONY: build
# build the rabbit binary
build: all
	@echo "Building rabbit"
	@echo "VERSION: $(VERSION)"
	@echo "BUILD_TIME: $(BUILD_TIME)"
	@echo "AUTHOR: $(AUTHOR)"
	@echo "AUTHOR_EMAIL: $(AUTHOR_EMAIL)"
	go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" -o bin/rabbit main.go

.PHONY: build-exe
# build the rabbit binary for windows
build-exe: all
	@echo "Building rabbit"
	@echo "VERSION: $(VERSION)"
	@echo "BUILD_TIME: $(BUILD_TIME)"
	@echo "AUTHOR: $(AUTHOR)"
	@echo "AUTHOR_EMAIL: $(AUTHOR_EMAIL)"
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" -o bin/rabbit.exe main.go

.PHONY: clean
# clean the rabbit binary
clean:
	@echo "Cleaning rabbit"
	rm -rf ./bin

.PHONY: dev
# run the rabbit binary in development mode
dev:
	@echo "Running rabbit in development mode"
	go run . run all --log-level=DEBUG

.PHONY: gen
# generate the gen files
gen:
	@echo "Generating gen files"
	go test -tags=generate ./internal/data/impl/do -run TestGenerate

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