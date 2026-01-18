package server

import (
	"context"
	"log"
	"sync"
	"time"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NodeStatus struct {
	Info     *cp.NodeInfo
	LastSeen time.Time
}

type ControlPlane struct {
	cp.UnimplementedControlPlaneServer
	mu sync.RWMutex

	nodes    map[string]*NodeStatus // list of all nodes that are/were in chain
	chain    []*cp.NodeInfo         // list of all active nodes in chain
	lastSeen map[string]time.Time   // NodeID -> Timestamp

	logFunc func(string, ...any)
}

func (s *ControlPlane) SetLogger(fn func(string, ...any)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logFunc = fn
}

func (s *ControlPlane) log(format string, v ...any) {
	if s.logFunc != nil {
		s.logFunc(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// GetState returns a snapshot of the current state for UI
func (s *ControlPlane) GetState() ([]*cp.NodeInfo, map[string]*NodeStatus) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// deep copy to avoid race conditions in UI
	chainCopy := make([]*cp.NodeInfo, len(s.chain))
	copy(chainCopy, s.chain)

	nodesCopy := make(map[string]*NodeStatus)
	for k, v := range s.nodes {
		nodesCopy[k] = &NodeStatus{
			Info:     v.Info,
			LastSeen: v.LastSeen,
		}
	}

	return chainCopy, nodesCopy
}

func (s *ControlPlane) GetClusterState(ctx context.Context, _ *emptypb.Empty) (*pb.GetClusterStateResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.chain) == 0 {
		return nil, status.Error(codes.Unavailable, "No nodes registered")
	}

	headNode := s.chain[0]
	tailNode := s.chain[len(s.chain)-1]

	return &pb.GetClusterStateResponse{
		Head: &pb.NodeInfo{
			NodeId:  headNode.NodeId,
			Address: headNode.Address,
		},
		Tail: &pb.NodeInfo{
			NodeId:  tailNode.NodeId,
			Address: tailNode.Address,
		},
	}, nil
}

func NewControlPlane() *ControlPlane {
	cpServer := &ControlPlane{
		nodes:    make(map[string]*NodeStatus),
		chain:    make([]*cp.NodeInfo, 0),
		lastSeen: make(map[string]time.Time),
	}

	go cpServer.startMonitoring()
	return cpServer
}

func (s *ControlPlane) startMonitoring() {
	for {
		time.Sleep(2 * time.Second)
		s.mu.Lock()

		var newChain []*cp.NodeInfo
		changed := false

		for _, node := range s.chain {
			lastHeartbeat := s.lastSeen[node.NodeId]

			// if the node's heartbeat is older than 5 seconds, it's removed
			if time.Since(lastHeartbeat) < 5*time.Second {
				newChain = append(newChain, node)
			} else {
				s.log("Vozlišče %s je poteklo (zadnjič videno pred %v) - odstranjujem iz verige",
					node.NodeId, time.Since(lastHeartbeat).Round(time.Second))
				changed = true
				// NOTE: Node isn't deleted from s.nodes, to preserve history for s.RegisterNode
			}
		}

		if changed {
			s.chain = newChain
			s.log("Veriga posodobljena. Nova dolžina: %d", len(s.chain))
		}
		s.mu.Unlock()
	}
}

// RegisterNode is called by new nodes, but it is also used as a heartbeat called every 2 seconds
func (s *ControlPlane) RegisterNode(ctx context.Context, req *cp.RegisterRequest) (*cp.ChainState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeID := req.Node.NodeId
	s.lastSeen[nodeID] = time.Now()

	// check if current node already exists in node history s.nodes
	_, existedBefore := s.nodes[nodeID]
	if !existedBefore {
		s.log("Registering NEW node: %s (%s)", nodeID, req.Node.Address)
		s.nodes[nodeID] = &NodeStatus{
			Info:     req.Node,
			LastSeen: time.Now(),
		}
	} else {
		s.nodes[nodeID].LastSeen = time.Now()
		s.nodes[nodeID].Info = req.Node
	}

	// check if the node is in active chain
	inChain := false
	for _, n := range s.chain {
		if n.NodeId == nodeID {
			inChain = true
			break
		}
	}

	// replace active chain with new chain
	if !inChain {
		if existedBefore {
			s.log("Node %s RETURNED into chain", nodeID)
		} else {
			s.log("Node %s ADDED to the tail", nodeID)
		}
		s.chain = append(s.chain, req.Node)
	}

	return &cp.ChainState{Chain: s.chain}, nil
}

// GetChainState returns current chain state
func (s *ControlPlane) GetChainState(ctx context.Context, _ *emptypb.Empty) (*cp.ChainState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &cp.ChainState{Chain: s.chain}, nil
}
