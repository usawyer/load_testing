run:
	go run cmd/main.go
.PHONY: run

linter:
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .
	gofmt -s -w .
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --out-format colored-line-number -v
.PHONY: linter
