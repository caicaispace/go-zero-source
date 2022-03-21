.PHONY: build clean tool lint help

all: build

# analyse
vet:
	@go vet ./breaker/...; true
	@go vet ./shedding/...; true
	@go vet ./limit/...; true
	@go vet ./balancer/...; true

# fotmat
# go install mvdan.cc/gofumpt@latest
fmt:
	@gofumpt -l -w ./breaker/; true
	@gofumpt -l -w ./shedding/; true
	@gofumpt -l -w ./limit/; true
	@gofumpt -l -w ./balancer/; true

clean:
	go clean -i .

help:
	@echo "make vet: run specified go vet"
	@echo "make fmt: gofumpt -l -w ./service/"
	@echo "make clean: remove object files and cached files"