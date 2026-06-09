.PHONY: build test vet fmt fmt-check clean dogfood

BINARY ?= bin/docs-cli
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION_PACKAGE := github.com/jhl-labs/docs-cli/internal/version

build:
	mkdir -p bin
	go build -trimpath \
		-ldflags "-s -w -X $(VERSION_PACKAGE).Version=$(VERSION) -X $(VERSION_PACKAGE).Commit=$(COMMIT) -X $(VERSION_PACKAGE).Date=$(DATE)" \
		-o $(BINARY) ./cmd/docs-cli

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './.git/*')

fmt-check:
	test -z "$(shell gofmt -l $(shell find . -name '*.go' -not -path './.git/*'))"

# Regenerate docs-cli's own standardized docs and skill (dogfooding).
dogfood: build
	$(BINARY) init . --force
	$(BINARY) --generate-skill --output skill.md
	$(BINARY) validate .

clean:
	rm -rf bin dist coverage.out docs/_site docs/_xml
