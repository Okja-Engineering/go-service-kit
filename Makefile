VERSION ?= 0.1.0

REPO_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

GOLINT_PATH := $(REPO_DIR)/.tools/golangci-lint
AIR_PATH := $(REPO_DIR)/.tools/air
MDOX_PATH := $(REPO_DIR)/.tools/mdox

.PHONY: help lint lint-fix clean test clean

.DEFAULT_GOAL := help

help:
	@figlet $@ || true
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools:
	@figlet $@ || true
	@$(GOLINT_PATH) > /dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./.tools
	@$(AIR_PATH) -v > /dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b ./.tools
	@$(MDOX_PATH) -v > /dev/null 2>&1 || GOBIN=$(REPO_DIR)/.tools go install github.com/bwplotka/mdox@latest

lint: install-tools
	@figlet $@ || true
	$(GOLINT_PATH) run

lint-fix: install-tools
	@figlet $@ || true
	$(GOLINT_PATH) run --fix

test:
	@figlet $@ || true
	go test -v -count=1 ./pkg/...

docs: install-tools
	@echo "Generating docs with mdox..."
	mdox execgen ./pkg/*/README.md

clean:
	rm -f server
