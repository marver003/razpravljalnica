package ui

import (
	"context"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ChainClient redirects calls to Head or Tail based on the operation type
type ChainClient struct {
	Head pb.MessageBoardClient
	Tail pb.MessageBoardClient
}

// WRITES (Go to Head)
func (c *ChainClient) CreateUser(ctx context.Context, in *pb.CreateUserRequest, opts ...grpc.CallOption) (*pb.User, error) {
	return c.Head.CreateUser(ctx, in, opts...)
}
func (c *ChainClient) CreateTopic(ctx context.Context, in *pb.CreateTopicRequest, opts ...grpc.CallOption) (*pb.Topic, error) {
	return c.Head.CreateTopic(ctx, in, opts...)
}
func (c *ChainClient) PostMessage(ctx context.Context, in *pb.PostMessageRequest, opts ...grpc.CallOption) (*pb.Message, error) {
	return c.Head.PostMessage(ctx, in, opts...)
}
func (c *ChainClient) LikeMessage(ctx context.Context, in *pb.LikeMessageRequest, opts ...grpc.CallOption) (*pb.Message, error) {
	return c.Head.LikeMessage(ctx, in, opts...)
}

// READS (Go to Tail)
func (c *ChainClient) ListTopics(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListTopicsResponse, error) {
	return c.Tail.ListTopics(ctx, in, opts...)
}
func (c *ChainClient) GetMessages(ctx context.Context, in *pb.GetMessagesRequest, opts ...grpc.CallOption) (*pb.GetMessagesResponse, error) {
	return c.Tail.GetMessages(ctx, in, opts...)
}
func (c *ChainClient) SubscribeTopic(ctx context.Context, in *pb.SubscribeTopicRequest, opts ...grpc.CallOption) (pb.MessageBoard_SubscribeTopicClient, error) {
	return c.Tail.SubscribeTopic(ctx, in, opts...)
}
func (c *ChainClient) GetSubcscriptionNode(ctx context.Context, in *pb.SubscriptionNodeRequest, opts ...grpc.CallOption) (*pb.SubscriptionNodeResponse, error) {
	return c.Tail.GetSubcscriptionNode(ctx, in, opts...)
}