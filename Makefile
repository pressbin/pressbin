.PHONY: help dev run setup-dev reset-dev check keys migrate build release release-bundle clean tidy vet test install-air

# Monorepo dev home — same layout as ~/.pressbin on consumer machines.
DEV_HOME     ?= $(CURDIR)/.pressbin-dev
DEV_CONFIG   := $(DEV_HOME)/config.yml
DEV_SITE_URL ?= http://127.0.0.1:8080
DEV_LOG_LEVEL ?= debug

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?##"}; {printf "  %-16s %s\n", $$1, $$2}'

setup-dev: ## Create .pressbin-dev (config, data/, assets/, keys) via pressbin setup
	@echo "→ $(DEV_HOME) (same layout as ~/.pressbin)"
	@go run . setup --home "$(DEV_HOME)" --site-url "$(DEV_SITE_URL)" \
		--site-title "Pressbin Dev" --host 0.0.0.0 --port 8080 --log-level "$(DEV_LOG_LEVEL)"
	@$(MAKE) keys

reset-dev: ## Wipe .pressbin-dev and run setup-dev again
	rm -rf "$(DEV_HOME)"
	@$(MAKE) setup-dev

dev: ## Live reload with Air (runs setup-dev if missing)
	@test -f "$(DEV_CONFIG)" || $(MAKE) setup-dev
	@command -v air >/dev/null 2>&1 || { echo "Air not found. Run: make install-air"; exit 1; }
	air

run: ## Run server against .pressbin-dev/config.yml
	@test -f "$(DEV_CONFIG)" || { echo "Run: make setup-dev"; exit 1; }
	go run . serve --config "$(DEV_CONFIG)"

check: ## Preflight .pressbin-dev
	@test -f "$(DEV_CONFIG)" || { echo "Run: make setup-dev"; exit 1; }
	go run . check --config "$(DEV_CONFIG)"

keys: ## Print Bruno env hints (adminKey / syncKey from key files)
	@echo "baseUrl: $(DEV_SITE_URL)"
	@test -f "$(DEV_HOME)/admin.key" && echo "adminKey: $$(tr -d '\n' < "$(DEV_HOME)/admin.key")" || echo "adminKey: (run make setup-dev)"
	@test -f "$(DEV_HOME)/sync.key" && echo "syncKey: $$(tr -d '\n' < "$(DEV_HOME)/sync.key")" || echo "syncKey: (run make setup-dev)"
	@echo ""
	@echo "Copy into bruno/environments/local.bru"

migrate: ## Apply migrations to .pressbin-dev database
	@test -f "$(DEV_CONFIG)" || { echo "Run: make setup-dev"; exit 1; }
	go run ./cmd/migrate --config "$(DEV_CONFIG)"

build: ## Build production binary to ./bin/pressbin
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/pressbin .

release: ## Build all platform binaries under dist/
	chmod +x scripts/build-release.sh
	./scripts/build-release.sh v0.0.0-local

release-bundle: ## Build linux/amd64 tarball with config example + sync workflow
	chmod +x .github/scripts/package-release.sh
	.github/scripts/package-release.sh v0.0.0-local linux amd64

clean: ## Remove build artifacts and .pressbin-dev
	rm -rf tmp bin dist "$(DEV_HOME)"

tidy: ## go mod tidy
	go mod tidy

vet: ## go vet
	go vet ./...

test: ## Run unit tests
	go test ./...

install-air: ## Install github.com/air-verse/air (add GOPATH/bin to PATH)
	go install github.com/air-verse/air@latest
	@echo "Ensure $$(go env GOPATH)/bin is on your PATH."
