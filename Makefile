deps:
	go mod tidy

build:
	go build -o bin/qbt cmd/qbt/main.go