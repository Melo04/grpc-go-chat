package client

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	pb "github.com/Melo04/grpc-chat/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr         = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

func createChatServer(client pb.ChatServerClient, serverName string) string {
	resp, err := client.CreateChatServer(context.Background(), &pb.CreateChatServerRequest{ServerName: serverName})
	if err != nil {
		log.Fatalf("failed to create chat server: %v", err)
	}
	log.Printf("Chat server created with id: %s", resp.ServerId)
	return resp.ServerId
}

func joinChatServer(client pb.ChatServerClient, serverID, username string) {
	resp, err := client.JoinChatServer(context.Background(), &pb.JoinChatServerRequest{
		ServerId: serverID,
		Username: username,
	})
	if err != nil {
		log.Fatalf("failed to join chat server: %v", err)
	}
	log.Printf("User %s: ", resp.WelcomeMessage)
}

func leaveChatServer(client pb.ChatServerClient, serverID, username string) {
	resp, err := client.LeaveChatServer(context.Background(), &pb.LeaveChatServerRequest{
		ServerId: serverID,
		Username: username,
	})
	if err != nil {
		log.Fatalf("failed to leave chat server: %v", err)
	}
	log.Printf("Left server: %s", resp.GoodbyeMessage)
}

func listMessages(client pb.ChatServerClient, serverID, channelId string) {
	stream, err := client.ListMessages(context.Background(), &pb.ListMessagesRequest{
		ServerId: serverID,
		ChannelId: channelId,
	})
	if err != nil {
		log.Fatalf("failed to list messages: %v", err)
	}
	
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error receiving message: %v", err)
		}
		log.Printf("%s ==> Message from %s: %s", msg.Timestamp, msg.Username, msg.Text)
	}
}

func sendMessages(client pb.ChatServerClient, serverID, channelID, username, text string) {
	stream, err := client.SendMessages(context.Background())
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := stream.Send(&pb.SendMessageRequest{
			ServerId: serverID,
			ChannelId: channelID,
			Username: username,
			Text: text,
		}); err != nil {
			log.Fatalf("failed to send message: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive response: %v", err)
	}
	log.Printf("Messages sent: %d", reply.MessageCount)
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServerClient(conn)

	// Create a chat server
	serverID := createChatServer(client, "chat-server")

	// Join the chat server
	joinChatServer(client, serverID, "user1")

	// List messages
	go listMessages(client, serverID, "general")
	go sendMessages(client, serverID, "general", "user1", "Hello, World!")
}