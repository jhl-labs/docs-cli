.PHONY: build test cover vet fmt fmt-check clean dogfood

COVERAGE_MIN ?= 90

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

# Run tests with coverage and fail if the total is below COVERAGE_MIN (default 90%).
cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {print $$3}' | tr -d '%'); \
	awk -v c="$$total" -v min="$(COVERAGE_MIN)" 'BEGIN { if (c+0 < min+0) { printf "coverage %.1f%% is below %s%%\n", c, min; exit 1 } printf "coverage %.1f%% meets %s%% gate\n", c, min }'

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
