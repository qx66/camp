GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
API_PROTO_FILES=$(shell find api -name *.proto)

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@v0.6.1

.PHONY: errors
# generate errors code
errors:
	protoc --proto_path=. \
               --proto_path=./third_party \
               --go_out=paths=source_relative:. \
               --go-errors_out=paths=source_relative:. \
               $(API_PROTO_FILES)

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=. \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:. \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=. \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:. \
 	       --go-http_out=paths=source_relative:. \
 	       --go-grpc_out=paths=source_relative:. \
 	       --openapi_out==paths=source_relative:. \
	       $(API_PROTO_FILES)


.PHONY: submodule
# submodule
submodule:
	git submodule update --remote


.PHONY: build
# build
build:
	mkdir -p bin/mac   && CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/mac/   ./commander/
	mkdir -p bin/linux && CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/linux/ ./commander/


	mkdir -p bin/mac   && CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/mac/   ./soldier/
	mkdir -p bin/linux && CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/linux/ ./soldier/

	cd app && fyne package -os android -icon icon.png -appID cn.com.startops.camp
	mkdir bin/android/ && mv app/app.apk bin/app.apk


.PHONY: buildTool
# buildTool
buildTool:
	mkdir -p bin/mac   && CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/mac/   ./cmd/tools/oss/upload/upload.go
	mkdir -p bin/linux && CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/linux/ ./cmd/tools/oss/upload/upload.go

.PHONY: buildCronJob
# buildTool
buildCronJob:
	mkdir -p bin/mac   && CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/mac/   ./cmd/cronJob
	mkdir -p bin/linux && CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/linux/ ./cmd/cronJob


.PHONY: generate
# generate
generate:
	go generate ./...

.PHONY: all
# generate all
all:
	make api;
	make errors;
	make config;
	make generate;

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
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
