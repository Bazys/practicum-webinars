#!/bin/sh

protoc -I ./proto --go_out=./service --go_opt=paths=source_relative \
  --validate_out="lang=go:./service" --validate_opt=paths=source_relative \
  --go-grpc_out=./service --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./service --grpc-gateway_opt=paths=source_relative \
  proto/service.proto
