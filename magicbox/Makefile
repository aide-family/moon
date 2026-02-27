ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	PROTO_FILES=$(shell $(Git_Bash) -c "find ../proto/magicbox -name *.proto")
else
	PROTO_FILES=$(shell find ../proto/magicbox -name *.proto)
endif

.PHONY: init
init:
	@echo "Initializing magicbox environment"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/moon-monitor/stringer@latest
	go install github.com/protoc-gen/i18n-gen@latest

.PHONY: proto
# generate proto
proto:
	protoc --proto_path=../proto \
           --proto_path=../proto/third_party \
           --go_out=. --go_opt=module=github.com/aide-family/magicbox \
           --go-grpc_out=. --go-grpc_opt=module=github.com/aide-family/magicbox \
           --go-errors_out=. --go-errors_opt=module=github.com/aide-family/magicbox \
           --go-http_out=. --go-http_opt=module=github.com/aide-family/magicbox \
           $(PROTO_FILES)

