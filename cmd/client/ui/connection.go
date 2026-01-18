package ui

import (
	"context"
	"fmt"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ChainClient redirects calls to Head or Tail based on the operation type
type ChainClient struct {
	Cpc  cp.ControlPlaneClient
	Head pb.MessageBoardClient
	Tail pb.MessageBoardClient

	headAddr string
	tailAddr string
}

// RefreshTopology refreshes HEAD or/and TAIL. This is called when grpc call returns error
func (c *ChainClient) RefreshTopology(ctx context.Context) error {
	state, err := c.Cpc.GetChainState(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	if len(state.Chain) == 0 {
		return fmt.Errorf("veriga je prazna")
	}

	newHeadAddr := state.Chain[0].Address
	newTailAddr := state.Chain[len(state.Chain)-1].Address

	if newHeadAddr != c.headAddr {
		conn, err := grpc.NewClient(newHeadAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			c.Head = pb.NewMessageBoardClient(conn)
			c.headAddr = newHeadAddr
		}
	}

	if newTailAddr != c.tailAddr {
		conn, err := grpc.NewClient(newTailAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			c.Tail = pb.NewMessageBoardClient(conn)
			c.tailAddr = newTailAddr
		}
	}
	return nil
}

func (c *ChainClient) CreateUser(ctx context.Context, in *pb.CreateUserRequest, opts ...grpc.CallOption) (*pb.User, error) {
	res, err := c.Head.CreateUser(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Head.CreateUser(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) CreateTopic(ctx context.Context, in *pb.CreateTopicRequest, opts ...grpc.CallOption) (*pb.Topic, error) {
	res, err := c.Head.CreateTopic(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Head.CreateTopic(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) PostMessage(ctx context.Context, in *pb.PostMessageRequest, opts ...grpc.CallOption) (*pb.Message, error) {
	res, err := c.Head.PostMessage(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Head.PostMessage(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) LikeMessage(ctx context.Context, in *pb.LikeMessageRequest, opts ...grpc.CallOption) (*pb.Message, error) {
	res, err := c.Head.LikeMessage(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Head.LikeMessage(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) ListTopics(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListTopicsResponse, error) {
	res, err := c.Tail.ListTopics(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Tail.ListTopics(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) GetMessages(ctx context.Context, in *pb.GetMessagesRequest, opts ...grpc.CallOption) (*pb.GetMessagesResponse, error) {
	res, err := c.Tail.GetMessages(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Tail.GetMessages(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) SubscribeTopic(ctx context.Context, in *pb.SubscribeTopicRequest, opts ...grpc.CallOption) (pb.MessageBoard_SubscribeTopicClient, error) {
	res, err := c.Tail.SubscribeTopic(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Tail.SubscribeTopic(ctx, in, opts...)
	}
	return res, err
}

func (c *ChainClient) GetSubcscriptionNode(ctx context.Context, in *pb.SubscriptionNodeRequest, opts ...grpc.CallOption) (*pb.SubscriptionNodeResponse, error) {
	res, err := c.Tail.GetSubcscriptionNode(ctx, in, opts...)
	if err != nil {
		_ = c.RefreshTopology(ctx)
		return c.Tail.GetSubcscriptionNode(ctx, in, opts...)
	}
	return res, err
}
