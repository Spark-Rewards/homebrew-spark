.PHONY: build install clean test release-dry

VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/BrianBFarias/homebrew-spark/cmd.Version=$(VERSION) \
	-X github.com/BrianBFarias/homebrew-spark/cmd.Commit=$(COMMIT) \
	-X github.com/BrianBFarias/homebrew-spark/cmd.Date=$(DATE)

build:
	go build -ldflags '$(LDFLAGS)' -o bin/spark main.go

install: build
	cp bin/spark /usr/local/bin/spark
	@echo "spark installed to /usr/local/bin/spark"

clean:
	rm -rf bin/ dist/

test:
	go test ./...

release-dry:
	goreleaser release --snapshot --clean
