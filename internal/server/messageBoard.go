package server

// TODO: Tuki se implementira vse RPC metode
import (
	"context"
	"strings"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements pb.MessageBoardServer
type MessageBoardServer struct {
	pb.UnimplementedMessageBoardServer
	store   *storage.Storage
	subMgr  *subscription.Manager
	address string
	nodeId  string
}

// NewServerMessageBoard returns a pointer of MessageBoardServer struct
func NewServerMessageBoard(storePtr *storage.Storage, subMgrPtr *subscription.Manager, address, nodeId string) *MessageBoardServer {
	return &MessageBoardServer{
		store:   storePtr,
		subMgr:  subMgrPtr,
		address: address,
		nodeId:  nodeId,
	}
}

// CreateUser creates a new user and assigns an id
func (s *MessageBoardServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "User name cannot be empty")
	}

	user := s.store.CreateUser(req.Name)

	return &pb.User{
		Id:   user.Id,
		Name: user.Name,
	}, nil
}

// CreateTopic creates a new topic to which users can post messages
func (s *MessageBoardServer) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.Topic, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Topic name cannot be empty")
	}

	user := s.store.CreateTopic(req.Name)

	return &pb.Topic{
		Id:   user.Id,
		Name: user.Name,
	}, nil
}

// PostMessage posts a message to a topic; succeeds only if the User and the Topic exist
func (s *MessageBoardServer) PostMessage(ctx context.Context, req *pb.PostMessageRequest) (*pb.Message, error) {
	if !s.store.TopicExists(req.TopicId) {
		return nil, notFound("Topic doesn't exist")
	}

	if !s.store.UserExists(req.UserId) {
		return nil, notFound("User doesn't exist")
	}

	if strings.TrimSpace(req.Text) == "" {
		return nil, status.Error(codes.InvalidArgument, "Message cannot be empty")
	}

	message := s.store.PostMessage(req.TopicId, req.UserId, req.Text)

	pbMessage := &pb.Message{
		Id:        message.Id,
		TopicId:   message.TopicId,
		UserId:    message.UserId,
		Text:      message.Text,
		CreatedAt: timestamppb.New(message.CreatedAt),
		Likes:     message.Likes,
	}

	event := &pb.MessageEvent{
		SequenceNumber: 1,
		Op:             pb.OpType_OP_POST,
		Message:        pbMessage,
		EventAt:        timestamppb.Now(),
	}

	s.subMgr.Broadcast(event)

	return pbMessage, nil
}

// LikeMessage likes an existing message and returns message with updated likes
func (s *MessageBoardServer) LikeMessage(ctx context.Context, req *pb.LikeMessageRequest) (*pb.Message, error) {
	if !s.store.TopicExists(req.TopicId) {
		return nil, notFound("Topic doesn't exist")
	}

	if !s.store.UserExists(req.UserId) {
		return nil, notFound("User doesn't exist")
	}

	if !s.store.MessageExists(req.MessageId) {
		return nil, notFound("Message doesn't exist")
	}

	message := s.store.LikeMessage(req.MessageId)

	return &pb.Message{
		Id:        message.Id,
		TopicId:   message.TopicId,
		UserId:    message.UserId,
		Text:      message.Text,
		CreatedAt: timestamppb.New(message.CreatedAt),
		Likes:     message.Likes,
	}, nil
}

// ListTopics returns all topics
func (s *MessageBoardServer) ListTopics(ctx context.Context, _ *emptypb.Empty) (*pb.ListTopicsResponse, error) {
	topics := s.store.ListTopics()

	result := make([]*pb.Topic, 0, len(topics))
	for _, t := range topics {
		result = append(result, &pb.Topic{
			Id:   t.Id,
			Name: t.Name,
		})
	}

	return &pb.ListTopicsResponse{
		Topics: result,
	}, nil
}

// GetMessages returns messages in a topic starting from from_message_id (inclusive) up to limit
func (s *MessageBoardServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	if !s.store.TopicExists(req.TopicId) {
		return nil, notFound("Topic doesn't exist")
	}

	messages := s.store.GetMessages(req.TopicId, req.FromMessageId, req.Limit)

	result := make([]*pb.Message, 0, len(messages))
	for _, m := range messages {
		result = append(result, &pb.Message{
			Id:        m.Id,
			TopicId:   m.TopicId,
			UserId:    m.UserId,
			Text:      m.Text,
			CreatedAt: timestamppb.New(m.CreatedAt),
			Likes:     m.Likes,
		})
	}

	return &pb.GetMessagesResponse{
		Messages: result,
	}, nil
}

// GetSubcscriptionNode (note: generated proto has typo GetSubcscriptionNode)
// For a single-node deployment we return this node info and a subscribe token
func (s *MessageBoardServer) GetSubcscriptionNode(ctx context.Context, req *pb.SubscriptionNodeRequest) (*pb.SubscriptionNodeResponse, error) {
	return &pb.SubscriptionNodeResponse{
		SubscribeToken: "token",
		Node: &pb.NodeInfo{
			NodeId:  s.nodeId,
			Address: s.address,
		},
	}, nil
}

// SubscribeTopic: register subscriber, send historical messages from from_message_id, then stream live events
func (s *MessageBoardServer) SubscribeTopic(req *pb.SubscribeTopicRequest, stream pb.MessageBoard_SubscribeTopicServer) error {
	topics := make(map[int64]bool)
	for _, t := range req.TopicId {
		topics[t] = true
	}

	sub := &subscription.Subscriber{
		UserID:   req.UserId,
		TopicIDs: topics,
		Stream:   stream,
	}

	s.subMgr.Add(sub)
	defer s.subMgr.Remove(req.UserId)

	<-stream.Context().Done()
	return nil
}

// helper: simple not found error
func notFound(err string) error {
	return status.Error(codes.NotFound, err)
}
