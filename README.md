## gRPC Chat Server Golang
This is a real-time chat server built using gRPC and Golang. It serves as a fundamental project for me to learn gRPC, concurrency, mutex and channels. The concept of the chat server is similar to Discord, allowing users to create servers, join servers, send messages, and engage in real-time communication.

### Features
- Unary RPC: Login, CreateChatServer, JoinChatServer, LeaveChatServer
- Server-side streaming RPC: SendMessages
- Client-side streaming RPC: ListMessages
- Bidirectional streaming RPC: Chat

### How to run
1. Clone the repository
```bash
git clone https://github.com/Melo04/grpc-go-chat.git
```
2. Install the dependencies
```bash
make install
```
3. Run the server
```bash
make server
```
4. Open a new terminal and run the client
```bash
make client
```