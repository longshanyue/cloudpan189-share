PROJECT_NAME=cloudpan189-share
MODULE_NAME=github.com/xxcheng123/cloudpan189-share
VAR_COMMIT ?= $(shell git rev-parse HEAD)
VAR_BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VAR_GIT_SUMMARY ?= $(shell git describe --tags --dirty --always)
VAR_GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

# 输出配置
OUTPUT_DIR=output
BINARY_NAME=share
DOCKER_IMAGE=$(PROJECT_NAME):latest

.PHONY: build build-frontend build-backend build-multi-arch clean clean-all
.PHONY: docker-build docker-run docker-stop docker-clean docker-logs
.PHONY: dev test lint help

# 主构建目标
build: build-frontend build-backend
	@echo "✅ Build completed successfully!"

# 前端构建
build-frontend:
	@echo "🎨 Building frontend..."
	@if [ -d "fe" ]; then \
		cd fe && npm install && npm run build; \
		echo "✅ Frontend build completed"; \
	else \
		echo "⚠️  Frontend directory not found, skipping..."; \
	fi

# 后端构建
build-backend:
	@echo "🔨 Building backend..."
	@mkdir -p $(OUTPUT_DIR)
	go mod tidy
	GOOS=linux GOARCH=amd64 go build \
		-ldflags="-X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "✅ Backend build completed: $(OUTPUT_DIR)/$(BINARY_NAME)"

# 多架构构建
build-multi-arch:
	@echo "🔨 Building for multiple architectures..."
	@mkdir -p $(OUTPUT_DIR)
	go mod tidy
	@echo "📦 Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w -X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/main.go
	@echo "📦 Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build \
		-ldflags="-s -w -X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/main.go
	@echo "📦 Building for Linux ARMv7a..."
	GOOS=linux GOARCH=arm GOARM=7 go build \
		-ldflags="-s -w -X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-armv7a ./cmd/main.go
	@echo "📦 Building for Windows AMD64..."
	GOOS=windows GOARCH=amd64 go build \
		-ldflags="-s -w -X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/main.go
	@echo "📦 Building for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 go build \
		-ldflags="-s -w -X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/main.go
	@echo "📦 Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 go build \
		-ldflags="-s -w -X $(MODULE_NAME)/configs.Commit=$(VAR_COMMIT) \
		          -X $(MODULE_NAME)/configs.BuildDate=$(VAR_BUILD_DATE) \
		          -X $(MODULE_NAME)/configs.GitSummary=$(VAR_GIT_SUMMARY) \
		          -X $(MODULE_NAME)/configs.GitBranch=$(VAR_GIT_BRANCH)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/main.go
	@echo "✅ Multi-architecture build completed!"
	@ls -la $(OUTPUT_DIR)/

# 清理构建产物
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@if [ -d "fe" ]; then rm -rf fe/dist fe/node_modules; fi

# 清理 linter 缓存
lint-clean:
	@echo "🧹 Cleaning linter cache..."
	golangci-lint cache clean

# Docker 构建
docker-build:
	@echo "🐳 Building Docker image..."
	docker build \
		--build-arg MODULE_NAME=$(MODULE_NAME) \
		--build-arg VAR_COMMIT=$(VAR_COMMIT) \
		--build-arg VAR_BUILD_DATE=$(VAR_BUILD_DATE) \
		--build-arg VAR_GIT_SUMMARY=$(VAR_GIT_SUMMARY) \
		--build-arg VAR_GIT_BRANCH=$(VAR_GIT_BRANCH) \
		-t $(DOCKER_IMAGE) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE)"

# 运行 Docker 容器
docker-run: docker-stop
	@echo "🚀 Starting Docker container..."
	docker run -d \
		-p 12395:12395 \
		--name $(PROJECT_NAME) \
		$(DOCKER_IMAGE)
	@echo "✅ Container started: http://localhost:12395"

# 停止 Docker 容器
docker-stop:
	@echo "🛑 Stopping Docker container..."
	@docker stop $(PROJECT_NAME) 2>/dev/null || true
	@docker rm $(PROJECT_NAME) 2>/dev/null || true

# 查看 Docker 日志
docker-logs:
	@echo "📋 Docker container logs:"
	docker logs -f $(PROJECT_NAME)

# 清理 Docker 资源
docker-clean: docker-stop
	@echo "🧹 Cleaning Docker resources..."
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@docker image prune -f

# 完整清理
clean-all: clean docker-clean lint-clean
	@echo "✅ Complete cleanup finished!"

# 开发模式
dev:
	@echo "🔧 Starting development server..."
	go run ./cmd/main.go

# 运行测试
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# 运行 linter
lint:
	@echo "🔍 Running linter..."
	golangci-lint run

# 显示构建信息
info:
	@echo "📊 Build Information:"
	@echo "  Project: $(PROJECT_NAME)"
	@echo "  Module:  $(MODULE_NAME)"
	@echo "  Commit:  $(VAR_COMMIT)"
	@echo "  Date:    $(VAR_BUILD_DATE)"
	@echo "  Summary: $(VAR_GIT_SUMMARY)"
	@echo "  Branch:  $(VAR_GIT_BRANCH)"
	@echo "  Output:  $(OUTPUT_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "🏗️ Supported Architectures:"
	@echo "  - linux/amd64"
	@echo "  - linux/arm64"
	@echo "  - linux/arm/v7 (ARMv7a)"
	@echo "  - windows/amd64"
	@echo "  - darwin/amd64"
	@echo "  - darwin/arm64"

# 帮助信息
help:
	@echo "🚀 Available commands:"
	@echo ""
	@echo "📦 Build Commands:"
	@echo "  build           - Build frontend and backend"
	@echo "  build-frontend  - Build frontend only"
	@echo "  build-backend   - Build backend only"
	@echo "  build-multi-arch - Build for multiple architectures"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-stop     - Stop Docker container"
	@echo "  docker-logs     - Show container logs"
	@echo "  docker-clean    - Clean Docker resources"
	@echo ""
	@echo "🧹 Cleanup Commands:"
	@echo "  clean           - Clean build artifacts"
	@echo "  lint-clean      - Clean linter cache"
	@echo "  clean-all       - Complete cleanup"
	@echo ""
	@echo "🔧 Development Commands:"
	@echo "  dev             - Start development server"
	@echo "  test            - Run tests"
	@echo "  lint            - Run linter"
	@echo "  info            - Show build information"
	@echo ""

# 默认目标
.DEFAULT_GOAL := help