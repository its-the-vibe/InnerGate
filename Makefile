BINARY   := innergate
PKG      := ./...

.PHONY: build test lint run clean

## build: compile the binary
build:
	go build -o $(BINARY) .

## test: run all unit tests with race detector and coverage
test:
	go test -v -race -cover $(PKG)

## lint: run golangci-lint (install separately: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run $(PKG)

## run: build and execute the binary (requires config.json)
run: build
	./$(BINARY)

## clean: remove the compiled binary
clean:
	rm -f $(BINARY)
