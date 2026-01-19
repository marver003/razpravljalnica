package main

import (
	"context"
	"log"
	"net"
	"time"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	repl "github.com/marver003/razpravljalnica/api/replication"
	"github.com/marver003/razpravljalnica/cmd/node/ui"
	"github.com/marver003/razpravljalnica/internal/server"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port   string
	cpAddr string
	nodeID string
)

var rootCmd = &cobra.Command{
	Use:   "node",
	Short: "Razpravljalnica Message Board Node",
	Long: `A node in the Razpravljalnica distributed message board system.
Stores messages, handles replication, and participates in the chain.`,
	Run: runNode,
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "54321", "Port to listen on")
	rootCmd.Flags().StringVarP(&cpAddr, "controlplane", "c", "localhost:12345", "Control Plane address (host:port)")
	rootCmd.Flags().StringVarP(&nodeID, "id", "i", "node-1", "Unique ID for this node")
}

func runNode(cmd *cobra.Command, args []string) {
	myAddr := "localhost:" + port

	// init storage and node
	store := storage.NewStorage()
	subManager := subscription.NewManager()
	node := server.NewNode(store, nodeID, myAddr, subManager)

	// run everything else in a goroutine so UI can take over main thread
	go func() {
		// registering at control plane
		conn, err := grpc.NewClient(cpAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("Could not connect to Control Plane: %v", err)
		}
		cpClient := cp.NewControlPlaneClient(conn)

		resp, err := cpClient.RegisterNode(context.Background(), &cp.RegisterRequest{
			Node: &cp.NodeInfo{
				NodeId:  nodeID,
				Address: myAddr,
			},
		})
		if err != nil {
			log.Fatalf("Failed to register: %v", err)
		}

		// updating chain state (head/tail/next)
		node.UpdateNodeChainState(resp)

		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterMessageBoardServer(s, node)
		repl.RegisterReplicationServer(s, node)

		log.Printf("Node %s gRPC server listening on port %s", nodeID, port)

		go func() {
			for {
				time.Sleep(2 * time.Second) // check every 2 seconds

				resp, err := cpClient.RegisterNode(context.Background(), &cp.RegisterRequest{
					Node: &cp.NodeInfo{
						NodeId:  nodeID,
						Address: myAddr,
					},
				})

				if err == nil {
					// this forces the node to update nextInChain, isHead, and isTail
					node.UpdateNodeChainState(resp)
				}
			}
		}()

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// start UI
	ui.Start(node)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
