.PHONY: install gen run signup create join leave

install:
	@echo "Installing dependencies ..."
	go mod tidy
	@echo "Dependencies installed successfully :)"

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
		echo "Error: Password is required!!! Example: 'make client username=John password=1234'"; \
		exit 1; \
	fi
	@echo "Signing up..."
	go run client/client.go -username=${username} -password=${password} -command=login

create:
	if [ -z ${server} ]; then \
		echo "Error: Server name is required!!! Example: 'make create server=gochat'"; \
		exit 1; \
	fi
	@echo "Creating chat server..."
	go run client/client.go -server=${server} -command=create

join:
	if [ -z ${server} ]; then \
		echo "Error: Server name is required!!! Example: 'make join server=gochat'"; \
		exit 1; \
	fi
	@echo "Joining chat server..."
	go run client/client.go -server=${server} -command=join

leave:
	if [ -z ${server} ]; then \
		echo "Error: Server name is required!!! Example: 'make leave server=gochat'"; \
		exit 1; \
	fi
	@echo "Leaving chat server..."
	go run client/client.go -server=${server} -command=leave

send:
	if [ -z ${server} ]; then \
		echo "Error: Server name is required!!! Example: 'make send server=gochat message=Hello'"; \
		exit 1; \
	fi
	if [ -z ${message} ]; then \
		echo "Error: Message is required!!! Example: 'make send server=gochat message=Hello'"; \
		exit 1; \
	fi
	@echo "Sending message..."
	go run client/client.go -server=${server} -message=${message} -command=send

list:
	if [ -z ${server} ]; then \
		echo "Error: Server name is required!!! Example: 'make leave server=gochat'"; \
		exit 1; \
	fi
	@echo "Listing messages in server..."
	go run client/client.go -server=${server} -command=list

help:
	@echo "Description                       | Command"
	@echo "==================================|=========================================="
	@echo "Install dependencies              | make install"
	@echo "Run the chat server               | make run"
	@echo "Signup a user                     | make signup username=John password=1234"
	@echo "Create a chat server              | make create server=gochat"
	@echo "Join a chat server                | make join server=gochat"
	@echo "Leave a chat server               | make leave server=gochat"
	@echo "Send a message to a chat server   | make send server=gochat message=Hello"
	@echo "List messages in a chat server    | make list server=gochat"