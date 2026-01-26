GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "v0.0.0")

ifeq ($(GOHOSTOS), windows)
	# 直接指定正确的Git Bash路径（无需自动提取）
	Git_Bash="D:/Git/bin/bash.exe"
	# 查找Proto文件
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name '*.proto' 2>/dev/null")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name '*.proto' 2>/dev/null")
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

.PHONY: config
# generate internal proto
config:
	# 【修改3】检查Proto文件是否存在，无文件则提示并正常退出（避免protoc报错）
	@if [ -z "$(INTERNAL_PROTO_FILES)" ]; then \
		echo "WARNING: No proto files found in internal directory, skip config generate."; \
		exit 0; \
	fi
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	# 【修改3】同config目标，增加文件检查
	@if [ -z "$(API_PROTO_FILES)" ]; then \
		echo "WARNING: No proto files found in api directory, skip api generate."; \
		exit 0; \
	fi
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...



.PHONY: validate
# generate validate proto
validate:
	protoc --proto_path=. \
		   --proto_path=./third_party \
		   --go_out=paths=source_relative:. \
		   --validate_out=paths=source_relative,lang=go:. \
		   $(API_PROTO_FILES)

.PHONY: errors
errors:
	protoc --proto_path=. \
			 --proto_path=./third_party \
			 --go_out=paths=source_relative:. \
			 --go-errors_out=paths=source_relative:. \
			 $(API_PROTO_FILES)

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	make api
	make config
	make generate

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

