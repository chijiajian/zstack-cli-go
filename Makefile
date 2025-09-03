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
