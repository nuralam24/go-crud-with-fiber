GO ?= go
GOMODCACHE ?= $(PWD)/.gomodcache
GOPATH ?= $(PWD)/.gopath
TOOLS_BIN ?= $(PWD)/.tools/bin
AIR ?= $(TOOLS_BIN)/air
SQLC ?= $(TOOLS_BIN)/sqlc
GOLANGCI_LINT ?= $(TOOLS_BIN)/golangci-lint

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run build test fmt lint migrate watch sqlc

run:
	GOMODCACHE=$(GOMODCACHE) GOPATH=$(GOPATH) $(GO) run ./cmd/api

build:
	GOMODCACHE=$(GOMODCACHE) GOPATH=$(GOPATH) $(GO) build ./...

test:
	GOMODCACHE=$(GOMODCACHE) GOPATH=$(GOPATH) $(GO) test ./...

fmt:
	GOMODCACHE=$(GOMODCACHE) GOPATH=$(GOPATH) $(GO) fmt ./...

lint:
	@mkdir -p "$(TOOLS_BIN)"
	@test -x "$(GOLANGCI_LINT)" || GOBIN="$(TOOLS_BIN)" GOMODCACHE="$(GOMODCACHE)" GOPATH="$(GOPATH)" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	GOMODCACHE=$(GOMODCACHE) GOPATH=$(GOPATH) "$(GOLANGCI_LINT)" run ./...

migrate:
	psql "$(PG_DSN)" -f migrations/001_init.sql

watch:
	@mkdir -p "$(TOOLS_BIN)"
	@test -x "$(AIR)" || GOBIN="$(TOOLS_BIN)" GOMODCACHE="$(GOMODCACHE)" GOPATH="$(GOPATH)" $(GO) install github.com/air-verse/air@latest
	@"$(AIR)" -c .air.toml

sqlc:
	@mkdir -p "$(TOOLS_BIN)"
	@test -x "$(SQLC)" || GOBIN="$(TOOLS_BIN)" GOMODCACHE="$(GOMODCACHE)" GOPATH="$(GOPATH)" $(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@"$(SQLC)" generate
