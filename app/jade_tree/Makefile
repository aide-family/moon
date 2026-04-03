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

.PHONY: build-exe
# build the jade_tree binary for windows
build-exe: all
	@echo "Building jade_tree (Windows)"
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" -o bin/jade_tree.exe main.go

.PHONY: build-mac
# build the jade_tree binary for macOS (darwin)
build-mac: all
	@echo "Building jade_tree (macOS)"
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" -o bin/jade_tree main.go

.PHONY: build-linux
# build the jade_tree binary for linux
build-linux: all
	@echo "Building jade_tree (Linux)"
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" -o bin/jade_tree main.go

.PHONY: build-docker
build-docker:
	@echo "Building jade_tree docker image"
	# Build with repo root as context so Docker can access `proto/` and `magicbox/`
	# referenced by this app's Makefile (e.g. `../../proto`).
	docker build -f Dockerfile -t jade_tree:$(VERSION) ../..

.PHONY: dev
dev:
	@echo "Running jade_tree in development mode"
	go run -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Author=$(AUTHOR) -X main.Email=$(AUTHOR_EMAIL) -X main.Repo=$(REPO)" . run all --log-level=DEBUG

.PHONY: clean rpm srpm rpm-clean rpm-docker rpm-docker-build srpm-docker
clean:
	rm -rf ./bin

# --- RPM packaging (artifacts under ./rpmbuild/) ---
RPM_TOPDIR := $(abspath rpmbuild)
RPM_VERSION ?= $(shell v=$$(git describe --tags --match 'v*' --abbrev=0 2>/dev/null | sed 's/^v//'); echo $${v:-0.0.1})
RPM_RELEASE ?= 1
MK_TARBALL := packaging/rpm/mk-source-tarball.sh
RPM_SPEC := packaging/rpm/jade-tree.spec
RPM_DOCKERFILE := packaging/rpm/Dockerfile.rpmbuild
RPM_BUILD_IMAGE ?= jade-tree-rpmbuild:local

rpm-clean:
	rm -rf $(RPM_TOPDIR)

rpm:
	@command -v rpmbuild >/dev/null || { \
		echo "rpmbuild not found. Use make rpm-docker on macOS."; \
		exit 1; \
	}
	@command -v rsync >/dev/null || { echo "rsync not found. Please install rsync first."; exit 1; }
	@$(MAKE) rpm-clean
	mkdir -p $(RPM_TOPDIR)/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	bash $(MK_TARBALL) "$(RPM_VERSION)" "$(CURDIR)" "$(RPM_TOPDIR)/SOURCES"
	cp $(RPM_SPEC) $(RPM_TOPDIR)/SPECS/jade-tree.spec
	rpmbuild --define "_topdir $(RPM_TOPDIR)" \
		--define "jade_tree_version $(RPM_VERSION)" \
		--define "jade_tree_release $(RPM_RELEASE)" \
		-ba $(RPM_TOPDIR)/SPECS/jade-tree.spec
	@echo "RPM artifacts:"
	@ls -1 $(RPM_TOPDIR)/RPMS/*/*.rpm 2>/dev/null || true

srpm:
	@command -v rpmbuild >/dev/null || { \
		echo "rpmbuild not found. Use make srpm-docker on macOS."; \
		exit 1; \
	}
	@command -v rsync >/dev/null || { echo "rsync not found. Please install rsync first."; exit 1; }
	@$(MAKE) rpm-clean
	mkdir -p $(RPM_TOPDIR)/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	bash $(MK_TARBALL) "$(RPM_VERSION)" "$(CURDIR)" "$(RPM_TOPDIR)/SOURCES"
	cp $(RPM_SPEC) $(RPM_TOPDIR)/SPECS/jade-tree.spec
	rpmbuild --define "_topdir $(RPM_TOPDIR)" \
		--define "jade_tree_version $(RPM_VERSION)" \
		--define "jade_tree_release $(RPM_RELEASE)" \
		-bs $(RPM_TOPDIR)/SPECS/jade-tree.spec
	@echo "SRPM artifact:"
	@ls -1 $(RPM_TOPDIR)/SRPMS/*.src.rpm 2>/dev/null || true

rpm-docker-build:
	@command -v docker >/dev/null || { echo "docker not found. Please install Docker first."; exit 1; }
	docker build -t $(RPM_BUILD_IMAGE) -f $(RPM_DOCKERFILE) $(dir $(RPM_DOCKERFILE))

rpm-docker: rpm-docker-build
	docker run --rm \
		-v "$(CURDIR):/src:rw" \
		-w /src \
		-e RPM_VERSION="$(RPM_VERSION)" \
		-e RPM_RELEASE="$(RPM_RELEASE)" \
		-e GOPROXY="$(GOPROXY)" \
		-e GOSUMDB="$(GOSUMDB)" \
		-e GOPRIVATE="$(GOPRIVATE)" \
		$(RPM_BUILD_IMAGE) \
		make rpm

srpm-docker: rpm-docker-build
	docker run --rm \
		-v "$(CURDIR):/src:rw" \
		-w /src \
		-e RPM_VERSION="$(RPM_VERSION)" \
		-e RPM_RELEASE="$(RPM_RELEASE)" \
		-e GOPROXY="$(GOPROXY)" \
		-e GOSUMDB="$(GOSUMDB)" \
		-e GOPRIVATE="$(GOPRIVATE)" \
		$(RPM_BUILD_IMAGE) \
		make srpm

.DEFAULT_GOAL := build
