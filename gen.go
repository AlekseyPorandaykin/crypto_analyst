package main

//go:generate protoc --go_out=./internal/client/loader/specification  --go-grpc_out=./internal/client/loader/specification  api/grpc/EventService.proto
