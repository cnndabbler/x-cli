BINARY_NAME=x
VERSION=0.1.0
LDFLAGS=-ldflags "-s -w"

.PHONY: build clean test all

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

all: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*
