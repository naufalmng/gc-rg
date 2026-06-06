# gc-rg — Makefile

VERSION := $(shell cat VERSION | tr -d '[:space:]')
ROOT := $(shell pwd)
DIST := $(ROOT)/dist
SHELL := bash

.PHONY: help version build clean test install uninstall standalone release

help: ## show this help
	@awk 'BEGIN{FS=":.*##"; printf "\n  \033[1mgc-rg v$(VERSION)\033[0m\n\n"} /^[a-zA-Z_-]+:.*##/{ printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo

version: ## print version
	@echo $(VERSION)

build: ## build dist binaries and installer
	@bash scripts/build.sh

clean: ## remove build artifacts
	@rm -rf $(DIST) bin
	@echo "  cleaned $(DIST) bin"

test: build ## run Go tests and installer smoke test
	@go test ./...
	@bash tests/installer_smoke.sh

install: build ## build + install via apt (requires Linux root)
	@sudo bash $(DIST)/gc-rg.sh install --yes

uninstall: ## remove via apt
	@sudo apt-get remove -y gc-rg

standalone: build ## create ./gc-rg-standalone using release assets
	@bash $(DIST)/gc-rg.sh standalone --yes --force

release: ## tag current VERSION and push
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "tagged v$(VERSION); GitHub Actions will publish the release"
