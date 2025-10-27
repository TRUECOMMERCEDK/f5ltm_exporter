.PHONY: build cover deploy start test test-integration

build:
	docker build -t f5ltmexporter .

cover:
	go tool cover -html=cover.out

deploy:
	docker tag f5ltmexporter:latest truecommercedk/f5ltmexporter:latest
	docker push truecommercedk/f5ltmexporter:latest

start:
	go run cmd/f5ltm_exporter/*.go

test:
	go test -coverprofile=cover.out -short ./...

test-integration:
	go test -coverprofile=cover.out -p 1 ./...