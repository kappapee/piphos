install:
	@go mod tidy
	@go install golang.org/x/tools/cmd/goimports@latest

format:
	@go fmt ./...
	@goimports -w .

check:
	@go vet ./...
	@golangci-lint run ./... --fix

test:
	@go test ./...

run:
	@go run .

build:
	@mkdir -p ./bin
	@rm -f ./bin/*
	GOOS=darwin GOARCH=amd64 go build -o ./bin/piphos-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o ./bin/piphos-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o ./bin/piphos-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o ./bin/piphos-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o ./bin/piphos-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o ./bin/piphos-windows-arm64.exe .

