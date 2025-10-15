.PHONY: help build build-race test test-int run-dev run-stg run-prod

# Setup these variables to be used by commands later.
# --------------------------------------------------------------------------------------------------
SERVICE_NAME         := communication-service

.EXPORT_ALL_VARIABLES:
GOPRIVATE              = github.com/hivemindd/*
CGO_ENABLED           ?= 0
CGO_CFLAGS             = -g -O2 -Wno-return-local-addr
ENV                   ?= dev 

# Sets a default for make
.DEFAULT_GOAL := help

include .envrc-local

# Inject the git commit as the version into the go binary
VERSION=$(shell git rev-parse --short HEAD)


help:; ## Output help
	@printf "%s\\n" \
		"The following targets are available:" \
		""
	@awk 'BEGIN {FS = ":.*?## "} /^[\/.%a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m - %s\n", $$1, $$2}' ${MAKEFILE_LIST}

	@printf "%s\\n" "" "" \
		"Examples:" \
		"make build" \
		"    build the service"

build: ## Build binary
	go generate ./...
	go build -ldflags="-s -w -X 'main.version=${VERSION}'"

build-docker: ## Build binary
	go generate ./...
	CGO_ENABLED=1 go build -ldflags="-s -w"

build-race: ## Build with race detector turned on.
	go generate ./...
	CGO_ENABLED=1 go build -race -ldflags="-s -w -X 'main.version=${VERSION}'"

test: ## Test all packages
	ENV=test go test ./...

test-int: ## Test all packages using special tags: make test/int
	go test --tags=integration ./... 

run: build
	./communication-service

run-local: build
	ENV=local ./communication-service

run-dev: build ## Run local instance of the service pointing to dev region using: make run/dev
	ENV=dev ./communication-service

run-stg: build ## Run local instance of the service pointing to stg region using: make run/stg
	ENV=stg ./communication-service

run-prod: build ## Run local instance of the service pointing to prod region using: make run/prod
	ENV=prod ./communication-service
	