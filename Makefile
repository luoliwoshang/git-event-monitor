.PHONY: build test test-github test-gitee clean help

# 默认目标
help:
	@echo "Available targets:"
	@echo "  build       - Build the CLI tool"
	@echo "  test        - Run all tests"
	@echo "  test-github - Run GitHub API tests"
	@echo "  test-gitee  - Run Gitee API tests"
	@echo "  clean       - Clean build artifacts"

# 构建 CLI 工具
build:
	@echo "Building git-event-monitor..."
	@go build -o bin/git-event-monitor ./cmd/git-event-monitor

# 运行所有测试
test:
	@echo "Running all tests..."
	@go test -v ./internal/...

# 运行 GitHub 测试
test-github:
	@echo "Running GitHub API tests..."
	@go test -v ./internal/api/github

# 运行 Gitee 测试
test-gitee:
	@echo "Running Gitee API tests..."
	@go test -v ./internal/api/gitee

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf bin/

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 检查代码
lint:
	@echo "Running go vet..."
	@go vet ./...