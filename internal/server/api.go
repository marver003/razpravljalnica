package server

import (
	"context"
	"log"
	"strings"
	"sync/atomic"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	repl "github.com/marver003/razpravljalnica/api/replication"
	"github.com/marver003/razpravljalnica/internal/subscription"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (n *Node) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "User name cannot be empty")
	}

	// check if user already exists
	if u := n.storage.GetUserByName(req.Name); u != nil {
		return u, nil
	}

	n.mu.RLock()
	isHead := n.isHead
	next := n.nextInChain
	n.mu.RUnlock()

	if !isHead {
		return nil, status.Error(codes.PermissionDenied, "Requests must be sent to the Head node.")
	}

	// create the operation
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_CREATE_USER,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// propagate down the chain
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Replication failed: %v", err)
		}
	}

	// apply locally
	n.storage.Apply(op)

	// return the result from storage (where ID == op.Index)
	return n.storage.GetUser(op.Index), nil
}

// CreateTopic - Only the HEAD handles the initial creation request
func (n *Node) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.Topic, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Topic name cannot be empty")
	}

	n.mu.RLock()
	isHead := n.isHead
	next := n.nextInChain
	n.mu.RUnlock()

	if !isHead {
		return nil, status.Error(codes.PermissionDenied, "Requests must be sent to the Head node.")
	}

	// create the operation
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_CREATE_TOPIC,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// propagate
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Replication failed: %v", err)
		}
	}

	// apply locally
	n.storage.Apply(op)

	return n.storage.GetTopic(op.Index), nil
}

func (n *Node) PostMessage(ctx context.Context, req *pb.PostMessageRequest) (*pb.Message, error) {
	n.mu.RLock()
	isHead := n.isHead
	next := n.nextInChain
	n.mu.RUnlock()

	if !isHead {
		return nil, status.Error(codes.PermissionDenied, "Writes allowed only on head node")
	}

	if !n.storage.TopicExists(req.TopicId) {
		return nil, notFound("Topic doesn't exist")
	}

	if !n.storage.UserExists(req.UserId) {
		return nil, notFound("User doesn't exist")
	}

	if strings.TrimSpace(req.Text) == "" {
		return nil, status.Error(codes.InvalidArgument, "Message cannot be empty")
	}

	// head specifies index and time
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_POST_MESSAGE,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// propagate
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Napaka pri replikaciji: %v", err)
		}
	}

	// apply locally
	n.storage.Apply(op)
	n.broadcastOperation(op)

	return n.storage.GetMessage(op.Index), nil
}

func (n *Node) LikeMessage(ctx context.Context, req *pb.LikeMessageRequest) (*pb.Message, error) {
	n.mu.RLock()
	isHead := n.isHead
	next := n.nextInChain
	n.mu.RUnlock()

	// Guard: Only Head accepts state-changing operations
	if !isHead {
		return nil, status.Error(codes.PermissionDenied, "Writes allowed only on head node")
	}

	if !n.storage.TopicExists(req.TopicId) {
		return nil, notFound("Topic doesn't exist")
	}

	if !n.storage.UserExists(req.UserId) {
		return nil, notFound("User doesn't exist")
	}

	if !n.storage.MessageExists(req.MessageId) {
		return nil, notFound("Message doesn't exist")
	}

	// Create the Operation
	// Note: nextIndex counter is used for all types of operations
	// to keep the log consistent across the cluster.
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_LIKE_MESSAGE,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// propagate to the next node in the chain
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Replication of Like failed: %v", err)
		}
	}

	// apply locally
	n.storage.Apply(op)
	n.broadcastOperation(op)

	// return the updated message from storage
	updatedMsg := n.storage.GetMessage(req.MessageId)
	if updatedMsg == nil {
		return nil, status.Error(codes.NotFound, "Message not found after applying like")
	}

	return updatedMsg, nil
}

func (n *Node) ListTopics(ctx context.Context, _ *emptypb.Empty) (*pb.ListTopicsResponse, error) {
	n.mu.RLock()
	isTail := n.isTail
	n.mu.RUnlock()

	if !isTail {
		return nil, status.Error(codes.PermissionDenied, "Read requests must be sent to the Tail node.")
	}

	topics := n.storage.ListTopics()

	return &pb.ListTopicsResponse{Topics: topics}, nil
}

func (n *Node) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	n.mu.RLock()
	isTail := n.isTail
	n.mu.RUnlock()

	if !isTail {
		return nil, status.Error(codes.PermissionDenied, "Read requests must be sent to the Tail node.")
	}

	if !n.storage.TopicExists(req.TopicId) {
		return nil, status.Error(codes.NotFound, "Topic doesn't exist")
	}

	messages := n.storage.GetMessages(req.TopicId, req.FromMessageId, req.Limit)
	return &pb.GetMessagesResponse{Messages: messages}, nil
}

func (n *Node) GetSubcscriptionNode(ctx context.Context, req *pb.SubscriptionNodeRequest) (*pb.SubscriptionNodeResponse, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !n.isTail {
		return nil, status.Error(codes.FailedPrecondition, "Contact the Tail node for subscriptions")
	}

	if len(n.chain) == 0 {
		return nil, status.Error(codes.Unavailable, "No nodes in chain")
	}

	// simple load balancing: Hash = (TopicID + UserID) to pick a node
	var targetNode *pb.NodeInfo
	if len(req.TopicId) > 0 {
		hashKey := int(req.TopicId[0]) + int(req.UserId)
		idx := hashKey % len(n.chain)
		cpNode := n.chain[idx]
		targetNode = &pb.NodeInfo{
			NodeId:  cpNode.NodeId,
			Address: cpNode.Address,
		}
	} else {
		// default to self if no topic specified
		targetNode = &pb.NodeInfo{
			NodeId:  n.ID,
			Address: n.address,
		}
	}

	return &pb.SubscriptionNodeResponse{
		SubscribeToken: "distributed-token",
		Node:           targetNode,
	}, nil
}

func (n *Node) SubscribeTopic(req *pb.SubscribeTopicRequest, stream pb.MessageBoard_SubscribeTopicServer) error {
	n.mu.RLock()
	// allow subscription on ANY node
	log.Printf("NODE %s accepting subscription for user %d topics %v", n.ID, req.UserId, req.TopicId)
	n.mu.RUnlock()

	topics := make(map[int64]bool)
	for _, t := range req.TopicId {
		topics[t] = true
	}

	sub := &subscription.Subscriber{
		UserID:   req.UserId,
		TopicIDs: topics,
		Stream:   stream,
	}

	n.subMgr.Add(sub)
	defer n.subMgr.Remove(req.UserId)

	<-stream.Context().Done()
	return nil
}

// helper: simple not found error
func notFound(err string) error {
	return status.Error(codes.NotFound, err)
}
