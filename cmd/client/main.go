package main

import (
	"context"
	"flag"
	"log"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"

	"github.com/marver003/razpravljalnica/cmd/client/ui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	cpAddr := flag.String("cp", "localhost:12345", "Control Plane address")
	flag.Parse()

	// 1. Connect to Control Plane
	conn, err := grpc.NewClient(*cpAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("CP connection failed: %v", err)
	}

	cpClient := cp.NewControlPlaneClient(conn)

	// 2. Get the Chain State to find Head and Tail
	state, err := cpClient.GetChainState(context.Background(), &emptypb.Empty{})
	if err != nil || len(state.Chain) == 0 {
		log.Fatal("Could not get chain state or chain is empty")
	}

	headAddr := state.Chain[0].Address
	tailAddr := state.Chain[len(state.Chain)-1].Address

	// 3. Create connections to both
	hConn, _ := grpc.NewClient(headAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	tConn, _ := grpc.NewClient(tailAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 4. Initialize our Smart Client
	smartClient := &ui.ChainClient{
		Head: pb.NewMessageBoardClient(hConn),
		Tail: pb.NewMessageBoardClient(tConn),
	}

	// 5. Start UI - unchanged, it thinks smartClient is a normal client!
	ui.Start(context.Background(), smartClient)

}
