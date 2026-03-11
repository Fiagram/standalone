VERSION := $(shell cat VERSION)
COMMIT_HASH := $(shell git rev-parse HEAD)
PROJECT_NAME := standalone

all: generate build-all

.PHONY: init
init:
	go install github.com/rubenv/sql-migrate/...@v1.8.1
	go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: generate
generate:
	@echo "--- Generating OpenAPI server and types ---"
	@go tool oapi-codegen --config=oapi_codegen.yml docs/openapi.yml

.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_linux_amd64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_linux_arm64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-macos-amd64
build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_macos_amd64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-macos-arm64
build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_macos_arm64 cmd/$(PROJECT_NAME)/*.go

.PHONY: build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_windows_amd64.exe cmd/$(PROJECT_NAME)/*.go

.PHONY: build-windows-arm64
build-windows-arm64:
	GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME)_windows_arm64.exe cmd/$(PROJECT_NAME)/*.go

.PHONY: build-all
build-all:
	make build-linux-amd64
	make build-linux-arm64
	make build-macos-amd64
	make build-macos-arm64
	make build-windows-amd64
	make build-windows-arm64

.PHONY: build
build:
	go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/$(PROJECT_NAME) \
		cmd/$(PROJECT_NAME)/*.go

.PHONY: clean
clean:
	rm -rf build/
	go clean -cache
	go clean -testcache

.PHONY: docker-compose-up-test
docker-compose-up-test:
	docker-compose -f deployments/docker-compose.test.yml up -d --force-recreate

.PHONY: docker-compose-down-test
docker-compose-down-test:
	docker-compose -f deployments/docker-compose.test.yml down

.PHONY: docker-compose-ps-test
docker-compose-ps-test:
	docker-compose -f deployments/docker-compose.test.yml ps

.PHONY: docker-compose-logs-test
docker-compose-logs-test:
	docker-compose -f deployments/docker-compose.test.yml logs

.PHONY: run-server
run-server:
	go run cmd/$(PROJECT_NAME)/*.go

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: test
test:
	go test -v ./test/configs \
				./test/dao/cache/redis \
				./test/dao/database \
				./test/logic/account \
				./test/logic/token

.PHONY: lint
lint:
	golangci-lint run ./... 

.PHONY: migrate-up-dev
migrate-up-dev:
	sql-migrate up -config=internal/dataaccess/database/migrations/sql-migrate-config.yaml

.PHONY: migrate-down-dev
migrate-down-dev:
	sql-migrate down -config=internal/dataaccess/database/migrations/sql-migrate-config.yaml 

.PHONY: migrate-new
migrate-new:
	sql-migrate new -config=internal/dataaccess/database/migrations/sql-migrate-config.yaml "$(word 2,$(MAKECMDGOALS))" 
%:
	@true

.PHONY: migrate-status
migrate-status:
	sql-migrate status -config=internal/dataaccess/database/migrations/sql-migrate-config.yaml