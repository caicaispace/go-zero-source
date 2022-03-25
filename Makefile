.PHONY: build clean tool lint help

all: build

# analyse
vet:
	@go vet ./code/...; true

# fotmat
# go install mvdan.cc/gofumpt@latest
fmt:
	@gofumpt -l -w ./code/; true

clean:
	go clean -i .

help:
	@echo "make vet: run specified go vet"
	@echo "make fmt: gofumpt -l -w ./service/"
	@echo "make clean: remove object files and cached files"