package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	repl "github.com/marver003/razpravljalnica/api/replication"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type Node struct {
	pb.UnimplementedMessageBoardServer
	repl.UnimplementedReplicationServer

	ID      string
	address string
	mu      sync.RWMutex

	// chain state
	isHead      bool
	isTail      bool
	nextInChain repl.ReplicationClient

	storage   *storage.Storage
	subMgr    *subscription.Manager
	nextIndex int64 // counter used by HEAD
	synced    bool

	chain []*cp.NodeInfo // full view of the chain for load balancing

	logFunc func(string, ...any)
}

type NodeUIState struct {
	ID        string
	Address   string
	IsHead    bool
	IsTail    bool
	Synced    bool
	LastIndex int64
	SubCount  int
	RecentOps []*repl.Operation
}

func (n *Node) SetLogger(fn func(string, ...any)) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.logFunc = fn
}

func (n *Node) log(format string, v ...any) {
	if n.logFunc != nil {
		n.logFunc(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

func (n *Node) GetUIState() NodeUIState {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return NodeUIState{
		ID:        n.ID,
		Address:   n.address,
		IsHead:    n.isHead,
		IsTail:    n.isTail,
		Synced:    n.synced,
		LastIndex: n.storage.GetLastIndex(),
		SubCount:  n.subMgr.Count(),
		RecentOps: n.storage.GetRecentOperations(10), // Get last 10 ops
	}
}

func NewNode(s *storage.Storage, id string, addr string, subMgrPtr *subscription.Manager) *Node {
	return &Node{
		ID:      id,
		address: addr,
		storage: s,
		subMgr:  subMgrPtr,
		// we set isHead and isTail to true, control plane will update the node later
		isHead: true,
		isTail: true,
	}
}

func (n *Node) broadcastOperation(op *repl.Operation) {
	var msg *pb.Message
	var msgType pb.OpType

	switch op.Type {
	case repl.OperationType_OP_POST_MESSAGE:
		msg = n.storage.GetMessage(op.Index)
		msgType = pb.OpType_OP_POST

	case repl.OperationType_OP_LIKE_MESSAGE:
		msgType = pb.OpType_OP_LIKE
		var likeReq pb.LikeMessageRequest
		if err := proto.Unmarshal(op.Payload, &likeReq); err == nil {
			msg = n.storage.GetMessage(likeReq.MessageId)
		}

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
		panic(fmt.Sprintf("failed to marshal protobuf message: %v", err))
	}
	return data
}

func (n *Node) syncWithNode(targetAddr string) {
	n.log("Starting sync from node %s...", targetAddr)
	conn, err := grpc.NewClient(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		n.log("Failed to connect to %s for sync: %v", targetAddr, err)
		return
	}
	defer conn.Close()

	client := repl.NewReplicationClient(conn)
	lastIdx := n.storage.GetLastIndex()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.GetLogFrom(ctx, &repl.GetLogRequest{FromIndex: lastIdx + 1})
	if err != nil {
		n.log("Failed to call GetLogFrom: %v", err)
		return
	}

	count := 0
	for {
		op, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			n.log("Error receiving stream: %v", err)
			break
		}
		n.storage.Apply(op)
		count++
	}
	n.log("Synced %d operations from %s", count, targetAddr)
}

func (n *Node) UpdateNodeChainState(state *cp.ChainState) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.chain = state.Chain // store the full chain

	myIndex := -1
	for i, nodeInfo := range state.Chain {
		if nodeInfo.NodeId == n.ID {
			myIndex = i
			break
		}
	}

	if myIndex == -1 {
		n.log("I (%s) am not in the current chain!", n.ID)
		return
	}

	// sync with previoud node
	if myIndex > 0 && !n.synced {
		prevNode := state.Chain[myIndex-1]
		n.syncWithNode(prevNode.Address)
		n.synced = true
	}

	n.isHead = (myIndex == 0)
	n.isTail = (myIndex == len(state.Chain)-1)

	if !n.isTail {
		nextAddr := state.Chain[myIndex+1].Address
		conn, err := grpc.NewClient(nextAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			n.nextInChain = repl.NewReplicationClient(conn)
		}
	} else {
		n.nextInChain = nil
	}
}
