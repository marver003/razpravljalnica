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

	// Seznam vseh aktivnih vozlišč v verigi
	nodes    map[string]*NodeStatus
	chain    []*cp.NodeInfo
	lastSeen map[string]time.Time // NodeID -> Timestamp
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

	// Opcijsko: Zaženi gorutino, ki preverja izpadle node (za oceno 9-10)
	// go cpServer.monitorNodes()
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

			// Če je vozlišče tiho več kot 5 sekund, ga odstranimo
			if time.Since(lastHeartbeat) < 5*time.Second {
				newChain = append(newChain, node)
			} else {
				log.Printf("Vozlišče %s je poteklo (zadnjič videno pred %v) - odstranjujem iz verige",
					node.NodeId, time.Since(lastHeartbeat).Round(time.Second))
				changed = true
				// OPOMBA: Vozlišča NE brišemo iz s.nodes, da ohranimo zgodovino za s.RegisterNode!
			}
		}

		if changed {
			s.chain = newChain
			log.Printf("Veriga posodobljena. Nova dolžina: %d", len(s.chain))
		}
		s.mu.Unlock()
	}
}

func (s *ControlPlane) RegisterNode(ctx context.Context, req *cp.RegisterRequest) (*cp.ChainState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeID := req.Node.NodeId
	s.lastSeen[nodeID] = time.Now()

	// 1. Preverimo, če vozlišče že obstaja v našem zgodovinskem seznamu (s.nodes)
	_, existedBefore := s.nodes[nodeID]

	if !existedBefore {
		log.Printf("Registracija NOVEGA vozlišča: %s (%s)", nodeID, req.Node.Address)
		s.nodes[nodeID] = &NodeStatus{
			Info:     req.Node,
			LastSeen: time.Now(),
		}
	} else {
		s.nodes[nodeID].LastSeen = time.Now()
		s.nodes[nodeID].Info = req.Node
	}

	// 2. Preverimo, če je trenutno v AKTIVNI verigi
	inChain := false
	for _, n := range s.chain {
		if n.NodeId == nodeID {
			inChain = true
			break
		}
	}

	// 3. Dodajanje v verigo z natančnim logiranjem
	if !inChain {
		if existedBefore {
			log.Printf("Vozlišče %s se je VRNILO v verigo po padcu/timeoutu", nodeID)
		} else {
			log.Printf("Vozlišče %s je bilo dodano na konec verige", nodeID)
		}
		s.chain = append(s.chain, req.Node)
	}

	return &cp.ChainState{Chain: s.chain}, nil
}

// SendHeartbeat vozlišče pokliče vsakih nekaj sekund
func (s *ControlPlane) SendHeartbeat(ctx context.Context, req *cp.Heartbeat) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status, exists := s.nodes[req.NodeId]; exists {
		status.LastSeen = time.Now()
	}

	return &emptypb.Empty{}, nil
}

// GetChainState vrne trenutni vrstni red verige
func (s *ControlPlane) GetChainState(ctx context.Context, _ *emptypb.Empty) (*cp.ChainState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &cp.ChainState{Chain: s.chain}, nil
}

// monitorNodes (Osnova za oceno 9-10)
func (s *ControlPlane) monitorNodes() {
	for {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		// Tukaj bi preveril, če je time.Since(status.LastSeen) > 5 * time.Second
		// In če je, bi odstranil node iz s.chain in rekonfiguriral sosede.
		s.mu.Unlock()
	}
}
