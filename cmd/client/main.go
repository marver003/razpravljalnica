package main

import (
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	//"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"github.com/marver003/razpravljalnica/cmd/client/ui"
)

// TODO: CLI client, se poveže s serverjem, demonstracija programa se nahaja tuki

func main() {
	addrPtr := flag.String("u", "localhost", "server address")
	portPtr := flag.Int("p", 12345, "server port number")
	//subTestPtr := flag.Bool("s", false, "run client subscription test")

	flag.Parse()

	url := fmt.Sprintf("%s:%d", *addrPtr, *portPtr)

	fmt.Printf("Connecting to server at %s\n", url)

	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	grpcClient := pb.NewMessageBoardClient(conn)

	ui.Start(context.Background(), grpcClient)

	// if *subTestPtr {
	// 	clientSubTest(ctx, grpcClient)
	// } else {
	// 	clientTest(ctx, grpcClient)
	// }

}

/*
func clientSubTest(ctx context.Context, grpcClient pb.MessageBoardClient) {

	if _, err := grpcClient.CreateUser(ctx, &pb.CreateUserRequest{Name: "Test User 1"}); err != nil {
		panic(err)
	}
	fmt.Println("User created successfully")

	if _, err := grpcClient.CreateTopic(ctx, &pb.CreateTopicRequest{Name: "Test Topic 1"}); err != nil {
		panic(err)
	}
	fmt.Println("Topic created successfully")

	// ask server where to subscribe
	nodeResp, err := grpcClient.GetSubcscriptionNode(ctx, &pb.SubscriptionNodeRequest{
		UserId:  1,
		TopicId: []int64{1},
	})
	if err != nil {
		panic(err)
	}

	_ = nodeResp // single-node: same client, ignore address/token

	stream, err := grpcClient.SubscribeTopic(ctx, &pb.SubscribeTopicRequest{
		UserId:         1,
		TopicId:        []int64{1},
		FromMessageId:  0,
		SubscribeToken: nodeResp.SubscribeToken,
	})
	if err != nil {
		panic(err)
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			return // stream closed or context cancelled
		}

		fmt.Printf(
			"[EVENT] seq=%d op=%s topic=%d msg=%d text=%q likes=%d\n",
			event.SequenceNumber,
			event.Op.String(),
			event.Message.TopicId,
			event.Message.Id,
			event.Message.Text,
			event.Message.Likes,
		)
	}
}

func clientTest(context context.Context, grpcClient pb.MessageBoardClient) {
	var err error
	fmt.Println("TESTIRANJE:")

	if _, err := grpcClient.CreateUser(context, &pb.CreateUserRequest{Name: "Test User 2"}); err != nil {
		panic(err)
	}
	fmt.Println("User created successfully")

	if _, err := grpcClient.CreateUser(context, &pb.CreateUserRequest{Name: "Test User 3"}); err != nil {
		panic(err)
	}
	fmt.Println("User created successfully")

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

	if _, err := grpcClient.PostMessage(context, &pb.PostMessageRequest{TopicId: 1, UserId: 1, Text: "Testiranje 1"}); err != nil {
		panic(err)
	}
	fmt.Println("Message posted successfully")

	if _, err := grpcClient.PostMessage(context, &pb.PostMessageRequest{TopicId: 1, UserId: 2, Text: "Testiranje 2"}); err != nil {
		panic(err)
	}
	fmt.Println("Message posted successfully")

	if _, err := grpcClient.PostMessage(context, &pb.PostMessageRequest{TopicId: 1, UserId: 3, Text: "Testiranje 3"}); err != nil {
		panic(err)
	}
	fmt.Println("Message posted successfully")

	var messages *pb.GetMessagesResponse
	if messages, err = grpcClient.GetMessages(context, &pb.GetMessagesRequest{TopicId: 1}); err != nil {
		panic(err)
	}
	fmt.Println("Messages in Topic ID 3:")
	for _, message := range messages.Messages {
		fmt.Printf(" - ID: %d, UserID: %d, Text: %s, CreatedAt: %s, Likes: %d\n", message.Id, message.UserId, message.Text, message.CreatedAt.AsTime().String(), message.Likes)
	}
	fmt.Println("Listed messages successfully")

	if _, err := grpcClient.LikeMessage(context, &pb.LikeMessageRequest{TopicId: 1, MessageId: 2, UserId: 2}); err != nil {
		panic(err)
	}
	fmt.Println("Message liked successfully")

	if messages, err = grpcClient.GetMessages(context, &pb.GetMessagesRequest{TopicId: 1}); err != nil {
		panic(err)
	}
	fmt.Println("Messages in Topic ID 3 after liking message ID 2:")
	for _, message := range messages.Messages {
		fmt.Printf(" - ID: %d, UserID: %d, Text: %s, CreatedAt: %s, Likes: %d\n", message.Id, message.UserId, message.Text, message.CreatedAt.AsTime().String(), message.Likes)
	}
	fmt.Println("Listed messages successfully after liking")

	fmt.Println("\nAll tests completed successfully.")
}
*/
