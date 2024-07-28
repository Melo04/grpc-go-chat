package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/Melo04/grpc-chat/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedChatServerServer
	mu       sync.Mutex
	servers  map[string]*ChatServer
	messages map[string][]*pb.Message
}

type ChatServer struct {
	ID   string
	Name string
}

func NewServer() *server {
	return &server{
		servers:  make(map[string]*ChatServer),
		messages: make(map[string][]*pb.Message),
	}
}

func (s *server) CreateChatServer(ctx context.Context, req *pb.CreateChatServerRequest) (*pb.CreateChatServerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//generate server id dynamically
	serverID := uuid.New().String()

	chatServer := &ChatServer{
		ID:   serverID,
		Name: req.GetServerName(),
	}

	s.servers[serverID] = chatServer

	return &pb.CreateChatServerResponse{ServerId: serverID}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChatServerServer(grpcServer, NewServer())

	log.Println("Starting server on port", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
