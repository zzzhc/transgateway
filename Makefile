# 项目名称
APP_NAME=transgateway
VERSION=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD)

# Go 相关配置
GO=go
GOFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 目标平台
PLATFORMS=windows linux darwin
ARCHITECTURES=amd64 arm64

# 默认目标
.PHONY: all
all: build

# 安装依赖
.PHONY: deps
deps:
	$(GO) mod download

# 清理编译文件
.PHONY: clean
clean:
	rm -rf build/
	rm -f $(APP_NAME)

# 构建所有平台
.PHONY: build-all
build-all: $(foreach platform,$(PLATFORMS),$(foreach arch,$(ARCHITECTURES),build-$(platform)-$(arch)))

# 构建特定平台
define build-platform
.PHONY: build-$(1)-$(2)
build-$(1)-$(2):
	@echo "Building for $(1)/$(2)..."
	@mkdir -p build/$(1)/$(2)
	@GOOS=$(1) GOARCH=$(2) $(GO) build $(GOFLAGS) -o build/$(1)/$(2)/$(APP_NAME)$(if $(filter windows,$(1)),.exe,) .
endef

# 为每个平台和架构生成构建规则
$(foreach platform,$(PLATFORMS),$(foreach arch,$(ARCHITECTURES),$(eval $(call build-platform,$(platform),$(arch)))))

# 构建当前平台
.PHONY: build
build: deps
	$(GO) build $(GOFLAGS) -o $(APP_NAME)$(if $(filter windows,$(shell go env GOOS)),.exe,) .

# 运行测试
.PHONY: test
test:
	$(GO) test -v ./...

# 运行
.PHONY: run
run: build
	./$(APP_NAME)$(if $(filter windows,$(shell go env GOOS)),.exe,)

# 帮助信息
.PHONY: help
help:
	@echo "可用命令:"
	@echo "  make all          - 构建当前平台"
	@echo "  make build-all    - 构建所有平台"
	@echo "  make build        - 构建当前平台"
	@echo "  make clean        - 清理编译文件"
	@echo "  make deps         - 安装依赖"
	@echo "  make test         - 运行测试"
	@echo "  make run          - 运行程序"
	@echo "  make help         - 显示帮助信息" 