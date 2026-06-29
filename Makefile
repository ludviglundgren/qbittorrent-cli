.PHONY: deps build docs

deps:
	go mod tidy

build:
	go build -o bin/qbt cmd/qbt/main.go

docs:
	go run ./tools/gen-docs
