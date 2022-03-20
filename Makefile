.PHONY: build clean tool lint help

all: build

build:
	# @go build -v .
	@go build -ldflags="-s -w" -o tmp/main main.go; true

# 减小 Go 代码编译后的二进制体积 https://geektutu.com/post/hpg-reduce-size.html
# wsl2 安装方式如下：
# 	apt install upx
build-upx:
	@go build -ldflags="-s -w" -o tmp/main main.go && upx -9 tmp/main; true

build-escape:
	# @go build -ldflags="-s -w -X main.escape=true" -o tmp/main main.go
	@go build -gcflags=-m main.go; true

update:
	@go get -u all && go mod tidy && go mod vendor; true

update-show:
	@go list -u -m -mod=readonly all; true

vendor:
	@go mod tidy && go mod vendor; true

# 静态检测
vet:
	@go vet ./breaker/...; true
	@go vet ./shedding/...; true
	@go vet ./limit/...; true

# 静态检测 go install github.com/mgechev/revive@latest
revive:
	@revive -config revive.toml -formatter friendly ./code/service/...; true

# 格式化 go install mvdan.cc/gofumpt@latest
fmt:
	@gofumpt -l -w ./breaker/; true
	@gofumpt -l -w ./shedding/; true
	@gofumpt -l -w ./limit/; true

clean:
	go clean -i .

git-push:
	./cmd.sh git push; true

git-clear:
	./cmd.sh git clear; true

help:
	@echo "make: compile packages and dependencies"
	@echo "make vet: run specified go vet"
	@echo "make lint: golint ./..."
	@echo "make fmt: gofumpt -l -w ./service/"
	@echo "make clean: remove object files and cached files"