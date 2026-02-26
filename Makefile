.PHONY: all build test test-verbose test-cover lint fmt vet clean help

# ─── Variables ────────────────────────────────────────────────────────────────

MODULE   := $(shell go list -m)
BINARY   := bin/isoserver
CMD      := ./cmd/isoserver
COVERAGE := coverage.out

GO       := go
GOFLAGS  :=

# ─── Default ──────────────────────────────────────────────────────────────────

all: fmt vet test build

# ─── Build ────────────────────────────────────────────────────────────────────

build:
	@echo "» building $(BINARY)"
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD)

# ─── Test ─────────────────────────────────────────────────────────────────────

test:
	@echo "» running tests"
	$(GO) test $(GOFLAGS) ./...

test-verbose:
	@echo "» running tests (verbose)"
	$(GO) test $(GOFLAGS) -v ./...

test-cover:
	@echo "» running tests with coverage"
	$(GO) test $(GOFLAGS) -coverprofile=$(COVERAGE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE)

cover-html: test-cover
	@echo "» opening coverage report in browser"
	$(GO) tool cover -html=$(COVERAGE)

# ─── Code Quality ─────────────────────────────────────────────────────────────

fmt:
	@echo "» formatting"
	$(GO) fmt ./...

vet:
	@echo "» vetting"
	$(GO) vet ./...

lint:
	@echo "» linting"
	@which golangci-lint > /dev/null 2>&1 || { \
		echo "golangci-lint not found — install from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	}
	golangci-lint run ./...

# ─── Maintenance ──────────────────────────────────────────────────────────────

tidy:
	@echo "» tidying modules"
	$(GO) mod tidy

clean:
	@echo "» cleaning"
	@rm -rf bin $(COVERAGE)

# ─── Help ─────────────────────────────────────────────────────────────────────

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  all           fmt + vet + test + build (default)"
	@echo "  build         compile cmd/isoserver to bin/isoserver"
	@echo "  test          run all tests"
	@echo "  test-verbose  run all tests with -v"
	@echo "  test-cover    run tests and print coverage summary"
	@echo "  cover-html    run tests and open HTML coverage report"
	@echo "  fmt           gofmt all packages"
	@echo "  vet           go vet all packages"
	@echo "  lint          run golangci-lint (must be installed)"
	@echo "  tidy          go mod tidy"
	@echo "  clean         remove bin/ and coverage.out"
