# Makefile for bubbles-cn project
# 本 Makefile 用于管理 bubbles-cn 项目的构建、测试、依赖管理等任务

# 默认目标
.DEFAULT_GOAL := help

# Go 命令
GO := go

# 代码检查工具
GOLANGCI_LINT := golangci-lint

# 格式化工具
GOFMT := gofmt

# 项目根目录
ROOT_DIR := $(shell pwd)

# 包路径
PACKAGE := github.com/purpose168/bubbles-cn

# 组件列表
COMPONENTS := \
	cursor \
	timer \
	textinput \
	textarea \
	viewport \
	table \
	stopwatch \
	spinner \
	runeutil \
	progress \
	list \
	paginator \
	key \
	help \
	filepicker

# 帮助信息
.PHONY: help
help:
	@echo "bubbles-cn 项目的 Makefile 命令："
	@echo "  make init      - 初始化项目的模块"
	@echo "  make test      - 运行项目的测试"
	@echo "  make tidy      - 整理项目的依赖"
	@echo "  make format    - 格式化项目的代码"
	@echo "  make lint      - 检查项目的代码质量"
	@echo "  make clean     - 清理项目的构建产物"
	@echo "  make verify    - 验证项目的依赖"
	@echo "  make deptree   - 显示项目的依赖树"
	@echo "  make build     - 构建项目"
	@echo "  make help      - 显示此帮助信息"
	@echo ""
	@echo "可用的组件："
	@for component in $(COMPONENTS); do \
		echo "  $$component"; \
	 done

# 初始化模块
.PHONY: init
init:
	@echo "初始化 bubbles-cn 项目的模块..."
	@if [ -f "go.mod" ]; then \
		echo "go.mod 文件已存在，跳过初始化..."; \
		echo "运行 make tidy 整理依赖..."; \
		$(GO) mod tidy; \
	else \
		$(GO) mod init $(PACKAGE); \
		echo "模块初始化完成，运行 make tidy 添加依赖..."; \
		$(GO) mod tidy; \
	fi

# 运行测试
.PHONY: test
test:
	@echo "运行 bubbles-cn 项目的测试..."
	@$(GO) test -v ./...

# 整理依赖
.PHONY: tidy
tidy:
	@echo "整理 bubbles-cn 项目的依赖..."
	@$(GO) mod tidy

# 格式化代码
.PHONY: format
format:
	@echo "格式化 bubbles-cn 项目的代码..."
	@$(GOFMT) -s -w .

# 检查代码质量
.PHONY: lint
lint:
	@echo "检查 bubbles-cn 项目的代码质量..."
	@if command -v $(GOLANGCI_LINT) > /dev/null; then \
		$(GOLANGCI_LINT) run; \
	else \
		echo "警告: golangci-lint 未安装，使用 go vet 进行检查..."; \
		$(GO) vet ./...; \
	fi

# 清理构建产物
.PHONY: clean
clean:
	@echo "清理 bubbles-cn 项目的构建产物..."
	@$(GO) clean ./...

# 验证依赖
.PHONY: verify
verify:
	@echo "验证 bubbles-cn 项目的依赖..."
	@$(GO) mod verify

# 显示依赖树
.PHONY: deptree
deptree:
	@echo "显示 bubbles-cn 项目的依赖树..."
	@$(GO) list -m -u -json all | gojq -r '(.Path + " " + .Version)'

# 构建项目
.PHONY: build
build:
	@echo "构建 bubbles-cn 项目..."
	@for component in $(COMPONENTS); do \
		echo "构建 $$component 组件..."; \
		cd $(ROOT_DIR)/$$component && $(GO) build ./...; \
	 done
	@echo "项目构建完成！"
