package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	pb "github.com/Melo04/grpc-chat/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
	username   = flag.String("username", "", "username for login")
	password   = flag.String("password", "", "password for login")
)

var serverIDMap = make(map[string]string)

func createChatServer(ctx context.Context, client pb.ChatServerClient, serverName string) string {
	resp, err := client.CreateChatServer(ctx, &pb.CreateChatServerRequest{ServerName: serverName})
	if err != nil {
		log.Fatalf("Failed to create chat server: %v", err)
	}
	serverIDMap[serverName] = resp.ServerId
	log.Printf("Chat server created with id: %s", resp.ServerId)
	return resp.ServerId
}

func joinChatServer(ctx context.Context, client pb.ChatServerClient, serverID, username string) {
	resp, err := client.JoinChatServer(ctx, &pb.JoinChatServerRequest{
		ServerId: serverID,
		Username: username,
	})
	if err != nil {
		log.Fatalf("Failed to join chat server: Server does not exist")
	}
	log.Printf("User %s", resp.WelcomeMessage)
}

func leaveChatServer(ctx context.Context, client pb.ChatServerClient, serverID, username string) {
	resp, err := client.LeaveChatServer(ctx, &pb.LeaveChatServerRequest{
		ServerId: serverID,
		Username: username,
	})
	if err != nil {
		log.Fatalf("Failed to leave chat server: %v", err)
	}
	log.Printf("Left server: %s", resp.GoodbyeMessage)
}

func createChannel(ctx context.Context, client pb.ChatServerClient, serverID, channelName string) {
	resp, err := client.CreateChannel(ctx, &pb.CreateChannelRequest{
		ServerId:    serverID,
		ChannelName: channelName,
	})
	if err != nil {
		log.Fatalf("Failed to create channel: %v", err)
	}
	log.Printf("Channel created with id: %s", resp.ChannelId)
}

func listMessages(ctx context.Context, client pb.ChatServerClient, serverID, channelId string) {
	stream, err := client.ListMessages(ctx, &pb.ListMessagesRequest{
		ServerId:  serverID,
		ChannelId: channelId,
	})
	if err != nil {
		log.Fatalf("Failed to list messages: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error receiving message: %v", err)
		}
		log.Printf("Message from %s: %s", msg.Username, msg.Text)
	}
}

func getServerIDByName(serverName string) string {
	return serverIDMap[serverName]
}

func sendMessages(ctx context.Context, client pb.ChatServerClient, serverID, channelID, username, text string) {
	stream, err := client.SendMessages(ctx)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	if err := stream.Send(&pb.SendMessageRequest{
		ServerId:  serverID,
		ChannelId: channelID,
		Username:  username,
		Text:      text,
	}); err != nil {
		log.Fatalf("failed to send message: %v", err)
	}
	time.Sleep(1 * time.Second)

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to receive response: %v", err)
	}
	log.Printf("Messages sent: %d", reply.MessageCount)
}

func chat(ctx context.Context, client pb.ChatServerClient, serverID, channelID, username string) {
	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("failed to start chat: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	messageChan := make(chan struct{})
	quitChan := make(chan struct{})

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("Enter message (enter q to stop): ")
			scanner.Scan()
			text := scanner.Text()
			if text == "q" {
				close(quitChan)
				break
			}
			if err := stream.Send(&pb.ChatMessage{
				ServerId:  serverID,
				ChannelId: channelID,
				Username:  username,
				Text:      text,
			}); err != nil {
				log.Fatalf("failed to send message: %v", err)
			}
			//Signal that a message was sent
			messageChan <- struct{}{}
		}

		stream.CloseSend()
		close(messageChan)
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-messageChan:
				continue
			case <-quitChan:
				return
			default:
				msg, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("error receiving message: %v", err)
				}
				log.Printf("Message received from %s: %s", msg.Username, msg.Text)
			}
			
		}
	}()

	wg.Wait()
}

func main() {
	flag.Parse()

	if *username == "" {
		log.Fatalf("Please provide a username with -username flag")
		os.Exit(1)
	}

	if *password == "" {
		log.Fatalf("Please provide a password with -password flag")
		os.Exit(1)
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServerClient(conn)

	//Login
	res, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: *username,
		Password: *password,
	})
	if err != nil {
		log.Fatalf("failed to login: %v", err)
	}
	log.Printf("Login response: %s", res.GetMessage())

	// token for authenticated requests
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", res.GetToken()))

	for {
		fmt.Println("=====> gRPC Chat Server <===== \n1. create server\n2. join server\n3. leave server\n4. create channels\n5. list messages\n6. send message\n7. chat (send and receive messages)\n8. exit\nEnter number to activate command: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		command, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Please enter a valid integer")
		}

		switch command {
		case 1:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			createChatServer(ctx, client, serverName)
		case 2:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			serverID := getServerIDByName(serverName)
			joinChatServer(ctx, client, serverID, *username)
		case 3:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			serverID := getServerIDByName(serverName)
			leaveChatServer(ctx, client, serverID, *username)
		case 4:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			serverID := getServerIDByName(serverName)
			fmt.Println("Enter channel name: ")
			scanner.Scan()
			channelName := scanner.Text()
			createChannel(ctx, client, serverID, channelName)
		case 5:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			fmt.Println("Enter channel name: ")
			scanner.Scan()
			channelName := scanner.Text()
			serverID := getServerIDByName(serverName)
			listMessages(ctx, client, serverID, channelName)
		case 6:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			serverID := getServerIDByName(serverName)
			fmt.Println("Enter channel name: ")
			scanner.Scan()
			channelName := scanner.Text()
			// send multiple messages to the same channel
			for {
				fmt.Print("Enter message (enter q to stop): ")
				scanner.Scan()
				message := scanner.Text()
				if message == "q" {
					break
				}
				sendMessages(ctx, client, serverID, channelName, *username, message)
			}
		case 7:
			fmt.Println("Enter server name: ")
			scanner.Scan()
			serverName := scanner.Text()
			serverID := getServerIDByName(serverName)
			fmt.Println("Enter channel name: ")
			scanner.Scan()
			channelName := scanner.Text()
			chat(ctx, client, serverID, channelName, *username)
		case 8:
			fmt.Println("Exiting...")
			os.Exit(0)
			break
		default:
			fmt.Println("Invalid command")
		}
	}
}
