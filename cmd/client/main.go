package main

import (
	"context"
	"log"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"

	"github.com/marver003/razpravljalnica/cmd/client/ui"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var cpAddr string

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Razpravljalnica Message Board Client",
	Long: `A client application for the Razpravljalnica distributed message board system.
Connects to the control plane to discover nodes and interact with the message board.`,
	Run: runClient,
}

func init() {
	rootCmd.Flags().StringVarP(&cpAddr, "controlplane", "c", "localhost:12345", "Control Plane address (host:port)")
}

func runClient(cmd *cobra.Command, args []string) {
	// connect to control plane
	conn, err := grpc.NewClient(cpAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("CP connection failed: %v", err)
	}

	cpClient := cp.NewControlPlaneClient(conn)

	// get the chain state to find head and tail
	state, err := cpClient.GetChainState(context.Background(), &emptypb.Empty{})
	if err != nil || len(state.Chain) == 0 {
		log.Fatal("Could not get chain state or chain is empty")
	}

	headAddr := state.Chain[0].Address
	tailAddr := state.Chain[len(state.Chain)-1].Address

	// create connections to both
	hConn, _ := grpc.NewClient(headAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	tConn, _ := grpc.NewClient(tailAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// initialize smart client
	smartClient := &ui.ChainClient{
		Cpc:  cpClient,
		Head: pb.NewMessageBoardClient(hConn),
		Tail: pb.NewMessageBoardClient(tConn),
	}

	// start UI
	ui.Start(context.Background(), smartClient)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
