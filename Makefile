.PHONY: install gen run signup

install:
	@echo "Installing dependencies ..."
	go mod tidy
	@echo "Dependencies installed successfully :)"

# for generating gRPC code
gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pb/app.proto

run:
	go run server/server.go

signup:
	@if [ -z ${username} ]; then \
		echo "Error: Username is required!!! Example: 'make signup username=John password=1234'"; \
		exit 1; \
	fi
	@if [ -z ${password} ]; then \
		echo "Error: Password is required!!! Example: 'make signup username=John password=1234'"; \
		exit 1; \
	fi
	@echo "Signing up..."
	go run client/client.go -username=${username} -password=${password}