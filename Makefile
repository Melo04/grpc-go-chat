.PHONY: install gen server client

install:
	go mod tidy

gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/app.proto

server:
	go run server/server.go

client:
	go run client/client.go