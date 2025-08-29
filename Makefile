install:
	@go mod tidy

dev:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
	@go install github.com/goreleaser/goreleaser/v2@latest

check:
	@golangci-lint run

format:
	@golangci-lint fmt

test:
	@go test ./...

run:
	@go run .

build:
	@goreleaser release --snapshot --clean
