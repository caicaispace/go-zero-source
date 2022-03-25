.PHONY: build clean tool lint help

all: proto-gen

# proto gen
proto-gen:
	@cd ./proto && protoc --go_out=plugins=grpc:. balancer.proto