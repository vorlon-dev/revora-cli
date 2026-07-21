BINARY_NAME = revora
BUILD_DIR = build
GO = go

.PHONY: all build clean test lint install

all: build

build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/revora

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GO) test ./...

lint:
	golangci-lint run

install:
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/