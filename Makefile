# Makefile for zstack-cli

BIN_NAME := zstack-cli
SRC := main.go
OS := $(shell uname | tr '[:upper:]' '[:lower:]')

.PHONY: all build install uninstall clean

all: build


build:
	@echo "Checking Go module..."
	@test -f go.mod || (echo "go.mod not found, please run 'go mod init' first" && exit 1)
	@echo "Building $(BIN_NAME)..."
	go build -o $(BIN_NAME) $(SRC)


install: build
	@echo "Running install.sh..."
	@chmod +x ./install.sh
	@./install.sh


uninstall:
	@echo "Running uninstall.sh..."
	@chmod +x ./uninstall.sh
	@./uninstall.sh


clean:
	@echo "Cleaning..."
	@rm -f $(BIN_NAME)

.PHONY: test-smoke
# Local runner tests only; never contacts a cloud environment.
test-smoke:
	$(PYTHON) -m unittest discover -s scripts -p 'test_*.py' -v

GO ?= go
PYTHON ?= python3
TEST_TIMEOUT ?= 5m

.PHONY: test test-race vet check-build verify

test:
	$(GO) test -count=1 -timeout $(TEST_TIMEOUT) ./...

test-race:
	$(GO) test -race -count=1 -timeout $(TEST_TIMEOUT) ./...

vet:
	$(GO) vet ./...

check-build:
	$(GO) build ./...

# CI and developers share these checks; no live environment credentials required.
verify: vet test-race test-smoke check-build

# Generate user/agent references using only CLI --help; no cloud access.
.PHONY: docs docs-check
docs: build
	$(PYTHON) scripts/generate_cli_docs.py --cli ./$(BIN_NAME)

docs-check: build
	$(PYTHON) scripts/generate_cli_docs.py --cli ./$(BIN_NAME) --check

.PHONY: skill-check skill-package
skill-check: build
	$(PYTHON) skills/zstack-operations/scripts/check_cli.py --cli ./$(BIN_NAME) --profile query
	$(PYTHON) skills/zstack-operations/scripts/check_cli.py --cli ./$(BIN_NAME) --profile lifecycle
	$(PYTHON) skills/zstack-operations/scripts/check_cli.py --cli ./$(BIN_NAME) --profile iam
	$(PYTHON) -m unittest discover -s scripts -p 'test_skill_check.py' -v

skill-package:
	mkdir -p dist
	COPYFILE_DISABLE=1 tar --exclude='._*' --exclude='__pycache__' --exclude='*.pyc' -czf dist/zstack-operations-skill.tar.gz -C skills zstack-operations
