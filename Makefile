# Makefile for zstack-cli

BIN_NAME := zstack-cli
SRC := main.go
OS := $(shell uname | tr '[:upper:]' '[:lower:]')

.PHONY: all build install uninstall clean

all: build

# 构建二进制
build:
	@echo "Checking Go module..."
	@test -f go.mod || (echo "go.mod not found, please run 'go mod init' first" && exit 1)
	@echo "Building $(BIN_NAME)..."
	go build -o $(BIN_NAME) $(SRC)

# 安装 CLI 和补全
install: build
	@echo "Running install.sh..."
	@chmod +x ./install.sh
	@./install.sh

# 卸载 CLI 和补全
uninstall:
	@echo "Running uninstall.sh..."
	@chmod +x ./uninstall.sh
	@./uninstall.sh

# 清理生成的二进制
clean:
	@echo "Cleaning..."
	@rm -f $(BIN_NAME)
