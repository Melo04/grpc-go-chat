## gRPC Chat Server Golang
This is a real-time chat server built using gRPC and Golang. It serves as a fundamental project for me to learn gRPC, concurrency, mutex and channels.
The concept of the chat server is inspired by Discord, allowing users to login, create servers, join servers, send messages, and engage in real-time communication. 

### Features
- Unary RPC: Login, CreateChatServer, JoinChatServer, LeaveChatServer, CreateChannel
- Server-side streaming RPC: SendMessages
- Client-side streaming RPC: ListMessages
- Bidirectional streaming RPC: Chat (Send and Receive messages)

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
make run
```
4. Open a new terminal and start up the chat server by signing up with your username and password.
```bash
make signup username=<your_username> password=<your_password>
```