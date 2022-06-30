#!/usr/bin/env bash

# Generate the Go-language protocol buffers and gRPC code.
# NOTE! This should be run from the root directory of the project.
#
# You will need to install the protocol buffers binaries and tools:
#
## Install protobuf compiler:
# brew update
# brew install protobuf
# protoc --version
#
## Install Go plugins for protoc:
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#
## Update PATH so the `protoc` compiler can find the plugins:
# export PATH="$PATH:$(go env GOPATH)/bin"
mkdir -p internal/api
protoc \
  --proto_path=api/proto/v1 \
  --go_out=internal/api \
  --go-grpc_out=internal/api \
  ppdd_service.proto
