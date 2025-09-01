VERSION ?= 0.2.0

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

version: ## Show current version
	@echo "Current version: $(VERSION)"

release-patch: ## Create a patch release (0.1.0 -> 0.1.1)
	@echo "Creating patch release..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "✅ Released v$(VERSION)"

release-minor: ## Create a minor release (0.1.0 -> 0.2.0)
	@echo "Creating minor release..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "✅ Released v$(VERSION)"

release-major: ## Create a major release (1.0.0 -> 2.0.0)
	@echo "Creating major release..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "✅ Released v$(VERSION)"

next-version: ## Show what the next patch version would be
	@echo "Current: $(VERSION)"
	@echo "Next patch: $(shell echo $(VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}')"
