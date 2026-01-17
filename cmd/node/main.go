package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	repl "github.com/marver003/razpravljalnica/api/replication"
	"github.com/marver003/razpravljalnica/internal/server" // tvoj Node struct
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Nastavitve preko zastavic
	port := flag.String("port", "54321", "Port for this node")
	cpAddr := flag.String("cp", "localhost:12345", "Control Plane address")
	nodeID := flag.String("id", "node-1", "Unique ID for this node")
	flag.Parse()

	myAddr := "localhost:" + *port

	// 2. Inicializacija Storage-a in Node-a
	store := storage.NewStorage()
	subManager := subscription.NewManager()
	node := server.NewNode(store, *nodeID, myAddr, subManager) // Funkcija, ki jo moraš dodati v Node

	// 3. Registracija pri Control Plane
	conn, err := grpc.NewClient(*cpAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Could not connect to Control Plane: %v", err)
	}
	cpClient := cp.NewControlPlaneClient(conn)

	resp, err := cpClient.RegisterNode(context.Background(), &cp.RegisterRequest{
		Node: &cp.NodeInfo{
			NodeId:  *nodeID,
			Address: myAddr,
		},
	})
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// 4. Posodobitev stanja verige (Head/Tail/Next)
	node.UpdateNodeChainState(resp)

	// 5. Zagon gRPC strežnika
	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMessageBoardServer(s, node)
	repl.RegisterReplicationServer(s, node)

	go func() {
		for {
			time.Sleep(2 * time.Second) // Check every 5 seconds

			resp, err := cpClient.RegisterNode(context.Background(), &cp.RegisterRequest{
				Node: &cp.NodeInfo{
					NodeId:  *nodeID,
					Address: myAddr,
				},
			})

			if err == nil {
				// This forces the node to update nextInChain, isHead, and isTail
				node.UpdateNodeChainState(resp)
			}
		}
	}()

	log.Printf("Node %s listening at %v", *nodeID, lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
