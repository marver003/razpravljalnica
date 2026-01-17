package server

// TODO: Tuki se implementira vse RPC metode
import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	repl "github.com/marver003/razpravljalnica/api/replication"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

type Node struct {
	// gRPC embedanje
	pb.UnimplementedMessageBoardServer
	repl.UnimplementedReplicationServer

	// Identifikacija
	ID      string
	address string
	mu      sync.RWMutex

	// Stanje v verigi
	isHead      bool
	isTail      bool
	nextInChain repl.ReplicationClient

	// Podatki
	storage   *storage.Storage
	subMgr    *subscription.Manager
	nextIndex int64 // Števec, ki ga uporablja samo Head
}

func NewNode(s *storage.Storage, id string, addr string, subMgrPtr *subscription.Manager) *Node {
	return &Node{
		ID:      id,
		address: addr,
		storage: s,
		subMgr:  subMgrPtr,
		// Na začetku predpostavimo, da smo sami v verigi (Head in Tail hkrati),
		// dokler nas Control Plane ne posodobi.
		isHead: true,
		isTail: true,
	}
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

// CreateUser - Only the HEAD handles the initial creation request
func (n *Node) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "User name cannot be empty")
	}

	n.mu.RLock()
	isHead := n.isHead
	next := n.nextInChain
	n.mu.RUnlock()

	if !isHead {
		return nil, status.Error(codes.PermissionDenied, "Requests must be sent to the Head node.")
	}

	// 1. Create the Operation
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_CREATE_USER,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// 2. Propagate down the chain
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Replication failed: %v", err)
		}
	}

	// 3. Apply locally
	n.storage.Apply(op)

	// 4. Return the result from storage (where ID == op.Index)
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

	// 1. Create the Operation
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_CREATE_TOPIC,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// 2. Propagate
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Replication failed: %v", err)
		}
	}

	// 3. Apply locally
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

	// 1. Head določi index in čas
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_POST_MESSAGE,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// 2. Pošlji naslednjemu v verigi (Replicate)
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Napaka pri replikaciji: %v", err)
		}
	}

	// 3. Ko vsi v verigi potrdijo, zapiši še k sebi
	n.storage.Apply(op)

	// Vrni sporočilo iz storage-a
	return n.storage.GetMessage(op.Index), nil
}

func (n *Node) LikeMessage(ctx context.Context, req *pb.LikeMessageRequest) (*pb.Message, error) {
	n.mu.RLock()
	isHead := n.isHead
	next := n.nextInChain
	n.mu.RUnlock()

	// 1. Guard: Only Head accepts state-changing operations
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

	// 2. Create the Operation
	// Note: We use the same nextIndex counter for all types of operations
	// to keep the log consistent across the cluster.
	op := &repl.Operation{
		Index:     atomic.AddInt64(&n.nextIndex, 1),
		Type:      repl.OperationType_OP_LIKE_MESSAGE,
		Payload:   marshalOrPanic(req),
		Timestamp: timestamppb.Now(),
	}

	// 3. Propagate to the next node in the chain
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Replication of Like failed: %v", err)
		}
	}

	// 4. Apply locally (This increments the like count in storage)
	n.storage.Apply(op)

	// 5. Return the updated message from storage
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
	// No need to manually map if your storage already uses pb.Topic
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

	// For a distributed system, we always point the subscriber to the TAIL
	// But we need to know the Tail's address from the Control Plane state
	// For now, if THIS node is the tail, return itself.
	// If not, the client should have asked the Tail directly.
	if !n.isTail {
		return nil, status.Error(codes.FailedPrecondition, "Contact the Tail node for subscriptions")
	}

	return &pb.SubscriptionNodeResponse{
		SubscribeToken: "distributed-token",
		Node: &pb.NodeInfo{
			NodeId:  n.ID,
			Address: n.address, // Ensure you store your own address in the Node struct
		},
	}, nil
}

func (n *Node) SubscribeTopic(req *pb.SubscribeTopicRequest, stream pb.MessageBoard_SubscribeTopicServer) error {
	n.mu.RLock()
	isTail := n.isTail
	n.mu.RUnlock()

	if !isTail {
		return status.Error(codes.PermissionDenied, "Subscriptions only allowed on the Tail node.")
	}

	// Your existing subMgr logic
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

func (n *Node) Replicate(ctx context.Context, op *repl.Operation) (*repl.AckMessage, error) {
	n.mu.RLock()
	isTail := n.isTail
	next := n.nextInChain
	n.mu.RUnlock()

	// 1. Forward to next node if not Tail
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			log.Printf("Replication to next node failed: %v. Waiting for reconfiguration...", err)
			return nil, status.Error(codes.Unavailable, "Chain broken, try again in a moment")
		}
	}

	// 2. Apply locally to storage
	n.storage.Apply(op)

	// 3. Trigger Broadcast ONLY if we are the Tail
	if isTail {
		n.broadcastOperation(op)
	}

	return &repl.AckMessage{Index: op.Index}, nil
}

func (n *Node) broadcastOperation(op *repl.Operation) {
	var msg *pb.Message
	var msgType pb.OpType // Ensure this matches your proto (e.g., pb.OpType_OP_POST)

	switch op.Type {
	case repl.OperationType_OP_POST_MESSAGE:
		msg = n.storage.GetMessage(op.Index)
		msgType = pb.OpType_OP_POST

	case repl.OperationType_OP_LIKE_MESSAGE:
		msgType = pb.OpType_OP_LIKE
		// IMPORTANT: Unmarshal the payload to find the message that was liked!
		var likeReq pb.LikeMessageRequest
		if err := proto.Unmarshal(op.Payload, &likeReq); err == nil {
			msg = n.storage.GetMessage(likeReq.MessageId)
		}

	// You can add OP_CREATE_TOPIC or others here if your TUI supports live topic updates
	default:
		return
	}

	if msg != nil && n.subMgr != nil {
		n.subMgr.Broadcast(&pb.MessageEvent{
			SequenceNumber: op.Index,
			Op:             msgType,
			Message:        msg,
			EventAt:        op.Timestamp,
		})
	}
}

func marshalOrPanic(m proto.Message) []byte {
	data, err := proto.Marshal(m)
	if err != nil {
		// Če pride tuki do panic je neki zlo narobe
		panic(fmt.Sprintf("failed to marshal protobuf message: %v", err))
	}
	return data
}

func (n *Node) UpdateNodeChainState(state *cp.ChainState) {
	n.mu.Lock()
	defer n.mu.Unlock()

	logString := n.address
	if n.isHead {
		logString += " HEAD"
	}
	if n.isTail {
		logString += " TAIL"
	}
	log.Println(logString)

	myIndex := -1
	for i, nodeInfo := range state.Chain {
		if nodeInfo.NodeId == n.ID {
			myIndex = i
			break
		}
	}

	if myIndex == -1 {
		log.Printf("I (%s) am not in the current chain!", n.ID)
		return
	}

	n.isHead = (myIndex == 0)
	n.isTail = (myIndex == len(state.Chain)-1)

	if !n.isTail {
		nextAddr := state.Chain[myIndex+1].Address
		// Only create a new client if the address changed
		// You might want to store currentNextAddr in your Node struct
		conn, err := grpc.NewClient(nextAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			n.nextInChain = repl.NewReplicationClient(conn)
		}
	} else {
		n.nextInChain = nil
	}
}
