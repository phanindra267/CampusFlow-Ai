# Makefile for CampusCare AI

GO ?= go
NPM ?= npm
FRONTEND_DIR := frontend
BIN_DIR := bin

.PHONY: help run build test vet lint fmt clean \
        migrate-up migrate-down \
        frontend-install frontend-dev frontend-build frontend-lint frontend-typecheck

help:
	@echo "Backend"
	@echo "  make run               Start the API server"
	@echo "  make build             Compile the API binary into $(BIN_DIR)"
	@echo "  make test              Run the Go test suite"
	@echo "  make vet               Run go vet"
	@echo "  make lint              Run gofmt and go vet checks"
	@echo "  make fmt               Format Go sources"
	@echo "  make migrate-up        Apply pending database migrations"
	@echo "  make migrate-down      Roll back the most recent migration"
	@echo ""
	@echo "Frontend (Next.js)"
	@echo "  make frontend-install  Install frontend dependencies"
	@echo "  make frontend-dev      Start the Next.js dev server"
	@echo "  make frontend-build    Build the frontend for production"
	@echo "  make frontend-lint     Lint the frontend"
	@echo "  make frontend-typecheck Type-check the frontend"
	@echo ""
	@echo "  make clean             Remove build artefacts"

run:
	$(GO) run ./cmd/api

build:
	$(GO) build -o $(BIN_DIR)/api ./cmd/api

test:
	$(GO) test -count=1 ./cmd/... ./internal/... ./pkg/...

vet:
	$(GO) vet ./cmd/... ./internal/... ./pkg/... ./migrations/...

lint: vet
	@unformatted=$$(gofmt -l cmd internal pkg migrations); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt found unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

fmt:
	$(GO) fmt ./cmd/... ./internal/... ./pkg/... ./migrations/...

migrate-up:
	$(GO) run ./cmd/migrate -direction up

migrate-down:
	$(GO) run ./cmd/migrate -direction down -steps 1

frontend-install:
	$(NPM) --prefix $(FRONTEND_DIR) install

frontend-dev:
	$(NPM) --prefix $(FRONTEND_DIR) run dev

frontend-build:
	$(NPM) --prefix $(FRONTEND_DIR) run build

frontend-lint:
	$(NPM) --prefix $(FRONTEND_DIR) run lint

frontend-typecheck:
	$(NPM) --prefix $(FRONTEND_DIR) exec tsc -- --noEmit

clean:
	$(GO) clean
	rm -rf $(BIN_DIR)