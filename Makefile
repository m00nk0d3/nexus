VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := nexus
PKG     := github.com/m00nk0d3/nexus/internal/version
LDFLAGS := -X $(PKG).Version=$(VERSION)

.PHONY: build test lint clean install release snapshot \
        build-linux build-darwin build-windows

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/nexus

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-linux-amd64 ./cmd/nexus

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-amd64 ./cmd/nexus
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-darwin-arm64 ./cmd/nexus

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY).exe ./cmd/nexus

test:
	go test ./... -v -count=1

lint:
	golangci-lint run

clean:
	rm -f $(BINARY) $(BINARY)-linux-* $(BINARY)-darwin-* $(BINARY).exe

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/nexus

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean
