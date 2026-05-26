.PHONY: help dev run migrate build release release-bundle clean tidy vet install-air

CONFIG ?= config.yml

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?##"}; {printf "  %-14s %s\n", $$1, $$2}'

dev: ## Run with Air (live reload on .go / embedded assets / config schema changes)
	@command -v air >/dev/null 2>&1 || { echo "Air not found. Run: make install-air"; exit 1; }
	air

run: ## Run without Air
	go run . --config $(CONFIG)

migrate: ## Apply pending SQL migrations (uses DB path from config; does not start the server)
	go run ./cmd/migrate --config $(CONFIG)

build: ## Build production binary to ./bin/pressbin
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/pressbin .

release: ## Build all platform binaries under dist/
	chmod +x scripts/build-release.sh
	./scripts/build-release.sh v0.0.0-local

release-bundle: ## Build linux/amd64 tarball with config example + sync workflow
	chmod +x .github/scripts/package-release.sh
	.github/scripts/package-release.sh v0.0.0-local linux amd64

clean: ## Remove Air tmp dir, bin/, dist/, and local DB artifact
	rm -rf tmp bin dist
	rm -f pressbin.db pressbin.db-shm pressbin.db-wal

tidy: ## go mod tidy
	go mod tidy

vet: ## go vet
	go vet ./...

install-air: ## Install github.com/air-verse/air (add GOPATH/bin to PATH)
	go install github.com/air-verse/air@latest
	@echo "Ensure $$(go env GOPATH)/bin is on your PATH."
