# 主版本号，修改比较大，无法向后兼容
# todo ?=如果前面已经赋过值了，那么就使用前面赋的值
VERSION_MAJOR ?= 1
# 次版本号，新加功能，向后兼容
VERSION_MINOR ?= 0
# 相同代码只是对不同平台的不同编译
VERSION_BUILD ?= 0

# todo Makefile中变量的引用方法是(VERSION_MAJOR) ， 比 如 本 例 中 的 “ (变量名)”，比如本例中的“(变量名)”
VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)
VERSION_PACKAGE = $(REPOPATH/pkg/version)

# todo := 变量只使用已经定义好的 这是和=的区别
SHELL := /bin/bash
# todo GOOS表示目标平台  linux/windows
GOOS ?= $(shell go env GOOS)
# todo GOARCH=amd64表示64位x86
GOARCH ?= $(shell go env GOARCH)
ORG := yussgo
PROJECT := deploy
REGISTRY?=vgpu

REPOPATH ?= $(ORG)/$(PROJECT)
VERSION_PACKAGE = $(REPOPATH)/pkg/version

# todo -ldflags 'flag list' '-s -w': 压缩编译后的体积 -s: 去掉符号表 -w: 去掉调试信息，不能gdb调试了
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_LDFLAGS := '-extldflags "-static"
GO_LDFLAGS += -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -w -s # Drop debugging symbols.
GO_LDFLAGS += '

EXECUTOR_PACKAGE = $(REPOPATH)/cmd
BUILD_ARG ?=

# Force using Go Modules and always read the dependencies from
# the `vendor` folder.
export GO111MODULE = on
export GOFLAGS = -mod=vendor

# todo CGO_ENABLE=1 程序使用到了Cgo就打开是1  默认是0关闭
# todo -o 指定输出文件。
start: $(GO_FILES)
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags $(GO_LDFLAGS) -o $@ .
.PHONY: build
build:
	DOCKER_BUILDKIT=1 docker build ${BUILD_ARG} --build-arg=GOARCH=amd64 -t yu:latest -f hack/Dockerfile .
#	docker push $(REGISTRY)/executor:latest