package main

import (
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
)

// TODO: CLI client, se poveže s serverjem, demonstracija programa se nahaja tuki

func main() {
	addrPtr := flag.String("u", "localhost", "server address")
	portPtr := flag.Int("p", 12345, "server port number")

	flag.Parse()

	url := fmt.Sprintf("%s:%d", *addrPtr, *portPtr)

	fmt.Printf("Connecting to server at %s\n", url)

	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	context, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcClient := pb.NewMessageBoardClient(conn)

	fmt.Println("TESTIRANJE:")
	if _, err := grpcClient.CreateUser(context, &pb.CreateUserRequest{Name: "Test User"}); err != nil {
		panic(err)
	}
	fmt.Println("User created successfully")

	if _, err := grpcClient.CreateUser(context, &pb.CreateUserRequest{Name: "Test User 2"}); err != nil {
		panic(err)
	}
	fmt.Println("User created successfully")

	if _, err := grpcClient.CreateUser(context, &pb.CreateUserRequest{Name: "Test User 3"}); err != nil {
		panic(err)
	}
	fmt.Println("User created successfully")

	if _, err := grpcClient.CreateTopic(context, &pb.CreateTopicRequest{Name: "Test Topic"}); err != nil {
		panic(err)
	}
	fmt.Println("Topic created successfully")

	if _, err := grpcClient.CreateTopic(context, &pb.CreateTopicRequest{Name: "Test Topic 2"}); err != nil {
		panic(err)
	}
	fmt.Println("Topic created successfully")

	if _, err := grpcClient.CreateTopic(context, &pb.CreateTopicRequest{Name: "Test Topic 3"}); err != nil {
		panic(err)
	}
	fmt.Println("Topic created successfully")

	var topics *pb.ListTopicsResponse
	if topics, err = grpcClient.ListTopics(context, &emptypb.Empty{}); err != nil {
		panic(err)
	}
	fmt.Println("Topics:")
	for _, topic := range topics.Topics {
		fmt.Printf(" - ID: %d, Name: %s\n", topic.Id, topic.Name)
	}
	fmt.Println("Listed topics successfully")

	if _, err := grpcClient.PostMessage(context, &pb.PostMessageRequest{TopicId: 3, UserId: 1, Text: "Testiranje 1"}); err != nil {
		panic(err)
	}
	fmt.Println("Message posted successfully")

	if _, err := grpcClient.PostMessage(context, &pb.PostMessageRequest{TopicId: 3, UserId: 2, Text: "Testiranje 2"}); err != nil {
		panic(err)
	}
	fmt.Println("Message posted successfully")

	if _, err := grpcClient.PostMessage(context, &pb.PostMessageRequest{TopicId: 3, UserId: 3, Text: "Testiranje 3"}); err != nil {
		panic(err)
	}
	fmt.Println("Message posted successfully")

	var messages *pb.GetMessagesResponse
	if messages, err = grpcClient.GetMessages(context, &pb.GetMessagesRequest{TopicId: 3}); err != nil {
		panic(err)
	}
	fmt.Println("Messages in Topic ID 3:")
	for _, message := range messages.Messages {
		fmt.Printf(" - ID: %d, UserID: %d, Text: %s, CreatedAt: %s, Likes: %d\n", message.Id, message.UserId, message.Text, message.CreatedAt.AsTime().String(), message.Likes)
	}
	fmt.Println("Listed messages successfully")

	if _, err := grpcClient.LikeMessage(context, &pb.LikeMessageRequest{TopicId: 3, MessageId: 2, UserId: 2}); err != nil {
		panic(err)
	}
	fmt.Println("Message liked successfully")

	if messages, err = grpcClient.GetMessages(context, &pb.GetMessagesRequest{TopicId: 3}); err != nil {
		panic(err)
	}
	fmt.Println("Messages in Topic ID 3 after liking message ID 2:")
	for _, message := range messages.Messages {
		fmt.Printf(" - ID: %d, UserID: %d, Text: %s, CreatedAt: %s, Likes: %d\n", message.Id, message.UserId, message.Text, message.CreatedAt.AsTime().String(), message.Likes)
	}
	fmt.Println("Listed messages successfully after liking")

	fmt.Println("\nAll tests completed successfully.")

}
