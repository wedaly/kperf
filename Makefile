COMMANDS=kperf

BINARIES=$(addprefix bin/,$(COMMANDS))

# default recipe is build
.DEFAULT_GOAL := build

# Always build
ALWAYS:

bin/%: cmd/% ALWAYS
	@go build -o $@ ./$<

build: $(BINARIES) ## build binaries
	@echo "$@"

test: ## run test
	@go test -v ./...

lint: ## run lint
	@golangci-lint run --config .golangci.yml

.PHONY: clean
clean: ## clean up binaries
	@rm -f $(BINARIES)

.PHONY: help
help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-36s\033[0m%s\n", $$1, $$2}' $(MAKEFILE_LIST)
