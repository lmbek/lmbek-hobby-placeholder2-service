# Auth Service Makefile

.PHONY: test build run

test:
	@echo "==> Running tests for auth-service..."
	go test -v ./...

build:
	@echo "==> Building auth-service..."
	go build -o auth-service main.go

run:
	@echo "==> Running auth-service..."
	go run main.go
