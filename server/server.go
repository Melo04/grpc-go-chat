package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	pb "github.com/Melo04/grpc-chat/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var port = flag.Int("port", 50051, "The server port")

type server struct {
	pb.UnimplementedChatServerServer
	mu       sync.Mutex
	users    map[string]string
	servers  map[string]*ChatServer
	messages map[string][]*pb.Message
	channels map[string]map[string]string
}

type ChatServer struct {
	ID   string
	Name string
}

func NewServer() *server {
	return &server{
		servers:  make(map[string]*ChatServer),
		messages: make(map[string][]*pb.Message),
		users:    make(map[string]string),
		channels: make(map[string]map[string]string),
	}
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := uuid.New().String()
	s.users[req.GetUsername()] = token

	return &pb.LoginResponse{
		Token: token,
		Message: "Login successful",
	}, nil
}

func (s *server) CreateChatServer(ctx context.Context, req *pb.CreateChatServerRequest) (*pb.CreateChatServerResponse, error) {
	// Check if the user is authenticated
	if !s.authenticate(ctx) {
		return nil, grpc.Errorf(codes.Unauthenticated, "not authenticated")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	serverID := uuid.New().String()

	chatServer := &ChatServer{
		ID:   serverID,
		Name: req.GetServerName(),
	}

	s.servers[serverID] = chatServer

	return &pb.CreateChatServerResponse{ServerId: serverID}, nil
}

func (s *server) CreateChannel(ctx context.Context, req *pb.CreateChannelRequest) (*pb.CreateChannelResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	serverID := req.GetServerId()
	if _, exists := s.servers[req.GetServerId()]; !exists {
		return nil, grpc.Errorf(codes.NotFound, "chat server not found")
	}

	//generate channel id dynamically
	channelID := uuid.New().String()
	if s.channels[serverID] == nil {
		s.channels[serverID] = make(map[string]string)
	}

	s.channels[serverID][channelID] = req.GetChannelName()

	return &pb.CreateChannelResponse{ChannelId: channelID}, nil
}

func (s *server) JoinChatServer(ctx context.Context, req *pb.JoinChatServerRequest) (*pb.JoinChatServerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servers[req.GetServerId()]; !exists {
		return nil, grpc.Errorf(codes.NotFound, "chat server not found")
	}

	welcomeMessage := req.GetUsername() + " just slid into the server " + s.servers[req.GetServerId()].Name
	return &pb.JoinChatServerResponse{WelcomeMessage: welcomeMessage}, nil
}

func (s *server) LeaveChatServer(ctx context.Context, req *pb.LeaveChatServerRequest) (*pb.LeaveChatServerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.servers[req.GetServerId()]; !exists {
		return nil, grpc.Errorf(codes.NotFound, "chat server not found")
	}

	goodbyeMessage := req.GetUsername() + " just left the server"
	return &pb.LeaveChatServerResponse{GoodbyeMessage: goodbyeMessage}, nil
}

func (s *server) ListMessages(req *pb.ListMessagesRequest, stream pb.ChatServer_ListMessagesServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	serverID := req.GetServerId()
	if _, ok := s.servers[serverID]; !ok {
		return fmt.Errorf("chat server not found")
	}

	messages, ok := s.messages[req.GetChannelId()]
	if !ok {
		return fmt.Errorf("channel not found")
	}

	for _, message := range messages {
		if err := stream.Send(message); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) SendMessages(stream pb.ChatServer_SendMessagesServer) error {
	var messageCount int32

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.SendMessagesResponse{
				MessageCount: messageCount,
			})
		}

		if err != nil {
			return err
		}

		s.mu.Lock()
		s.messages[req.GetChannelId()] = append(s.messages[req.GetChannelId()], &pb.Message{
			Username:  req.GetUsername(),
			Text:      req.GetText(),
			Timestamp: timestamppb.Now(),
		})
		s.mu.Unlock()

		log.Printf("Message received from %s: %s", req.GetUsername(), req.GetText())
		messageCount++
	}
}

func (s *server) Chat(stream pb.ChatServer_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("\nMessage received from %s: %s", in.GetUsername(), in.GetText())

		s.mu.Lock()
		// ChatMessage doesnt have a channel id, so we use the username as the channel id
		s.messages[in.GetChannelId()] = append(s.messages[in.GetChannelId()], &pb.Message{
			Username:  in.GetUsername(),
			Text:      in.GetText(),
			Timestamp: timestamppb.Now(),
		})
		s.mu.Unlock()

		// Echo the message back to the client
		if err := stream.Send(&pb.ChatMessage{
			ServerId:  in.GetServerId(),
			ChannelId: in.GetChannelId(),
			Username:  in.GetUsername(),
			Text:      in.GetText(),
			Timestamp: timestamppb.Now(),
		}); err != nil {
			return err
		}
	}
}

func (s *server) authenticate(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	tokens := md["authorization"]
	if len(tokens) == 0 {
		return false
	}

	token := tokens[0]
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, storeToken := range s.users {
		if token == storeToken {
			return true
		}
	}

	return false
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
