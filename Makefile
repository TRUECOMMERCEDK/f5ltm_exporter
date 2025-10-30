.PHONY: build coverstart test test-integration

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X main.release=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o f5ltm_exporter .

cover:
	go tool cover -html=cover.out

start:
	go run cmd/f5ltm_exporter/*.go

test:
	go test -coverprofile=cover.out -short ./...

test-integration:
	go test -coverprofile=cover.out -p 1 ./...