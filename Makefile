BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
	golangci-lint run ./...

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-search-data-extractor

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-search-data-extractor
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-search-data-extractor

.PHONY: debug-run
debug-run:
	HUMAN_LOG=1 DEBUG=1 go run -tags 'debug' -race $(LDFLAGS) main.go

.PHONY: test
test:
	go test -count=1 -race -cover ./...

.PHONY: produce
produce:
	HUMAN_LOG=1 go run cmd/producer/main.go

.PHONY: convey
convey:
	goconvey ./...

.PHONY: test-component
test-component:
	go test -cover -race -coverpkg=github.com/ONSdigital/dp-search-data-extractor/... -component
