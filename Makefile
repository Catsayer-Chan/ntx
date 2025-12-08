# Makefile for NTX (Network Tools eXtended)
# 作者: Catsayer

# 变量定义
APP_NAME=ntx
VERSION=0.1.0
BUILD_DIR=bin
MAIN_PATH=./cmd/ntx
INTERNAL_PACKAGES=$(shell go list ./internal/...)
PKG_PACKAGES=$(shell go list ./pkg/...)
ALL_PACKAGES=$(shell go list ./...)

# 编译参数
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S') -s -w"
GCFLAGS=-gcflags "all=-trimpath=$(PWD)"
ASMFLAGS=-asmflags "all=-trimpath=$(PWD)"
BUILD_TAGS=-tags netgo
EXTRA_FLAGS=-installsuffix netgo

# 目标平台
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

# 颜色输出
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: all build build-small clean test coverage lint fmt vet deps help install uninstall cross-build

# 默认目标
all: clean deps fmt vet lint test build

# 帮助信息
help:
	@echo "$(GREEN)NTX Makefile 帮助$(NC)"
	@echo ""
	@echo "$(YELLOW)可用目标：$(NC)"
	@echo "  $(GREEN)make build$(NC)        - 编译项目"
	@echo "  $(GREEN)make build-small$(NC)  - 编译并使用 UPX 压缩（减小体积）"
	@echo "  $(GREEN)make test$(NC)         - 运行测试"
	@echo "  $(GREEN)make coverage$(NC)     - 生成测试覆盖率报告"
	@echo "  $(GREEN)make lint$(NC)         - 运行代码检查"
	@echo "  $(GREEN)make fmt$(NC)          - 格式化代码"
	@echo "  $(GREEN)make vet$(NC)          - 运行 go vet"
	@echo "  $(GREEN)make clean$(NC)        - 清理构建产物"
	@echo "  $(GREEN)make deps$(NC)         - 下载依赖"
	@echo "  $(GREEN)make install$(NC)      - 安装到系统"
	@echo "  $(GREEN)make uninstall$(NC)    - 从系统卸载"
	@echo "  $(GREEN)make cross-build$(NC)  - 交叉编译所有平台"
	@echo "  $(GREEN)make run$(NC)          - 编译并运行"
	@echo "  $(GREEN)make all$(NC)          - 执行完整的构建流程"

# 编译
build:
	@echo "$(GREEN)正在编译 $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ 编译完成: $(BUILD_DIR)/$(APP_NAME)$(NC)"
	@ls -lh $(BUILD_DIR)/$(APP_NAME)

# 编译小体积版本（使用 UPX 压缩）
build-small: build
	@echo "$(GREEN)使用 UPX 压缩二进制文件...$(NC)"
	@if command -v upx >/dev/null 2>&1; then \
		cp $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/$(APP_NAME).backup; \
		upx --best --lzma --force-macos $(BUILD_DIR)/$(APP_NAME) 2>/dev/null || \
		upx --best --lzma $(BUILD_DIR)/$(APP_NAME); \
		echo "$(GREEN)✓ 压缩完成$(NC)"; \
		echo "$(YELLOW)压缩前大小:$(NC)"; \
		ls -lh $(BUILD_DIR)/$(APP_NAME).backup; \
		echo "$(YELLOW)压缩后大小:$(NC)"; \
		ls -lh $(BUILD_DIR)/$(APP_NAME); \
		rm $(BUILD_DIR)/$(APP_NAME).backup; \
	else \
		echo "$(YELLOW)! UPX 未安装，无法压缩$(NC)"; \
		echo "$(YELLOW)  macOS 安装: brew install upx$(NC)"; \
		echo "$(YELLOW)  Linux 安装: apt-get install upx 或 yum install upx$(NC)"; \
	fi

# 交叉编译
cross-build:
	@echo "$(GREEN)开始交叉编译...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output_name=$(BUILD_DIR)/$(APP_NAME)-$$platform-$$arch; \
			if [ "$$platform" = "windows" ]; then \
				output_name=$$output_name.exe; \
			fi; \
			echo "$(YELLOW)编译 $$platform/$$arch...$(NC)"; \
			GOOS=$$platform GOARCH=$$arch go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $$output_name $(MAIN_PATH); \
			if [ $$? -eq 0 ]; then \
				echo "$(GREEN)✓ $$output_name$(NC)"; \
			else \
				echo "$(RED)✗ $$platform/$$arch 编译失败$(NC)"; \
			fi; \
		done; \
	done
	@echo "$(GREEN)✓ 交叉编译完成$(NC)"

# 运行
run: build
	@echo "$(GREEN)运行 $(APP_NAME)...$(NC)"
	@$(BUILD_DIR)/$(APP_NAME) --help

# 测试
test:
	@echo "$(GREEN)运行测试...$(NC)"
	@go test -v -race -timeout 300s $(ALL_PACKAGES)
	@echo "$(GREEN)✓ 测试完成$(NC)"

# 测试覆盖率
coverage:
	@echo "$(GREEN)生成测试覆盖率报告...$(NC)"
	@mkdir -p $(BUILD_DIR)/coverage
	@go test -v -race -coverprofile=$(BUILD_DIR)/coverage/coverage.out -covermode=atomic $(ALL_PACKAGES)
	@go tool cover -html=$(BUILD_DIR)/coverage/coverage.out -o $(BUILD_DIR)/coverage/coverage.html
	@echo "$(GREEN)✓ 覆盖率报告: $(BUILD_DIR)/coverage/coverage.html$(NC)"

# 快速测试（不包含竞态检测）
test-quick:
	@echo "$(GREEN)运行快速测试...$(NC)"
	@go test -v -timeout 60s $(ALL_PACKAGES)

# 基准测试
bench:
	@echo "$(GREEN)运行基准测试...$(NC)"
	@go test -bench=. -benchmem $(ALL_PACKAGES)

# 代码格式化
fmt:
	@echo "$(GREEN)格式化代码...$(NC)"
	@gofmt -s -w .
	@echo "$(GREEN)✓ 代码格式化完成$(NC)"

# go vet 检查
vet:
	@echo "$(GREEN)运行 go vet...$(NC)"
	@go vet $(ALL_PACKAGES)
	@echo "$(GREEN)✓ go vet 检查通过$(NC)"

# 代码检查
lint:
	@echo "$(GREEN)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ 代码检查完成$(NC)"; \
	else \
		echo "$(YELLOW)! golangci-lint 未安装，跳过代码检查$(NC)"; \
		echo "$(YELLOW)  安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

# 下载依赖
deps:
	@echo "$(GREEN)下载依赖...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ 依赖下载完成$(NC)"

# 更新依赖
deps-update:
	@echo "$(GREEN)更新依赖...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ 依赖更新完成$(NC)"

# 清理
clean:
	@echo "$(GREEN)清理构建产物...$(NC)"
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "$(GREEN)✓ 清理完成$(NC)"

# 安装到系统
install: build
	@echo "$(GREEN)安装 $(APP_NAME) 到系统...$(NC)"
	@go install $(MAIN_PATH)
	@echo "$(GREEN)✓ 安装完成$(NC)"

# 从系统卸载
uninstall:
	@echo "$(GREEN)卸载 $(APP_NAME)...$(NC)"
	@rm -f $(GOPATH)/bin/$(APP_NAME)
	@echo "$(GREEN)✓ 卸载完成$(NC)"

# 生成文档
docs:
	@echo "$(GREEN)生成文档...$(NC)"
	@mkdir -p docs/api
	@godoc -http=:6060 &
	@echo "$(GREEN)✓ 文档服务启动: http://localhost:6060$(NC)"

# 检查代码质量
check: fmt vet lint test
	@echo "$(GREEN)✓ 所有检查通过$(NC)"

# 版本信息
version:
	@echo "$(GREEN)$(APP_NAME) version $(VERSION)$(NC)"

# 开发模式（监听文件变化并重新编译）
dev:
	@echo "$(GREEN)启动开发模式（需要安装 air）...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(RED)✗ air 未安装$(NC)"; \
		echo "$(YELLOW)  安装: go install github.com/air-verse/air@latest$(NC)"; \
	fi

# Docker 构建
docker-build:
	@echo "$(GREEN)构建 Docker 镜像...$(NC)"
	@docker build -t $(APP_NAME):$(VERSION) .
	@echo "$(GREEN)✓ Docker 镜像构建完成$(NC)"