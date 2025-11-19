.PHONY: help build run test clean docker docker-build docker-run docker-stop lint fmt vet

# 默认目标
help:
	@echo "MoviePilot Go 开发工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build        构建应用"
	@echo "  run          运行应用"
	@echo "  test         运行测试"
	@echo "  test-cover   运行测试并生成覆盖率报告"
	@echo "  lint         代码检查"
	@echo "  fmt          格式化代码"
	@echo "  vet          静态分析"
	@echo "  clean        清理构建文件"
	@echo "  docker       构建Docker镜像"
	@echo "  docker-run   运行Docker容器"
	@echo "  docker-stop  停止Docker容器"
	@echo "  dev          启动开发环境"
	@echo "  prod         启动生产环境"
	@echo "  migrate      数据库迁移"
	@echo "  swagger      生成API文档"

# 变量定义
APP_NAME := moviepilot-go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# 构建应用
build:
	@echo "构建 $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/server/main.go

# 运行应用
run:
	@echo "运行 $(APP_NAME)..."
	go run cmd/server/main.go

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 运行测试并生成覆盖率报告
test-cover:
	@echo "运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 代码检查
lint:
	@echo "运行代码检查..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	goimports -w .

# 静态分析
vet:
	@echo "运行静态分析..."
	go vet ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

# 构建Docker镜像
docker:
	@echo "构建Docker镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# 运行Docker容器
docker-run:
	@echo "运行Docker容器..."
	docker-compose -f deployments/docker-compose.yml up -d

# 停止Docker容器
docker-stop:
	@echo "停止Docker容器..."
	docker-compose -f deployments/docker-compose.yml down

# 启动开发环境
dev:
	@echo "启动开发环境..."
	docker-compose -f deployments/docker-compose.dev.yml up -d

# 启动生产环境
prod:
	@echo "启动生产环境..."
	docker-compose -f deployments/docker-compose.prod.yml up -d

# 数据库迁移
migrate:
	@echo "运行数据库迁移..."
	go run cmd/migrate/main.go up

# 生成API文档
swagger:
	@echo "生成API文档..."
	swag init -g cmd/server/main.go -o docs/swagger

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# 生成mock文件
mock:
	@echo "生成mock文件..."
	go generate ./...

# 性能测试
bench:
	@echo "运行性能测试..."
	go test -bench=. -benchmem ./...

# 安全扫描
security:
	@echo "运行安全扫描..."
	gosec ./...

# 依赖检查
deps:
	@echo "检查依赖..."
	go list -u -m all

# 更新依赖
update-deps:
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy

# 生成protobuf文件
proto:
	@echo "生成protobuf文件..."
	protoc --go_out=. --go-grpc_out=. shared/proto/*.proto

# 完整的CI流程
ci: fmt vet lint test security
	@echo "CI流程完成"

# 发布准备
release-prep: clean fmt vet lint test docker
	@echo "发布准备完成"

# 开发环境设置
setup-dev:
	@echo "设置开发环境..."
	cp configs/config.yaml.sample configs/config.yaml
	$(MAKE) install-tools
	$(MAKE) dev
	@echo "开发环境设置完成"