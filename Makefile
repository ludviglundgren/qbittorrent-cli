deps:
	go mod tidy

build:
	go build -o bin/qbt cmd/qbt/main.go
	cp bin/qbt ~/.scripts/

buildW:
	env GOOS=windows GOARCH=amd64 go build -o bin/qbt.exe cmd/qbt/main.go
