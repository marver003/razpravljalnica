package server

// TODO: Tuki se implementira vse RPC metode
import (
	"context"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements pb.MessageBoardServer
type MessageBoardServer struct {
	pb.UnimplementedMessageBoardServer
	store *storage.Storage
	subs  *subscription.Manager
}

// NewServerMessageBoard returns a pointer of MessageBoardServer struct
func NewServerMessageBoard(storePtr *storage.Storage, subsPtr *subscription.Manager) *MessageBoardServer {
	return &MessageBoardServer{
		store: storePtr,
		subs:  subsPtr,
	}
}

// CreateUser creates a new user and assigns an id
func (s *MessageBoardServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// CreateTopic creates a new topic to which users can post messages
func (s *MessageBoardServer) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.Topic, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// PostMessage posts a message to a topic; succeeds only if the User and the Topic exist
func (s *MessageBoardServer) PostMessage(ctx context.Context, req *pb.PostMessageRequest) (*pb.Message, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// LikeMessage likes an existing message and returns message with updated likes
func (s *MessageBoardServer) LikeMessage(ctx context.Context, req *pb.LikeMessageRequest) (*pb.Message, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// GetSubcscriptionNode (note: generated proto has typo GetSubcscriptionNode)
// For a single-node deployment we return this node info and a subscribe token
func (s *MessageBoardServer) GetSubcscriptionNode(ctx context.Context, req *pb.SubscriptionNodeRequest) (*pb.SubscriptionNodeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// ListTopics returns all topics
func (s *MessageBoardServer) ListTopics(ctx context.Context, _ *emptypb.Empty) (*pb.ListTopicsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// GetMessages returns messages in a topic starting from from_message_id (inclusive) up to limit
func (s *MessageBoardServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not yet implemented")
}

// SubscribeTopic: register subscriber, send historical messages from from_message_id, then stream live events
func (s *MessageBoardServer) SubscribeTopic(req *pb.SubscribeTopicRequest, stream pb.MessageBoard_SubscribeTopicServer) error {
	return status.Error(codes.Unimplemented, "not yet implemented")
}

// helper: simple not found error
func notFound(err string) error {
	return status.Error(codes.NotFound, err)
}
